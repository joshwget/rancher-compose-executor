package config

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/docker/docker/pkg/urlutil"
	"github.com/docker/libcompose/utils"
	composeYaml "github.com/docker/libcompose/yaml"
	"github.com/rancher/go-rancher/v2"
	"github.com/rancher/rancher-compose-executor/template"
	"gopkg.in/yaml.v2"
)

var (
	noMerge = []string{
		"links",
		"volumes_from",
	}
)

// CreateRawConfig unmarshals contents to config and creates config based on version
func CreateRawConfig(contents []byte) (*RawConfig, error) {
	var rawConfig RawConfig
	if err := yaml.Unmarshal(contents, &rawConfig); err != nil {
		return nil, err
	}

	if rawConfig.Version != "2" {
		var baseRawServices RawServiceMap
		if err := yaml.Unmarshal(contents, &baseRawServices); err != nil {
			return nil, err
		}
		if _, ok := baseRawServices[".catalog"]; ok {
			delete(baseRawServices, ".catalog")
		}
		rawConfig.Services = baseRawServices
	}

	if rawConfig.Volumes == nil {
		rawConfig.Volumes = make(map[string]interface{})
	}
	if rawConfig.Networks == nil {
		rawConfig.Networks = make(map[string]interface{})
	}

	return &rawConfig, nil
}

// Merge merges a compose file into an existing set of service configs
func Merge(existingServices *ServiceConfigs, environmentLookup EnvironmentLookup, resourceLookup ResourceLookup, file string, contents []byte) (*Config, error) {
	var err error
	contents, err = template.Apply(contents, environmentLookup.Variables())
	if err != nil {
		return nil, err
	}

	rawConfig, err := CreateRawConfig(contents)
	if err != nil {
		return nil, err
	}
	baseRawServices := rawConfig.Services

	if err := InterpolateRawServiceMap(&baseRawServices, environmentLookup); err != nil {
		return nil, err
	}

	for k, v := range rawConfig.Volumes {
		if err := Interpolate(k, &v, environmentLookup); err != nil {
			return nil, err
		}
		rawConfig.Volumes[k] = v
	}

	for k, v := range rawConfig.Networks {
		if err := Interpolate(k, &v, environmentLookup); err != nil {
			return nil, err
		}
		rawConfig.Networks[k] = v
	}

	baseRawServices, err = PreprocessServiceMap(baseRawServices)
	if err != nil {
		return nil, err
	}

	var serviceConfigs map[string]*ServiceConfig
	if rawConfig.Version == "2" {
		var err error
		serviceConfigs, err = MergeServicesV2(existingServices, environmentLookup, resourceLookup, file, baseRawServices)
		if err != nil {
			return nil, err
		}
	} else {
		serviceConfigsV1, err := MergeServicesV1(existingServices, environmentLookup, resourceLookup, file, baseRawServices)
		if err != nil {
			return nil, err
		}
		serviceConfigs, err = ConvertServices(serviceConfigsV1)
		if err != nil {
			return nil, err
		}
	}

	adjustValues(serviceConfigs)

	var volumes map[string]*VolumeConfig
	var networks map[string]*NetworkConfig
	var hosts map[string]*client.Host
	if err := utils.Convert(rawConfig.Volumes, &volumes); err != nil {
		return nil, err
	}
	for i, volume := range volumes {
		if volume == nil {
			volumes[i] = &VolumeConfig{}
		}
	}
	if err := utils.Convert(rawConfig.Networks, &networks); err != nil {
		return nil, err
	}
	if err := utils.Convert(rawConfig.Hosts, &hosts); err != nil {
		return nil, err
	}

	return &Config{
		Services: serviceConfigs,
		Volumes:  volumes,
		Networks: networks,
		Hosts:    hosts,
	}, nil
}

func InterpolateRawServiceMap(baseRawServices *RawServiceMap, environmentLookup EnvironmentLookup) error {
	for k, v := range *baseRawServices {
		for k2, v2 := range v {
			if err := Interpolate(k2, &v2, environmentLookup); err != nil {
				return err
			}
			(*baseRawServices)[k][k2] = v2
		}
	}
	return nil
}

func adjustValues(configs map[string]*ServiceConfig) {
	// yaml parser turns "no" into "false" but that is not valid for a restart policy
	for _, v := range configs {
		if v.Restart == "false" {
			v.Restart = "no"
		}
	}
}

func readEnvFile(resourceLookup ResourceLookup, inFile string, serviceData RawService) (RawService, error) {
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

func mergeConfig(baseService, serviceData RawService) RawService {
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
