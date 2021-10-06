package api

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/juju/client"

	"github.com/juju/errors"

	"github.com/juju/juju/api/base"
	"github.com/juju/juju/api/modelmanager"
)

type ModelsAPI struct {
	client *client.Client
}

func NewModelsAPI(client *client.Client) *ModelsAPI {
	return &ModelsAPI{
		client: client,
	}
}

func (s *ModelsAPI) Models() ([]base.UserModel, error) {
	accountDetails, err := s.client.AccountDetails()
	if err != nil {
		return nil, errors.Trace(err)
	}

	root, err := s.client.NewAPIRoot()
	if err != nil {
		return nil, errors.Trace(err)
	}



	return modelAPI.ListModels(accountDetails.User)
}
