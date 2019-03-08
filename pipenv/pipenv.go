package pipenv

import (
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

		if output, err := n.runner.RunWithOutput("python", layer.Root, "-m", "pip", "install", "pipenv", "--find-links="+layer.Root); err != nil {
			return errors.Wrap(err, "python failed to run pip : "+string(output))
		} else {
			n.context.Logger.Info(string(output))
		}

		return n.runner.Run("pipenv", layer.Root, "lock", "--requirements")
	}, layers.Build, layers.Cache)
}
