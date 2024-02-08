import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak/bard"
)

type Detect struct {
	Logger bard.Logger
}

func (d Detect) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	// Detect Vite in package.json
	packageJSONPath := filepath.Join(context.Application.Path, "package.json")
	file, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return libcnb.DetectResult{}, fmt.Errorf("failed to read package.json: %w", err)
	}

	var packageJSON map[string]interface{}
	if err := json.Unmarshal(file, &packageJSON); err != nil {
		return libcnb.DetectResult{}, fmt.Errorf("failed to parse package.json: %w", err)
	}

	viteDetected := false
	if _, ok := packageJSON["vite"]; ok {
		viteDetected = true
	}

	// Adjust Build Plan based on Detection
	buildPlans := []libcnb.BuildPlan{}

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
