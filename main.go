package main

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/rancher/go-rancher/v3"
	"github.com/rancher/rancher-compose-executor/project"
	"github.com/rancher/rancher-compose-executor/project/options"
	_ "github.com/rancher/rancher-compose-executor/resources"
)

const (
	cattleURLEnv       = "RANCHER_URL"
	cattleAccessKeyEnv = "RANCHER_ACCESS_KEY"
	cattleSecretKeyEnv = "RANCHER_SECRET_KEY"
)

func main() {
	//executor.Main()

	b, err := ioutil.ReadFile("compose.yml")
	if err != nil {
		panic(err)
	}

	rancherClient, err := createRancherClient()
	if err != nil {
		panic(err)
	}

	p := project.NewProject("s", rancherClient)
	if err := p.Load(map[string]interface{}{
		"compose.yml": b,
	}, map[string]string{}); err != nil {
		panic(err)
	}

	if err := p.Create(context.Background(), options.Options{}); err != nil {
		panic(err)
	}
}

func createRancherClient() (*client.RancherClient, error) {
	url, err := client.NormalizeUrl(os.Getenv(cattleURLEnv))
	if err != nil {
		return nil, err
	}
	return client.NewRancherClient(&client.ClientOpts{
		Url:       url,
		AccessKey: os.Getenv(cattleAccessKeyEnv),
		SecretKey: os.Getenv(cattleSecretKeyEnv),
	})
}
