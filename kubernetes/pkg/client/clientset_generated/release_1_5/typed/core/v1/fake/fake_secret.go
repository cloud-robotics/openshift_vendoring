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

// FakeSecrets implements SecretInterface
type FakeSecrets struct {
	Fake *FakeCoreV1
	ns   string
}

var secretsResource = unversioned.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}

func (c *FakeSecrets) Create(secret *v1.Secret) (result *v1.Secret, err error) {
	obj, err := c.Fake.
		Invokes(core.NewCreateAction(secretsResource, c.ns, secret), &v1.Secret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Secret), err
}

func (c *FakeSecrets) Update(secret *v1.Secret) (result *v1.Secret, err error) {
	obj, err := c.Fake.
		Invokes(core.NewUpdateAction(secretsResource, c.ns, secret), &v1.Secret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Secret), err
}

func (c *FakeSecrets) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(core.NewDeleteAction(secretsResource, c.ns, name), &v1.Secret{})

	return err
}

func (c *FakeSecrets) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := core.NewDeleteCollectionAction(secretsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1.SecretList{})
	return err
}

func (c *FakeSecrets) Get(name string) (result *v1.Secret, err error) {
	obj, err := c.Fake.
		Invokes(core.NewGetAction(secretsResource, c.ns, name), &v1.Secret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Secret), err
}

func (c *FakeSecrets) List(opts v1.ListOptions) (result *v1.SecretList, err error) {
	obj, err := c.Fake.
		Invokes(core.NewListAction(secretsResource, c.ns, opts), &v1.SecretList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := core.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.SecretList{}
	for _, item := range obj.(*v1.SecretList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested secrets.
func (c *FakeSecrets) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(core.NewWatchAction(secretsResource, c.ns, opts))

}

// Patch applies the patch and returns the patched secret.
func (c *FakeSecrets) Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *v1.Secret, err error) {
	obj, err := c.Fake.
		Invokes(core.NewPatchSubresourceAction(secretsResource, c.ns, name, data, subresources...), &v1.Secret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.Secret), err
}
