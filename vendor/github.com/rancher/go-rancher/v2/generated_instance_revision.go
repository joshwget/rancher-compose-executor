package client

const (
	INSTANCE_REVISION_TYPE = "instanceRevision"
)

type InstanceRevision struct {
	Resource

	AccountId string `json:"accountId,omitempty" yaml:"account_id,omitempty"`

	Config *Container `json:"config,omitempty" yaml:"config,omitempty"`

	Created string `json:"created,omitempty" yaml:"created,omitempty"`

	Data map[string]interface{} `json:"data,omitempty" yaml:"data,omitempty"`

	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	InstanceId string `json:"instanceId,omitempty" yaml:"instance_id,omitempty"`

	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	RemoveTime string `json:"removeTime,omitempty" yaml:"remove_time,omitempty"`

	Removed string `json:"removed,omitempty" yaml:"removed,omitempty"`

	State string `json:"state,omitempty" yaml:"state,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type InstanceRevisionCollection struct {
	Collection
	Data   []InstanceRevision `json:"data,omitempty"`
	client *InstanceRevisionClient
}

type InstanceRevisionClient struct {
	rancherClient *RancherClient
}

type InstanceRevisionOperations interface {
	List(opts *ListOpts) (*InstanceRevisionCollection, error)
	Create(opts *InstanceRevision) (*InstanceRevision, error)
	Update(existing *InstanceRevision, updates interface{}) (*InstanceRevision, error)
	ById(id string) (*InstanceRevision, error)
	Delete(container *InstanceRevision) error
}

func newInstanceRevisionClient(rancherClient *RancherClient) *InstanceRevisionClient {
	return &InstanceRevisionClient{
		rancherClient: rancherClient,
	}
}

func (c *InstanceRevisionClient) Create(container *InstanceRevision) (*InstanceRevision, error) {
	resp := &InstanceRevision{}
	err := c.rancherClient.doCreate(INSTANCE_REVISION_TYPE, container, resp)
	return resp, err
}

func (c *InstanceRevisionClient) Update(existing *InstanceRevision, updates interface{}) (*InstanceRevision, error) {
	resp := &InstanceRevision{}
	err := c.rancherClient.doUpdate(INSTANCE_REVISION_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *InstanceRevisionClient) List(opts *ListOpts) (*InstanceRevisionCollection, error) {
	resp := &InstanceRevisionCollection{}
	err := c.rancherClient.doList(INSTANCE_REVISION_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *InstanceRevisionCollection) Next() (*InstanceRevisionCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &InstanceRevisionCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *InstanceRevisionClient) ById(id string) (*InstanceRevision, error) {
	resp := &InstanceRevision{}
	err := c.rancherClient.doById(INSTANCE_REVISION_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *InstanceRevisionClient) Delete(container *InstanceRevision) error {
	return c.rancherClient.doResourceDelete(INSTANCE_REVISION_TYPE, &container.Resource)
}
