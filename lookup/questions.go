package lookup

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/docker/libcompose/utils"
	"github.com/rancher/rancher-catalog-service/model"
	"github.com/rancher/rancher-compose-executor/config"
	"gopkg.in/yaml.v2"
)

// Lookup from default answers to questions
func NewQuestionLookup(file string, parent config.EnvironmentLookup) (*CommonLookup, error) {
	contents, err := ioutil.ReadFile(file)
	if os.IsNotExist(err) {
		return &CommonLookup{
			variables: map[string]interface{}{},
			parent:    parent,
		}, nil
	} else if err != nil {
		return nil, err
	}

	questions, err := ParseQuestions(contents)
	if err != nil {
		return nil, err
	}

	variables := map[string]interface{}{}
	for key, question := range questions {
		answer := ask(question)
		if answer != "" {
			variables[key] = answer
		}
	}

	return &CommonLookup{
		variables: variables,
		parent:    parent,
	}, nil
}

func ParseQuestions(contents []byte) (map[string]model.Question, error) {
	catalogConfig, err := ParseCatalogConfig(contents)
	if err != nil {
		return nil, err
	}

	questions := map[string]model.Question{}
	for _, question := range catalogConfig.Questions {
		questions[question.Variable] = question
	}

	return questions, nil
}

// TODO: this code is duplicated in catalog service
func ParseCatalogConfig(contents []byte) (*model.RancherCompose, error) {
	rawConfig, err := config.CreateRawConfig(contents)
	if err != nil {
		return nil, err
	}
	var rawCatalogConfig interface{}

	if rawConfig.Version == "2" && rawConfig.Services[".catalog"] != nil {
		rawCatalogConfig = rawConfig.Services[".catalog"]
	}

	var data map[string]interface{}
	if err := yaml.Unmarshal(contents, &data); err != nil {
		return nil, err
	}

	if data["catalog"] != nil {
		rawCatalogConfig = data["catalog"]
	} else if data[".catalog"] != nil {
		rawCatalogConfig = data[".catalog"]
	}

	if rawCatalogConfig != nil {
		var catalogConfig model.RancherCompose
		if err := utils.Convert(rawCatalogConfig, &catalogConfig); err != nil {
			return nil, err
		}

		return &catalogConfig, nil
	}

	return &model.RancherCompose{}, nil
}

func ask(question model.Question) string {
	if len(question.Description) > 0 {
		fmt.Println(question.Description)
	}
	fmt.Printf("%s %s[%s]: ", question.Label, question.Variable, question.Default)

	answer, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return ""
	}

	answer = strings.TrimSpace(answer)
	if answer == "" {
		answer = question.Default
	}

	return answer
}
