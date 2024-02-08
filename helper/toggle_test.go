package helper_test

import (
	"io/ioutil"
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/initializ-buildpacks/static-buildpack/helper"
)

func testToggle(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		toggle = helper.Toggle{}
	)

	it("does not install npm dependency if Vite is not detected", func() {
		// Create a temporary package.json file without Vite
		tmpFile, err := ioutil.TempFile("", "package.json")
		Expect(err).ToNot(HaveOccurred())
		defer os.Remove(tmpFile.Name())

		err = ioutil.WriteFile(tmpFile.Name(), []byte(`{"name": "test"}`), 0644)
		Expect(err).ToNot(HaveOccurred())

		// Set the current directory to the directory containing the temporary package.json file
		err = os.Chdir(tmpFile.Name())
		Expect(err).ToNot(HaveOccurred())

		// Execute the toggle
		_, err = toggle.Execute()
		Expect(err).ToNot(HaveOccurred())
	})

	it("installs 'serve' npm dependency if Vite is detected", func() {
		// Create a temporary package.json file with Vite
		tmpFile, err := ioutil.TempFile("", "package.json")
		Expect(err).ToNot(HaveOccurred())
		defer os.Remove(tmpFile.Name())

		err = ioutil.WriteFile(tmpFile.Name(), []byte(`{"name": "test", "dependencies": {"vite": "^2.0.0"}}`), 0644)
		Expect(err).ToNot(HaveOccurred())

		// Set the current directory to the directory containing the temporary package.json file
		err = os.Chdir(tmpFile.Name())
		Expect(err).ToNot(HaveOccurred())

		// Execute the toggle
		_, err = toggle.Execute()
		Expect(err).ToNot(HaveOccurred())

		// Verify that the 'serve' npm dependency is installed in node_modules
		_, err = os.Stat("node_modules/serve")
		Expect(err).ToNot(HaveOccurred(), "'serve' npm dependency not found in node_modules directory")
	})

	it("does not install 'serve' npm dependency if Vite is not detected", func() {
		// Create a temporary package.json file without Vite
		tmpFile, err := ioutil.TempFile("", "package.json")
		Expect(err).ToNot(HaveOccurred())
		defer os.Remove(tmpFile.Name())

		err = ioutil.WriteFile(tmpFile.Name(), []byte(`{"name": "test", "dependencies": {"another_dependency": "^1.0.0"}}`), 0644)
		Expect(err).ToNot(HaveOccurred())

		// Set the current directory to the directory containing the temporary package.json file
		err = os.Chdir(tmpFile.Name())
		Expect(err).ToNot(HaveOccurred())

		// Execute the toggle
		_, err = toggle.Execute()
		Expect(err).ToNot(HaveOccurred())

		// Verify that the 'serve' npm dependency is not installed in node_modules
		_, err = os.Stat("node_modules/serve")
		Expect(os.IsNotExist(err)).To(BeTrue(), "'serve' npm dependency found in node_modules directory")
	})
}
