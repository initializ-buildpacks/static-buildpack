package helper

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/heroku/color"
	"github.com/paketo-buildpacks/libpak/bard"
)

type Toggle struct {
	Logger bard.Logger
}

func (t Toggle) Execute() (map[string]string, error) {
	t.Logger.Infof(color.CyanString("Vite detection and npm dependency installation process started..."))

	// Check if Vite is present in package.json
	viteDetected := false
	packageJSONPath := filepath.Join(".", "package.json")
	file, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read package.json: %w", err)
	}
	if strings.Contains(string(file), `"vite"`) {
		viteDetected = true
		t.Logger.Infof(color.GreenString("Vite detected in package.json"))
	}

	// Install npm dependency 'serve' if Vite is detected
	if viteDetected {
		t.Logger.Infof(color.CyanString("Installing 'serve' npm dependency..."))

		// Run npm install to install the 'serve' dependency
		cmd := exec.Command("npm", "install", "serve")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return nil, fmt.Errorf("failed to install 'serve' npm dependency: %w", err)
		}

		t.Logger.Infof(color.GreenString("'serve' npm dependency installed successfully"))
	}

	// If Vite is not detected or npm dependency is installed successfully, return nil
	return nil, nil
}
