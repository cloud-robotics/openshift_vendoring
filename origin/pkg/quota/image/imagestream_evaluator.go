package image

import (
	"github.com/openshift/kubernetes/pkg/admission"
	kapi "github.com/openshift/kubernetes/pkg/api"
	kquota "github.com/openshift/kubernetes/pkg/quota"
	"github.com/openshift/kubernetes/pkg/quota/generic"
	"github.com/openshift/kubernetes/pkg/runtime"

	oscache "github.com/openshift/origin/pkg/client/cache"
	imageapi "github.com/openshift/origin/pkg/image/api"
)

const imageStreamEvaluatorName = "Evaluator.ImageStream"

// NewImageStreamEvaluator computes resource usage of ImageStreams. Instantiating this is necessary for
// resource quota admission controller to properly work on image stream related objects.
func NewImageStreamEvaluator(store *oscache.StoreToImageStreamLister) kquota.Evaluator {
	allResources := []kapi.ResourceName{
		imageapi.ResourceImageStreams,
	}

	return &generic.GenericEvaluator{
		Name:              imageStreamEvaluatorName,
		InternalGroupKind: imageapi.Kind("ImageStream"),
		InternalOperationResources: map[admission.Operation][]kapi.ResourceName{
			admission.Create: allResources,
		},
		MatchedResourceNames: allResources,
		MatchesScopeFunc:     generic.MatchesNoScopeFunc,
		ConstraintsFunc:      generic.ObjectCountConstraintsFunc(imageapi.ResourceImageStreams),
		UsageFunc:            generic.ObjectCountUsageFunc(imageapi.ResourceImageStreams),
		ListFuncByNamespace: func(namespace string, options kapi.ListOptions) ([]runtime.Object, error) {
			list, err := store.ImageStreams(namespace).List(options.LabelSelector)
			if err != nil {
				return nil, err
			}
			results := make([]runtime.Object, 0, len(list))
			for _, is := range list {
				results = append(results, is)
			}
			return results, nil
		},
	}
}
