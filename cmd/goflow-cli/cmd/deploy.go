package cmd

import (
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/run"
	"github.com/spf13/cobra"
)

var (
	handlersPath string
	local        bool
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy workerpool with Redis broker and compiled plugins",
	RunE: func(_ *cobra.Command, _ []string) error {
		conf, err := config.Get()
		if err != nil {
			return err
		}

		return run.Deploy(conf)
	},
}

func init() {
	deployCmd.Flags().StringVarP(&handlersPath, "path-to-handlers", "p", "", "The full path to the location of your custom handlers (required)")
	deployCmd.Flags().BoolVarP(&local, "local", "l", false, "Whether to run the application locally (with docker containers) or connect to a kubernetes cluster")

	rootCmd.AddCommand(deployCmd)
}
