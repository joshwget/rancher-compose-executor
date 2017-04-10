package lookup

import (
	"strings"

	"github.com/docker/docker/runconfig/opts"
	"github.com/rancher/rancher-compose-executor/config"
)

// Lookup variables from an environment variable file
func NewEnvFileLookup(path string, parent config.EnvironmentLookup) (*CommonLookup, error) {
	envs, err := opts.ParseEnvFile(path)
	if err != nil {
		return nil, err
	}
	variables := map[string]interface{}{}
	for _, env := range envs {
		split := strings.SplitN(env, "=", 2)
		if len(split) > 1 {
			variables[split[0]] = split[1]
		}
	}
	return &CommonLookup{
		variables: variables,
		parent:    parent,
	}, nil
}
