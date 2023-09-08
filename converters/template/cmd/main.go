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
	"fmt"
	"os"
	"path/filepath"

	sourceapis "github.com/crossplane-contrib/provider-aws/apis"
	targetapis "github.com/upbound/provider-aws/apis"
	"github.com/upbound/upjet/pkg/migration"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/upbound/extensions-migration/converters/common"
	provideraws "github.com/upbound/extensions-migration/converters/provider-aws"
)

func main() {
	// Common CLI Flags
	// They can be extended according to requirements
	var (
		app               = kingpin.New(filepath.Base(os.Args[0]), "Upbound migration plan generator for migrating Kubernetes objects from community providers to official providers.").DefaultEnvars()
		planPath          = app.Flag("plan-path", "Path where the generated migration plan will be stored").Short('p').Default("migration_plan.yaml").String()
		sourcePath        = app.Flag("source-path", "Path of the root directory for the filesystem source. If this flag is not specified, Kubernetes source will be used.").Short('s').String()
		kubeconfigPath    = app.Flag("kubeconfig", "Path of the kubernetes config file. Defaults to ~/.kube/config ").String()
		skipGVKsPath      = app.Flag("skip-gvks", "Path of the file containing the GVKs to skip").String()
		setProviderConfig = app.Flag("set-provider-config", "Used to set a ProviderConfig Reference to all Managed Resources. The string specified for this flag is added as a ProviderConfig Reference to all MRs to be converted.").String()
	)
	if len(*kubeconfigPath) == 0 {
		homeDir, err := os.UserHomeDir()
		kingpin.FatalIfError(err, "Failed to get user's home directory")
		*kubeconfigPath = filepath.Join(homeDir, ".kube/config")
	}
	kingpin.MustParse(app.Parse(os.Args[1:]))

	// Registry initialization
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

	// Register ProviderConfigPreProcessor
	if *setProviderConfig != "" {
		pc := common.NewProviderConfigPreProcessor(*setProviderConfig)
		registry.RegisterResourcePreProcessor(migration.ResourcePreProcessor(pc.SetProviderConfig))
	}

	// Initialize Source for reading resources
	var source migration.Source
	var err error
	if len(*sourcePath) > 0 { // FileSystem Source
		fmt.Println("Using filesystem source")
		source, err = migration.NewFileSystemSource(*sourcePath)
		kingpin.FatalIfError(err, "Failed to initialize a Filesystem source")
	} else { // Kubernetes Source
		fmt.Println("Using kubernetes source")
		source, err = migration.NewKubernetesSourceFromKubeConfig(*kubeconfigPath, migration.WithRegistry(registry))
		kingpin.FatalIfError(err, "Failed to initialize a Kubernetes source")
	}

	// Calculate Abs Path for the migration plan and generated manifests
	absPath, err := filepath.Abs(*planPath)
	kingpin.FatalIfError(err, "Failed to get the absolute path for the migration plan output: %s", *planPath)
	planDir := filepath.Dir(absPath)

	// Initialize Target for writing resources
	target := migration.NewFileSystemTarget(migration.WithParentDirectory(planDir))

	var skipGVKs []schema.GroupVersionKind
	// Skipped GVKs
	if *skipGVKsPath != "" {
		skipGVKs, err = readSkipFile(*skipGVKsPath)
		if err != nil {
			kingpin.FatalIfError(err, "Failed to read skip GVK list")
		}
	}

	// Generate Plan
	pg := migration.PlanGenerator{}
	switch source.(type) {
	case *migration.FileSystemSource:
		pg = migration.NewPlanGenerator(registry, source, target, migration.WithEnableOnlyFileSystemAPISteps(), migration.WithSkipGVKs(skipGVKs...))
	case *migration.KubernetesSource:
		pg = migration.NewPlanGenerator(registry, source, target, migration.WithSkipGVKs(skipGVKs...))
	}

	err = pg.GeneratePlan()
	kingpin.FatalIfError(err, "Failed to generate the migration plan")

	// Write plan to the target plan path
	buff, err := yaml.Marshal(pg.Plan)
	kingpin.FatalIfError(err, "Failed to marshal the migration plan into YAML")
	kingpin.FatalIfError(os.WriteFile(*planPath, buff, 0600), "Failed to store the migration plan: %s", planPath)
}

func readSkipFile(path string) ([]schema.GroupVersionKind, error) {
	buff, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var skipGVKs []schema.GroupVersionKind
	var skipGVKMapSlice []map[string]string

	if err := yaml.Unmarshal(buff, &skipGVKMapSlice); err != nil {
		return nil, err
	}

	for _, skipGVK := range skipGVKMapSlice {
		skipGVKs = append(skipGVKs, schema.GroupVersionKind{
			Group:   skipGVK["group"],
			Version: skipGVK["version"],
			Kind:    skipGVK["kind"],
		})
	}

	return skipGVKs, nil
}
