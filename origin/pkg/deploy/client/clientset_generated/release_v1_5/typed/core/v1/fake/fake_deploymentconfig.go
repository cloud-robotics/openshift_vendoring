package fake

import (
	v1 "github.com/openshift/origin/pkg/deploy/api/v1"
	api "github.com/openshift/kubernetes/pkg/api"
	unversioned "github.com/openshift/kubernetes/pkg/api/unversioned"
	api_v1 "github.com/openshift/kubernetes/pkg/api/v1"
	core "github.com/openshift/kubernetes/pkg/client/testing/core"
	labels "github.com/openshift/kubernetes/pkg/labels"
	watch "github.com/openshift/kubernetes/pkg/watch"
)

// FakeDeploymentConfigs implements DeploymentConfigInterface
type FakeDeploymentConfigs struct {
	Fake *FakeCoreV1
	ns   string
}

var deploymentconfigsResource = unversioned.GroupVersionResource{Group: "", Version: "v1", Resource: "deploymentconfigs"}

func (c *FakeDeploymentConfigs) Create(deploymentConfig *v1.DeploymentConfig) (result *v1.DeploymentConfig, err error) {
	obj, err := c.Fake.
		Invokes(core.NewCreateAction(deploymentconfigsResource, c.ns, deploymentConfig), &v1.DeploymentConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.DeploymentConfig), err
}

func (c *FakeDeploymentConfigs) Update(deploymentConfig *v1.DeploymentConfig) (result *v1.DeploymentConfig, err error) {
	obj, err := c.Fake.
		Invokes(core.NewUpdateAction(deploymentconfigsResource, c.ns, deploymentConfig), &v1.DeploymentConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.DeploymentConfig), err
}

func (c *FakeDeploymentConfigs) UpdateStatus(deploymentConfig *v1.DeploymentConfig) (*v1.DeploymentConfig, error) {
	obj, err := c.Fake.
		Invokes(core.NewUpdateSubresourceAction(deploymentconfigsResource, "status", c.ns, deploymentConfig), &v1.DeploymentConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.DeploymentConfig), err
}

func (c *FakeDeploymentConfigs) Delete(name string, options *api_v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(core.NewDeleteAction(deploymentconfigsResource, c.ns, name), &v1.DeploymentConfig{})

	return err
}

func (c *FakeDeploymentConfigs) DeleteCollection(options *api_v1.DeleteOptions, listOptions api_v1.ListOptions) error {
	action := core.NewDeleteCollectionAction(deploymentconfigsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1.DeploymentConfigList{})
	return err
}

func (c *FakeDeploymentConfigs) Get(name string) (result *v1.DeploymentConfig, err error) {
	obj, err := c.Fake.
		Invokes(core.NewGetAction(deploymentconfigsResource, c.ns, name), &v1.DeploymentConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.DeploymentConfig), err
}

func (c *FakeDeploymentConfigs) List(opts api_v1.ListOptions) (result *v1.DeploymentConfigList, err error) {
	obj, err := c.Fake.
		Invokes(core.NewListAction(deploymentconfigsResource, c.ns, opts), &v1.DeploymentConfigList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := core.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.DeploymentConfigList{}
	for _, item := range obj.(*v1.DeploymentConfigList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested deploymentConfigs.
func (c *FakeDeploymentConfigs) Watch(opts api_v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(core.NewWatchAction(deploymentconfigsResource, c.ns, opts))

}

// Patch applies the patch and returns the patched deploymentConfig.
func (c *FakeDeploymentConfigs) Patch(name string, pt api.PatchType, data []byte, subresources ...string) (result *v1.DeploymentConfig, err error) {
	obj, err := c.Fake.
		Invokes(core.NewPatchSubresourceAction(deploymentconfigsResource, c.ns, name, data, subresources...), &v1.DeploymentConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.DeploymentConfig), err
}
