package client

import (
	"net/url"

	"github.com/openshift/kubernetes/pkg/api/errors"
	"github.com/openshift/kubernetes/pkg/api/unversioned"
	"github.com/openshift/kubernetes/pkg/client/restclient"
	"github.com/openshift/kubernetes/pkg/client/typed/discovery"
)

// DiscoveryClient implements the functions that discovery server-supported API groups,
// versions and resources.
type DiscoveryClient struct {
	*discovery.DiscoveryClient
}

// ServerResourcesForGroupVersion returns the supported resources for a group and version.
func (d *DiscoveryClient) ServerResourcesForGroupVersion(groupVersion string) (resources *unversioned.APIResourceList, err error) {
	parentList, err := d.DiscoveryClient.ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		return parentList, err
	}

	if groupVersion != "v1" {
		return parentList, nil
	}

	// we request v1, we must combine the parent list with the list from /oapi

	url := url.URL{}
	url.Path = "/oapi/" + groupVersion
	originResources := &unversioned.APIResourceList{}
	err = d.RESTClient().Get().AbsPath(url.String()).Do().Into(originResources)
	if err != nil {
		// ignore 403 or 404 error to be compatible with an v1.0 server.
		if groupVersion == "v1" && (errors.IsNotFound(err) || errors.IsForbidden(err)) {
			return parentList, nil
		}
		return nil, err
	}

	parentList.APIResources = append(parentList.APIResources, originResources.APIResources...)
	return parentList, nil
}

// ServerResources returns the supported resources for all groups and versions.
func (d *DiscoveryClient) ServerResources() (map[string]*unversioned.APIResourceList, error) {
	apiGroups, err := d.ServerGroups()
	if err != nil {
		return nil, err
	}
	groupVersions := unversioned.ExtractGroupVersions(apiGroups)
	result := map[string]*unversioned.APIResourceList{}
	for _, groupVersion := range groupVersions {
		resources, err := d.ServerResourcesForGroupVersion(groupVersion)
		if err != nil {
			return nil, err
		}
		result[groupVersion] = resources
	}
	return result, nil
}

// New creates a new DiscoveryClient for the given RESTClient.
func NewDiscoveryClient(c restclient.Interface) *DiscoveryClient {
	return &DiscoveryClient{discovery.NewDiscoveryClient(c)}
}
