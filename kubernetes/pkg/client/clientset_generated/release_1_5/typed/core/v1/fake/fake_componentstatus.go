/*
Copyright 2017 The Kubernetes Authors.

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

package fake

import (
	api "github.com/openshift/kubernetes/pkg/api"
	unversioned "github.com/openshift/kubernetes/pkg/api/unversioned"
	v1 "github.com/openshift/kubernetes/pkg/api/v1"
	core "github.com/openshift/kubernetes/pkg/client/testing/core"
	labels "github.com/openshift/kubernetes/pkg/labels"
	watch "github.com/openshift/kubernetes/pkg/watch"
)

// FakeComponentStatuses implements ComponentStatusInterface
type FakeComponentStatuses struct {
	Fake *FakeCoreV1
}

var componentstatusesResource = unversioned.GroupVersionResource{Group: "", Version: "v1", Resource: "componentstatuses"}

func (c *FakeComponentStatuses) Create(componentStatus *v1.ComponentStatus) (result *v1.ComponentStatus, err error) {
	obj, err := c.Fake.
		Invokes(core.NewRootCreateAction(componentstatusesResource, componentStatus), &v1.ComponentStatus{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1.ComponentStatus), err
}

func (c *FakeComponentStatuses) Update(componentStatus *v1.ComponentStatus) (result *v1.ComponentStatus, err error) {
	obj, err := c.Fake.
		Invokes(core.NewRootUpdateAction(componentstatusesResource, componentStatus), &v1.ComponentStatus{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1.ComponentStatus), err
}

func (c *FakeComponentStatuses) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(core.NewRootDeleteAction(componentstatusesResource, name), &v1.ComponentStatus{})
	return err
}

func (c *FakeComponentStatuses) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := core.NewRootDeleteCollectionAction(componentstatusesResource, listOptions)

	_, err := c.Fake.Invokes(action, &v1.ComponentStatusList{})
	return err
}

func (c *FakeComponentStatuses) Get(name string) (result *v1.ComponentStatus, err error) {
	obj, err := c.Fake.
		Invokes(core.NewRootGetAction(componentstatusesResource, name), &v1.ComponentStatus{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1.ComponentStatus), err
}

func (c *FakeComponentStatuses) List(opts v1.ListOptions) (result *v1.ComponentStatusList, err error) {
	obj, err := c.Fake.
		Invokes(core.NewRootListAction(componentstatusesResource, opts), &v1.ComponentStatusList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := core.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.ComponentStatusList{}
	for _, item := range obj.(*v1.ComponentStatusList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested componentStatuses.
func (c *FakeComponentStatuses) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(core.NewRootWatchAction(componentstatusesResource, opts))
}

// Patch applies the patch and returns the patched componentStatus.
func (c *FakeComponentStatuses) Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *v1.ComponentStatus, err error) {
	obj, err := c.Fake.
		Invokes(core.NewRootPatchSubresourceAction(componentstatusesResource, name, data, subresources...), &v1.ComponentStatus{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1.ComponentStatus), err
}
