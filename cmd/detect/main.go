package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/pipenv-cnb/pipenv"
)

func main() {
	context, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create default detect context: %s", err)
		os.Exit(100)
	}

	code, err := runDetect(context)
	if err != nil {
		context.Logger.Info(err.Error())
	}

	os.Exit(code)
}

func runDetect(context detect.Detect) (int, error) {
	exists, err := helper.FileExists(filepath.Join(context.Application.Root, "Pipfile"))
	if err != nil {
		return detect.FailStatusCode, err
	} else if !exists {
		context.Logger.Info("no Pipfile found")
		return detect.FailStatusCode, nil
	}

	if exists, err := helper.FileExists(filepath.Join(context.Application.Root, "requirements.txt")); err != nil {
		return detect.FailStatusCode, err
	} else if exists {
		context.Logger.Error("found Pipfile + requirements.txt")
		return detect.FailStatusCode, nil
	}

	return context.Pass(buildplan.BuildPlan{
		pipenv.PythonLayer: buildplan.Dependency{
			Metadata: buildplan.Metadata{"build": true, "launch": true},
		},
		pipenv.Layer: buildplan.Dependency{
			Metadata: buildplan.Metadata{"build": true},
		},
		pipenv.PythonPackagesLayer: buildplan.Dependency{
			Metadata: buildplan.Metadata{"build": true, "launch": true},
		},
	})
}
