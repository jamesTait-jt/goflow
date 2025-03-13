package docker

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/jamesTait-jt/goflow/cmd/cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/cli/internal/infrastructure"
	"github.com/jamesTait-jt/goflow/pkg/log"
)

type DeploymentManager struct {
	conf   *config.Config
	logger log.Logger
	client *Docker
}

func NewDeploymentManager(conf *config.Config, logger log.Logger) (*DeploymentManager, error) {
	dockerClient, err := New()
	if err != nil {
		return nil, fmt.Errorf("error creating Docker client: %v", err)
	}

	return &DeploymentManager{
		conf:   conf,
		logger: logger,
		client: dockerClient,
	}, nil
}

func (d *DeploymentManager) DeployNamespace() error {
	return d.client.CreateNetwork(d.conf.Docker.NetworkID)
}

func (d *DeploymentManager) DeployMessageBroker() error {
	containerInfo, err := d.client.ContainerInfo(infrastructure.RedisContainerName)
	if err != nil {
		return err
	}

	if containerInfo.Running {
		fmt.Println("Redis container already started")

		return nil
	}

	if !containerInfo.Exists {
		if err = d.client.PullImage(d.conf.Redis.Image); err != nil {
			return fmt.Errorf("failed to pull redis image: %v", err)
		}

		containerPort, err := nat.NewPort("tcp", strconv.Itoa(int(infrastructure.RedisPort)))
		if err != nil {
			return err
		}

		containerInfo.ID, err = d.client.CreateContainer(
			&container.Config{
				Image: d.conf.Redis.Image,
			},
			&container.HostConfig{
				PortBindings: nat.PortMap{
					containerPort: []nat.PortBinding{
						{
							HostIP:   d.conf.Redis.IP,
							HostPort: strconv.Itoa(int(d.conf.Redis.Port)),
						},
					},
				},
			},
			d.conf.Docker.NetworkID,
			infrastructure.RedisContainerName,
		)

		if err != nil {
			return fmt.Errorf("error creating Redis container: %v", err)
		}
	}

	if err = d.client.StartContainer(containerInfo.ID); err != nil {
		return fmt.Errorf("error starting redis container: %v", err)
	}

	fmt.Println("Redis container started successfully")

	return nil
}

func (d *DeploymentManager) DeployGRPCServer() error {
	containerPort, err := nat.NewPort("tcp", strconv.Itoa(int(infrastructure.GRPCContainerPort)))
	if err != nil {
		return err
	}

	containerID, err := d.client.CreateContainer(
		&container.Config{
			Image: d.conf.GoFlowServer.Image,
			Cmd: []string{
				"--broker-type", "redis",
				"--broker-addr", fmt.Sprintf("%s:%d", infrastructure.RedisContainerName, infrastructure.RedisPort),
			},
		},
		&container.HostConfig{
			PortBindings: nat.PortMap{
				containerPort: []nat.PortBinding{
					{
						HostIP:   d.conf.GoFlowServer.IP,
						HostPort: strconv.Itoa(int(d.conf.GoFlowServer.Port)),
					},
				},
			},
		},
		d.conf.Docker.NetworkID,
		infrastructure.GRPCServerContainerName,
	)

	if err != nil {
		return fmt.Errorf("failed to create goflow container: %v", err)
	}

	if err := d.client.StartContainer(containerID); err != nil {
		return fmt.Errorf("error starting goflow container: %v", err)
	}

	fmt.Println("goflow container started successfully")

	return nil
}

func (d *DeploymentManager) DeployWorkerpools() error {
	pluginBuilderContainerID, err := d.client.CreateContainer(
		&container.Config{
			Image: d.conf.Workerpool.PluginBuilderImage,
			Cmd:   []string{"handlers"},
		},
		&container.HostConfig{
			Binds:      []string{fmt.Sprintf("%s:%s", d.conf.Workerpool.PathToHandlers, infrastructure.WorkerpoolHandlersLocation)},
			AutoRemove: true,
		},
		"",
		"",
	)

	if err != nil {
		return fmt.Errorf("failed to create plugin-builder container: %v", err)
	}

	if err = d.client.StartContainer(pluginBuilderContainerID); err != nil {
		return fmt.Errorf("failed to start plugin-builder container: %v", err)
	}

	err = d.client.WaitForContainerToFinish(pluginBuilderContainerID)
	if err != nil {
		return fmt.Errorf("failed to wait for plugin-builder to finish: %v", err)
	}

	containerPassed, err := d.client.ContainerPassed(pluginBuilderContainerID)
	if err != nil {
		return fmt.Errorf("failed to check plugin-builder exit status: %v", err)
	}

	if !containerPassed {
		return errors.New("plugin-builder container failed")
	}

	hostConfig := &container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:%s", d.conf.Workerpool.PathToHandlers, infrastructure.WorkerpoolHandlersLocation)},
	}

	containerID, err := d.client.CreateContainer(
		&container.Config{
			Image: d.conf.Workerpool.Image,
			Cmd: []string{
				"--broker-type", "redis",
				"--broker-addr", fmt.Sprintf("%s:6379", infrastructure.RedisContainerName),
				"--handlers-path", fmt.Sprintf("%s/compiled", infrastructure.WorkerpoolHandlersLocation),
			},
		},
		hostConfig,
		d.conf.Docker.NetworkID,
		infrastructure.WorkerpoolContainerName,
	)

	if err != nil {
		return fmt.Errorf("failed to create workerpool container: %v", err)
	}

	if err := d.client.StartContainer(containerID); err != nil {
		return fmt.Errorf("error starting workerpool container: %v", err)
	}

	return nil
}

func (d *DeploymentManager) DestroyAll() error {
	for _, containerID := range []string{
		infrastructure.RedisContainerName,
		infrastructure.GRPCServerContainerName,
		infrastructure.WorkerpoolContainerName,
	} {
		fmt.Printf("Destroying container '%s'\n", containerID)

		if err := d.client.DestroyContainer(containerID); err != nil {
			return fmt.Errorf("failed to destroy container '%s': %v", containerID, err)
		}
	}

	fmt.Println("Destroying Docker network...")

	if err := d.client.DestroyNetwork(d.conf.Docker.NetworkID); err != nil {
		return fmt.Errorf("failed to destroy network '%s': %v", d.conf.Docker.NetworkID, err)
	}

	fmt.Println("Done!")

	return nil
}
