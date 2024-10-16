package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var dockerNetworkName = "goflow-network"
var redisContainerName = "redis-server"

func main() {
	// Define Cobra root command
	rootCmd := &cobra.Command{
		Use:   "goflow",
		Short: "Goflow CLI tool to deploy workerpool and plugins using Docker",
	}

	// Define deploy command
	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy workerpool with Redis broker and compiled plugins",
		Run: func(cmd *cobra.Command, args []string) {
			err := deploy()
			if err != nil {
				log.Fatalf("Error during deployment: %v", err)
			}
		},
	}

	// Add flags for the deploy command
	// deployCmd.Flags().StringVarP(&brokerType, "broker", "b", "redis", "Specify the broker type (default: redis)")

	// Add deploy command to root
	rootCmd.AddCommand(deployCmd)

	// Execute Cobra root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func deploy() error {
	fmt.Println("deploying...")

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("error creating Docker client: %v", err)
	}
	defer dockerClient.Close()

	fmt.Println("Creating Docker network...")
	err = createNetwork(dockerClient, dockerNetworkName)
	if err != nil {
		return err
	}

	fmt.Println("Starting Redis container...")
	err = startRedis(dockerClient)
	if err != nil {
		return err
	}

	fmt.Println("Compiling plugins...")
	err = compilePlugins(dockerClient)
	if err != nil {
		return err
	}

	// fmt.Println("Starting WorkerPool container...")
	// err = startWorkerPool(cli)
	// if err != nil {
	// 	return err
	// }

	// fmt.Println("Deployment successful!")
	// return nil

	return nil
}

func createNetwork(dockerClien *client.Client, networkName string) error {
	_, err := dockerClien.NetworkInspect(context.Background(), networkName, types.NetworkInspectOptions{})
	if err == nil {
		fmt.Println("Network already exists")
		return nil
	}

	_, err = dockerClien.NetworkCreate(context.Background(), networkName, types.NetworkCreate{})
	if err != nil {
		return fmt.Errorf("error creating network: %v", err)
	}

	fmt.Println("Network created successfully")

	return nil
}

func startRedis(dockerClient *client.Client) error {
	_, err := dockerClient.ContainerInspect(context.Background(), redisContainerName)
	if err == nil {
		fmt.Println("Redis container is already running")
		return nil
	}

	resp, err := dockerClient.ContainerCreate(
		context.Background(),
		&container.Config{
			Image: "redis:latest",
		},
		nil,
		&network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				dockerNetworkName: {},
			},
		},
		nil,
		redisContainerName,
	)

	if err != nil {
		return fmt.Errorf("error creating Redis container: %v", err)
	}

	if err := dockerClient.ContainerStart(
		context.Background(),
		resp.ID,
		container.StartOptions{},
	); err != nil {
		return fmt.Errorf("error starting Redis container: %v", err)
	}

	fmt.Println("Redis container started successfully")

	return nil
}

func compilePlugins(cli *client.Client) error {
	containerConfig := &container.Config{
		Image: "plugin-builder",     // Docker image to use
		Cmd:   []string{"handlers"}, // Command to run inside the container
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	handlersPath := fmt.Sprintf("%s/handlers", cwd)

	hostConfig := &container.HostConfig{
		Binds:      []string{fmt.Sprintf("%s:/app/handlers", handlersPath)}, // Bind mount the handlers directory
		AutoRemove: true,                                                    // Automatically remove the container when it exits
	}

	// Create the container for the plugin-builder
	resp, err := cli.ContainerCreate(
		context.Background(),
		containerConfig,
		hostConfig,
		nil,                        // No networking configuration needed here
		nil,                        // No platform-specific configuration needed
		"plugin-builder-container", // Name of the container
	)
	if err != nil {
		return fmt.Errorf("failed to create plugin-builder container: %v", err)
	}

	// Start the plugin-builder container
	if err := cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start plugin-builder container: %v", err)
	}

	// Wait for the container to finish
	statusCh, errCh := cli.ContainerWait(context.Background(), resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("error waiting for plugin-builder container: %v", err)
		}
	case <-statusCh:
	}

	// Check the exit code of the plugin-builder container
	containerInspect, err := cli.ContainerInspect(context.Background(), resp.ID)
	if err != nil {
		return fmt.Errorf("failed to inspect plugin-builder container: %v", err)
	}
	if containerInspect.State.ExitCode != 0 {
		return fmt.Errorf("plugin-builder container exited with code %d", containerInspect.State.ExitCode)
	}

	return nil
}
