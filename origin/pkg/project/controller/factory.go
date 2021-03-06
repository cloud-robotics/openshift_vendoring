package controller

import (
	"time"

	kapi "github.com/openshift/kubernetes/pkg/api"
	"github.com/openshift/kubernetes/pkg/client/cache"
	kclientset "github.com/openshift/kubernetes/pkg/client/clientset_generated/internalclientset"
	"github.com/openshift/kubernetes/pkg/runtime"
	"github.com/openshift/kubernetes/pkg/util/flowcontrol"
	utilruntime "github.com/openshift/kubernetes/pkg/util/runtime"
	"github.com/openshift/kubernetes/pkg/watch"

	osclient "github.com/openshift/origin/pkg/client"
	controller "github.com/openshift/origin/pkg/controller"
)

type NamespaceControllerFactory struct {
	// Client is an OpenShift client.
	Client osclient.Interface
	// KubeClient is a Kubernetes client.
	KubeClient *kclientset.Clientset
}

// Create creates a NamespaceController.
func (factory *NamespaceControllerFactory) Create() controller.RunnableController {
	namespaceLW := &cache.ListWatch{
		ListFunc: func(options kapi.ListOptions) (runtime.Object, error) {
			return factory.KubeClient.Core().Namespaces().List(options)
		},
		WatchFunc: func(options kapi.ListOptions) (watch.Interface, error) {
			return factory.KubeClient.Core().Namespaces().Watch(options)
		},
	}
	queue := cache.NewResyncableFIFO(cache.MetaNamespaceKeyFunc)
	cache.NewReflector(namespaceLW, &kapi.Namespace{}, queue, 1*time.Minute).Run()

	namespaceController := &NamespaceController{
		Client:     factory.Client,
		KubeClient: factory.KubeClient,
	}

	return &controller.RetryController{
		Queue: queue,
		RetryManager: controller.NewQueueRetryManager(
			queue,
			cache.MetaNamespaceKeyFunc,
			func(obj interface{}, err error, retries controller.Retry) bool {
				utilruntime.HandleError(err)
				if _, isFatal := err.(fatalError); isFatal {
					return false
				}
				if retries.Count > 0 {
					return false
				}
				return true
			},
			flowcontrol.NewTokenBucketRateLimiter(1, 10),
		),
		Handle: func(obj interface{}) error {
			namespace := obj.(*kapi.Namespace)
			return namespaceController.Handle(namespace)
		},
	}
}
