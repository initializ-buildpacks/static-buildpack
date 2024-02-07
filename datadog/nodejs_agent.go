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

	// Detect Vite Configuration Files
	viteConfigJS := filepath.Join(n.ApplicationPath, "vite.config.js")
	viteConfigTS := filepath.Join(n.ApplicationPath, "vite.config.ts")

	if _, err := os.Stat(viteConfigJS); err == nil {
		// Vite config file detected, install related dependencies
		if err := n.Executor.Execute(effect.Execution{
			Command: "npm",
			Args:    []string{"install", "--no-save", "vite", "serve", "other_dependency"},
			Dir:     layer.Path,
			Stdout:  n.Logger.InfoWriter(),
			Stderr:  n.Logger.InfoWriter(),
		}); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to install Vite related dependencies\n%w", err)
		}
	} else if _, err := os.Stat(viteConfigTS); err == nil {
		// Vite TypeScript config file detected, install related dependencies
		if err := n.Executor.Execute(effect.Execution{
			Command: "npm",
			Args:    []string{"install", "--no-save", "vite", "serve", "typescript", "other_dependency"},
			Dir:     layer.Path,
			Stdout:  n.Logger.InfoWriter(),
			Stderr:  n.Logger.InfoWriter(),
		}); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to install Vite TypeScript related dependencies\n%w", err)
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
			return libcnb.Layer{}, fmt.Errorf("unable to run npm install\n%w", err)
		}

		layer.LaunchEnvironment.Prepend("NODE_PATH", string(os.PathListSeparator), filepath.Join(layer.Path, "node_modules"))

		return layer, nil
	})
	if err != nil {
		return libcnb.Layer{}, fmt.Errorf("unable to install node module\n%w", err)
	}

	m, err := sherpa.NodeJSMainModule(n.ApplicationPath)
	if err != nil {
		return libcnb.Layer{}, fmt.Errorf("unable to find main module in %s\n%w", n.ApplicationPath, err)
	}

	file := filepath.Join(n.ApplicationPath, m)
	c, err := ioutil.ReadFile(file)
	if err != nil {
		return libcnb.Layer{}, fmt.Errorf("unable to read contents of %s\n%w", file, err)
	}

	if !regexp.MustCompile(`require\(['"]dd-trace['"]\)\.init\(\)`).Match(c) {
		n.Logger.Header("Requiring 'dd-trace' module")

		if err := ioutil.WriteFile(file, append([]byte("require('dd-trace').init();\n"), c...), 0644); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to write main module %s\n%w", file, err)
		}
	}

	return layer, nil
}

func (n NodeJSAgent) Name() string {
	return n.LayerContributor.LayerName()
}
