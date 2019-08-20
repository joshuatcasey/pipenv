package pipenv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/logger"

	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/cloudfoundry/libcfbuildpack/runner"
	"github.com/pkg/errors"
)

const (
	Dependency               = "pipenv"
	PythonLayer              = "python"
	PythonPackagesLayer      = "python_packages"
	RequirementsLayer        = "requirements"
	PythonPackagesCacheLayer = "python_packages_cache"
	Pipfile                  = "Pipfile"
	LockFile                 = "Pipfile.lock"
	RequirementsFile         = "requirements.txt"
)

type PipfileLock struct {
	Meta struct {
		Requires struct {
			Version string `json:"python_version"`
		} `json:"requires"`
		Sources []struct {
			URL string
		}
	} `json:"_meta"`
	Default map[string]struct {
		Version string
	}
}

type Contributor struct {
	requirementsMetadata logger.Identifiable
	context              build.Build
	runner               runner.Runner
	pipenvLayer          layers.DependencyLayer
	requirementsLayer    layers.Layer
	packagesCacheLayer   layers.Layer
	buildContribution    bool
}

type Metadata struct {
	Name string
	Hash string
}

func (m Metadata) Identity() (name string, version string) {
	return m.Name, m.Hash
}

func NewContributor(context build.Build, runner runner.Runner) (Contributor, bool, error) {
	plan, willContribute, err := context.Plans.GetShallowMerged(Dependency)
	if err != nil || !willContribute {
		return Contributor{}, false, err
	}

	contributor := Contributor{
		context:              context,
		runner:               runner,
		requirementsLayer:    context.Layers.Layer(RequirementsLayer),
		requirementsMetadata: Metadata{RequirementsLayer, strconv.FormatInt(time.Now().UnixNano(), 16)},
		// TODO: something cache
	}

	contributor.buildContribution, _ = plan.Metadata["build"].(bool)

	return contributor, true, nil
}

func GetPythonVersionFromPipfileLock(fullPath string) (string, error) {
	file, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return "", err
	}
	pipfileLock := PipfileLock{}
	err = json.Unmarshal(file, &pipfileLock)
	if err != nil {
		return "", err
	}

	return pipfileLock.Meta.Requires.Version, nil

}

func (n Contributor) ContributePipenv() error {
	deps, err := n.context.Buildpack.Dependencies()
	if err != nil {
		return err
	}

	dep, err := deps.Best(Dependency, "*", n.context.Stack)
	if err != nil {
		return err
	}

	layer := n.context.Layers.DependencyLayer(dep)

	return layer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
		layer.Logger.SubsequentLine("Expanding to %s", layer.Root)
		if err := helper.ExtractTarGz(artifact, layer.Root, 0); err != nil {
			return errors.Wrap(err, "problem extracting")
		}

		if err := n.runner.Run("python", layer.Root, "-m", "pip", "install", "pipenv", "--find-links="+layer.Root); err != nil {
			return errors.Wrap(err, "problem installing pipenv")
		}

		return nil
	}, layers.Build, layers.Launch, layers.Cache)
}

func (n Contributor) ContributeRequirementsTxt() error {
	n.context.Logger.Info("Generating requirements.txt")

	lockPath := filepath.Join(n.context.Application.Root, "Pipfile.lock")
	present, err := helper.FileExists(lockPath)
	if err != nil {
		return err
	}
	var requirements []byte
	if present {
		n.context.Logger.Info("Generating requirements.txt from Pipfile.lock")
		requirements, err = pipfileLockToRequirementsTxt(lockPath)
		if err != nil {
			return errors.Wrap(err, "problem generating requirements.txt from Pipfile.lock")
		}
	} else {
		if err := n.runner.Run("pipenv", n.context.Application.Root, "lock", "--requirements"); err != nil {
			return errors.Wrap(err, "problem generating initial Pipfile.lock")
		}

		// When we run this a second time, we get the output we care about without extraneous logging
		requirements, err = n.runner.RunWithOutput("pipenv", n.context.Application.Root, "lock", "--requirements")
		if err != nil {
			return errors.Wrap(err, "problem with reading requirements from Pipfile.lock")
		}
	}

	return n.requirementsLayer.Contribute(n.requirementsMetadata, func(layer layers.Layer) error {
		layer.Touch()
		layer.Logger.SubsequentLine("Writing requirements.txt to %s", layer.Root)

		if err = helper.WriteFile(filepath.Join(layer.Root, "requirements.txt"), 0644, "%s", requirements); err != nil {
			return errors.Wrap(err, "problem writing requirements")
		}

		return nil
	}, layers.Build)
}

func pipfileLockToRequirementsTxt(pipfileLockPath string) ([]byte, error) {
	lockContents, err := ioutil.ReadFile(pipfileLockPath)
	if err != nil {
		return []byte{}, err
	}

	lockFile := PipfileLock{}
	err = json.Unmarshal(lockContents, &lockFile)
	if err != nil {
		return []byte{}, err
	}

	buf := &bytes.Buffer{}

	for _, source := range lockFile.Meta.Sources {
		fmt.Fprintf(buf, "-i %s\n", source.URL)
	}

	for pkg, obj := range lockFile.Default {
		fmt.Fprintf(buf, "%s%s\n", pkg, obj.Version)
	}

	return buf.Bytes(), nil
}
