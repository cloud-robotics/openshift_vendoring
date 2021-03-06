/*
Copyright 2015 The Kubernetes Authors.

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

package pod

import (
	"testing"

	"github.com/openshift/kubernetes/pkg/api"
	"github.com/openshift/kubernetes/pkg/util/intstr"
)

func TestFindPort(t *testing.T) {
	testCases := []struct {
		name       string
		containers []api.Container
		port       intstr.IntOrString
		expected   int
		pass       bool
	}{{
		name:       "valid int, no ports",
		containers: []api.Container{{}},
		port:       intstr.FromInt(93),
		expected:   93,
		pass:       true,
	}, {
		name: "valid int, with ports",
		containers: []api.Container{{Ports: []api.ContainerPort{{
			Name:          "",
			ContainerPort: 11,
			Protocol:      "TCP",
		}, {
			Name:          "p",
			ContainerPort: 22,
			Protocol:      "TCP",
		}}}},
		port:     intstr.FromInt(93),
		expected: 93,
		pass:     true,
	}, {
		name:       "valid str, no ports",
		containers: []api.Container{{}},
		port:       intstr.FromString("p"),
		expected:   0,
		pass:       false,
	}, {
		name: "valid str, one ctr with ports",
		containers: []api.Container{{Ports: []api.ContainerPort{{
			Name:          "",
			ContainerPort: 11,
			Protocol:      "UDP",
		}, {
			Name:          "p",
			ContainerPort: 22,
			Protocol:      "TCP",
		}, {
			Name:          "q",
			ContainerPort: 33,
			Protocol:      "TCP",
		}}}},
		port:     intstr.FromString("q"),
		expected: 33,
		pass:     true,
	}, {
		name: "valid str, two ctr with ports",
		containers: []api.Container{{}, {Ports: []api.ContainerPort{{
			Name:          "",
			ContainerPort: 11,
			Protocol:      "UDP",
		}, {
			Name:          "p",
			ContainerPort: 22,
			Protocol:      "TCP",
		}, {
			Name:          "q",
			ContainerPort: 33,
			Protocol:      "TCP",
		}}}},
		port:     intstr.FromString("q"),
		expected: 33,
		pass:     true,
	}, {
		name: "valid str, two ctr with same port",
		containers: []api.Container{{}, {Ports: []api.ContainerPort{{
			Name:          "",
			ContainerPort: 11,
			Protocol:      "UDP",
		}, {
			Name:          "p",
			ContainerPort: 22,
			Protocol:      "TCP",
		}, {
			Name:          "q",
			ContainerPort: 22,
			Protocol:      "TCP",
		}}}},
		port:     intstr.FromString("q"),
		expected: 22,
		pass:     true,
	}, {
		name: "valid str, invalid protocol",
		containers: []api.Container{{}, {Ports: []api.ContainerPort{{
			Name:          "a",
			ContainerPort: 11,
			Protocol:      "snmp",
		},
		}}},
		port:     intstr.FromString("a"),
		expected: 0,
		pass:     false,
	}, {
		name: "valid hostPort",
		containers: []api.Container{{}, {Ports: []api.ContainerPort{{
			Name:          "a",
			ContainerPort: 11,
			HostPort:      81,
			Protocol:      "TCP",
		},
		}}},
		port:     intstr.FromString("a"),
		expected: 11,
		pass:     true,
	},
		{
			name: "invalid hostPort",
			containers: []api.Container{{}, {Ports: []api.ContainerPort{{
				Name:          "a",
				ContainerPort: 11,
				HostPort:      -1,
				Protocol:      "TCP",
			},
			}}},
			port:     intstr.FromString("a"),
			expected: 11,
			pass:     true,
			//this should fail but passes.
		},
		{
			name: "invalid ContainerPort",
			containers: []api.Container{{}, {Ports: []api.ContainerPort{{
				Name:          "a",
				ContainerPort: -1,
				Protocol:      "TCP",
			},
			}}},
			port:     intstr.FromString("a"),
			expected: -1,
			pass:     true,
			//this should fail but passes
		},
		{
			name: "HostIP Address",
			containers: []api.Container{{}, {Ports: []api.ContainerPort{{
				Name:          "a",
				ContainerPort: 11,
				HostIP:        "192.168.1.1",
				Protocol:      "TCP",
			},
			}}},
			port:     intstr.FromString("a"),
			expected: 11,
			pass:     true,
		},
	}

	for _, tc := range testCases {
		port, err := FindPort(&api.Pod{Spec: api.PodSpec{Containers: tc.containers}},
			&api.ServicePort{Protocol: "TCP", TargetPort: tc.port})
		if err != nil && tc.pass {
			t.Errorf("unexpected error for %s: %v", tc.name, err)
		}
		if err == nil && !tc.pass {
			t.Errorf("unexpected non-error for %s: %d", tc.name, port)
		}
		if port != tc.expected {
			t.Errorf("wrong result for %s: expected %d, got %d", tc.name, tc.expected, port)
		}
	}
}
