package client

const (
	SERVICE_REVISION_TYPE = "serviceRevision"
)

type ServiceRevision struct {
	Resource

	AccountId string `json:"accountId,omitempty" yaml:"account_id,omitempty"`

	Configs map[string]interface{} `json:"configs,omitempty" yaml:"configs,omitempty"`

	Created string `json:"created,omitempty" yaml:"created,omitempty"`

	Data map[string]interface{} `json:"data,omitempty" yaml:"data,omitempty"`

	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	RemoveTime string `json:"removeTime,omitempty" yaml:"remove_time,omitempty"`

	Removed string `json:"removed,omitempty" yaml:"removed,omitempty"`

	ServiceId string `json:"serviceId,omitempty" yaml:"service_id,omitempty"`

	State string `json:"state,omitempty" yaml:"state,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type ServiceRevisionCollection struct {
	Collection
	Data   []ServiceRevision `json:"data,omitempty"`
	client *ServiceRevisionClient
}

type ServiceRevisionClient struct {
	rancherClient *RancherClient
}

type ServiceRevisionOperations interface {
	List(opts *ListOpts) (*ServiceRevisionCollection, error)
	Create(opts *ServiceRevision) (*ServiceRevision, error)
	Update(existing *ServiceRevision, updates interface{}) (*ServiceRevision, error)
	ById(id string) (*ServiceRevision, error)
	Delete(container *ServiceRevision) error
}

func newServiceRevisionClient(rancherClient *RancherClient) *ServiceRevisionClient {
	return &ServiceRevisionClient{
		rancherClient: rancherClient,
	}
}

func (c *ServiceRevisionClient) Create(container *ServiceRevision) (*ServiceRevision, error) {
	resp := &ServiceRevision{}
	err := c.rancherClient.doCreate(SERVICE_REVISION_TYPE, container, resp)
	return resp, err
}

func (c *ServiceRevisionClient) Update(existing *ServiceRevision, updates interface{}) (*ServiceRevision, error) {
	resp := &ServiceRevision{}
	err := c.rancherClient.doUpdate(SERVICE_REVISION_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ServiceRevisionClient) List(opts *ListOpts) (*ServiceRevisionCollection, error) {
	resp := &ServiceRevisionCollection{}
	err := c.rancherClient.doList(SERVICE_REVISION_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *ServiceRevisionCollection) Next() (*ServiceRevisionCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ServiceRevisionCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ServiceRevisionClient) ById(id string) (*ServiceRevision, error) {
	resp := &ServiceRevision{}
	err := c.rancherClient.doById(SERVICE_REVISION_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *ServiceRevisionClient) Delete(container *ServiceRevision) error {
	return c.rancherClient.doResourceDelete(SERVICE_REVISION_TYPE, &container.Resource)
}
