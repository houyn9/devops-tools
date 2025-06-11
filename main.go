package main

import (
	"devops_tools/cmd/clusterCmd"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:     "devops-tool",
	Short:   "devops-tool is a CLI tool",
	Version: "V1.0.0",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func Execute(rootCmd *cobra.Command) error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(clusterCmd.ClusterCmd())
}

func main() {
	err := Execute(rootCmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
