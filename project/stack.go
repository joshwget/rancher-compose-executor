package project

import (
	"github.com/rancher/rancher-compose-executor/config"
	"golang.org/x/net/context"
)

type Stacks interface {
	Initialize(ctx context.Context) error
}

type StacksFactory interface {
	Create(projectName string, stackConfigs map[string]*config.StackConfig) (Stacks, error)
}
