/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package limitrange

import (
	"fmt"

	"github.com/openshift/kubernetes/pkg/api"
	"github.com/openshift/kubernetes/pkg/api/validation"
	"github.com/openshift/kubernetes/pkg/fields"
	"github.com/openshift/kubernetes/pkg/labels"
	"github.com/openshift/kubernetes/pkg/runtime"
	"github.com/openshift/kubernetes/pkg/storage"
	"github.com/openshift/kubernetes/pkg/util/uuid"
	"github.com/openshift/kubernetes/pkg/util/validation/field"
)

type limitrangeStrategy struct {
	runtime.ObjectTyper
	api.NameGenerator
}

// Strategy is the default logic that applies when creating and updating
// LimitRange objects via the REST API.
var Strategy = limitrangeStrategy{api.Scheme, api.SimpleNameGenerator}

func (limitrangeStrategy) NamespaceScoped() bool {
	return true
}

func (limitrangeStrategy) PrepareForCreate(ctx api.Context, obj runtime.Object) {
	limitRange := obj.(*api.LimitRange)
	if len(limitRange.Name) == 0 {
		limitRange.Name = string(uuid.NewUUID())
	}
}

func (limitrangeStrategy) PrepareForUpdate(ctx api.Context, obj, old runtime.Object) {
}

func (limitrangeStrategy) Validate(ctx api.Context, obj runtime.Object) field.ErrorList {
	limitRange := obj.(*api.LimitRange)
	return validation.ValidateLimitRange(limitRange)
}

// Canonicalize normalizes the object after validation.
func (limitrangeStrategy) Canonicalize(obj runtime.Object) {
}

func (limitrangeStrategy) AllowCreateOnUpdate() bool {
	return true
}

func (limitrangeStrategy) ValidateUpdate(ctx api.Context, obj, old runtime.Object) field.ErrorList {
	limitRange := obj.(*api.LimitRange)
	return validation.ValidateLimitRange(limitRange)
}

func (limitrangeStrategy) AllowUnconditionalUpdate() bool {
	return true
}

func LimitRangeToSelectableFields(limitRange *api.LimitRange) fields.Set {
	return nil
}

func (limitrangeStrategy) Export(api.Context, runtime.Object, bool) error {
	// Copied from OpenShift exporter
	// TODO: this needs to be fixed
	//  limitrange.Strategy.PrepareForCreate(ctx, obj)
	return nil
}

func MatchLimitRange(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label: label,
		Field: field,
		GetAttrs: func(obj runtime.Object) (labels.Set, fields.Set, error) {
			lr, ok := obj.(*api.LimitRange)
			if !ok {
				return nil, nil, fmt.Errorf("given object is not a limit range.")
			}
			return labels.Set(lr.ObjectMeta.Labels), LimitRangeToSelectableFields(lr), nil
		},
	}
}
