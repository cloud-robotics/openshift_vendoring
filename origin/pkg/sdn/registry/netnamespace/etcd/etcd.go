package etcd

import (
	"github.com/openshift/kubernetes/pkg/fields"
	"github.com/openshift/kubernetes/pkg/labels"
	"github.com/openshift/kubernetes/pkg/registry/generic/registry"
	"github.com/openshift/kubernetes/pkg/runtime"
	"github.com/openshift/kubernetes/pkg/storage"

	"github.com/openshift/origin/pkg/sdn/api"
	"github.com/openshift/origin/pkg/sdn/registry/netnamespace"
	"github.com/openshift/origin/pkg/util/restoptions"
)

// rest implements a RESTStorage for sdn against etcd
type REST struct {
	registry.Store
}

// NewREST returns a RESTStorage object that will work against netnamespaces
func NewREST(optsGetter restoptions.Getter) (*REST, error) {
	store := &registry.Store{
		NewFunc:     func() runtime.Object { return &api.NetNamespace{} },
		NewListFunc: func() runtime.Object { return &api.NetNamespaceList{} },
		ObjectNameFunc: func(obj runtime.Object) (string, error) {
			return obj.(*api.NetNamespace).NetName, nil
		},
		PredicateFunc: func(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
			return netnamespace.Matcher(label, field)
		},
		QualifiedResource: api.Resource("netnamespaces"),

		CreateStrategy: netnamespace.Strategy,
		UpdateStrategy: netnamespace.Strategy,
	}

	if err := restoptions.ApplyOptions(optsGetter, store, false, storage.NoTriggerPublisher); err != nil {
		return nil, err
	}

	return &REST{*store}, nil
}
