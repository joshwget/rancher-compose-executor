package rancher

import (
	"time"

	"github.com/rancher/go-rancher/v2"
)

func (r *RancherService) Wait(service *client.Service) error {
	return r.WaitFor(&service.Resource, service, func() string {
		return service.Transitioning
	})
}

func (r *RancherService) waitInstance(instance *client.Instance) error {
	return r.WaitFor(&instance.Resource, instance, func() string {
		return instance.Transitioning
	})
}

func (r *RancherService) WaitFor(resource *client.Resource, output interface{}, transitioning func() string) error {
	return WaitFor(r.Client(), resource, output, transitioning)
}

func (r *RancherContainer) Wait(container *client.Container) error {
	return WaitFor(r.Client(), &container.Resource, container, func() string {
		return container.Transitioning
	})
}

func WaitFor(client *client.RancherClient, resource *client.Resource, output interface{}, transitioning func() string) error {
	for {
		if transitioning() != "yes" {
			return nil
		}

		time.Sleep(150 * time.Millisecond)

		err := client.Reload(resource, output)
		if err != nil {
			return err
		}
	}
}
