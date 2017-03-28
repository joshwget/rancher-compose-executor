package lookup

import (
	"fmt"

	"github.com/rancher/rancher-compose-executor/config"
)

type MapEnvLookup struct {
	Env map[string]interface{}
}

func (m *MapEnvLookup) Lookup(key string, config *config.ServiceConfig) []string {
	if v, ok := m.Env[key]; ok {
		return []string{fmt.Sprintf("%s=%v", key, v)}
	}
	return []string{}
}

func (m *MapEnvLookup) Variables() map[string]interface{} {
	return m.Env
}
