package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/docker/docker/pkg/urlutil"
	"github.com/rancher/go-rancher/catalog"
	"github.com/rancher/go-rancher/v3"
	"github.com/rancher/rancher-compose-executor/config"
	"github.com/rancher/rancher-compose-executor/lookup"
	"github.com/rancher/rancher-compose-executor/parser/interpolation"
	"github.com/rancher/rancher-compose-executor/template"
	"github.com/rancher/rancher-compose-executor/utils"
	composeYaml "github.com/rancher/rancher-compose-executor/yaml"
)

var (
	noMerge = []string{
		"links",
		"volumes_from",
	}
)

// Merge merges a compose file into an existing set of service configs
func Merge(existingServices map[string]*config.ServiceConfig, vars map[string]string, resourceLookup lookup.ResourceLookup, templateVersion *catalog.TemplateVersion, cluster *client.Cluster, file string, contents []byte) (*config.Config, error) {
	var err error
	contents, err = template.Apply(contents, templateVersion, cluster, vars)
	if err != nil {
		return nil, err
	}

	rawConfig, err := createRawConfig(contents)
	if err != nil {
		return nil, err
	}

	baseRawServices := rawConfig.Services
	baseRawContainers := rawConfig.Containers

	// TODO: just interpolate at the map level earlier
	if err := interpolateRawServiceMap(&baseRawServices, vars); err != nil {
		return nil, err
	}
	if err := interpolateRawServiceMap(&baseRawContainers, vars); err != nil {
		return nil, err
	}

	for k, v := range rawConfig.Volumes {
		if err := interpolation.Interpolate(k, &v, vars); err != nil {
			return nil, err
		}
		rawConfig.Volumes[k] = v
	}

	for k, v := range rawConfig.Networks {
		if err := interpolation.Interpolate(k, &v, vars); err != nil {
			return nil, err
		}
		rawConfig.Networks[k] = v
	}

	baseRawServices = preProcessServiceMap(baseRawServices)
	baseRawContainers = preProcessServiceMap(baseRawContainers)

	var serviceConfigs map[string]*config.ServiceConfig
	if rawConfig.Version == "2" {
		var err error
		serviceConfigs, err = mergeServicesV2(resourceLookup, file, baseRawServices)
		if err != nil {
			return nil, err
		}
	} else {
		serviceConfigsV1, err := mergeServicesV1(resourceLookup, file, baseRawServices)
		if err != nil {
			return nil, err
		}
		serviceConfigs, err = convertServices(serviceConfigsV1)
		if err != nil {
			return nil, err
		}
	}

	for name, serviceConfig := range serviceConfigs {
		if existingServiceConfig, ok := existingServices[name]; ok {
			var rawService config.RawService
			if err := utils.Convert(serviceConfig, &rawService); err != nil {
				return nil, err
			}
			var rawExistingService config.RawService
			if err := utils.Convert(existingServiceConfig, &rawExistingService); err != nil {
				return nil, err
			}

			rawService = mergeConfig(rawExistingService, rawService)
			if err := utils.Convert(rawService, &serviceConfig); err != nil {
				return nil, err
			}
		}
	}

	var containerConfigs map[string]*config.ServiceConfig
	if rawConfig.Version == "2" {
		var err error
		containerConfigs, err = mergeServicesV2(resourceLookup, file, baseRawContainers)
		if err != nil {
			return nil, err
		}
	}

	adjustValues(serviceConfigs)
	adjustValues(containerConfigs)

	var dependencies map[string]*config.DependencyConfig
	var volumes map[string]*config.VolumeConfig
	var networks map[string]*config.NetworkConfig
	var secrets map[string]*config.SecretConfig
	var hosts map[string]*config.HostConfig
	if err := utils.Convert(rawConfig.Dependencies, &dependencies); err != nil {
		return nil, err
	}
	if err := utils.Convert(rawConfig.Volumes, &volumes); err != nil {
		return nil, err
	}
	for i, volume := range volumes {
		if volume == nil {
			volumes[i] = &config.VolumeConfig{}
		}
	}
	if err := utils.Convert(rawConfig.Networks, &networks); err != nil {
		return nil, err
	}
	if err := utils.Convert(rawConfig.Hosts, &hosts); err != nil {
		return nil, err
	}
	if err := utils.Convert(rawConfig.Secrets, &secrets); err != nil {
		return nil, err
	}

	return &config.Config{
		Services:            serviceConfigs,
		Containers:          containerConfigs,
		Dependencies:        dependencies,
		Volumes:             volumes,
		Networks:            networks,
		Secrets:             secrets,
		Hosts:               hosts,
		KubernetesResources: rawConfig.KubernetesResources,
	}, nil
}

func interpolateRawServiceMap(baseRawServices *config.RawServiceMap, vars map[string]string) error {
	for k, v := range *baseRawServices {
		for k2, v2 := range v {
			if err := interpolation.Interpolate(k2, &v2, vars); err != nil {
				return err
			}
			(*baseRawServices)[k][k2] = v2
		}
	}
	return nil
}

func adjustValues(configs map[string]*config.ServiceConfig) {
	// yaml parser turns "no" into "false" but that is not valid for a restart policy
	for _, v := range configs {
		if v.Restart == "false" {
			v.Restart = "no"
		}
	}
}

func readEnvFile(resourceLookup lookup.ResourceLookup, inFile string, serviceData config.RawService) (config.RawService, error) {
	if _, ok := serviceData["env_file"]; !ok {
		return serviceData, nil
	}

	var envFiles composeYaml.Stringorslice

	if err := utils.Convert(serviceData["env_file"], &envFiles); err != nil {
		return nil, err
	}

	if len(envFiles) == 0 {
		return serviceData, nil
	}

	if resourceLookup == nil {
		return nil, fmt.Errorf("Can not use env_file in file %s no mechanism provided to load files", inFile)
	}

	var vars composeYaml.MaporEqualSlice

	if _, ok := serviceData["environment"]; ok {
		if err := utils.Convert(serviceData["environment"], &vars); err != nil {
			return nil, err
		}
	}

	for i := len(envFiles) - 1; i >= 0; i-- {
		envFile := envFiles[i]
		content, _, err := resourceLookup.Lookup(envFile, inFile)
		if err != nil {
			return nil, err
		}

		if err != nil {
			return nil, err
		}

		scanner := bufio.NewScanner(bytes.NewBuffer(content))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())

			if len(line) > 0 && !strings.HasPrefix(line, "#") {
				key := strings.SplitAfter(line, "=")[0]

				found := false
				for _, v := range vars {
					if strings.HasPrefix(v, key) {
						found = true
						break
					}
				}

				if !found {
					vars = append(vars, line)
				}
			}
		}

		if scanner.Err() != nil {
			return nil, scanner.Err()
		}
	}

	serviceData["environment"] = vars

	delete(serviceData, "env_file")

	return serviceData, nil
}

func mergeConfig(baseService, serviceData config.RawService) config.RawService {
	for k, v := range serviceData {
		// Image and build are mutually exclusive in merge
		if k == "image" {
			delete(baseService, "build")
		} else if k == "build" {
			delete(baseService, "image")
		}
		existing, ok := baseService[k]
		if ok {
			baseService[k] = merge(existing, v)
		} else {
			baseService[k] = v
		}
	}

	return baseService
}

// IsValidRemote checks if the specified string is a valid remote (for builds)
func IsValidRemote(remote string) bool {
	return urlutil.IsGitURL(remote) || urlutil.IsURL(remote)
}
