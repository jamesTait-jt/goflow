package cmd

import (
	"github.com/jamesTait-jt/goflow/cmd/cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/cli/internal/service"
	"github.com/jamesTait-jt/goflow/pkg/log"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy workerpool with Redis broker and compiled plugins",
	RunE: func(_ *cobra.Command, _ []string) error {
		conf, err := config.Get()
		if err != nil {
			return err
		}

		logger := log.NewConsoleLogger()

		deploymentManager, err := service.NewDeploymentManager(conf, logger)
		if err != nil {
			return err
		}

		deploymentService := service.NewDeploymentService(deploymentManager)

		return deploymentService.Deploy()
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
