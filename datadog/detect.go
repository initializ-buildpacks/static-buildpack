package datadog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
)

type Detect struct {
	Logger bard.Logger
}

func (d Detect) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	cr, err := libpak.NewConfigurationResolver(context.Buildpack, &d.Logger)
	if err != nil {
		return libcnb.DetectResult{}, fmt.Errorf("unable to create configuration resolver: %w", err)
	}

	if !cr.ResolveBool("BP_DATADOG_ENABLED") {
		d.Logger.Info("SKIPPED: variable 'BP_DATADOG_ENABLED' not set to true")
		return libcnb.DetectResult{Pass: false}, nil
	}

	// Check for Vite in package.json
	if viteDetected, err := isViteInPackageJSON(context.Application.Path); err != nil {
		return libcnb.DetectResult{}, fmt.Errorf("failed to detect Vite in package.json: %w", err)
	} else if viteDetected {
		d.Logger.Info("Vite detected in package.json, installing serve package")
		return libcnb.DetectResult{
			Pass: true,
			Plans: []libcnb.BuildPlan{
				{
					Provides: []libcnb.BuildPlanProvide{
						{Name: "serve"},
					},
				},
			},
		}, nil
	}

	// Vite not detected in package.json, continue with existing detection logic...

	buildPlans := []libcnb.BuildPlan{
		{
			Provides: []libcnb.BuildPlanProvide{
				{Name: "datadog-java"},
			},
			Requires: []libcnb.BuildPlanRequire{
				{Name: "datadog-java"},
				{Name: "jvm-application"},
			},
		},
		{
			Provides: []libcnb.BuildPlanProvide{
				{Name: "datadog-nodejs"},
			},
			Requires: []libcnb.BuildPlanRequire{
				{Name: "datadog-nodejs"},
				{Name: "node_modules"},
				{Name: "node", Metadata: map[string]interface{}{"build": true}},
			},
		},
	}

	return libcnb.DetectResult{
		Pass:  true,
		Plans: buildPlans,
	}, nil
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
