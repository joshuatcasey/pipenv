package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/dagger"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

func TestIntegration(t *testing.T) {
	spec.Run(t, "Integration", testIntegration, spec.Report(report.Terminal{}))
}

func testIntegration(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
		Expect(dagger.BuildCFLinuxFS3()).To(Succeed())
	})

	when("building a simple pipenv app", func() {
		it("builds and runs", func() {
			pipenvBPPath, err := dagger.PackageBuildpack()
			Expect(err).ToNot(HaveOccurred())

			pythonBPPath, err := dagger.GetLatestBuildpack("python-cnb")
			Expect(err).NotTo(HaveOccurred())

			//pipBPPath, err := dagger.GetLatestBuildpack("pip-cnb")
			//Expect(err).NotTo(HaveOccurred())
			pipBPPath := "/tmp/pip-cnb-83d544ccc223c057d2bf80d3"

			app, err := dagger.PackBuild(filepath.Join("testdata", "pipfile_lock"), pythonBPPath, pipenvBPPath, pipBPPath)
			Expect(err).NotTo(HaveOccurred())

			err = app.Start()
			if err != nil {
				_, err = fmt.Fprintf(os.Stderr, "App failed to start: %v\n", err)
			}

			body, _, err := app.HTTPGet("/")
			Expect(err).ToNot(HaveOccurred())
			Expect(body).To(ContainSubstring("Hello, World with pipenv!"))

			//Expect(app.Destroy()).To(Succeed())

		})
	})
}
