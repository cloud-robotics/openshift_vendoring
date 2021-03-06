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

package metaonly

import (
	"github.com/openshift/kubernetes/pkg/api/unversioned"
	"github.com/openshift/kubernetes/pkg/api/v1"
)

// MetadataOnlyObject allows decoding only the apiVersion, kind, and metadata fields of
// JSON data.
// TODO: enable meta-only decoding for protobuf.
type MetadataOnlyObject struct {
	unversioned.TypeMeta `json:",inline"`
	// +optional
	v1.ObjectMeta `json:"metadata,omitempty"`
}

// MetadataOnlyObjectList allows decoding from JSON data only the typemeta and metadata of
// a list, and those of the enclosing objects.
// TODO: enable meta-only decoding for protobuf.
type MetadataOnlyObjectList struct {
	unversioned.TypeMeta `json:",inline"`
	// +optional
	unversioned.ListMeta `json:"metadata,omitempty"`

	Items []MetadataOnlyObject `json:"items"`
}
