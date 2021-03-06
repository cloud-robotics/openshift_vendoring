package etcd

import (
	"github.com/openshift/kubernetes/pkg/fields"
	"github.com/openshift/kubernetes/pkg/labels"
	"github.com/openshift/kubernetes/pkg/registry/generic/registry"
	"github.com/openshift/kubernetes/pkg/runtime"
	"github.com/openshift/kubernetes/pkg/storage"

	"github.com/openshift/origin/pkg/template/api"
	tregistry "github.com/openshift/origin/pkg/template/registry"
	"github.com/openshift/origin/pkg/util/restoptions"
)

// REST implements a RESTStorage for templates against etcd
type REST struct {
	*registry.Store
}

// NewREST returns a RESTStorage object that will work against templates.
func NewREST(optsGetter restoptions.Getter) (*REST, error) {
	store := &registry.Store{
		NewFunc:     func() runtime.Object { return &api.Template{} },
		NewListFunc: func() runtime.Object { return &api.TemplateList{} },
		ObjectNameFunc: func(obj runtime.Object) (string, error) {
			return obj.(*api.Template).Name, nil
		},
		PredicateFunc: func(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
			return tregistry.Matcher(label, field)
		},
		QualifiedResource: api.Resource("templates"),

		CreateStrategy: tregistry.Strategy,
		UpdateStrategy: tregistry.Strategy,

		ReturnDeletedObject: true,
	}

	if err := restoptions.ApplyOptions(optsGetter, store, true, storage.NoTriggerPublisher); err != nil {
		return nil, err
	}

	return &REST{store}, nil
}
