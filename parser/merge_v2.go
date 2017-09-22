package parser

import (
	"fmt"
	"path"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/rancher-compose-executor/config"
	"github.com/rancher/rancher-compose-executor/lookup"
	"github.com/rancher/rancher-compose-executor/utils"
)

// mergeServicesV2 merges a v2 compose file into an existing set of service configs
func mergeServicesV2(resourceLookup lookup.ResourceLookup, file string, datas config.RawServiceMap) (map[string]*config.ServiceConfig, error) {
	if err := validateV2(datas); err != nil {
		return nil, err
	}

	for name, data := range datas {
		var err error
		datas[name], err = parseV2(resourceLookup, file, data)
		if err != nil {
			logrus.Errorf("Failed to parse service %s: %v", name, err)
			return nil, err
		}
	}

	serviceConfigs := make(map[string]*config.ServiceConfig)
	if err := utils.Convert(datas, &serviceConfigs); err != nil {
		return nil, err
	}

	return serviceConfigs, nil
}

func parseV2(resourceLookup lookup.ResourceLookup, inFile string, serviceData config.RawService) (config.RawService, error) {
	serviceData, err := readEnvFile(resourceLookup, inFile, serviceData)
	if err != nil {
		return nil, err
	}
	serviceData = resolveContextV2(inFile, serviceData)
	return serviceData, nil
}

func resolveContextV2(inFile string, serviceData config.RawService) config.RawService {
	if _, ok := serviceData["build"]; !ok {
		return serviceData
	}

	var build map[interface{}]interface{}
	if buildAsString, ok := serviceData["build"].(string); ok {
		build = map[interface{}]interface{}{
			"context": buildAsString,
		}
	} else {
		build = serviceData["build"].(map[interface{}]interface{})
	}
	context := asString(build["context"])
	if context == "" {
		return serviceData
	}

	if IsValidRemote(context) {
		return serviceData
	}

	current := path.Dir(inFile)

	if context == "." {
		context = current
	} else {
		current = path.Join(current, context)
	}

	build["context"] = current

	return serviceData
}
