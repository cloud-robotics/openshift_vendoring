package restoptions

import (
	"fmt"
	"strings"

	kapi "github.com/openshift/kubernetes/pkg/api"
	"github.com/openshift/kubernetes/pkg/registry/generic/registry"
	"github.com/openshift/kubernetes/pkg/storage"
)

// DefaultKeyFunctions sets the default behavior for storage key generation onto a Store.
func DefaultKeyFunctions(store *registry.Store, prefix string, isNamespaced bool) {
	if isNamespaced {
		if store.KeyRootFunc == nil {
			store.KeyRootFunc = func(ctx kapi.Context) string {
				return registry.NamespaceKeyRootFunc(ctx, prefix)
			}
		}
		if store.KeyFunc == nil {
			store.KeyFunc = func(ctx kapi.Context, name string) (string, error) {
				return registry.NamespaceKeyFunc(ctx, prefix, name)
			}
		}
	} else {
		if store.KeyRootFunc == nil {
			store.KeyRootFunc = func(ctx kapi.Context) string {
				return prefix
			}
		}
		if store.KeyFunc == nil {
			store.KeyFunc = func(ctx kapi.Context, name string) (string, error) {
				return registry.NoNamespaceKeyFunc(ctx, prefix, name)
			}
		}
	}
}

// ApplyOptions updates the given generic storage from the provided rest options
// TODO: remove need for etcdPrefix once Decorator interface is refactored upstream
func ApplyOptions(optsGetter Getter, store *registry.Store, oldIsNamespaced bool, triggerFn storage.TriggerPublisherFunc) error {
	if store.QualifiedResource.Empty() {
		return fmt.Errorf("store must have a non-empty qualified resource")
	}
	if store.NewFunc == nil {
		return fmt.Errorf("store for %s must have NewFunc set", store.QualifiedResource.String())
	}
	if store.NewListFunc == nil {
		return fmt.Errorf("store for %s must have NewListFunc set", store.QualifiedResource.String())
	}
	if store.CreateStrategy == nil {
		return fmt.Errorf("store for %s must have CreateStrategy set", store.QualifiedResource.String())
	}

	isNamespaced := store.CreateStrategy.NamespaceScoped()
	if isNamespaced != oldIsNamespaced { // TODO(soltysh): oldIsNamespaced should be completely removed in #12541
		return fmt.Errorf("CreateStrategy has %v for namespace scope but user specified %v as namespace scope", isNamespaced, oldIsNamespaced)
	}

	opts, err := optsGetter.GetRESTOptions(store.QualifiedResource)
	if err != nil {
		return fmt.Errorf("error building RESTOptions for %s store: %v", store.QualifiedResource.String(), err)
	}

	// Resource prefix must come from the underlying factory
	prefix := opts.ResourcePrefix
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}

	DefaultKeyFunctions(store, prefix, isNamespaced)

	store.DeleteCollectionWorkers = opts.DeleteCollectionWorkers
	store.Storage, store.DestroyFunc = opts.Decorator(
		opts.StorageConfig,
		UseConfiguredCacheSize,
		store.NewFunc(),
		prefix,
		store.CreateStrategy,
		store.NewListFunc,
		triggerFn,
	)
	return nil
}
