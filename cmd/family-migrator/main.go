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

// main package for the family-migrator tool...
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/alecthomas/kong"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/pkg/errors"
	"github.com/upbound/upjet/pkg/migration"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/upbound/extensions-migration/pkg/converter/configuration"
)

const (
	defaultKubeConfig = ".kube/config"

	stepByStepChoice    = "Step-by-Step"
	noInteractionChoice = "No Interaction"

	automaticallyChoice = "Automatically"
	manuallyChoice      = "Manually"
	skipChoice          = "Skip"
	retryChoice         = "Retry"
	cancelChoice        = "Cancel"

	providerAwsChoice   = "provider-aws"
	providerAzureChoice = "provider-azure"
	providerGcpChoice   = "provider-gcp"

	configurationMode = "configuration"
	justMrMode        = "managed"
)

var monolithicToFamily = map[string]string{
	"provider-aws":   "provider-family-aws",
	"provider-azure": "provider-family-azure",
	"provider-gcp":   "provider-family-gcp",
}

// Options represents the available options for the family-migrator.
type Options struct {
	Generate struct {
		Configuration struct {
			SourceConfigurationPackage string `name:"source-configuration-package" help:"Migration source Configuration package's URL." survey:"source-configuration-package"`
			TargetConfigurationPackage string `name:"target-configuration-package" help:"Migration target Configuration package's URL." survey:"target-configuration-package"`

			PackageRoot   string `name:"package-root" help:"Source directory for the Crossplane Configuration package." survey:"package-root"`
			ExamplesRoot  string `name:"examples-root" help:"Path to Crossplane package examples directory." survey:"examples-root"`
			PackageOutput string `name:"package-output" help:"Path to store the updated configuration package." survey:"package-output"`
		} `kong:"cmd"`

		Managed struct {
			resourcePath string
		} `kong:"cmd"`

		RegistryOrg        string `name:"regorg" help:"<registry host>/<organization> for the provider family packages."`
		AWSFamilyVersion   string `name:"aws-family-version" help:"Version of the AWS provider family."`
		AzureFamilyVersion string `name:"azure-family-version" help:"Version of the Azure provider family."`
		GCPFamilyVersion   string `name:"gcp-family-version" help:"Version of the GCP provider family."`

		KubeConfig string `name:"kubeconfig" help:"Path to the kubeconfig to use."`
	} `kong:"cmd"`

	Execute struct{} `kong:"cmd"`

	PlanPath string `name:"plan-path" help:"Migration plan output path." survey:"plan-path"`

	Debug bool `name:"debug" short:"d" help:"Run with debug logging."`
}

func main() {
	opts := &Options{}
	kongCtx := kong.Parse(opts, kong.Name("family-migrator"),
		kong.Description("Upbound provider families migration tool"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact:   true,
			FlagsLast: true,
			Summary:   true,
		}))

	getCommonInputs(kongCtx, opts)
	absPath, err := filepath.Abs(opts.PlanPath)
	kongCtx.FatalIfErrorf(err, "Failed to get the absolute path for the migration plan output: %s", opts.PlanPath)
	planDir := filepath.Dir(absPath)

	c := strings.Split(kongCtx.Command(), " ")
	switch c[0] {
	case "generate":
		mode := c[1]
		getGenerateInputs(kongCtx, planDir, opts, mode)
		generatePlan(kongCtx, opts, planDir, mode)
	case "execute":
		executePlan(kongCtx, planDir, opts)
	}
}

func generatePlan(kongCtx *kong.Context, opts *Options, planDir string, mode string) {
	r := migration.NewRegistry(runtime.NewScheme())

	switch mode {
	case configurationMode:
		if err := registerConfigurationPackageConverters(opts, r); err != nil {
			kongCtx.FatalIfErrorf(err, "Failed to register converters")
		}
	case justMrMode:
		if err := registerManagedResourceConverters(opts, r); err != nil {
			kongCtx.FatalIfErrorf(err, "Failed to register converters")
		}
	}

	if len(opts.Generate.KubeConfig) == 0 {
		homeDir, err := os.UserHomeDir()
		kongCtx.FatalIfErrorf(err, "Failed to get user's home")
		opts.Generate.KubeConfig = filepath.Join(homeDir, defaultKubeConfig)
	}

	sources, err := initializeSources(mode, r, opts)
	if err != nil {
		kongCtx.FatalIfErrorf(err, "Failed to initialize sources")
	}

	pgOpts := []migration.PlanGeneratorOption{
		migration.WithMultipleSources(sources...),
		migration.WithSkipGVKs(schema.GroupVersionKind{}),
		migration.WithEnableConfigurationMigrationSteps(),
	}

	pg := migration.NewPlanGenerator(r, nil, migration.NewFileSystemTarget(migration.WithParentDirectory(planDir)), pgOpts...)
	kongCtx.FatalIfErrorf(pg.GeneratePlan(), "Failed to generate the migration plan for the provider families")

	if mode == configurationMode {
		setPkgParameters(&pg.Plan, *opts)
	}

	buff, err := yaml.Marshal(pg.Plan)
	kongCtx.FatalIfErrorf(err, "Failed to marshal the migration plan to YAML")
	kongCtx.FatalIfErrorf(os.WriteFile(opts.PlanPath, buff, 0600), "Failed to store the migration plan at path: %s", opts.PlanPath)

	var moveExecution bool
	moveExecutionPhaseQuestion := &survey.Confirm{
		Message: fmt.Sprintf("The migration plan has been generated at path: %s. The referred resource manifests and the patch documents can be found under: %s.\n"+
			"Would you like to proceed to the execution phase?", opts.PlanPath, planDir),
	}
	kongCtx.FatalIfErrorf(survey.AskOne(moveExecutionPhaseQuestion, &moveExecution))
	if moveExecution {
		executePlan(kongCtx, planDir, opts)
	}
}

func initializeSources(mode string, r *migration.Registry, opts *Options) ([]migration.Source, error) {
	kubeSource, err := migration.NewKubernetesSourceFromKubeConfig(opts.Generate.KubeConfig, migration.WithRegistry(r), migration.WithCategories([]migration.Category{migration.CategoryManaged}))
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to initialize the migration Kubernetes source from kubeconfig: %s", opts.Generate.KubeConfig)
	}
	sources := []migration.Source{kubeSource}

	if mode == configurationMode {
		fsSource, err := migration.NewFileSystemSource(opts.Generate.Configuration.PackageRoot)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to initialize the migration FileSystem source from path: %s", opts.Generate.Configuration.PackageRoot)
		}
		return []migration.Source{fsSource, kubeSource}, nil
	}
	return sources, nil
}

func executePlan(kongCtx *kong.Context, planDir string, opts *Options) {
	plan := &migration.Plan{}
	buff, err := os.ReadFile(opts.PlanPath)
	kongCtx.FatalIfErrorf(err, "Failed to read the migration plan from path: %s", opts.PlanPath)
	kongCtx.FatalIfErrorf(yaml.Unmarshal(buff, plan), "Failed to unmarshal the migration plan: %s", opts.PlanPath)

	stepByStep := askExecutionSteps(kongCtx, plan, opts, planDir)
	zl := zap.New(zap.UseDevMode(opts.Debug))
	log := logging.NewLogrLogger(zl.WithName("fork-executor"))
	executor := migration.NewForkExecutor(migration.WithWorkingDir(planDir), migration.WithLogger(log))
	// TODO: we need to load the plan back from the filesystem as it may
	// have been modified.
	var cb migration.ExecutorCallback
	cb = &loggerCallback{
		logger: logging.NewLogrLogger(zl.WithName("family-migrator")),
	}
	if stepByStep {
		cb = &executionCallback{
			logger: logging.NewLogrLogger(zl.WithName("family-migrator")),
		}
	}
	planExecutor := migration.NewPlanExecutor(*plan, []migration.Executor{executor},
		migration.WithExecutorCallback(cb))
	backupDir := filepath.Join(planDir, "backup")
	kongCtx.FatalIfErrorf(os.MkdirAll(backupDir, 0o700), "Failed to mkdir backup directory: %s", backupDir)
	kongCtx.FatalIfErrorf(planExecutor.Execute(), "Failed to execute the migration plan at path: %s", opts.PlanPath)
}

func setPkgParameters(plan *migration.Plan, opts Options) {
	for i, s := range plan.Spec.Steps {
		// TODO: consider exporting step constants. But the idea is
		// to introduce the concept of a migration Scenario that
		// encapsulated both the converters and the steps involved.
		if s.Name == "push-configuration" || s.Name == "build-configuration" {
			s.Exec.Args[1] = strings.ReplaceAll(s.Exec.Args[1], "{{TARGET_CONFIGURATION_PACKAGE}}", opts.Generate.Configuration.TargetConfigurationPackage)
			s.Exec.Args[1] = strings.ReplaceAll(s.Exec.Args[1], "{{PKG_PATH}}", opts.Generate.Configuration.PackageOutput)
			s.Exec.Args[1] = strings.ReplaceAll(s.Exec.Args[1], "{{PKG_ROOT}}", opts.Generate.Configuration.PackageRoot)
			s.Exec.Args[1] = strings.ReplaceAll(s.Exec.Args[1], "{{EXAMPLES_ROOT}}", opts.Generate.Configuration.ExamplesRoot)
			migration.AddManualExecution(&s)
			plan.Spec.Steps[i] = s
		}
	}
}

func getGenerateInputs(kongCtx *kong.Context, planDir string, opts *Options, mode string) {
	registryOrgValidator := func(ans interface{}) error {
		re := regexp.MustCompile(`(?i)([a-z0-9]+/[a-z0-9]+)`)
		if !re.MatchString(ans.(string)) {
			return errors.Errorf("The answer %q does not match with the format of reg-org. Format: <registry host>/<organization>", ans)
		}
		return nil
	}

	if opts.Generate.RegistryOrg == "" {
		regOrgQuestion := &survey.Input{
			Message: "Please provide the registry and organization for the provider family packages",
			Help:    "Input Format: <registry host>/<organization> Example xpkg.upbound.io/upbound",
		}
		kongCtx.FatalIfErrorf(survey.AskOne(regOrgQuestion, &opts.Generate.RegistryOrg, survey.WithValidator(registryOrgValidator)))
	}

	if opts.Generate.AWSFamilyVersion == "" && opts.Generate.AzureFamilyVersion == "" && opts.Generate.GCPFamilyVersion == "" {
		var selectedProviders []string
		providerSelection := &survey.MultiSelect{
			Message: "Please select the providers that will be migrated",
			Options: []string{
				providerAwsChoice,
				providerAzureChoice,
				providerGcpChoice,
			},
		}
		kongCtx.FatalIfErrorf(survey.AskOne(providerSelection, &selectedProviders))

		versionValidator := func(ans interface{}) error {
			re := regexp.MustCompile(`v[0-9]+\.[0-9]+\.[0-9]+`)
			if !re.MatchString(ans.(string)) {
				return errors.Errorf("The answer %q does not match with the format of version. Format: v0.x.y", ans)
			}
			return nil
		}

		for _, sp := range selectedProviders {
			versionQuestion := &survey.Input{
				Message: fmt.Sprintf("Please specify the version of the %s family. Possible versions:\n%s", sp, listFamilyProviderVersions(getFamilyProviderVersions(kongCtx, sp))),
				Help:    "Format: v0.x.y",
			}
			switch sp {
			case providerAwsChoice:
				kongCtx.FatalIfErrorf(survey.AskOne(versionQuestion, &opts.Generate.AWSFamilyVersion, survey.WithValidator(versionValidator)))
			case providerAzureChoice:
				kongCtx.FatalIfErrorf(survey.AskOne(versionQuestion, &opts.Generate.AzureFamilyVersion, survey.WithValidator(versionValidator)))
			case providerGcpChoice:
				kongCtx.FatalIfErrorf(survey.AskOne(versionQuestion, &opts.Generate.GCPFamilyVersion, survey.WithValidator(versionValidator)))
			}
		}
	}

	if mode == configurationMode {
		var packageAndPathQuestions []*survey.Question
		if opts.Generate.Configuration.SourceConfigurationPackage == "" {
			packageAndPathQuestions = append(packageAndPathQuestions, &survey.Question{
				Name: "source-configuration-package",
				Prompt: &survey.Input{
					Message: "Please enter the URL of the migration source Configuration package",
					Help:    "Example: xpkg.upbound.io/upbound/platform-ref-gcp:v0.3.0",
				},
			})
		}
		if opts.Generate.Configuration.TargetConfigurationPackage == "" {
			packageAndPathQuestions = append(packageAndPathQuestions, &survey.Question{
				Name: "target-configuration-package",
				Prompt: &survey.Input{
					Message: "Please enter the URL of the migration target Configuration package",
					Help:    "Example: xpkg.upbound.io/upbound/platform-ref-gcp:v0.4.0",
				},
			})
		}
		if opts.Generate.Configuration.PackageRoot == "" {
			packageAndPathQuestions = append(packageAndPathQuestions, &survey.Question{
				Name: "package-root",
				Prompt: &survey.Input{
					Message: "Please specify the source directory for the Crossplane Configuration package",
				},
			})
		}
		if opts.Generate.Configuration.ExamplesRoot == "" {
			packageAndPathQuestions = append(packageAndPathQuestions, &survey.Question{
				Name: "examples-root",
				Prompt: &survey.Input{
					Message: "Please specify the path to the directory containing the Crossplane package examples",
				},
			})
		}
		if opts.Generate.Configuration.PackageOutput == "" {
			opts.Generate.Configuration.PackageOutput = filepath.Join(planDir, "updated-package.pkg")
		}
		kongCtx.FatalIfErrorf(survey.Ask(packageAndPathQuestions, &opts.Generate.Configuration))
	}
}

func getCommonInputs(kongCtx *kong.Context, opts *Options) {
	if opts.PlanPath == "" {
		outputQuestion := &survey.Input{
			Message: "Please specify the path for the migration plan",
		}
		kongCtx.FatalIfErrorf(survey.AskOne(outputQuestion, &opts.PlanPath))
	}
}

func askExecutionSteps(kongCtx *kong.Context, plan *migration.Plan, opts *Options, planDir string) bool {
	var isReviewed bool
	reviewMigration := &survey.Confirm{
		Message: fmt.Sprintf("The migration file is here: %s. The referred resource manifests and the patch documents can be found under: %s. "+
			"Please review the migraiton plan and continue to the execution step.\n"+
			"Did you review the generated migration plan?", opts.PlanPath, planDir),
	}
	kongCtx.FatalIfErrorf(survey.AskOne(reviewMigration, &isReviewed))

	var displaySteps bool
	manualExecutionSteps := &survey.Confirm{
		Message: "The migration plan has manualExecution instructions. " +
			"Do you want the instructions to be listed?",
	}
	kongCtx.FatalIfErrorf(survey.AskOne(manualExecutionSteps, &displaySteps))
	if displaySteps {
		for _, s := range plan.Spec.Steps {
			for _, c := range s.ManualExecution {
				fmt.Println(c)
			}
		}
	}

	var moveExecutionChoice string
	moveExecutionChoiceQuestion := &survey.Select{
		Message: "Do you want to execute the migration plan with step-by-step confirmation or no interaction",
		Options: []string{
			stepByStepChoice,
			noInteractionChoice,
		},
	}
	kongCtx.FatalIfErrorf(survey.AskOne(moveExecutionChoiceQuestion, &moveExecutionChoice))
	switch moveExecutionChoice {
	case stepByStepChoice:
		return true
	default: // "No Interaction"
		return false
	}
}

func registerManagedResourceConverters(opts *Options, r *migration.Registry) error {
	if err := r.AddCrossplanePackageTypes(); err != nil {
		return errors.Wrap(err, "Failed to register the Provider package types with the migration registry")
	}
	mp := configuration.NewMRPreProcessor()
	r.RegisterPreProcessor(migration.CategoryManaged, migration.PreProcessor(mp.GetSSOPNameFromManagedResource))
	r.RegisterProviderPackageConverter(regexp.MustCompile(`xpkg.upbound.io/upbound/provider-aws:.+`), &configuration.ProviderPkgFamilyConfigParameters{
		FamilyVersion: opts.Generate.AWSFamilyVersion,
	})
	r.RegisterProviderPackageConverter(regexp.MustCompile(`xpkg.upbound.io/upbound/provider-azure:.+`), &configuration.ProviderPkgFamilyConfigParameters{
		FamilyVersion: opts.Generate.AzureFamilyVersion,
	})
	r.RegisterProviderPackageConverter(regexp.MustCompile(`xpkg.upbound.io/upbound/provider-gcp:.+`), &configuration.ProviderPkgFamilyConfigParameters{
		FamilyVersion: opts.Generate.GCPFamilyVersion,
	})
	// register converters for the family resource packages
	r.RegisterProviderPackageConverter(regexp.MustCompile(`xpkg.upbound.io/upbound/provider-aws:.+`), &configuration.ProviderPkgFamilyParameters{
		FamilyVersion:            opts.Generate.AWSFamilyVersion,
		Monolith:                 "provider-aws",
		ManagedResourceProcessor: mp,
	})
	r.RegisterProviderPackageConverter(regexp.MustCompile(`xpkg.upbound.io/upbound/provider-azure:.+`), &configuration.ProviderPkgFamilyParameters{
		FamilyVersion:            opts.Generate.AzureFamilyVersion,
		Monolith:                 "provider-azure",
		ManagedResourceProcessor: mp,
	})
	r.RegisterProviderPackageConverter(regexp.MustCompile(`xpkg.upbound.io/upbound/provider-gcp:.+`), &configuration.ProviderPkgFamilyParameters{
		FamilyVersion:            opts.Generate.GCPFamilyVersion,
		Monolith:                 "provider-gcp",
		ManagedResourceProcessor: mp,
	})
	return nil
}

func registerConfigurationPackageConverters(opts *Options, r *migration.Registry) error {
	if err := r.AddCrossplanePackageTypes(); err != nil {
		return errors.Wrap(err, "Failed to register the Provider package types with the migration registry")
	}
	cp := configuration.NewCompositionPreProcessor()
	r.RegisterPreProcessor(migration.CategoryComposition, migration.PreProcessor(cp.GetSSOPNameFromComposition))
	r.RegisterConfigurationMetadataConverter(migration.AllConfigurations, &configuration.ConfigMetaParameters{
		FamilyVersion:        opts.Generate.AWSFamilyVersion,
		Monolith:             "provider-aws",
		CompositionProcessor: cp,
	})
	r.RegisterConfigurationMetadataConverter(migration.AllConfigurations, &configuration.ConfigMetaParameters{
		FamilyVersion:        opts.Generate.AzureFamilyVersion,
		Monolith:             "provider-azure",
		CompositionProcessor: cp,
	})
	r.RegisterConfigurationMetadataConverter(migration.AllConfigurations, &configuration.ConfigMetaParameters{
		FamilyVersion:        opts.Generate.GCPFamilyVersion,
		Monolith:             "provider-gcp",
		CompositionProcessor: cp,
	})
	r.RegisterConfigurationPackageConverter(regexp.MustCompile(opts.Generate.Configuration.SourceConfigurationPackage), &configuration.ConfigPkgParameters{
		PackageURL: opts.Generate.Configuration.TargetConfigurationPackage,
	})
	// TODO: should we also handle missing registry (xpkg.upbound.io),
	// i.e., is it the default?
	// register converters for the family config packages
	r.RegisterProviderPackageConverter(regexp.MustCompile(`xpkg.upbound.io/upbound/provider-aws:.+`), &configuration.ProviderPkgFamilyConfigParameters{
		FamilyVersion: opts.Generate.AWSFamilyVersion,
	})
	r.RegisterProviderPackageConverter(regexp.MustCompile(`xpkg.upbound.io/upbound/provider-azure:.+`), &configuration.ProviderPkgFamilyConfigParameters{
		FamilyVersion: opts.Generate.AzureFamilyVersion,
	})
	r.RegisterProviderPackageConverter(regexp.MustCompile(`xpkg.upbound.io/upbound/provider-gcp:.+`), &configuration.ProviderPkgFamilyConfigParameters{
		FamilyVersion: opts.Generate.GCPFamilyVersion,
	})
	// register converters for the family resource packages
	r.RegisterProviderPackageConverter(regexp.MustCompile(`xpkg.upbound.io/upbound/provider-aws:.+`), &configuration.ProviderPkgFamilyParameters{
		FamilyVersion:        opts.Generate.AWSFamilyVersion,
		Monolith:             "provider-aws",
		CompositionProcessor: cp,
	})
	r.RegisterProviderPackageConverter(regexp.MustCompile(`xpkg.upbound.io/upbound/provider-azure:.+`), &configuration.ProviderPkgFamilyParameters{
		FamilyVersion:        opts.Generate.AzureFamilyVersion,
		Monolith:             "provider-azure",
		CompositionProcessor: cp,
	})
	r.RegisterProviderPackageConverter(regexp.MustCompile(`xpkg.upbound.io/upbound/provider-gcp:.+`), &configuration.ProviderPkgFamilyParameters{
		FamilyVersion:        opts.Generate.GCPFamilyVersion,
		Monolith:             "provider-gcp",
		CompositionProcessor: cp,
	})
	r.RegisterPackageLockConverter(migration.CrossplaneLockName, &configuration.LockParameters{
		PackageURL: opts.Generate.Configuration.SourceConfigurationPackage,
	})
	if err := r.AddCompositionTypes(); err != nil {
		return errors.Wrap(err, "Failed to register the Crossplane Composition types with the migration registry")
	}
	return nil
}

func getFamilyProviderVersions(kongCtx *kong.Context, providerName string) []interface{} {
	resp, err := http.Get(fmt.Sprintf("https://api.upbound.io/v1/packageMetadata/upbound/%s", monolithicToFamily[providerName]))
	if err != nil {
		fmt.Println("No suitable version found in the marketplace for the specified provider. Please check manually.")
	}
	if resp.StatusCode != 200 {
		return nil
	}
	defer func() {
		kongCtx.FatalIfErrorf(resp.Body.Close())
	}()
	var v interface{}
	err = json.NewDecoder(resp.Body).Decode(&v)
	if err != nil {
		fmt.Println("No suitable version found in the marketplace for the specified provider. Please check manually.")
		return nil
	}
	return v.(map[string]interface{})["versions"].([]interface{})
}

func listFamilyProviderVersions(versions []interface{}) string {
	r := make([]string, len(versions))
	for i, v := range versions {
		r[i] = v.(string)
	}
	return strings.Join(r, "\n")
}

type loggerCallback struct {
	logger logging.Logger
}

func (cb *loggerCallback) StepToExecute(s migration.Step, index int) migration.CallbackResult {
	cb.logger.Info("Executing step...", "index", index, "name", s.Name)
	return migration.CallbackResult{Action: migration.ActionContinue}
}

func (cb *loggerCallback) StepSucceeded(s migration.Step, index int, diagnostics any) migration.CallbackResult {
	cb.logger.Info("Step succeeded", "diagnostics", fmt.Sprintf("%s", diagnostics), "index", index, "name", s.Name)
	return migration.CallbackResult{Action: migration.ActionContinue}
}

func (cb *loggerCallback) StepFailed(s migration.Step, index int, diagnostics any, err error) migration.CallbackResult {
	cb.logger.Info("Step failed", "diagnostics", fmt.Sprintf("%s", diagnostics), "index", index, "name", s.Name, "err", err)
	return migration.CallbackResult{Action: migration.ActionCancel}
}

type executionCallback struct {
	logger logging.Logger
}

func (cb *executionCallback) StepToExecute(s migration.Step, index int) migration.CallbackResult {
	var executionChoice string
	buff := strings.Builder{}
	for _, c := range s.ManualExecution {
		buff.WriteString(c)
		buff.WriteString("\n")
	}
	executionChoiceQuestion := &survey.Select{
		Message: fmt.Sprintf("Step (with name %q at index %d) to execute:\n%s\n"+
			"What is your execution preference?", s.Name, index, buff.String()),
		Help: "Automatically: Commands will be executed automatically and the output will be shown.\n" +
			"Manually: Commands will not be executed and you will be prompted for confirmation that you have successfully run the command.\n" +
			"Skip: This step will be skipped.",
		Options: []string{
			automaticallyChoice,
			manuallyChoice,
			skipChoice,
		},
	}
	if err := survey.AskOne(executionChoiceQuestion, &executionChoice); err != nil {
		cb.logger.Info("Execution choice question could not ask or get answer", "index", index, "name", s.Name)
		return migration.CallbackResult{Action: migration.ActionCancel}
	}
	switch executionChoice {
	case manuallyChoice:
		isDone := false
		for !isDone {
			reviewMigration := &survey.Confirm{
				Message: "Manually execution was selected. Did you run the command?",
			}
			if err := survey.AskOne(reviewMigration, &isDone); err != nil {
				cb.logger.Info("Manual execution confirmation could not get", "index", index, "name", s.Name)
				return migration.CallbackResult{Action: migration.ActionCancel}
			}
		}
		return migration.CallbackResult{Action: migration.ActionSkip}
	case skipChoice:
		cb.logger.Info("Execution of this step skipped", "index", index, "name", s.Name)
		return migration.CallbackResult{Action: migration.ActionSkip}
	default: // "Automatically"
		cb.logger.Info("Executing step...", "index", index, "name", s.Name)
		return migration.CallbackResult{Action: migration.ActionContinue}
	}
}

func (cb *executionCallback) StepSucceeded(s migration.Step, index int, diagnostics any) migration.CallbackResult {
	cb.logger.Info("Step succeeded", "diagnostics", fmt.Sprintf("%s", diagnostics), "index", index, "name", s.Name)
	return migration.CallbackResult{Action: migration.ActionContinue}
}

func (cb *executionCallback) StepFailed(s migration.Step, index int, diagnostics any, err error) migration.CallbackResult {
	cb.logger.Info("Step failed", "diagnostics", fmt.Sprintf("%s", diagnostics), "index", index, "name", s.Name, "err", err)
	var failChoice string
	retryChoiceQuestion := &survey.Select{
		Message: fmt.Sprintf("Execution of this step failed: %s\n"+
			"What is your choice to continue?", s.ManualExecution),
		Help: "Retry: Command will be retried.\n" +
			"Skip: This step will be skipped.\n" +
			"Cancel: Execution of plan will be canceled.",
		Options: []string{
			retryChoice,
			skipChoice,
			cancelChoice,
		},
	}
	if err := survey.AskOne(retryChoiceQuestion, &failChoice); err != nil {
		cb.logger.Info("Retry choice question could not ask or get answer", "index", index, "name", s.Name)
		return migration.CallbackResult{Action: migration.ActionCancel}
	}
	switch failChoice {
	case skipChoice:
		cb.logger.Info("Step skipped", "index", index, "name", s.Name, "err", err)
		return migration.CallbackResult{Action: migration.ActionSkip}
	case cancelChoice:
		cb.logger.Info("Execution of plan canceled", "index", index, "name", s.Name, "err", err)
		return migration.CallbackResult{Action: migration.ActionCancel}
	default: // "Retry"
		cb.logger.Info("Step will be run again", "index", index, "name", s.Name, "err", err)
		return migration.CallbackResult{Action: migration.ActionRepeat}
	}

}
