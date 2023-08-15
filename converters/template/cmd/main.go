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

// Package main contains a template for the migrators.
// This template is a place for migrator developers to use as a starting point.
// Related developer can accelerate migrator development by using this template.
// Initializing the Registry, adding schemas, adding converters to the Registry,
// and many other high-level definitions are done here.
package main

import (
	sourceapis "github.com/crossplane-contrib/provider-aws/apis"
	"github.com/upbound/extensions-migration/converters/common"
	provideraws "github.com/upbound/extensions-migration/converters/provider-aws"
	targetapis "github.com/upbound/provider-aws/apis"
	"github.com/upbound/upjet/pkg/migration"
	"k8s.io/apimachinery/pkg/runtime"
)

func main() {
	registry := migration.NewRegistry(runtime.NewScheme())

	// Register source and target API Groups to schema
	// Example for AWS
	sourceF := sourceapis.AddToScheme
	targetF := targetapis.AddToScheme
	if err := common.AddToScheme(registry, sourceF, targetF); err != nil {
		panic(err)
	}

	// Register Composition, Claim, and Composite Types to the Register
	//
	if err := registry.AddCompositionTypes(); err != nil {
		panic(err)
	}
	// Registry.AddClaimType(...)
	// Registry.AddCompositeType(...)

	// Register all known API converters for the community AWS provider
	provideraws.RegisterAllKnownConverters(registry)
}
