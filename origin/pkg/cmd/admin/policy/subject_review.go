package policy

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/openshift/github.com/spf13/cobra"

	kapi "github.com/openshift/kubernetes/pkg/api"
	"github.com/openshift/kubernetes/pkg/api/meta"
	"github.com/openshift/kubernetes/pkg/kubectl"
	kcmdutil "github.com/openshift/kubernetes/pkg/kubectl/cmd/util"
	"github.com/openshift/kubernetes/pkg/kubectl/resource"
	"github.com/openshift/kubernetes/pkg/runtime"
	"github.com/openshift/kubernetes/pkg/serviceaccount"
	utilerrors "github.com/openshift/kubernetes/pkg/util/errors"

	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/templates"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	securityapi "github.com/openshift/origin/pkg/security/api"
)

var (
	subjectReviewLong = templates.LongDesc(`Check whether a User, Service Account or a Group can create a Pod.
	It returns a list of Security Context Constraints that will admit the resource.
	If User is specified but not Groups, it is interpreted as "What if User is not a member of any groups".
	If User and Groups are empty, then the check is performed using the current user
	`)
	subjectReviewExamples = templates.Examples(`# Check whether user bob can create a pod specified in myresource.yaml
	$ %[1]s -u bob -f myresource.yaml

	# Check whether user bob who belongs to projectAdmin group can create a pod specified in myresource.yaml
	$ %[1]s -u bob -g projectAdmin -f myresource.yaml

	# Check whether ServiceAccount specified in podTemplateSpec in myresourcewithsa.yaml can create the Pod
	$  %[1]s -f myresourcewithsa.yaml `)
)

const SubjectReviewRecommendedName = "scc-subject-review"

type sccSubjectReviewOptions struct {
	sccSubjectReviewClient     client.PodSecurityPolicySubjectReviewsNamespacer
	sccSelfSubjectReviewClient client.PodSecurityPolicySelfSubjectReviewsNamespacer
	namespace                  string
	enforceNamespace           bool
	out                        io.Writer
	mapper                     meta.RESTMapper
	typer                      runtime.ObjectTyper
	RESTClientFactory          func(mapping *meta.RESTMapping) (resource.RESTClient, error)
	printer                    sccSubjectReviewPrinter
	FilenameOptions            resource.FilenameOptions
	User                       string
	Groups                     []string
	serviceAccount             string
}

func NewCmdSccSubjectReview(name, fullName string, f *clientcmd.Factory, out io.Writer) *cobra.Command {
	o := &sccSubjectReviewOptions{}
	cmd := &cobra.Command{
		Use:     name,
		Long:    subjectReviewLong,
		Short:   "Check whether a user or a ServiceAccount can create a Pod.",
		Example: fmt.Sprintf(subjectReviewExamples, fullName, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			kcmdutil.CheckErr(o.Complete(f, args, cmd, out))
			kcmdutil.CheckErr(o.Run(args))
		},
	}

	cmd.Flags().StringVarP(&o.User, "user", "u", o.User, "Review will be performed on behalf of this user")
	cmd.Flags().StringSliceVarP(&o.Groups, "groups", "g", o.Groups, "Comma separated, list of groups. Review will be performed on behalf of these groups")
	cmd.Flags().StringVarP(&o.serviceAccount, "serviceaccount", "z", o.serviceAccount, "service account in the current namespace to use as a user")
	kcmdutil.AddFilenameOptionFlags(cmd, &o.FilenameOptions, "Filename, directory, or URL to a file identifying the resource to get from a server.")
	kcmdutil.AddPrinterFlags(cmd)
	return cmd
}

func (o *sccSubjectReviewOptions) Complete(f *clientcmd.Factory, args []string, cmd *cobra.Command, out io.Writer) error {
	if len(args) == 0 && len(o.FilenameOptions.Filenames) == 0 {
		return kcmdutil.UsageError(cmd, cmd.Use)
	}
	if len(o.User) > 0 && len(o.serviceAccount) > 0 {
		return fmt.Errorf("--user and --serviceaccount are mutually exclusive")
	}
	if len(o.serviceAccount) > 0 { // check whether user supplied a list of SA
		if len(strings.Split(o.serviceAccount, ",")) > 1 {
			return fmt.Errorf("only one Service Account is supported")
		}
		if strings.HasPrefix(o.serviceAccount, serviceaccount.ServiceAccountUsernamePrefix) {
			_, user, err := serviceaccount.SplitUsername(o.serviceAccount)
			if err != nil {
				return err
			}
			o.serviceAccount = user
		}
	}
	var err error
	o.namespace, o.enforceNamespace, err = f.DefaultNamespace()
	if err != nil {
		return err
	}
	oclient, _, err := f.Clients()
	if err != nil {
		return fmt.Errorf("unable to obtain client: %v", err)
	}
	o.sccSubjectReviewClient = oclient
	o.sccSelfSubjectReviewClient = oclient
	o.mapper, o.typer = f.Object()
	o.RESTClientFactory = f.ClientForMapping

	if len(kcmdutil.GetFlagString(cmd, "output")) != 0 {
		clientConfig, err := f.ClientConfig()
		if err != nil {
			return err
		}
		version, err := kcmdutil.OutputVersion(cmd, clientConfig.GroupVersion)
		if err != nil {
			return err
		}
		p, _, err := kcmdutil.PrinterForCommand(cmd)
		if err != nil {
			return err
		}
		o.printer = &sccSubjectReviewOutputPrinter{kubectl.NewVersionedPrinter(p, kapi.Scheme, version)}
	} else {
		o.printer = &sccSubjectReviewHumanReadablePrinter{noHeaders: kcmdutil.GetFlagBool(cmd, "no-headers")}
	}
	o.out = out
	return nil
}

func (o *sccSubjectReviewOptions) Run(args []string) error {
	userOrSA := o.User
	if len(o.serviceAccount) > 0 {
		userOrSA = o.serviceAccount
	}
	r := resource.NewBuilder(o.mapper, o.typer, resource.ClientMapperFunc(o.RESTClientFactory), kapi.Codecs.UniversalDecoder()).
		NamespaceParam(o.namespace).
		FilenameParam(o.enforceNamespace, &o.FilenameOptions).
		ResourceTypeOrNameArgs(true, args...).
		ContinueOnError().
		Flatten().
		Do()
	err := r.Err()
	if err != nil {
		return err
	}

	allErrs := []error{}
	err = r.Visit(func(info *resource.Info, err error) error {
		if err != nil {
			return err
		}
		var response runtime.Object
		objectName := info.Name
		podTemplateSpec, err := GetPodTemplateForObject(info.Object)
		if err != nil {
			return fmt.Errorf(" %q cannot create pod: %v", objectName, err)
		}
		err = CheckStatefulSetWithWolumeClaimTemplates(info.Object)
		if err != nil {
			return err
		}
		if len(userOrSA) > 0 || len(o.Groups) > 0 {
			response, err = o.pspSubjectReview(userOrSA, podTemplateSpec)
		} else {
			response, err = o.pspSelfSubjectReview(podTemplateSpec)
		}
		if err != nil {
			return fmt.Errorf("unable to compute Pod Security Policy Subject Review for %q: %v", objectName, err)
		}
		if err := o.printer.print(info, response, o.out); err != nil {
			allErrs = append(allErrs, err)
		}
		return nil
	})
	allErrs = append(allErrs, err)
	return utilerrors.NewAggregate(allErrs)
}

func (o *sccSubjectReviewOptions) pspSubjectReview(userOrSA string, podTemplateSpec *kapi.PodTemplateSpec) (*securityapi.PodSecurityPolicySubjectReview, error) {
	podSecurityPolicySubjectReview := &securityapi.PodSecurityPolicySubjectReview{
		Spec: securityapi.PodSecurityPolicySubjectReviewSpec{
			Template: *podTemplateSpec,
			User:     userOrSA,
			Groups:   o.Groups,
		},
	}
	return o.sccSubjectReviewClient.PodSecurityPolicySubjectReviews(o.namespace).Create(podSecurityPolicySubjectReview)
}

func (o *sccSubjectReviewOptions) pspSelfSubjectReview(podTemplateSpec *kapi.PodTemplateSpec) (*securityapi.PodSecurityPolicySelfSubjectReview, error) {
	podSecurityPolicySelfSubjectReview := &securityapi.PodSecurityPolicySelfSubjectReview{
		Spec: securityapi.PodSecurityPolicySelfSubjectReviewSpec{
			Template: *podTemplateSpec,
		},
	}
	return o.sccSelfSubjectReviewClient.PodSecurityPolicySelfSubjectReviews(o.namespace).Create(podSecurityPolicySelfSubjectReview)
}

type sccSubjectReviewPrinter interface {
	print(*resource.Info, runtime.Object, io.Writer) error
}

type sccSubjectReviewOutputPrinter struct {
	kubectl.ResourcePrinter
}

var _ sccSubjectReviewPrinter = &sccSubjectReviewOutputPrinter{}

func (s *sccSubjectReviewOutputPrinter) print(unused *resource.Info, obj runtime.Object, out io.Writer) error {
	return s.ResourcePrinter.PrintObj(obj, out)
}

type sccSubjectReviewHumanReadablePrinter struct {
	noHeaders bool
}

var _ sccSubjectReviewPrinter = &sccSubjectReviewHumanReadablePrinter{}

const (
	sccSubjectReviewTabWriterMinWidth = 0
	sccSubjectReviewTabWriterWidth    = 7
	sccSubjectReviewTabWriterPadding  = 3
	sccSubjectReviewTabWriterPadChar  = ' '
	sccSubjectReviewTabWriterFlags    = 0
)

func (s *sccSubjectReviewHumanReadablePrinter) print(info *resource.Info, obj runtime.Object, out io.Writer) error {
	w := tabwriter.NewWriter(out, sccSubjectReviewTabWriterMinWidth, sccSubjectReviewTabWriterWidth, sccSubjectReviewTabWriterPadding, sccSubjectReviewTabWriterPadChar, sccSubjectReviewTabWriterFlags)
	defer w.Flush()
	if s.noHeaders == false {
		columns := []string{"RESOURCE", "ALLOWED BY"}
		fmt.Fprintf(w, "%s\t\n", strings.Join(columns, "\t"))
		s.noHeaders = true // printed only the first time if requested
	}
	gvk, _, err := kapi.Scheme.ObjectKind(info.Object)
	if err != nil {
		return err
	}
	kind := gvk.Kind
	allowedBy, err := getAllowedBy(obj)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%s/%s\t%s\t\n", kind, info.Name, allowedBy)
	if err != nil {
		return err
	}
	return nil
}

func getAllowedBy(obj runtime.Object) (string, error) {
	value := "<none>"
	switch review := obj.(type) {
	case *securityapi.PodSecurityPolicySelfSubjectReview:
		if review.Status.AllowedBy != nil {
			value = review.Status.AllowedBy.Name
		}
	case *securityapi.PodSecurityPolicySubjectReview:
		if review.Status.AllowedBy != nil {
			value = review.Status.AllowedBy.Name
		}
	default:
		return value, fmt.Errorf("unexpected object %T", obj)
	}
	return value, nil
}
