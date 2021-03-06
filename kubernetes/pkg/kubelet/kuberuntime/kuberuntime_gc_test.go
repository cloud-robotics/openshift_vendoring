/*
Copyright 2016 The Kubernetes Authors.

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

package kuberuntime

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/openshift/github.com/golang/mock/gomock"
	"github.com/openshift/github.com/stretchr/testify/assert"
	"github.com/openshift/kubernetes/pkg/api"
	runtimeApi "github.com/openshift/kubernetes/pkg/kubelet/api/v1alpha1/runtime"
	kubecontainer "github.com/openshift/kubernetes/pkg/kubelet/container"
	containertest "github.com/openshift/kubernetes/pkg/kubelet/container/testing"
)

func TestSandboxGC(t *testing.T) {
	fakeRuntime, _, m, err := createTestRuntimeManager()
	assert.NoError(t, err)

	pods := []*api.Pod{
		makeTestPod("foo1", "new", "1234", []api.Container{
			makeTestContainer("bar1", "busybox"),
			makeTestContainer("bar2", "busybox"),
		}),
		makeTestPod("foo2", "new", "5678", []api.Container{
			makeTestContainer("bar3", "busybox"),
		}),
	}

	for c, test := range []struct {
		description string              // description of the test case
		sandboxes   []sandboxTemplate   // templates of sandboxes
		containers  []containerTemplate // templates of containers
		minAge      time.Duration       // sandboxMinGCAge
		remain      []int               // template indexes of remaining sandboxes
	}{
		{
			description: "sandbox with no containers should be garbage collected.",
			sandboxes: []sandboxTemplate{
				{pod: pods[0], state: runtimeApi.PodSandboxState_SANDBOX_NOTREADY},
			},
			containers: []containerTemplate{},
			remain:     []int{},
		},
		{
			description: "running sandbox should not be garbage collected.",
			sandboxes: []sandboxTemplate{
				{pod: pods[0], state: runtimeApi.PodSandboxState_SANDBOX_READY},
			},
			containers: []containerTemplate{},
			remain:     []int{0},
		},
		{
			description: "sandbox with containers should not be garbage collected.",
			sandboxes: []sandboxTemplate{
				{pod: pods[0], state: runtimeApi.PodSandboxState_SANDBOX_NOTREADY},
			},
			containers: []containerTemplate{
				{pod: pods[0], container: &pods[0].Spec.Containers[0], state: runtimeApi.ContainerState_CONTAINER_EXITED},
			},
			remain: []int{0},
		},
		{
			description: "sandbox within min age should not be garbage collected.",
			sandboxes: []sandboxTemplate{
				{pod: pods[0], createdAt: time.Now().UnixNano(), state: runtimeApi.PodSandboxState_SANDBOX_NOTREADY},
				{pod: pods[1], createdAt: time.Now().Add(-2 * time.Hour).UnixNano(), state: runtimeApi.PodSandboxState_SANDBOX_NOTREADY},
			},
			containers: []containerTemplate{},
			minAge:     time.Hour, // assume the test won't take an hour
			remain:     []int{0},
		},
		{
			description: "multiple sandboxes should be handled properly.",
			sandboxes: []sandboxTemplate{
				// running sandbox.
				{pod: pods[0], attempt: 1, state: runtimeApi.PodSandboxState_SANDBOX_READY},
				// exited sandbox with containers.
				{pod: pods[1], attempt: 1, state: runtimeApi.PodSandboxState_SANDBOX_NOTREADY},
				// exited sandbox without containers.
				{pod: pods[1], attempt: 0, state: runtimeApi.PodSandboxState_SANDBOX_NOTREADY},
			},
			containers: []containerTemplate{
				{pod: pods[1], container: &pods[1].Spec.Containers[0], sandboxAttempt: 1, state: runtimeApi.ContainerState_CONTAINER_EXITED},
			},
			remain: []int{0, 1},
		},
	} {
		t.Logf("TestCase #%d: %+v", c, test)
		fakeSandboxes := makeFakePodSandboxes(t, m, test.sandboxes)
		fakeContainers := makeFakeContainers(t, m, test.containers)
		fakeRuntime.SetFakeSandboxes(fakeSandboxes)
		fakeRuntime.SetFakeContainers(fakeContainers)

		err := m.containerGC.evictSandboxes(test.minAge)
		assert.NoError(t, err)
		realRemain, err := fakeRuntime.ListPodSandbox(nil)
		assert.NoError(t, err)
		assert.Len(t, realRemain, len(test.remain))
		for _, remain := range test.remain {
			status, err := fakeRuntime.PodSandboxStatus(fakeSandboxes[remain].GetId())
			assert.NoError(t, err)
			assert.Equal(t, &fakeSandboxes[remain].PodSandboxStatus, status)
		}
	}
}

func TestContainerGC(t *testing.T) {
	fakeRuntime, _, m, err := createTestRuntimeManager()
	assert.NoError(t, err)

	fakePodGetter := m.containerGC.podGetter.(*fakePodGetter)
	makeGCContainer := func(podName, containerName string, attempt int, createdAt int64, state runtimeApi.ContainerState) containerTemplate {
		container := makeTestContainer(containerName, "test-image")
		pod := makeTestPod(podName, "test-ns", podName, []api.Container{container})
		if podName != "deleted" {
			// initialize the pod getter, explicitly exclude deleted pod
			fakePodGetter.pods[pod.UID] = pod
		}
		return containerTemplate{
			pod:       pod,
			container: &container,
			attempt:   attempt,
			createdAt: createdAt,
			state:     state,
		}
	}
	defaultGCPolicy := kubecontainer.ContainerGCPolicy{MinAge: time.Hour, MaxPerPodContainer: 2, MaxContainers: 6}

	for c, test := range []struct {
		description string                           // description of the test case
		containers  []containerTemplate              // templates of containers
		policy      *kubecontainer.ContainerGCPolicy // container gc policy
		remain      []int                            // template indexes of remaining containers
	}{
		{
			description: "all containers should be removed when max container limit is 0",
			containers: []containerTemplate{
				makeGCContainer("foo", "bar", 0, 0, runtimeApi.ContainerState_CONTAINER_EXITED),
			},
			policy: &kubecontainer.ContainerGCPolicy{MinAge: time.Minute, MaxPerPodContainer: 1, MaxContainers: 0},
			remain: []int{},
		},
		{
			description: "max containers should be complied when no max per pod container limit is set",
			containers: []containerTemplate{
				makeGCContainer("foo", "bar", 4, 4, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo", "bar", 3, 3, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo", "bar", 2, 2, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo", "bar", 1, 1, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo", "bar", 0, 0, runtimeApi.ContainerState_CONTAINER_EXITED),
			},
			policy: &kubecontainer.ContainerGCPolicy{MinAge: time.Minute, MaxPerPodContainer: -1, MaxContainers: 4},
			remain: []int{0, 1, 2, 3},
		},
		{
			description: "no containers should be removed if both max container and per pod container limits are not set",
			containers: []containerTemplate{
				makeGCContainer("foo", "bar", 2, 2, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo", "bar", 1, 1, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo", "bar", 0, 0, runtimeApi.ContainerState_CONTAINER_EXITED),
			},
			policy: &kubecontainer.ContainerGCPolicy{MinAge: time.Minute, MaxPerPodContainer: -1, MaxContainers: -1},
			remain: []int{0, 1, 2},
		},
		{
			description: "recently started containers should not be removed",
			containers: []containerTemplate{
				makeGCContainer("foo", "bar", 2, time.Now().UnixNano(), runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo", "bar", 1, time.Now().UnixNano(), runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo", "bar", 0, time.Now().UnixNano(), runtimeApi.ContainerState_CONTAINER_EXITED),
			},
			remain: []int{0, 1, 2},
		},
		{
			description: "oldest containers should be removed when per pod container limit exceeded",
			containers: []containerTemplate{
				makeGCContainer("foo", "bar", 2, 2, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo", "bar", 1, 1, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo", "bar", 0, 0, runtimeApi.ContainerState_CONTAINER_EXITED),
			},
			remain: []int{0, 1},
		},
		{
			description: "running containers should not be removed",
			containers: []containerTemplate{
				makeGCContainer("foo", "bar", 2, 2, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo", "bar", 1, 1, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo", "bar", 0, 0, runtimeApi.ContainerState_CONTAINER_RUNNING),
			},
			remain: []int{0, 1, 2},
		},
		{
			description: "no containers should be removed when limits are not exceeded",
			containers: []containerTemplate{
				makeGCContainer("foo", "bar", 1, 1, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo", "bar", 0, 0, runtimeApi.ContainerState_CONTAINER_EXITED),
			},
			remain: []int{0, 1},
		},
		{
			description: "max container count should apply per (UID, container) pair",
			containers: []containerTemplate{
				makeGCContainer("foo", "bar", 2, 2, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo", "bar", 1, 1, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo", "bar", 0, 0, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo1", "baz", 2, 2, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo1", "baz", 1, 1, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo1", "baz", 0, 0, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo2", "bar", 2, 2, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo2", "bar", 1, 1, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo2", "bar", 0, 0, runtimeApi.ContainerState_CONTAINER_EXITED),
			},
			remain: []int{0, 1, 3, 4, 6, 7},
		},
		{
			description: "max limit should apply and try to keep from every pod",
			containers: []containerTemplate{
				makeGCContainer("foo", "bar", 1, 1, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo", "bar", 0, 0, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo1", "bar1", 1, 1, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo1", "bar1", 0, 0, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo2", "bar2", 1, 1, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo2", "bar2", 0, 0, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo3", "bar3", 1, 1, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo3", "bar3", 0, 0, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo4", "bar4", 1, 1, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo4", "bar4", 0, 0, runtimeApi.ContainerState_CONTAINER_EXITED),
			},
			remain: []int{0, 2, 4, 6, 8},
		},
		{
			description: "oldest pods should be removed if limit exceeded",
			containers: []containerTemplate{
				makeGCContainer("foo", "bar", 2, 2, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo", "bar", 1, 1, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo1", "bar1", 2, 2, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo1", "bar1", 1, 1, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo2", "bar2", 1, 1, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo3", "bar3", 0, 0, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo4", "bar4", 1, 1, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo5", "bar5", 0, 0, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo6", "bar6", 2, 2, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo7", "bar7", 1, 1, runtimeApi.ContainerState_CONTAINER_EXITED),
			},
			remain: []int{0, 2, 4, 6, 8, 9},
		},
		{
			description: "containers for deleted pods should be removed",
			containers: []containerTemplate{
				makeGCContainer("foo", "bar", 1, 1, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("foo", "bar", 0, 0, runtimeApi.ContainerState_CONTAINER_EXITED),
				// deleted pods still respect MinAge.
				makeGCContainer("deleted", "bar1", 2, time.Now().UnixNano(), runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("deleted", "bar1", 1, 1, runtimeApi.ContainerState_CONTAINER_EXITED),
				makeGCContainer("deleted", "bar1", 0, 0, runtimeApi.ContainerState_CONTAINER_EXITED),
			},
			remain: []int{0, 1, 2},
		},
	} {
		t.Logf("TestCase #%d: %+v", c, test)
		fakeContainers := makeFakeContainers(t, m, test.containers)
		fakeRuntime.SetFakeContainers(fakeContainers)

		if test.policy == nil {
			test.policy = &defaultGCPolicy
		}
		err := m.containerGC.evictContainers(*test.policy, true)
		assert.NoError(t, err)
		realRemain, err := fakeRuntime.ListContainers(nil)
		assert.NoError(t, err)
		assert.Len(t, realRemain, len(test.remain))
		for _, remain := range test.remain {
			status, err := fakeRuntime.ContainerStatus(fakeContainers[remain].GetId())
			assert.NoError(t, err)
			assert.Equal(t, &fakeContainers[remain].ContainerStatus, status)
		}
	}
}

// Notice that legacy container symlink is not tested since it may be deprecated soon.
func TestPodLogDirectoryGC(t *testing.T) {
	_, _, m, err := createTestRuntimeManager()
	assert.NoError(t, err)
	fakeOS := m.osInterface.(*containertest.FakeOS)
	fakePodGetter := m.containerGC.podGetter.(*fakePodGetter)

	// pod log directories without corresponding pods should be removed.
	fakePodGetter.pods["123"] = makeTestPod("foo1", "new", "123", nil)
	fakePodGetter.pods["456"] = makeTestPod("foo2", "new", "456", nil)
	files := []string{"123", "456", "789", "012"}
	removed := []string{filepath.Join(podLogsRootDirectory, "789"), filepath.Join(podLogsRootDirectory, "012")}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	fakeOS.ReadDirFn = func(string) ([]os.FileInfo, error) {
		var fileInfos []os.FileInfo
		for _, file := range files {
			mockFI := containertest.NewMockFileInfo(ctrl)
			mockFI.EXPECT().Name().Return(file)
			fileInfos = append(fileInfos, mockFI)
		}
		return fileInfos, nil
	}

	// allSourcesReady == true, pod log directories without corresponding pod should be removed.
	err = m.containerGC.evictPodLogsDirectories(true)
	assert.NoError(t, err)
	assert.Equal(t, removed, fakeOS.Removes)

	// allSourcesReady == false, pod log directories should not be removed.
	fakeOS.Removes = []string{}
	err = m.containerGC.evictPodLogsDirectories(false)
	assert.NoError(t, err)
	assert.Empty(t, fakeOS.Removes)
}
