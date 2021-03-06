package identity

import (
	"fmt"

	kapi "github.com/openshift/kubernetes/pkg/api"
	"github.com/openshift/kubernetes/pkg/fields"
	"github.com/openshift/kubernetes/pkg/labels"
	"github.com/openshift/kubernetes/pkg/runtime"
	kstorage "github.com/openshift/kubernetes/pkg/storage"
	"github.com/openshift/kubernetes/pkg/util/validation/field"

	"github.com/openshift/origin/pkg/user/api"
	"github.com/openshift/origin/pkg/user/api/validation"
)

// identityStrategy implements behavior for Identities
type identityStrategy struct {
	runtime.ObjectTyper
}

// Strategy is the default logic that applies when creating and updating Identity
// objects via the REST API.
var Strategy = identityStrategy{kapi.Scheme}

func (identityStrategy) PrepareForUpdate(ctx kapi.Context, obj, old runtime.Object) {}

// NamespaceScoped is false for users
func (identityStrategy) NamespaceScoped() bool {
	return false
}

func (identityStrategy) GenerateName(base string) string {
	return base
}

func (identityStrategy) PrepareForCreate(ctx kapi.Context, obj runtime.Object) {
	identity := obj.(*api.Identity)
	identity.Name = identityName(identity.ProviderName, identity.ProviderUserName)
}

// Validate validates a new user
func (identityStrategy) Validate(ctx kapi.Context, obj runtime.Object) field.ErrorList {
	identity := obj.(*api.Identity)
	return validation.ValidateIdentity(identity)
}

// AllowCreateOnUpdate is false for identity
func (identityStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (identityStrategy) AllowUnconditionalUpdate() bool {
	return false
}

// Canonicalize normalizes the object after validation.
func (identityStrategy) Canonicalize(obj runtime.Object) {
}

// ValidateUpdate is the default update validation for an identity
func (identityStrategy) ValidateUpdate(ctx kapi.Context, obj, old runtime.Object) field.ErrorList {
	return validation.ValidateIdentityUpdate(obj.(*api.Identity), old.(*api.Identity))
}

// Matcher returns a generic matcher for a given label and field selector.
func Matcher(label labels.Selector, field fields.Selector) kstorage.SelectionPredicate {
	return kstorage.SelectionPredicate{
		Label: label,
		Field: field,
		GetAttrs: func(o runtime.Object) (labels.Set, fields.Set, error) {
			obj, ok := o.(*api.Identity)
			if !ok {
				return nil, nil, fmt.Errorf("not an Identity")
			}
			return labels.Set(obj.Labels), SelectableFields(obj), nil
		},
	}
}

// SelectableFields returns a field set that can be used for filter selection
func SelectableFields(obj *api.Identity) fields.Set {
	return api.IdentityToSelectableFields(obj)
}
