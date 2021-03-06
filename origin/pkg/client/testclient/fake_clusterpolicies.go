package testclient

import (
	kapi "github.com/openshift/kubernetes/pkg/api"
	"github.com/openshift/kubernetes/pkg/api/unversioned"
	"github.com/openshift/kubernetes/pkg/client/testing/core"
	"github.com/openshift/kubernetes/pkg/watch"

	authorizationapi "github.com/openshift/origin/pkg/authorization/api"
)

// FakeClusterPolicies implements ClusterPolicyInterface. Meant to be embedded into a struct to get a default
// implementation. This makes faking out just the methods you want to test easier.
type FakeClusterPolicies struct {
	Fake *Fake
}

var clusterPoliciesResource = unversioned.GroupVersionResource{Group: "", Version: "", Resource: "clusterpolicies"}

func (c *FakeClusterPolicies) Get(name string) (*authorizationapi.ClusterPolicy, error) {
	obj, err := c.Fake.Invokes(core.NewRootGetAction(clusterPoliciesResource, name), &authorizationapi.ClusterPolicy{})
	if obj == nil {
		return nil, err
	}

	return obj.(*authorizationapi.ClusterPolicy), err
}

func (c *FakeClusterPolicies) List(opts kapi.ListOptions) (*authorizationapi.ClusterPolicyList, error) {
	obj, err := c.Fake.Invokes(core.NewRootListAction(clusterPoliciesResource, opts), &authorizationapi.ClusterPolicyList{})
	if obj == nil {
		return nil, err
	}

	return obj.(*authorizationapi.ClusterPolicyList), err
}

func (c *FakeClusterPolicies) Delete(name string) error {
	_, err := c.Fake.Invokes(core.NewRootDeleteAction(clusterPoliciesResource, name), &authorizationapi.ClusterPolicy{})
	return err
}

func (c *FakeClusterPolicies) Watch(opts kapi.ListOptions) (watch.Interface, error) {
	return c.Fake.InvokesWatch(core.NewRootWatchAction(clusterPoliciesResource, opts))
}
