package providers

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"go.borchero.com/cuckoo/utils"
	"go.borchero.com/typewriter"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/storage/driver"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // enable gcp auth providers
)

// HelmRelease describes the (potential) release of a Helm chart.
type HelmRelease struct {
	repo       string
	name       string
	namespace  string
	chart      string
	version    string
	appVersion string
	config     *action.Configuration
	settings   *cli.EnvSettings
	logger     typewriter.CLILogger
}

type helmChart struct {
	APIVersion   string        `yaml:"apiVersion"`
	Type         string        `yaml:"type"`
	Name         string        `yaml:"name"`
	Version      string        `yaml:"version"`
	AppVersion   string        `yaml:"appVersion"`
	Dependencies []interface{} `yaml:"dependencies"`
}

type templateValues struct {
	Name      string
	Namespace string
}

// NewHelmRelease initializes a new Helm release and sets its properties correctly.
func NewHelmRelease(
	repo, chart, version, name, namespace string, logger typewriter.CLILogger,
) (*HelmRelease, error) {
	os.Setenv("HELM_NAMESPACE", namespace)
	settings := cli.New()

	config := &action.Configuration{}
	err := config.Init(
		settings.RESTClientGetter(), namespace, "secrets",
		func(format string, values ...interface{}) {
			fmt.Printf(fmt.Sprintf("%s\n", format), values...)
		},
	)
	if err != nil {
		return nil, err
	}

	return &HelmRelease{
		repo:      repo,
		name:      name,
		namespace: namespace,
		chart:     chart,
		version:   version,
		config:    config,
		settings:  settings,
		logger:    logger,
	}, nil
}

// IsLocalChart returns whether the release refers to a proper local chart (including a templates
// directory).
func (release *HelmRelease) IsLocalChart() bool {
	if release.repo != "" {
		return false
	}
	templatesDir := fmt.Sprintf("%s/templates", release.chart)
	_, err := os.Stat(templatesDir)
	return err == nil
}

// Upgrade runs the helm upgrade command for the release (and optionally installs).
func (release *HelmRelease) Upgrade(valuesFiles []string, image, tag string, dryRun bool) error {
	// 1) Get Helm chart
	var chart *chart.Chart
	var values map[string]interface{}
	var err error

	if release.repo != "" {
		chart, values, err = release.getRemoteChart(valuesFiles)
	} else if release.IsLocalChart() {
		chart, values, err = release.getLocalChart(valuesFiles, image, tag)
	} else {
		chart, values, err = release.getLocalDir()
	}
	if err != nil {
		return err
	}

	// 2) Check if release already exists, install if not
	history := action.NewHistory(release.config)
	history.Max = 1
	if _, err := history.Run(release.name); err == driver.ErrReleaseNotFound {
		// 2.1) Release does not exist, install
		release.logger.Infof("Installing %s...", release.name)

		install := action.NewInstall(release.config)
		install.DryRun = dryRun
		install.DisableHooks = false
		install.Timeout = 15 * time.Minute
		install.Wait = true
		install.Atomic = true
		install.ReleaseName = release.name
		install.Namespace = release.namespace
		install.Version = release.version

		_, err := install.Run(chart, values)
		if err != nil {
			return fmt.Errorf("Unable to freshly install release: %s", err)
		}
		return nil
	}

	// 3) Otherwise, install
	release.logger.Infof("Upgrading %s...", release.name)

	upgrade := action.NewUpgrade(release.config)
	upgrade.DryRun = dryRun
	upgrade.DisableHooks = false
	upgrade.Timeout = 15 * time.Minute
	upgrade.Wait = true
	upgrade.Atomic = true
	upgrade.MaxHistory = 10
	upgrade.Namespace = release.namespace
	upgrade.Version = release.version

	_, err = upgrade.Run(release.name, chart, values)
	if err != nil {
		return fmt.Errorf("Unable to upgrade release: %s", err)
	}
	return nil
}

func (release *HelmRelease) getRemoteChart(
	valuesFiles []string,
) (*chart.Chart, map[string]interface{}, error) {
	// 1) Get (remote) location of chart
	chartPath, err := release.locateChart(release.chart)
	if err != nil {
		return nil, nil, err
	}

	// 2) Load chart
	chart, err := release.loadChart(chartPath)
	if err != nil {
		return nil, nil, err
	}

	// 3) Read values file
	valueOpts, err := release.readTemplatedValuesFiles(valuesFiles)
	if err != nil {
		return nil, nil, err
	}

	// 4) Get values
	values, err := release.getValuesFromOptions(valueOpts)
	if err != nil {
		return nil, nil, err
	}

	return chart, values, nil
}

func (release *HelmRelease) getLocalChart(
	valuesFiles []string, image, tag string,
) (*chart.Chart, map[string]interface{}, error) {
	// 1) Create Chart.yaml
	err := release.writeChartYaml(release.chart, tag)
	if err != nil {
		return nil, nil, err
	}

	// 2) Get location of chart
	chartPath, err := release.locateChart(release.chart)
	if err != nil {
		return nil, nil, err
	}

	// 3) Populate default values.yaml
	err = release.populateDefaultValuesTemplate()
	if err != nil {
		return nil, nil, err
	}

	// 4) Load chart
	chart, err := release.loadChart(chartPath)
	if err != nil {
		return nil, nil, err
	}

	// 5) Update dependencies
	chart, err = release.downloadDependencies(chart, chartPath)
	if err != nil {
		return nil, nil, err
	}

	// 6) Read values
	valueOpts, err := release.readTemplatedValuesFiles(valuesFiles)
	if err != nil {
		return nil, nil, err
	}

	// 7) Set values for image and tag
	valueOpts.Values = []string{
		fmt.Sprintf("image.name=%s", image),
		fmt.Sprintf("image.tag=%s", tag),
	}

	// 8) Get values
	values, err := release.getValuesFromOptions(valueOpts)
	if err != nil {
		return nil, nil, err
	}

	return chart, values, nil
}

func (release *HelmRelease) getLocalDir() (*chart.Chart, map[string]interface{}, error) {
	chartDir := fmt.Sprintf("%s/cuckoo-deploy", os.TempDir())
	templateDir := fmt.Sprintf("%s/templates", chartDir)
	defer os.RemoveAll(chartDir)

	// 1) Link all files to templates directory
	err := os.MkdirAll(templateDir, 0755)
	if err != nil {
		return nil, nil, fmt.Errorf("Cannot generate temporary directory to store chart: %s", err)
	}

	// 1.1) Check if chart is file instead of directory
	chartInfo, err := os.Stat(release.chart)
	if err != nil {
		return nil, nil, fmt.Errorf("Cannot determine whether chart is file or directory: %s", err)
	}

	// 1.2) Link file(s)
	if chartInfo.IsDir() {
		files, err := utils.GetMatchingFiles(".*", release.chart)
		if err != nil {
			return nil, nil, fmt.Errorf("Cannot link to files in chart directory: %s", err)
		}

		for _, file := range files {
			target := fmt.Sprintf("%s/%s", templateDir, strings.ReplaceAll(file, "/", "-"))
			err := os.Link(file, target)
			if err != nil {
				return nil, nil, fmt.Errorf("Cannot link to file '%s': %s", file, err)
			}
		}
	} else {
		target := fmt.Sprintf("%s/%s", templateDir, strings.ReplaceAll(release.chart, "/", "-"))
		err := os.Link(release.chart, target)
		if err != nil {
			return nil, nil, fmt.Errorf("Cannot link to chart file '%s': %s", release.chart, err)
		}
	}

	// 2) Create Chart.yaml
	err = release.writeChartYaml(chartDir, "0.0.0")
	if err != nil {
		return nil, nil, err
	}

	// 3) Get location of chart
	chartPath, err := release.locateChart(chartDir)
	if err != nil {
		return nil, nil, err
	}

	// 4) Load chart
	chart, err := release.loadChart(chartPath)
	if err != nil {
		return nil, nil, err
	}

	return chart, map[string]interface{}{}, nil
}

func (release *HelmRelease) writeChartYaml(chartDir, appVersion string) error {
	// 1) Generate Chart.yaml
	chartDefinition := helmChart{
		APIVersion: "v2", Type: "application", Name: release.name,
		Version: release.version, AppVersion: appVersion,
	}

	// 2) Add dependencies if file exists
	dependencyFile := fmt.Sprintf("%s/dependencies.yaml", chartDir)
	if _, err := os.Stat(dependencyFile); err == nil {
		// File exists, read and add to chart definition
		dependenciesContents, err := ioutil.ReadFile(dependencyFile)
		if err != nil {
			return fmt.Errorf("Cannot read existing dependencies file: %s", err)
		}

		err = yaml.Unmarshal(dependenciesContents, &chartDefinition.Dependencies)
		if err != nil {
			return fmt.Errorf("Cannot parse dependencies file: %s", err)
		}
	}

	// 3) Write Chart.yaml
	chartContents, _ := yaml.Marshal(&chartDefinition)
	chartsFile := fmt.Sprintf("%s/Chart.yaml", chartDir)
	err := ioutil.WriteFile(chartsFile, chartContents, 0644)
	if err != nil {
		return fmt.Errorf("Cannot write Chart.yaml: %s", err)
	}

	return nil
}

func (release *HelmRelease) locateChart(location string) (string, error) {
	pathOptions := action.ChartPathOptions{
		Version: release.version,
		RepoURL: release.repo,
	}

	chartPath, err := pathOptions.LocateChart(location, release.settings)
	if err != nil {
		return "", fmt.Errorf("Unable to find Chart: %s", err)
	}

	return chartPath, nil
}

func (release *HelmRelease) populateDefaultValuesTemplate() error {
	defaultValuesFile := fmt.Sprintf("%s/values.yaml", release.chart)
	if _, err := os.Stat(defaultValuesFile); os.IsNotExist(err) {
		return nil
	}

	valueTemplate := templateValues{release.name, release.namespace}
	err := utils.PopulateTemplateWrite(defaultValuesFile, defaultValuesFile, valueTemplate)
	if err != nil {
		return fmt.Errorf("Cannot replace template values in default values.yaml: %s", err)
	}

	return nil
}

func (release *HelmRelease) loadChart(path string) (*chart.Chart, error) {
	chart, err := loader.Load(path)
	if err != nil {
		return nil, fmt.Errorf("Unable to load chart: %s", err)
	}
	return chart, err
}

func (release *HelmRelease) downloadDependencies(
	chart *chart.Chart, chartPath string,
) (*chart.Chart, error) {
	if err := action.CheckDependencies(chart, chart.Metadata.Dependencies); err != nil {
		manager := &downloader.Manager{
			Out:              os.Stdout,
			ChartPath:        chartPath,
			SkipUpdate:       false,
			Getters:          getter.All(release.settings),
			RepositoryConfig: release.settings.RepositoryConfig,
			RepositoryCache:  release.settings.RepositoryCache,
		}
		err := manager.Update()
		if err != nil {
			return nil, fmt.Errorf("Failed to download dependencies: %s", err)
		}

		// 1.4.2) If needed, reload dependencies
		chart, err = loader.Load(chartPath)
		if err != nil {
			return nil, fmt.Errorf("Unable to load chart after downloading dependencies: %s", err)
		}
	}
	return chart, nil
}

func (release *HelmRelease) readTemplatedValuesFiles(files []string) (*values.Options, error) {
	valueOpts := &values.Options{ValueFiles: []string{}}

	for _, file := range files {
		valueTemplate := templateValues{release.name, release.namespace}
		err := utils.PopulateTemplateWrite(file, file, valueTemplate)
		if err != nil {
			return nil, fmt.Errorf("Cannot read values file: %s", err)
		}
		valueOpts.ValueFiles = append(valueOpts.ValueFiles, file)
	}

	return valueOpts, nil
}

func (release *HelmRelease) getValuesFromOptions(
	options *values.Options,
) (map[string]interface{}, error) {
	values, err := options.MergeValues(getter.All(release.settings))
	if err != nil {
		return nil, fmt.Errorf("Cannot correctly load all values: %s", err)
	}
	return values, nil
}
