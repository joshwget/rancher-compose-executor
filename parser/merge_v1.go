package parser

import (
	"fmt"
	"path"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/rancher-compose-executor/config"
	"github.com/rancher/rancher-compose-executor/lookup"
	"github.com/rancher/rancher-compose-executor/utils"
)

// mergeServicesV1 merges a v1 compose file into an existing set of service configs
func mergeServicesV1(resourceLookup lookup.ResourceLookup, file string, datas config.RawServiceMap) (map[string]*config.ServiceConfigV1, error) {
	if err := validate(datas); err != nil {
		return nil, err
	}

	for name, data := range datas {
		var err error
		datas[name], err = parseV1(resourceLookup, file, data)
		if err != nil {
			logrus.Errorf("Failed to parse service %s: %v", name, err)
			return nil, err
		}
	}

	serviceConfigs := make(map[string]*config.ServiceConfigV1)
	if err := utils.Convert(datas, &serviceConfigs); err != nil {
		return nil, err
	}

	return serviceConfigs, nil
}

func parseV1(resourceLookup lookup.ResourceLookup, inFile string, serviceData config.RawService) (config.RawService, error) {
	serviceData, err := readEnvFile(resourceLookup, inFile, serviceData)
	if err != nil {
		return nil, err
	}
	serviceData = resolveContextV1(inFile, serviceData)
	return serviceData, nil
}

func resolveContextV1(inFile string, serviceData config.RawService) config.RawService {
	context := asString(serviceData["build"])
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

	serviceData["build"] = current

	return serviceData
}
