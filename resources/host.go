package resources

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher/v3"
	"github.com/rancher/rancher-compose-executor/project"
	"github.com/rancher/rancher-compose-executor/project/options"
)

func HostsCreate(p *project.Project) (project.ResourceSet, error) {
	hosts := make([]*Host, 0, len(p.Config.Hosts))
	for name, config := range p.Config.Hosts {
		count := config.Count
		if count == 0 {
			count = 1
		}
		hosts = append(hosts, &Host{
			name:       name,
			project:    p,
			hostConfig: keysToCamelCase(config.Dynamic).(map[string]interface{}),
			count:      count,
			template:   config.Template,
		})
	}
	return &Hosts{
		hosts: hosts,
	}, nil
}

// Convert map keys from underscore separated to camel case
func keysToCamelCase(item interface{}) interface{} {
	switch typedDatas := item.(type) {

	case map[string]interface{}:
		newMap := make(map[string]interface{})

		for key, value := range typedDatas {
			newMap[toCamelCase(key)] = keysToCamelCase(value)
		}
		return newMap

	case map[interface{}]interface{}:
		newMap := make(map[string]interface{})

		for key, value := range typedDatas {
			stringKey := key.(string)
			newMap[toCamelCase(stringKey)] = keysToCamelCase(value)
		}
		return newMap

	case []interface{}:
		newArray := make([]interface{}, 0)

		for _, value := range typedDatas {
			newArray = append(newArray, keysToCamelCase(value))
		}
		return newArray

	default:
		return item
	}
}

func toCamelCase(s string) string {
	var buffer bytes.Buffer
	for i, c := range s {
		if i > 0 && s[i-1] == '_' {
			buffer.WriteString(strings.ToUpper(string(c)))
		} else {
			buffer.WriteRune(c)
		}
	}
	return strings.Replace(buffer.String(), "_", "", -1)
}

type Hosts struct {
	hosts   []*Host
}

func (h *Hosts) Initialize(ctx context.Context, _ options.Options) error {
	for _, host := range h.hosts {
		if err := host.EnsureItExists(ctx); err != nil {
			return err
		}
	}
	return nil
}

type Host struct {
	project    *project.Project
	name       string
	hostConfig map[string]interface{}
	count      int
	template   string
}

func (h *Host) EnsureItExists(ctx context.Context) error {
	if h.count == 0 {
		return nil
	}

	existingHosts, err := h.project.Client.Host.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"stackId": h.project.Stack.Id,
		},
	})
	if err != nil {
		return err
	}

	existingNames := map[string]bool{}
	for _, existingHost := range existingHosts.Data {
		existingNames[existingHost.Name] = true
	}

	var hostsToCreate []map[string]interface{}
	for i := 1; i < h.count+1; i++ {
		name := fmt.Sprintf("%s-%s-%d", h.project.Stack.Name, h.name, i)
		if _, ok := existingNames[name]; !ok {
			hostConfig, err := h.createHostConfig(name)
			if err != nil {
				return err
			}
			hostsToCreate = append(hostsToCreate, hostConfig)
		}
	}

	for _, host := range hostsToCreate {
		log.Infof("Creating host %s", host["name"])
		if err = h.project.Client.Create("host", host, &client.Host{}); err != nil {
			return err
		}
	}

	return nil
}

func (h *Host) createHostConfig(name string) (map[string]interface{}, error) {
	hostConfig := map[string]interface{}{}

	for k, v := range h.hostConfig {
		hostConfig[k] = v
	}

	hostConfig["name"] = name
	hostConfig["hostname"] = name
	hostConfig["stackId"] = h.project.Stack.Id

	if h.template != "" {
		existingHostTemplates, err := h.project.Client.HostTemplate.List(&client.ListOpts{
			Filters: map[string]interface{}{
				"name": h.template,
			},
		})
		if err != nil {
			return nil, err
		}

		if len(existingHostTemplates.Data) == 0 {
			return nil, fmt.Errorf("Failed to find host template %s", h.template)
		}

		hostConfig["hostTemplateId"] = existingHostTemplates.Data[0].Id
	}

	return hostConfig, nil
}
