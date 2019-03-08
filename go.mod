module github.com/cloudfoundry/pipenv-cnb

go 1.12

require (
	github.com/buildpack/libbuildpack v1.11.0
	github.com/cloudfoundry/cnb-tools v0.0.0
	github.com/cloudfoundry/dagger v0.0.0-20190219165033-d9b29e00cb50
	github.com/cloudfoundry/libcfbuildpack v1.47.0
	github.com/golang/mock v1.2.0
	github.com/onsi/gomega v1.4.3
	github.com/pkg/errors v0.8.1
	github.com/sclevine/spec v1.2.0
)

replace github.com/cloudfoundry/cnb-tools => /Users/pivotal/workspace/cnb-tools
