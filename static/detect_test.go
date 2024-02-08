package static

import (
	"os"
	"testing"

	"github.com/buildpacks/libcnb"
	"github.com/initializ-buildpacks/static-buildpack/static"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx    libcnb.DetectContext
		detect staticbuildpack.Detect
	)

	context("Vite is present in package.json", func() {
		it.Before(func() {
			// Prepare the test environment, simulate Vite being present in package.json
			Expect(os.WriteFile("package.json", []byte(`{"dependencies":{"vite":"^2.0.0"}}`), 0644)).To(Succeed())
		})

		it.After(func() {
			// Clean up the test environment after the test
			Expect(os.Remove("package.json")).To(Succeed())
		})

		it("passes detection with Vite detected", func() {
			Expect(detect.Detect(ctx)).To(Equal(libcnb.DetectResult{
				Pass: true,
				Plans: []libcnb.BuildPlan{
					{
						Provides: []libcnb.BuildPlanProvide{
							{Name: "serve"},
						},
					},
				},
			}))
		})
	})

	context("Vite is not present in package.json", func() {
		it.Before(func() {
			// Prepare the test environment, simulate Vite not being present in package.json
			Expect(os.WriteFile("package.json", []byte(`{"dependencies":{"react":"^17.0.2"}}`), 0644)).To(Succeed())
		})

		it.After(func() {
			// Clean up the test environment after the test
			Expect(os.Remove("package.json")).To(Succeed())
		})

		it("passes detection without Vite detected", func() {
			Expect(detect.Detect(ctx)).To(Equal(libcnb.DetectResult{}))
		})
	})
}
