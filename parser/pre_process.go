package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fatih/structs"
	"github.com/rancher/rancher-compose-executor/config"
)

func preProcessServiceMap(serviceMap config.RawServiceMap) (config.RawServiceMap, error) {
	newServiceMap := make(config.RawServiceMap)

	for k, v := range serviceMap {
		newServiceMap[k] = make(config.RawService)
		for k2, v2 := range v {
			if k2 == "environment" || k2 == "labels" {
				newServiceMap[k][k2] = preProcess(v2, true)
			} else {
				newServiceMap[k][k2] = preProcess(v2, false)
			}
		}
	}

	return newServiceMap, nil
}

func preProcess(item interface{}, replaceTypes bool) interface{} {
	switch typedDatas := item.(type) {

	case map[interface{}]interface{}:
		newMap := make(map[interface{}]interface{})

		for key, value := range typedDatas {
			newMap[key] = preProcess(value, replaceTypes)
		}
		return newMap

	case []interface{}:
		// newArray := make([]interface{}, 0) will cause golint to complain
		var newArray []interface{}
		newArray = make([]interface{}, 0)

		for _, value := range typedDatas {
			newArray = append(newArray, preProcess(value, replaceTypes))
		}
		return newArray

	default:
		if replaceTypes && item != nil {
			return fmt.Sprint(item)
		}
		return item
	}
}

func TryConvertStringsToInts(serviceMap config.RawServiceMap, fields map[string]bool) (config.RawServiceMap, error) {
	newServiceMap := make(config.RawServiceMap)

	for k, v := range serviceMap {
		newServiceMap[k] = make(config.RawService)

		for k2, v2 := range v {
			if _, ok := fields[k2]; ok {
				newServiceMap[k][k2] = tryConvertStringsToInts(v2, true)
			} else {
				newServiceMap[k][k2] = tryConvertStringsToInts(v2, false)
			}

		}
	}

	return newServiceMap, nil
}

func tryConvertStringsToInts(item interface{}, replaceTypes bool) interface{} {
	switch typedDatas := item.(type) {

	case map[interface{}]interface{}:
		newMap := make(map[interface{}]interface{})

		for key, value := range typedDatas {
			newMap[key] = tryConvertStringsToInts(value, replaceTypes)
		}
		return newMap

	case []interface{}:
		// newArray := make([]interface{}, 0) will cause golint to complain
		var newArray []interface{}
		newArray = make([]interface{}, 0)

		for _, value := range typedDatas {
			newArray = append(newArray, tryConvertStringsToInts(value, replaceTypes))
		}
		return newArray

	case string:
		lineAsInteger, err := strconv.Atoi(typedDatas)

		if replaceTypes && err == nil {
			return lineAsInteger
		}

		return item
	default:
		return item
	}
}

func getRancherConfigObjects() map[string]bool {
	rancherConfig := structs.New(config.RancherConfig{})
	fields := map[string]bool{}
	for _, field := range rancherConfig.Fields() {
		kind := field.Kind().String()
		if kind == "struct" || kind == "ptr" || kind == "slice" {
			split := strings.Split(field.Tag("yaml"), ",")
			fields[split[0]] = true
		}
	}
	return fields
}
