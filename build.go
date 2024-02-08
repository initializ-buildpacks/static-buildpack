import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/buildpacks/libcnb"
)

type Build struct {
	Logger bard.Logger
}

func (b Build) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	b.Logger.Title(context.Buildpack)
	result := libcnb.NewBuildResult()

	b.Logger.Info("Starting detection and installation process...")

	// Check if Vite is present in package.json
	viteDetected := false
	packageJSONPath := filepath.Join(context.Application.Path, "package.json")
	file, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("failed to read package.json: %w", err)
	}
	if containsViteDependency(file) {
		viteDetected = true
		b.Logger.Infof("Vite detected in package.json")
	}

	// Install npm dependency if Vite is detected
	if viteDetected {
		b.Logger.Info("Installing serve npm dependency...")

		// Run npm install to install the dependency
		cmd := exec.Command("npm", "install", "serve")
		cmd.Dir = context.Application.Path
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("failed to install npm dependency: %w", err)
		}

		b.Logger.Info("Serve npm dependency installed successfully")
	}

	// Return build result indicating successful completion
	return result, nil
}

// Function to check if "vite" dependency is present in package.json
func containsViteDependency(packageJSON []byte) bool {
	return bytes.Contains(packageJSON, []byte(`"vite"`))
}
