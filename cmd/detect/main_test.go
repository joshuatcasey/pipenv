package main

import (
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/test"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitDetect(t *testing.T) {
	spec.Run(t, "Detect", testDetect, spec.Report(report.Terminal{}))
}

func testDetect(t *testing.T, when spec.G, it spec.S) {
	var factory *test.DetectFactory

	it.Before(func() {
		RegisterTestingT(t)
		factory = test.NewDetectFactory(t)
	})

	when("there is no Pipfile", func() {
		it("should fail", func() {
			code, err := runDetect(factory.Detect)
			Expect(err).ToNot(HaveOccurred())
			Expect(code).To(Equal(detect.FailStatusCode))
		})
	})

	when("there is a Pipfile", func() {
		it.Before(func() {
			Expect(helper.WriteFile(filepath.Join(factory.Detect.Application.Root, "Pipfile"), 0666, "")).To(Succeed())
		})

		it("passes with pipenv and python for build and launch", func() {
			code, err := runDetect(factory.Detect)
			Expect(err).NotTo(HaveOccurred())
			Expect(code).To(Equal(detect.PassStatusCode))
			Expect(factory.Output).To(Equal(buildplan.BuildPlan{
				"pipenv": buildplan.Dependency{
					Metadata: buildplan.Metadata{
						"build":  true,
					},
				},
				"python": buildplan.Dependency{
					Metadata: buildplan.Metadata{
						"build":  true,
						"launch": true,
					},
				},
				"python_packages": buildplan.Dependency{
					Metadata: buildplan.Metadata{
						"build":  true,
						"launch": true,
					},
				},
			}))
		})
	})
}
