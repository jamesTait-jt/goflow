package service

import (
	"errors"
	"fmt"

	"github.com/jamesTait-jt/goflow/cmd/cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/cli/internal/infrastructure/docker"
	"github.com/jamesTait-jt/goflow/cmd/cli/internal/infrastructure/k8s"
	"github.com/jamesTait-jt/goflow/pkg/log"
)

type deploymentManager interface {
	DeployNamespace() error
	DeployMessageBroker() error
	DeployGRPCServer() error
	DeployWorkerpools() error
	DestroyAll() error
}

type DeploymentService struct {
	deploymentManager deploymentManager
}

func NewDeploymentService(d deploymentManager) *DeploymentService {
	return &DeploymentService{
		deploymentManager: d,
	}
}

func (d *DeploymentService) Deploy() error {
	if err := d.deploymentManager.DeployNamespace(); err != nil {
		return fmt.Errorf("failed to deploy namespace: %w", err)
	}

	if err := d.deploymentManager.DeployMessageBroker(); err != nil {
		return fmt.Errorf("failed to deploy message broker: %w", err)
	}

	if err := d.deploymentManager.DeployGRPCServer(); err != nil {
		return fmt.Errorf("failed to deploy gRPC server: %w", err)
	}

	if err := d.deploymentManager.DeployWorkerpools(); err != nil {
		return fmt.Errorf("failed to deploy worker pools: %w", err)
	}

	return nil
}

func (d *DeploymentService) Destroy() error {
	return d.deploymentManager.DestroyAll()
}

// NewDeploymentManager constructs a deploymentManager based on the provided config.
// Assumes that the config has been validated (e.g., only one of Docker or Kubernetes is specified).
func NewDeploymentManager(conf *config.Config, logger log.Logger) (deploymentManager, error) {
	if conf.Docker != nil {
		return docker.NewDeploymentManager(conf, logger)
	}

	if conf.Kubernetes != nil {
		return k8s.NewDeploymentManager(conf, logger)
	}

	return nil, errors.New("invalid configuration: neither Docker nor Kubernetes specified")
}
