package service

import (
	"testing"

	kapi "github.com/openshift/kubernetes/pkg/api"
	"github.com/openshift/kubernetes/pkg/api/errors"
	"github.com/openshift/kubernetes/pkg/client/clientset_generated/internalclientset/fake"
)

func TestServiceResolverCacheEmpty(t *testing.T) {
	fakeClient := fake.NewSimpleClientset(&kapi.Service{
		ObjectMeta: kapi.ObjectMeta{
			Name:      "foo",
			Namespace: kapi.NamespaceDefault,
		},
		Spec: kapi.ServiceSpec{
			Ports: []kapi.ServicePort{{Port: 80}},
		},
	})
	cache := NewServiceResolverCache(fakeClient.Core().Services("default").Get)
	if v, ok := cache.resolve("FOO_SERVICE_HOST"); v != "" || !ok {
		t.Errorf("unexpected cache item")
	}
	if len(fakeClient.Actions()) != 1 {
		t.Errorf("unexpected client actions: %#v", fakeClient.Actions())
	}
	cache.resolve("FOO_SERVICE_HOST")
	if len(fakeClient.Actions()) != 1 {
		t.Errorf("unexpected cache miss: %#v", fakeClient.Actions())
	}
	cache.resolve("FOO_SERVICE_PORT")
	if len(fakeClient.Actions()) != 1 {
		t.Errorf("unexpected cache miss: %#v", fakeClient.Actions())
	}
}

type fakeRetriever struct {
	service *kapi.Service
	err     error
}

func (r fakeRetriever) Get(name string) (*kapi.Service, error) {
	return r.service, r.err
}

func TestServiceResolverCache(t *testing.T) {
	c := fakeRetriever{
		err: errors.NewNotFound(kapi.Resource("Service"), "bar"),
	}
	cache := NewServiceResolverCache(c.Get)
	if v, ok := cache.resolve("FOO_SERVICE_HOST"); v != "" || ok {
		t.Errorf("unexpected cache item")
	}

	c = fakeRetriever{
		service: &kapi.Service{
			Spec: kapi.ServiceSpec{
				ClusterIP: "127.0.0.1",
				Ports:     []kapi.ServicePort{{Port: 80}},
			},
		},
	}
	cache = NewServiceResolverCache(c.Get)
	if v, ok := cache.resolve("FOO_SERVICE_HOST"); v != "127.0.0.1" || !ok {
		t.Errorf("unexpected cache item")
	}
	if v, ok := cache.resolve("FOO_SERVICE_PORT"); v != "80" || !ok {
		t.Errorf("unexpected cache item")
	}
	if _, err := cache.Defer("${UNKNOWN}"); err == nil {
		t.Errorf("unexpected non-error")
	}
	fn, err := cache.Defer("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v, ok := fn(); v != "test" || !ok {
		t.Errorf("unexpected cache item")
	}
	fn, err = cache.Defer("${FOO_SERVICE_HOST}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v, ok := fn(); v != "127.0.0.1" || !ok {
		t.Errorf("unexpected cache item")
	}
	if v, ok := fn(); v != "127.0.0.1" || !ok {
		t.Errorf("unexpected cache item")
	}
}
