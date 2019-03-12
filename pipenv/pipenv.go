package pipenv

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/cloudfoundry/libcfbuildpack/runner"
	"github.com/pkg/errors"
)

const (
	Layer               = "pipenv"
	PythonLayer         = "python"
	PythonPackagesLayer = "python_packages"
)

type Contributor struct {
	context build.Build
	runner  runner.Runner
}

func NewContributor(context build.Build, runner runner.Runner) (Contributor, bool, error) {
	_, willContribute := context.BuildPlan[Layer]
	if !willContribute {
		return Contributor{}, false, nil
	}

	contributor := Contributor{context: context, runner: runner}

	return contributor, true, nil
}

func (n Contributor) Contribute() error {
	deps, err := n.context.Buildpack.Dependencies()
	if err != nil {
		return err
	}

	dep, err := deps.Best(Layer, "*", n.context.Stack)
	if err != nil {
		return err
	}

	layer := n.context.Layers.DependencyLayer(dep)

	return layer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
		layer.Logger.SubsequentLine("Expanding to %s", layer.Root)
		if err := helper.ExtractTarGz(artifact, layer.Root, 0); err != nil {
			return err
		}

		if err := n.runner.Run("python", layer.Root, "-m", "pip", "install", "pipenv", "--find-links="+layer.Root); err != nil {
			return err
		}

		cmd := exec.Command("pipenv", "lock", "--requirements")
		cmd.Dir = n.context.Application.Root
		cmd.Env = append(os.Environ(), "VIRTUALENV_NEVER_DOWNLOAD=true")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return errors.Wrap(err, "something wrong "+string(output))
		}

		outputString := string(output)
		n.context.Logger.Info("%s\n", outputString)
		// Remove output due to virtualenv
		if strings.Contains(outputString, "virtualenv") {
			reqs := strings.SplitN(outputString, "\n", 2)
			if len(reqs) > 0 {
				outputString = reqs[1]
			}
		}

		if err = ioutil.WriteFile(filepath.Join(n.context.Application.Root, "requirements.txt"), []byte(outputString), 0644); err != nil {
			return err
		}

		return nil
	}, layers.Build, layers.Cache)
}
