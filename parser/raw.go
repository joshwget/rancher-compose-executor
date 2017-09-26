package parser

import (
	"strings"

	"github.com/fatih/structs"
	"github.com/rancher/go-rancher/v3"
	"github.com/rancher/rancher-compose-executor/config"
	"github.com/rancher/rancher-compose-executor/parser/kubernetes"
	"gopkg.in/yaml.v2"
)

func createRawConfig(contents []byte) (*config.RawConfig, error) {
	resources, err := kubernetes.GetResources(contents)
	if err != nil {
		return nil, err
	}
	if len(resources) > 0 {
		kubernetesResources := map[string]interface{}{}
		for _, resource := range resources {
			kubernetesResources[resource.CombinedName] = resource.ResourceContents
		}
		return &config.RawConfig{
			KubernetesResources: kubernetesResources,
		}, nil
	}

	var rawConfig config.RawConfig
	if err := yaml.Unmarshal(contents, &rawConfig); err != nil {
		return nil, err
	}

	if rawConfig.Version != "2" {
		var baseRawServices config.RawServiceMap
		if err := yaml.Unmarshal(contents, &baseRawServices); err != nil {
			return nil, err
		}
		if _, ok := baseRawServices[".catalog"]; ok {
			delete(baseRawServices, ".catalog")
		}
		rawConfig.Services = baseRawServices
	}

	if rawConfig.Services == nil {
		rawConfig.Services = make(config.RawServiceMap)
	}
	if rawConfig.Volumes == nil {
		rawConfig.Volumes = make(map[string]interface{})
	}
	if rawConfig.Hosts == nil {
		rawConfig.Hosts = make(map[string]interface{})
	}
	if rawConfig.Secrets == nil {
		rawConfig.Secrets = make(map[string]interface{})
	}

	// Merge other service types into primary service map
	for name, baseRawLoadBalancer := range rawConfig.LoadBalancers {
		rawConfig.Services[name] = baseRawLoadBalancer
		transferFields(baseRawLoadBalancer, rawConfig.Services[name], "lb_config", config.LBConfig{})
	}
	// TODO: validation will throw errors for fields directly under service
	for name, baseRawStorageDriver := range rawConfig.StorageDrivers {
		rawConfig.Services[name] = baseRawStorageDriver
		transferFields(baseRawStorageDriver, rawConfig.Services[name], "storage_driver", client.StorageDriver{})
	}
	// TODO: validation will throw errors for fields directly under service
	for name, baseRawNetworkDriver := range rawConfig.NetworkDrivers {
		rawConfig.Services[name] = baseRawNetworkDriver
		transferFields(baseRawNetworkDriver, rawConfig.Services[name], "network_driver", client.NetworkDriver{})
	}
	for name, baseRawVirtualMachine := range rawConfig.VirtualMachines {
		rawConfig.Services[name] = baseRawVirtualMachine
	}
	for name, baseRawExternalService := range rawConfig.ExternalServices {
		rawConfig.Services[name] = baseRawExternalService
		rawConfig.Services[name]["image"] = "rancher/external-service"
	}
	// TODO: container aliases
	for name, baseRawAlias := range rawConfig.Aliases {
		if serviceAliases, ok := baseRawAlias["services"]; ok {
			rawConfig.Services[name] = baseRawAlias
			rawConfig.Services[name]["image"] = "rancher/dns-service"
			rawConfig.Services[name]["links"] = serviceAliases
			delete(rawConfig.Services[name], "services")
		}
	}

	return &rawConfig, nil
}

func transferFields(from, to config.RawService, prefixField string, instance interface{}) {
	s := structs.New(instance)
	for _, f := range s.Fields() {
		field := strings.SplitN(f.Tag("yaml"), ",", 2)[0]
		if fieldValue, ok := from[field]; ok {
			if _, ok = to[prefixField]; !ok {
				to[prefixField] = map[interface{}]interface{}{}
			}
			to[prefixField].(map[interface{}]interface{})[field] = fieldValue
		}
	}
}

func convertVersion(baseRawServices config.RawServiceMap) (config.RawServiceMap, error) {

	for _, rawService := range baseRawServices {
		newServiceMap := make(config.RawServiceMap)
	}
	/*v2Services := make(map[string]*config.ServiceConfig)
	replacementFields := make(map[string]*config.ServiceConfig)

	for name, service := range v1Services {
		replacementFields[name] = &config.ServiceConfig{
			Build: yaml.Build{
				Context:    service.Build,
				Dockerfile: service.Dockerfile,
			},
			Logging: config.Log{
				Driver:  service.LogDriver,
				Options: service.LogOpt,
			},
			NetworkMode: service.Net,
		}

		v1Services[name].Build = ""
		v1Services[name].Dockerfile = ""
		v1Services[name].LogDriver = ""
		v1Services[name].LogOpt = nil
		v1Services[name].Net = ""
	}

	if err := utils.Convert(v1Services, &v2Services); err != nil {
		return nil, err
	}

	for name := range v2Services {
		v2Services[name].Build = replacementFields[name].Build
		v2Services[name].Logging = replacementFields[name].Logging
		v2Services[name].NetworkMode = replacementFields[name].NetworkMode
	}

	return v2Services, nil*/
}
