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

package install

import (
	"fmt"

	"github.com/golang/glog"

	core "github.com/openshift/kubernetes/federation/apis/core"
	core_v1 "github.com/openshift/kubernetes/federation/apis/core/v1"
	"github.com/openshift/kubernetes/pkg/api"
	"github.com/openshift/kubernetes/pkg/api/meta"
	"github.com/openshift/kubernetes/pkg/api/unversioned"
	"github.com/openshift/kubernetes/pkg/apimachinery"
	"github.com/openshift/kubernetes/pkg/apimachinery/registered"
	"github.com/openshift/kubernetes/pkg/runtime"
	"github.com/openshift/kubernetes/pkg/util/sets"
)

const importPrefix = "github.com/openshift/kubernetes/pkg/api"

var accessor = meta.NewAccessor()

// availableVersions lists all known external versions for this group from most preferred to least preferred
var availableVersions = []unversioned.GroupVersion{core_v1.SchemeGroupVersion}

func init() {
	registered.RegisterVersions(availableVersions)
	externalVersions := []unversioned.GroupVersion{}
	for _, v := range availableVersions {
		if registered.IsAllowedVersion(v) {
			externalVersions = append(externalVersions, v)
		}
	}
	if len(externalVersions) == 0 {
		glog.V(4).Infof("No version is registered for group %v", core.GroupName)
		return
	}

	if err := registered.EnableVersions(externalVersions...); err != nil {
		glog.V(4).Infof("%v", err)
		return
	}
	if err := enableVersions(externalVersions); err != nil {
		glog.V(4).Infof("%v", err)
		return
	}
}

// TODO: enableVersions should be centralized rather than spread in each API
// group.
// We can combine registered.RegisterVersions, registered.EnableVersions and
// registered.RegisterGroup once we have moved enableVersions there.
func enableVersions(externalVersions []unversioned.GroupVersion) error {
	addVersionsToScheme(externalVersions...)
	preferredExternalVersion := externalVersions[0]

	groupMeta := apimachinery.GroupMeta{
		GroupVersion:  preferredExternalVersion,
		GroupVersions: externalVersions,
		RESTMapper:    newRESTMapper(externalVersions),
		SelfLinker:    runtime.SelfLinker(accessor),
		InterfacesFor: interfacesFor,
	}

	if err := registered.RegisterGroup(groupMeta); err != nil {
		return err
	}
	return nil
}

// userResources is a group of resources mostly used by a kubectl user
var userResources = []string{"svc"}

func newRESTMapper(externalVersions []unversioned.GroupVersion) meta.RESTMapper {
	// the list of kinds that are scoped at the root of the api hierarchy
	// if a kind is not enumerated here, it is assumed to have a namespace scope
	rootScoped := sets.NewString(
		"Namespace",
	)

	// these kinds should be excluded from the list of resources
	ignoredKinds := sets.NewString(
		"ListOptions",
		"DeleteOptions",
		"Status")

	mapper := api.NewDefaultRESTMapperFromScheme(externalVersions, interfacesFor, importPrefix, ignoredKinds, rootScoped, core.Scheme)
	// setup aliases for groups of resources
	mapper.AddResourceAlias("all", userResources...)

	return mapper
}

// InterfacesFor returns the default Codec and ResourceVersioner for a given version
// string, or an error if the version is not known.
func interfacesFor(version unversioned.GroupVersion) (*meta.VersionInterfaces, error) {
	switch version {
	case core_v1.SchemeGroupVersion:
		return &meta.VersionInterfaces{
			ObjectConvertor:  core.Scheme,
			MetadataAccessor: accessor,
		}, nil
	default:
		g, _ := registered.Group(core.GroupName)
		return nil, fmt.Errorf("unsupported storage version: %s (valid: %v)", version, g.GroupVersions)
	}
}

func addVersionsToScheme(externalVersions ...unversioned.GroupVersion) {
	// add the internal version to Scheme
	if err := core.AddToScheme(core.Scheme); err != nil {
		// Programmer error, detect immediately
		panic(err)
	}
	// add the enabled external versions to Scheme
	for _, v := range externalVersions {
		if !registered.IsEnabledVersion(v) {
			glog.Errorf("Version %s is not enabled, so it will not be added to the Scheme.", v)
			continue
		}
		switch v {
		case core_v1.SchemeGroupVersion:
			if err := core_v1.AddToScheme(core.Scheme); err != nil {
				// Programmer error, detect immediately
				panic(err)
			}
		}
	}
}