package rancher

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	"github.com/rancher/go-rancher/v2"
	"github.com/rancher/rancher-compose-executor/config"
	"github.com/rancher/rancher-compose-executor/project"
)

type RancherStacksFactory struct {
	Context *Context
}

func (f *RancherStacksFactory) Create(projectName string, stackConfigs map[string]*config.StackConfig) (project.Stacks, error) {
	stacks := make([]*Stack, 0, len(stackConfigs))
	for name, config := range stackConfigs {
		stacks = append(stacks, &Stack{
			context:           f.Context,
			name:              name,
			description:       config.Description,
			projectName:       projectName,
			templateId:        config.TemplateId,
			templateVersionId: config.TemplateVersionId,
		})
	}
	return &Stacks{
		stacks: stacks,
	}, nil
}

type Stacks struct {
	stacks  []*Stack
	Context *Context
}

func (s *Stacks) Initialize(ctx context.Context) error {
	for _, stack := range s.stacks {
		if err := stack.EnsureItExists(ctx); err != nil {
			return err
		}
	}
	return nil
}

type Stack struct {
	context           *Context
	name              string
	description       string
	projectName       string
	templateId        string
	templateVersionId string
	answers           map[string]interface{}
}

func (s *Stack) EnsureItExists(ctx context.Context) error {
	stackName := fmt.Sprintf("%s-%s", s.projectName, s.name)

	var id string
	if s.templateId != "" {
		id = s.templateId
	} else if s.templateVersionId != "" {
		id = s.templateVersionId
	}

	existingStacks, err := s.context.Client.Stack.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"name": stackName,
		},
	})
	if err != nil {
		return err
	}

	if len(existingStacks.Data) == 0 {
		stack, err := s.context.Client.Stack.Create(&client.Stack{
			Name:          fmt.Sprintf("%s-%s", s.projectName, s.name),
			Description:   s.description,
			ExternalId:    fmt.Sprintf("catalog://%s", id),
			Answers:       s.answers,
			StartOnCreate: true,
		})
		if err != nil {
			return err
		}
		_, err = s.context.Client.Stack.ActionUpgrade(stack, &client.StackUpgrade{
			ExternalId: fmt.Sprintf("catalog://%s", id),
			Answers:    s.answers,
		})
	} else {
		time.Sleep(time.Second * 60)
		stack := existingStacks.Data[0]
		stack.
			_, err = s.context.Client.Stack.ActionUpgrade(&stack, &client.StackUpgrade{
			ExternalId: fmt.Sprintf("catalog://%s", id),
			Answers:    s.answers,
		})
	}

	return err
}
