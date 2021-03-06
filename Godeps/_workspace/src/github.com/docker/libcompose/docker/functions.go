package docker

import "github.com/samalba/dockerclient"

// GetContainersByFilter looks up the hosts containers with the specified filters and
// returns a list of container matching it, or an error.
func GetContainersByFilter(client dockerclient.Client, filter ...string) ([]dockerclient.Container, error) {
	filterResult := ""

	for _, value := range filter {
		if filterResult == "" {
			filterResult = value
		} else {
			filterResult = And(filterResult, value)
		}
	}

	return client.ListContainers(true, false, filterResult)
}

// GetContainerByName looks up the hosts containers with the specified name and
// returns it, or an error.
func GetContainerByName(client dockerclient.Client, name string) (*dockerclient.Container, error) {
	containers, err := client.ListContainers(true, false, NAME.Eq(name))
	if err != nil {
		return nil, err
	}

	if len(containers) == 0 {
		return nil, nil
	}

	return &containers[0], nil
}
