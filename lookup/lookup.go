package lookup

import (
	"fmt"

	"github.com/rancher/rancher-compose-executor/config"
	"github.com/rancher/rancher-compose-executor/utils"
)

type CommonLookup struct {
	parent    config.EnvironmentLookup
	variables map[string]interface{}
}

func NewCommonLookup(variables map[string]interface{}, parent config.EnvironmentLookup) *CommonLookup {
	return &CommonLookup{
		variables: variables,
		parent:    parent,
	}
}

func (l *CommonLookup) Lookup(key string, config *config.ServiceConfig) []string {
	variables := l.Variables()
	if v, ok := variables[key]; ok {
		return []string{fmt.Sprintf("%s=%v", key, v)}
	}
	return []string{}
}

func (l *CommonLookup) Variables() map[string]interface{} {
	if l.parent == nil {
		return l.variables
	}
	return utils.MapInterfaceUnion(l.parent.Variables(), l.variables)
}
