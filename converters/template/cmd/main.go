package main

import (
	sourceapis "github.com/crossplane-contrib/provider-aws/apis"
	"github.com/upbound/extensions-migration/converters/common"
	"github.com/upbound/extensions-migration/converters/provider-aws"
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

	provider_aws.RegisterAllKnownConverters(registry)
}
