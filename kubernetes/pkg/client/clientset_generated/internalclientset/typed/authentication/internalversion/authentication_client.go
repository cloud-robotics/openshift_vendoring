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

package internalversion

import (
	api "github.com/openshift/kubernetes/pkg/api"
	registered "github.com/openshift/kubernetes/pkg/apimachinery/registered"
	restclient "github.com/openshift/kubernetes/pkg/client/restclient"
)

type AuthenticationInterface interface {
	RESTClient() restclient.Interface
	TokenReviewsGetter
}

// AuthenticationClient is used to interact with features provided by the k8s.io/kubernetes/pkg/apimachinery/registered.Group group.
type AuthenticationClient struct {
	restClient restclient.Interface
}

func (c *AuthenticationClient) TokenReviews() TokenReviewInterface {
	return newTokenReviews(c)
}

// NewForConfig creates a new AuthenticationClient for the given config.
func NewForConfig(c *restclient.Config) (*AuthenticationClient, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := restclient.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &AuthenticationClient{client}, nil
}

// NewForConfigOrDie creates a new AuthenticationClient for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *restclient.Config) *AuthenticationClient {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new AuthenticationClient for the given RESTClient.
func New(c restclient.Interface) *AuthenticationClient {
	return &AuthenticationClient{c}
}

func setConfigDefaults(config *restclient.Config) error {
	// if authentication group is not registered, return an error
	g, err := registered.Group("authentication.k8s.io")
	if err != nil {
		return err
	}
	config.APIPath = "/apis"
	if config.UserAgent == "" {
		config.UserAgent = restclient.DefaultKubernetesUserAgent()
	}
	if config.GroupVersion == nil || config.GroupVersion.Group != g.GroupVersion.Group {
		copyGroupVersion := g.GroupVersion
		config.GroupVersion = &copyGroupVersion
	}
	config.NegotiatedSerializer = api.Codecs

	if config.QPS == 0 {
		config.QPS = 5
	}
	if config.Burst == 0 {
		config.Burst = 10
	}
	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *AuthenticationClient) RESTClient() restclient.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
