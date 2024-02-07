package datadog

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/effect"
	"github.com/paketo-buildpacks/libpak/sherpa"
)

type NodeJSAgent struct {
	ApplicationPath  string
	BuildpackPath    string
	Executor         effect.Executor
	LayerContributor libpak.DependencyLayerContributor
	Logger           bard.Logger
}

func NewNodeJSAgent(applicationPath string, buildpackPath string, dependency libpak.BuildpackDependency, cache libpak.DependencyCache, logger bard.Logger) NodeJSAgent {
	contributor, _ := libpak.NewDependencyLayer(dependency, cache, libcnb.LayerTypes{Launch: true})
	return NodeJSAgent{
		ApplicationPath:  applicationPath,
		BuildpackPath:    buildpackPath,
		Executor:         effect.NewExecutor(),
		LayerContributor: contributor,
		Logger:           logger,
	}
}

func (n NodeJSAgent) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	n.LayerContributor.Logger = n.Logger

	// Check for Vite in package.json
	if viteDetected, err := isViteInPackageJSON(n.ApplicationPath); err != nil {
		return libcnb.Layer{}, fmt.Errorf("failed to detect Vite in package.json: %w", err)
	} else if viteDetected {
		// Vite detected, install related dependencies
		if err := n.Executor.Execute(effect.Execution{
			Command: "npm",
			Args:    []string{"install", "--no-save", "vite", "serve", "other_dependency"},
			Dir:     layer.Path,
			Stdout:  n.Logger.InfoWriter(),
			Stderr:  n.Logger.InfoWriter(),
		}); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to install Vite related dependencies: %w", err)
		}
	}

	// Update Launch Environment if necessary
	layer.LaunchEnvironment.Default("VITE_ENV", "production")

	layer, err := n.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		n.Logger.Bodyf("Installing to %s", layer.Path)

		if err := n.Executor.Execute(effect.Execution{
			Command: "npm",
			Args:    []string{"install", "--no-save", artifact.Name()},
			Dir:     layer.Path,
			Stdout:  n.Logger.InfoWriter(),
			Stderr:  n.Logger.InfoWriter(),
		}); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to run npm install: %w", err)
		}

		layer.LaunchEnvironment.Prepend("NODE_PATH", string(os.PathListSeparator), filepath.Join(layer.Path, "node_modules"))

		return layer, nil
	})
	if err != nil {
		return libcnb.Layer{}, fmt.Errorf("unable to install node module: %w", err)
	}

	m, err := sherpa.NodeJSMainModule(n.ApplicationPath)
	if err != nil {
		return libcnb.Layer{}, fmt.Errorf("unable to find main module in %s: %w", n.ApplicationPath, err)
	}

	file := filepath.Join(n.ApplicationPath, m)
	c, err := ioutil.ReadFile(file)
	if err != nil {
		return libcnb.Layer{}, fmt.Errorf("unable to read contents of %s: %w", file, err)
	}

	if !regexp.MustCompile(`require\(['"]dd-trace['"]\)\.init\(\)`).Match(c) {
		n.Logger.Header("Requiring 'dd-trace' module")

		if err := ioutil.WriteFile(file, append([]byte("require('dd-trace').init();\n"), c...), 0644); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to write main module %s: %w", file, err)
		}
	}

	return layer, nil
}

func (n NodeJSAgent) Name() string {
	return n.LayerContributor.LayerName()
}

func isViteInPackageJSON(appPath string) (bool, error) {
	packageJSONPath := filepath.Join(appPath, "package.json")
	file, err := os.Open(packageJSONPath)
	if err != nil {
		return false, fmt.Errorf("failed to open package.json: %w", err)
	}
	defer file.Close()

	var packageJSON map[string]interface{}
	if err := json.NewDecoder(file).Decode(&packageJSON); err != nil {
		return false, fmt.Errorf("failed to decode package.json: %w", err)
	}

	dependencies, ok := packageJSON["dependencies"].(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("dependencies key not found in package.json")
	}

	_, viteFound := dependencies["vite"]
	return viteFound, nil
}
