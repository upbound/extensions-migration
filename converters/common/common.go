// Copyright 2022 Upbound Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	v1 "github.com/crossplane/crossplane/apis/apiextensions/v1"
	"github.com/pkg/errors"
	"github.com/upbound/upjet/pkg/migration"
	"k8s.io/apimachinery/pkg/runtime"
)

func AddToScheme(r *migration.Registry, sourceF, targetF func(s *runtime.Scheme) error) error {
	if err := r.AddToScheme(sourceF); err != nil {
		return err
	}
	if err := r.AddToScheme(targetF); err != nil {
		return err
	}
	return nil
}

func PtrFromString(s string) *string {
	return &s
}

func PtrFloat64(i *int32) *float64 {
	a := float64(*i)
	return &a
}

func SplittedResourcePatches(convertedTemplates []*v1.ComposedTemplate, resourceName string, patchesToAdd []v1.Patch) error {
	for i, cb := range convertedTemplates {
		if cb.Base.Raw != nil {
			u, err := migration.FromRawExtension(cb.Base)
			if err != nil {
				return errors.Wrap(err, "failed to convert ComposedTemplate base")
			}
			if u.GetKind() == resourceName {
				convertedTemplates[i].Patches = append(convertedTemplates[i].Patches, patchesToAdd...)
			}
		}
	}
	return nil
}
