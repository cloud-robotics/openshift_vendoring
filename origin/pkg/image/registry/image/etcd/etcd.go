package etcd

import (
	"github.com/openshift/kubernetes/pkg/fields"
	"github.com/openshift/kubernetes/pkg/labels"
	"github.com/openshift/kubernetes/pkg/registry/generic/registry"
	"github.com/openshift/kubernetes/pkg/runtime"
	"github.com/openshift/kubernetes/pkg/storage"

	"github.com/openshift/origin/pkg/image/api"
	"github.com/openshift/origin/pkg/image/registry/image"
	"github.com/openshift/origin/pkg/util/restoptions"
)

// REST implements a RESTStorage for images against etcd.
type REST struct {
	*registry.Store
}

// NewREST returns a new REST.
func NewREST(optsGetter restoptions.Getter) (*REST, error) {
	store := &registry.Store{
		NewFunc: func() runtime.Object { return &api.Image{} },

		// NewListFunc returns an object capable of storing results of an etcd list.
		NewListFunc: func() runtime.Object { return &api.ImageList{} },
		// Retrieve the name field of an image
		ObjectNameFunc: func(obj runtime.Object) (string, error) {
			return obj.(*api.Image).Name, nil
		},
		// Used to match objects based on labels/fields for list and watch
		PredicateFunc: func(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
			return image.Matcher(label, field)
		},
		QualifiedResource: api.Resource("images"),

		// Used to validate image creation
		CreateStrategy: image.Strategy,

		// Used to validate image updates
		UpdateStrategy: image.Strategy,

		ReturnDeletedObject: false,
	}

	if err := restoptions.ApplyOptions(optsGetter, store, false, storage.NoTriggerPublisher); err != nil {
		return nil, err
	}

	return &REST{store}, nil
}
