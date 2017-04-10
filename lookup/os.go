package lookup

import (
	"os"

	"github.com/docker/libcompose/yaml"
)

// Lookup from OS environment
func NewOsEnvLookup() *CommonLookup {
	env := yaml.MaporEqualSlice(os.Environ())
	envMap := env.ToMap()
	variables := map[string]interface{}{}
	for k, v := range envMap {
		variables[k] = v
	}
	return &CommonLookup{
		variables: variables,
	}
}
