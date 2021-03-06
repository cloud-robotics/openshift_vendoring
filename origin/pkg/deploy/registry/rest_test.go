package registry

import (
	"testing"
	"time"

	kapi "github.com/openshift/kubernetes/pkg/api"
	"github.com/openshift/kubernetes/pkg/api/unversioned"
	"github.com/openshift/kubernetes/pkg/client/clientset_generated/internalclientset/fake"
	"github.com/openshift/kubernetes/pkg/client/testing/core"
	"github.com/openshift/kubernetes/pkg/runtime"
	"github.com/openshift/kubernetes/pkg/watch"

	deployapi "github.com/openshift/origin/pkg/deploy/api"
)

func TestWaitForRunningDeploymentSuccess(t *testing.T) {
	fakeController := &kapi.ReplicationController{}
	fakeController.Name = "test-1"
	fakeController.Namespace = "test"

	kubeclient := fake.NewSimpleClientset([]runtime.Object{fakeController}...)
	fakeWatch := watch.NewFake()
	kubeclient.PrependWatchReactor("replicationcontrollers", core.DefaultWatchReactor(fakeWatch, nil))
	stopChan := make(chan struct{})

	go func() {
		defer close(stopChan)
		rc, ok, err := WaitForRunningDeployment(kubeclient.Core(), fakeController, 10*time.Second)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Errorf("expected to return success")
		}
		if rc == nil {
			t.Errorf("expected returned replication controller to not be nil")
		}
	}()

	fakeController.Annotations = map[string]string{deployapi.DeploymentStatusAnnotation: string(deployapi.DeploymentStatusRunning)}
	fakeWatch.Modify(fakeController)
	<-stopChan
}

func TestWaitForRunningDeploymentRestartWatch(t *testing.T) {
	fakeController := &kapi.ReplicationController{}
	fakeController.Name = "test-1"
	fakeController.Namespace = "test"

	kubeclient := fake.NewSimpleClientset([]runtime.Object{fakeController}...)
	fakeWatch := watch.NewFake()

	watchCalledChan := make(chan struct{})
	kubeclient.PrependWatchReactor("replicationcontrollers", func(action core.Action) (bool, watch.Interface, error) {
		fakeWatch.Reset()
		watchCalledChan <- struct{}{}
		return core.DefaultWatchReactor(fakeWatch, nil)(action)
	})

	getReceivedChan := make(chan struct{})
	kubeclient.PrependReactor("get", "replicationcontrollers", func(action core.Action) (bool, runtime.Object, error) {
		close(getReceivedChan)
		return true, fakeController, nil
	})

	stopChan := make(chan struct{})
	go func() {
		defer close(stopChan)
		rc, ok, err := WaitForRunningDeployment(kubeclient.Core(), fakeController, 10*time.Second)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Errorf("expected to return success")
		}
		if rc == nil {
			t.Errorf("expected returned replication controller to not be nil")
		}
	}()

	select {
	case <-watchCalledChan:
	case <-time.After(time.Second * 5):
		t.Fatalf("timeout waiting for the watch to start")
	}

	// Send the StatusReasonGone error to watcher which should trigger the watch restart.
	goneError := &unversioned.Status{Reason: unversioned.StatusReasonGone}
	fakeWatch.Error(goneError)

	// Make sure we observed the "get" action on replication controller, so the watch gets
	// the latest resourceVersion.
	select {
	case <-getReceivedChan:
	case <-time.After(time.Second * 5):
		t.Fatalf("timeout waiting for get on replication controllers")
	}

	// Wait for the watcher to restart and then transition the replication controller to
	// running state.
	select {
	case <-watchCalledChan:
		fakeController.Annotations = map[string]string{deployapi.DeploymentStatusAnnotation: string(deployapi.DeploymentStatusRunning)}
		fakeWatch.Modify(fakeController)
		<-stopChan
	case <-time.After(time.Second * 5):
		t.Fatalf("timeout waiting for the watch restart")
	}
}
