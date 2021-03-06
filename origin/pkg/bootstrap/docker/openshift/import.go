package openshift

import (
	"bytes"

	"github.com/golang/glog"
	"github.com/openshift/kubernetes/pkg/api/errors"
	cmdutil "github.com/openshift/kubernetes/pkg/kubectl/cmd/util"
	"github.com/openshift/kubernetes/pkg/kubectl/resource"

	"github.com/openshift/origin/pkg/bootstrap"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
)

// ImportObjects imports objects into OpenShift from a particular location
// into a given namespace
func ImportObjects(f *clientcmd.Factory, ns, location string) error {
	mapper, typer := f.Object()
	schema, err := f.Validator(false, "")
	if err != nil {
		return err
	}
	data, err := bootstrap.Asset(location)
	if err != nil {
		return err
	}
	glog.V(8).Infof("Importing data:\n%s\n", string(data))
	r := resource.NewBuilder(mapper, typer, resource.ClientMapperFunc(f.ClientForMapping), f.Decoder(true)).
		Schema(schema).
		ContinueOnError().
		NamespaceParam(ns).
		DefaultNamespace().
		Stream(bytes.NewBuffer(data), location).
		Flatten().
		Do()

	return r.Visit(func(info *resource.Info, err error) error {
		if err != nil {
			return err
		}
		glog.V(5).Infof("Creating %s/%s", info.Namespace, info.Name)
		if err = createAndRefresh(info); err != nil {
			return cmdutil.AddSourceToErr("creating", info.Source, err)
		}
		return nil
	})
}

func createAndRefresh(info *resource.Info) error {
	obj, err := resource.NewHelper(info.Client, info.Mapping).Create(info.Namespace, true, info.Object)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			glog.V(5).Infof("Object %s/%s already exists", info.Namespace, info.Name)
			return nil
		}
		return err
	}
	info.Refresh(obj, true)
	return nil
}
