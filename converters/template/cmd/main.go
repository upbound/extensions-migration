package main

import (
	"github.com/upbound/extensions-migration/converters/common"
	"github.com/upbound/extensions-migration/converters/template/config/null"
	"github.com/upbound/upjet/pkg/migration"
	"k8s.io/apimachinery/pkg/runtime"
)

func main() {
	registry := migration.NewRegistry(runtime.NewScheme())

	// Register source and target API Groups to schema
	//
	// sourceF := sourceapis.AddToScheme
	// targetF := targetapis.AddToScheme
	// common.AddToScheme(registry, sourceF, targetF)
	//
	//
	//
	// Register Composition, Claim, and Composite Types to the Register
	//
	// Registry.AddCompositionTypes()
	// Registry.AddClaimType(...)
	// Registry.AddCompositeType(...)

	rb := common.NewRegistryBuilder(registry)
	for _, c := range []func(builder *common.RegistryBuilder){
		null.ExampleMRKindConfigurator,
	} {
		c(rb)
	}
	rb.Register()
}
