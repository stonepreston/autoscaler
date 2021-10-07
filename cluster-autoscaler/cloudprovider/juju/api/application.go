package api

import (
	"github.com/juju/errors"
	"github.com/juju/juju/api/application"
	"github.com/juju/juju/apiserver/params"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/juju/client"
)

type ApplicationAPI struct {
	client *client.Client
}

func NewApplicationAPI(client *client.Client) *ApplicationAPI {
	return &ApplicationAPI{
		client: client,
	}
}

func (s *ApplicationAPI) AddUnit(applicationName string, units int) ([]string, error) {

	root, err := s.client.NewAPIRoot()
	if err != nil {
		return nil, errors.Trace(err)
	}
	applicationAPI := application.NewClient(root)
	defer applicationAPI.Close()

	args := application.AddUnitsParams{
		ApplicationName: applicationName,
		NumUnits:        units,
	}

	result, err := applicationAPI.AddUnits(args)

	if err != nil {
		return nil, err
	}

	return result, nil

}

func (s *ApplicationAPI) ScaleApplication(applicationName string, units int) (params.ScaleApplicationResult, error) {

	root, err := s.client.NewAPIRoot()
	if err != nil {
		return params.ScaleApplicationResult{}, errors.Trace(err)
	}
	applicationAPI := application.NewClient(root)
	defer applicationAPI.Close()

	args := application.ScaleApplicationParams{
		ApplicationName: applicationName,
		Scale:           units,
		Force:           true,
	}

	result, err := applicationAPI.ScaleApplication(args)

	if err != nil {
		return params.ScaleApplicationResult{}, err
	}

	return result, nil

}
