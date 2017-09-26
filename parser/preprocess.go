package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/fatih/structs"
	"github.com/rancher/rancher-compose-executor/config"
)

func preProcessServiceMap(serviceMap config.RawServiceMap) config.RawServiceMap {
	rancherFields := getRancherConfigObjects()
	newServiceMap := make(config.RawServiceMap)
	for k, v := range serviceMap {
		newServiceMap[k] = make(config.RawService)
		for k2, v2 := range v {
			if k2 == "environment" || k2 == "labels" {
				v2 = preProcess(v2, true)
			} else {
				v2 = preProcess(v2, false)
			}
			if _, ok := rancherFields[k2]; ok {
				newServiceMap[k][k2] = tryConvertStringsToInts(v2, true)
			} else {
				newServiceMap[k][k2] = tryConvertStringsToInts(v2, false)
			}
		}
	}
	return newServiceMap
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
		var newArray []interface{}
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

func tryConvertStringsToInts(item interface{}, replaceTypes bool) interface{} {
	switch typedDatas := item.(type) {
	case map[interface{}]interface{}:
		newMap := make(map[interface{}]interface{})
		for key, value := range typedDatas {
			newMap[key] = tryConvertStringsToInts(value, replaceTypes)
		}
		return newMap

	case []interface{}:
		var newArray []interface{}
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
