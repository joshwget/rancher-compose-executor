package client

const (
	CONVERT_TO_SERVICE_INPUT_TYPE = "convertToServiceInput"
)

type ConvertToServiceInput struct {
	Resource

	Name string `json:"name,omitempty" yaml:"name,omitempty"`
}

type ConvertToServiceInputCollection struct {
	Collection
	Data   []ConvertToServiceInput `json:"data,omitempty"`
	client *ConvertToServiceInputClient
}

type ConvertToServiceInputClient struct {
	rancherClient *RancherClient
}

type ConvertToServiceInputOperations interface {
	List(opts *ListOpts) (*ConvertToServiceInputCollection, error)
	Create(opts *ConvertToServiceInput) (*ConvertToServiceInput, error)
	Update(existing *ConvertToServiceInput, updates interface{}) (*ConvertToServiceInput, error)
	ById(id string) (*ConvertToServiceInput, error)
	Delete(container *ConvertToServiceInput) error
}

func newConvertToServiceInputClient(rancherClient *RancherClient) *ConvertToServiceInputClient {
	return &ConvertToServiceInputClient{
		rancherClient: rancherClient,
	}
}

func (c *ConvertToServiceInputClient) Create(container *ConvertToServiceInput) (*ConvertToServiceInput, error) {
	resp := &ConvertToServiceInput{}
	err := c.rancherClient.doCreate(CONVERT_TO_SERVICE_INPUT_TYPE, container, resp)
	return resp, err
}

func (c *ConvertToServiceInputClient) Update(existing *ConvertToServiceInput, updates interface{}) (*ConvertToServiceInput, error) {
	resp := &ConvertToServiceInput{}
	err := c.rancherClient.doUpdate(CONVERT_TO_SERVICE_INPUT_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ConvertToServiceInputClient) List(opts *ListOpts) (*ConvertToServiceInputCollection, error) {
	resp := &ConvertToServiceInputCollection{}
	err := c.rancherClient.doList(CONVERT_TO_SERVICE_INPUT_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *ConvertToServiceInputCollection) Next() (*ConvertToServiceInputCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ConvertToServiceInputCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ConvertToServiceInputClient) ById(id string) (*ConvertToServiceInput, error) {
	resp := &ConvertToServiceInput{}
	err := c.rancherClient.doById(CONVERT_TO_SERVICE_INPUT_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *ConvertToServiceInputClient) Delete(container *ConvertToServiceInput) error {
	return c.rancherClient.doResourceDelete(CONVERT_TO_SERVICE_INPUT_TYPE, &container.Resource)
}
