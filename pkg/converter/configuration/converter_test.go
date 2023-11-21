// Copyright 2023 Upbound Inc.
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

package configuration

import (
	"sort"
	"strings"
	"testing"

	xpmetav1 "github.com/crossplane/crossplane/apis/pkg/meta/v1"
	xpmetav1alpha1 "github.com/crossplane/crossplane/apis/pkg/meta/v1alpha1"
	xppkgv1 "github.com/crossplane/crossplane/apis/pkg/v1"
	xppkgv1beta1 "github.com/crossplane/crossplane/apis/pkg/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/upjet/pkg/migration"
	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	awsPackage = "xpkg.upbound.io/upbound/provider-aws"
)

var (
	unstructuredAwsVpc = map[string]interface{}{
		"apiVersion": "ec2.aws.upbound.io/v1beta1",
		"kind":       "VPC",
		"metadata": map[string]interface{}{
			"name": "sample-vpc",
		},
	}
	unstructuredAwsProviderConfig = map[string]interface{}{
		"apiVersion": "aws.upbound.io/v1beta1",
		"kind":       "ProviderConfig",
		"metadata": map[string]interface{}{
			"name": "sample-pc",
		},
	}
	unstructuredAzureZone = map[string]interface{}{
		"apiVersion": "network.azure.upbound.io/v1beta1",
		"kind":       "Zone",
		"metadata": map[string]interface{}{
			"name": "sample-zone",
		},
	}
	unstructuredAzureResourceGroup = map[string]interface{}{
		"apiVersion": "azure.upbound.io/v1beta1",
		"kind":       "ResourceGroup",
		"metadata": map[string]interface{}{
			"name": "example-resources",
		},
	}
	unstructuredGcpZone = map[string]interface{}{
		"apiVersion": "network.gcp.upbound.io/v1beta1",
		"kind":       "Zone",
		"metadata": map[string]interface{}{
			"name": "sample-zone",
		},
	}
	unstructuredGcpProviderConfig = map[string]interface{}{
		"apiVersion": "gcp.upbound.io/v1beta1",
		"kind":       "ProviderConfig",
		"metadata": map[string]interface{}{
			"name": "sample-pc",
		},
	}

	ap = xppkgv1.ManualActivation
)

func TestGetSSOPNameFromManagedResource(t *testing.T) {
	type args struct {
		u migration.UnstructuredWithMetadata
	}
	type want struct {
		providerNames map[string]struct{}
		err           error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Aws": {
			args: args{
				u: migration.UnstructuredWithMetadata{
					Object: unstructured.Unstructured{
						Object: unstructuredAwsVpc,
					},
				},
			},
			want: want{
				providerNames: map[string]struct{}{
					"provider-aws-ec2":    {},
					"provider-family-aws": {},
				},
				err: nil,
			},
		},
		"Family-Aws": {
			args: args{
				u: migration.UnstructuredWithMetadata{
					Object: unstructured.Unstructured{
						Object: unstructuredAwsProviderConfig,
					},
				},
			},
			want: want{
				providerNames: map[string]struct{}{
					"provider-family-aws": {},
				},
				err: nil,
			},
		},
		"Azure": {
			args: args{
				u: migration.UnstructuredWithMetadata{
					Object: unstructured.Unstructured{
						Object: unstructuredAzureZone,
					},
				},
			},
			want: want{
				providerNames: map[string]struct{}{
					"provider-azure-network": {},
					"provider-family-azure":  {},
				},
				err: nil,
			},
		},
		"Family-Azure": {
			args: args{
				u: migration.UnstructuredWithMetadata{
					Object: unstructured.Unstructured{
						Object: unstructuredAzureResourceGroup,
					},
				},
			},
			want: want{
				providerNames: map[string]struct{}{
					"provider-family-azure": {},
				},
				err: nil,
			},
		},
		"Gcp": {
			args: args{
				u: migration.UnstructuredWithMetadata{
					Object: unstructured.Unstructured{
						Object: unstructuredGcpZone,
					},
				},
			},
			want: want{
				providerNames: map[string]struct{}{
					"provider-family-gcp":  {},
					"provider-gcp-network": {},
				},
				err: nil,
			},
		},
		"Family-Gcp": {
			args: args{
				u: migration.UnstructuredWithMetadata{
					Object: unstructured.Unstructured{
						Object: unstructuredGcpProviderConfig,
					},
				},
			},
			want: want{
				providerNames: map[string]struct{}{
					"provider-family-gcp": {},
				},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mp := NewMRPreProcessor()
			err := mp.GetSSOPNameFromManagedResource(tc.u)
			if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("\nNext(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.providerNames, mp.ProviderNames); diff != "" {
				t.Errorf("\nNext(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestConfigurationMetadataV1(t *testing.T) {
	type args struct {
		c *xpmetav1.Configuration
	}
	type want struct {
		c   *xpmetav1.Configuration
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"WithAnotherProvider": {
			args: args{
				c: &xpmetav1.Configuration{
					Spec: xpmetav1.ConfigurationSpec{
						MetaSpec: xpmetav1.MetaSpec{
							DependsOn: []xpmetav1.Dependency{
								{
									Provider: ptrFromString("xpkg.upbound.io/upbound/provider-aws"),
									Version:  ">=v0.32.0",
								},
								{
									Provider: ptrFromString("xpkg.upbound.io/upbound/provider-helm"),
									Version:  ">=v0.15.0",
								},
							},
						},
					},
				},
			},
			want: want{
				c: &xpmetav1.Configuration{
					Spec: xpmetav1.ConfigurationSpec{
						MetaSpec: xpmetav1.MetaSpec{
							DependsOn: []xpmetav1.Dependency{
								{
									Provider: ptrFromString("xpkg.upbound.io/upbound/provider-aws-ec2"),
									Version:  ">=v0.33.0",
								},
								{
									Provider: ptrFromString("xpkg.upbound.io/upbound/provider-helm"),
									Version:  ">=v0.15.0",
								},
							},
						},
					},
				},
			},
		},
		"WithoutAnotherProvider": {
			args: args{
				c: &xpmetav1.Configuration{
					Spec: xpmetav1.ConfigurationSpec{
						MetaSpec: xpmetav1.MetaSpec{
							DependsOn: []xpmetav1.Dependency{
								{
									Provider: ptrFromString("xpkg.upbound.io/upbound/provider-aws"),
									Version:  ">=v0.32.0",
								},
							},
						},
					},
				},
			},
			want: want{
				c: &xpmetav1.Configuration{
					Spec: xpmetav1.ConfigurationSpec{
						MetaSpec: xpmetav1.MetaSpec{
							DependsOn: []xpmetav1.Dependency{
								{
									Provider: ptrFromString("xpkg.upbound.io/upbound/provider-aws-ec2"),
									Version:  ">=v0.33.0",
								},
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			cp := NewCompositionPreProcessor()
			cp.ProviderNames = map[string]struct{}{
				"provider-family-aws": {},
				"provider-aws-ec2":    {},
			}
			cm := ConfigMetaParameters{
				Monolith:             "provider-aws",
				FamilyVersion:        "v0.33.0",
				CompositionProcessor: cp,
			}
			err := cm.ConfigurationMetadataV1(tc.args.c)
			if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("\nNext(...): -want, +got:\n%s", diff)
			}
			sort.Slice(tc.args.c.Spec.DependsOn, func(i, j int) bool {
				return strings.Compare(*tc.args.c.Spec.DependsOn[i].Provider, *tc.args.c.Spec.DependsOn[j].Provider) == -1
			})
			if diff := cmp.Diff(tc.want.c, tc.args.c); diff != "" {
				t.Errorf("\nNext(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestConfigurationMetadataV1Alpha(t *testing.T) {
	type args struct {
		c *xpmetav1alpha1.Configuration
	}
	type want struct {
		c   *xpmetav1alpha1.Configuration
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"WithAnotherProvider": {
			args: args{
				c: &xpmetav1alpha1.Configuration{
					Spec: xpmetav1alpha1.ConfigurationSpec{
						MetaSpec: xpmetav1alpha1.MetaSpec{
							DependsOn: []xpmetav1alpha1.Dependency{
								{
									Provider: ptrFromString("xpkg.upbound.io/upbound/provider-aws"),
									Version:  ">=v0.32.0",
								},
								{
									Provider: ptrFromString("xpkg.upbound.io/upbound/provider-helm"),
									Version:  ">=v0.15.0",
								},
							},
						},
					},
				},
			},
			want: want{
				c: &xpmetav1alpha1.Configuration{
					Spec: xpmetav1alpha1.ConfigurationSpec{
						MetaSpec: xpmetav1alpha1.MetaSpec{
							DependsOn: []xpmetav1alpha1.Dependency{
								{
									Provider: ptrFromString("xpkg.upbound.io/upbound/provider-aws-ec2"),
									Version:  ">=v0.33.0",
								},
								{
									Provider: ptrFromString("xpkg.upbound.io/upbound/provider-helm"),
									Version:  ">=v0.15.0",
								},
							},
						},
					},
				},
			},
		},
		"WithoutAnotherProvider": {
			args: args{
				c: &xpmetav1alpha1.Configuration{
					Spec: xpmetav1alpha1.ConfigurationSpec{
						MetaSpec: xpmetav1alpha1.MetaSpec{
							DependsOn: []xpmetav1alpha1.Dependency{
								{
									Provider: ptrFromString("xpkg.upbound.io/upbound/provider-aws"),
									Version:  ">=v0.32.0",
								},
							},
						},
					},
				},
			},
			want: want{
				c: &xpmetav1alpha1.Configuration{
					Spec: xpmetav1alpha1.ConfigurationSpec{
						MetaSpec: xpmetav1alpha1.MetaSpec{
							DependsOn: []xpmetav1alpha1.Dependency{
								{
									Provider: ptrFromString("xpkg.upbound.io/upbound/provider-aws-ec2"),
									Version:  ">=v0.33.0",
								},
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			cp := NewCompositionPreProcessor()
			cp.ProviderNames = map[string]struct{}{
				"provider-family-aws": {},
				"provider-aws-ec2":    {},
			}
			cm := ConfigMetaParameters{
				Monolith:             "provider-aws",
				FamilyVersion:        "v0.33.0",
				CompositionProcessor: cp,
			}
			err := cm.ConfigurationMetadataV1Alpha1(tc.args.c)
			if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("\nNext(...): -want, +got:\n%s", diff)
			}
			sort.Slice(tc.args.c.Spec.DependsOn, func(i, j int) bool {
				return strings.Compare(*tc.args.c.Spec.DependsOn[i].Provider, *tc.args.c.Spec.DependsOn[j].Provider) == -1
			})
			if diff := cmp.Diff(tc.want.c, tc.args.c); diff != "" {
				t.Errorf("\nNext(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestConfigurationPackageV1(t *testing.T) {
	type args struct {
		pkg *xppkgv1.Configuration
	}
	type want struct {
		pkg *xppkgv1.Configuration
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"ConfigurationPkg": {
			args: args{
				pkg: &xppkgv1.Configuration{
					Spec: xppkgv1.ConfigurationSpec{
						PackageSpec: xppkgv1.PackageSpec{
							Package: "xpkg.upbound.io/upbound/provider-ref-aws:v0.1.0",
						},
					},
				},
			},
			want: want{
				pkg: &xppkgv1.Configuration{
					Spec: xppkgv1.ConfigurationSpec{
						PackageSpec: xppkgv1.PackageSpec{
							Package: "xpkg.upbound.io/upbound/provider-ref-aws:v0.1.0-ssop",
						},
					},
				},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			cp := ConfigPkgParameters{
				PackageURL: "xpkg.upbound.io/upbound/provider-ref-aws:v0.1.0-ssop",
			}
			err := cp.ConfigurationPackageV1(tc.args.pkg)
			if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("\nNext(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.pkg, tc.args.pkg); diff != "" {
				t.Errorf("\nNext(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestPackageLockV1Beta1(t *testing.T) {
	type args struct {
		lock *xppkgv1beta1.Lock
	}
	type want struct {
		lock *xppkgv1beta1.Lock
		err  error
	}

	cases := map[string]struct {
		args
		want
	}{
		"NeedRemoval": {
			args: args{
				lock: &xppkgv1beta1.Lock{
					Packages: []xppkgv1beta1.LockPackage{
						{
							Source: awsPackage,
						},
					},
				},
			},
			want: want{
				lock: &xppkgv1beta1.Lock{
					Packages: []xppkgv1beta1.LockPackage{
						{
							Source: "xpkg.upbound.io/upbound/provider-aws",
						},
					},
				},
				err: nil,
			},
		},
		"NoNeedRemoval": {
			args: args{
				lock: &xppkgv1beta1.Lock{
					Packages: []xppkgv1beta1.LockPackage{
						{
							Source: "xpkg.upbound.io/upbound/provider-helm",
						},
					},
				},
			},
			want: want{
				lock: &xppkgv1beta1.Lock{
					Packages: []xppkgv1beta1.LockPackage{
						{
							Source: "xpkg.upbound.io/upbound/provider-helm",
						},
					},
				},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			l := LockParameters{}
			err := l.PackageLockV1Beta1(tc.args.lock)
			if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("\nNext(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.lock, tc.args.lock); diff != "" {
				t.Errorf("\nNext(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestPackagePkgFamilyConfigParameters_ProviderPackageV1(t *testing.T) {
	type args struct {
		p xppkgv1.Provider
	}
	type want struct {
		providers []xppkgv1.Provider
		err       error
	}

	cases := map[string]struct {
		args
		want
	}{
		"AWSConf": {
			args: args{
				p: xppkgv1.Provider{
					ObjectMeta: metav1.ObjectMeta{
						Name: "provider-aws",
					},
					Spec: xppkgv1.ProviderSpec{
						PackageSpec: xppkgv1.PackageSpec{
							Package:                  "xpkg.upbound.io/upbound/provider-aws:v0.33.0",
							RevisionActivationPolicy: &ap,
						},
					},
				},
			},
			want: want{
				providers: []xppkgv1.Provider{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "upbound-provider-family-aws",
						},
						Spec: xppkgv1.ProviderSpec{
							PackageSpec: xppkgv1.PackageSpec{
								Package:                  "xpkg.upbound.io/upbound/provider-family-aws:v0.37.0",
								RevisionActivationPolicy: &ap,
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			pc := ProviderPkgFamilyConfigParameters{
				FamilyVersion: "v0.37.0",
			}
			providers, err := pc.ProviderPackageV1(tc.args.p)
			if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("\nNext(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.providers, providers); diff != "" {
				t.Errorf("\nNext(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestPackagePkgFamilyParameters_ProviderPackageV1(t *testing.T) {
	type args struct {
		p xppkgv1.Provider
	}
	type want struct {
		providers []xppkgv1.Provider
		err       error
	}

	cases := map[string]struct {
		args
		want
	}{
		"AWSFamily": {
			args: args{
				p: xppkgv1.Provider{
					ObjectMeta: metav1.ObjectMeta{
						Name: "provider-aws",
					},
					Spec: xppkgv1.ProviderSpec{
						PackageSpec: xppkgv1.PackageSpec{
							Package:                  "xpkg.upbound.io/upbound/provider-aws:v0.33.0",
							RevisionActivationPolicy: &ap,
						},
					},
				},
			},
			want: want{
				providers: []xppkgv1.Provider{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "upbound-provider-aws-ec2",
						},
						Spec: xppkgv1.ProviderSpec{
							PackageSpec: xppkgv1.PackageSpec{
								Package:                  "xpkg.upbound.io/upbound/provider-aws-ec2:v0.37.0",
								RevisionActivationPolicy: &ap,
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "upbound-provider-aws-eks",
						},
						Spec: xppkgv1.ProviderSpec{
							PackageSpec: xppkgv1.PackageSpec{
								Package:                  "xpkg.upbound.io/upbound/provider-aws-eks:v0.37.0",
								RevisionActivationPolicy: &ap,
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			cp := NewCompositionPreProcessor()
			cp.ProviderNames = map[string]struct{}{
				"provider-family-aws": {},
				"provider-aws-ec2":    {},
				"provider-aws-eks":    {},
			}
			pc := ProviderPkgFamilyParameters{
				FamilyVersion:        "v0.37.0",
				Monolith:             "provider-aws",
				CompositionProcessor: cp,
			}
			providers, err := pc.ProviderPackageV1(tc.args.p)
			if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("\nNext(...): -want, +got:\n%s", diff)
			}
			sort.Slice(providers, func(i, j int) bool {
				return strings.Compare(providers[i].Spec.PackageSpec.Package, providers[j].Spec.PackageSpec.Package) == -1
			})
			if diff := cmp.Diff(tc.want.providers, providers); diff != "" {
				t.Errorf("\nNext(...): -want, +got:\n%s", diff)
			}
		})
	}
}
