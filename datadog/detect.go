package datadog

import (
	"fmt"

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
		return libcnb.DetectResult{}, fmt.Errorf("unable to create configuration resolver\n%w", err)
	}

	if !cr.ResolveBool("BP_DATADOG_ENABLED") {
		d.Logger.Info("SKIPPED: variable 'BP_DATADOG_ENABLED' not set to true")
		return libcnb.DetectResult{Pass: false}, nil
	}

	// Detect Vite Configuration Files
	viteConfigJS := context.Application.Path.Join("vite.config.js")
	viteConfigTS := context.Application.Path.Join("vite.config.ts")

	viteDetected := false
	if _, err := context.Application.Path.Stat(viteConfigJS); err == nil {
		viteDetected = true
	} else if _, err := context.Application.Path.Stat(viteConfigTS); err == nil {
		viteDetected = true
	}

	// Adjust Build Plan based on Detection
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

	if viteDetected {
		buildPlans = append(buildPlans, libcnb.BuildPlan{
			Provides: []libcnb.BuildPlanProvide{
				{Name: "vite-configuration"},
			},
			Requires: []libcnb.BuildPlanRequire{
				{Name: "vite-configuration"},
			},
		})
	}

	return libcnb.DetectResult{
		Pass:   true,
		Plans:  buildPlans,
	}, nil
}
