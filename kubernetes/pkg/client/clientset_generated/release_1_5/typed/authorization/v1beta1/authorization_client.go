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

package v1beta1

import (
	fmt "fmt"
	api "github.com/openshift/kubernetes/pkg/api"
	unversioned "github.com/openshift/kubernetes/pkg/api/unversioned"
	registered "github.com/openshift/kubernetes/pkg/apimachinery/registered"
	restclient "github.com/openshift/kubernetes/pkg/client/restclient"
	serializer "github.com/openshift/kubernetes/pkg/runtime/serializer"
)

type AuthorizationV1beta1Interface interface {
	RESTClient() restclient.Interface
	LocalSubjectAccessReviewsGetter
	SelfSubjectAccessReviewsGetter
	SubjectAccessReviewsGetter
}

// AuthorizationV1beta1Client is used to interact with features provided by the k8s.io/kubernetes/pkg/apimachinery/registered.Group group.
type AuthorizationV1beta1Client struct {
	restClient restclient.Interface
}

func (c *AuthorizationV1beta1Client) LocalSubjectAccessReviews(namespace string) LocalSubjectAccessReviewInterface {
	return newLocalSubjectAccessReviews(c, namespace)
}

func (c *AuthorizationV1beta1Client) SelfSubjectAccessReviews() SelfSubjectAccessReviewInterface {
	return newSelfSubjectAccessReviews(c)
}

func (c *AuthorizationV1beta1Client) SubjectAccessReviews() SubjectAccessReviewInterface {
	return newSubjectAccessReviews(c)
}

// NewForConfig creates a new AuthorizationV1beta1Client for the given config.
func NewForConfig(c *restclient.Config) (*AuthorizationV1beta1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := restclient.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &AuthorizationV1beta1Client{client}, nil
}

// NewForConfigOrDie creates a new AuthorizationV1beta1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *restclient.Config) *AuthorizationV1beta1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new AuthorizationV1beta1Client for the given RESTClient.
func New(c restclient.Interface) *AuthorizationV1beta1Client {
	return &AuthorizationV1beta1Client{c}
}

func setConfigDefaults(config *restclient.Config) error {
	gv, err := unversioned.ParseGroupVersion("authorization.k8s.io/v1beta1")
	if err != nil {
		return err
	}
	// if authorization.k8s.io/v1beta1 is not enabled, return an error
	if !registered.IsEnabledVersion(gv) {
		return fmt.Errorf("authorization.k8s.io/v1beta1 is not enabled")
	}
	config.APIPath = "/apis"
	if config.UserAgent == "" {
		config.UserAgent = restclient.DefaultKubernetesUserAgent()
	}
	copyGroupVersion := gv
	config.GroupVersion = &copyGroupVersion

	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: api.Codecs}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *AuthorizationV1beta1Client) RESTClient() restclient.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
