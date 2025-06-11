package clusterCmd

import (
	"devops_tools/internal/api"
	"devops_tools/internal/cluster"
	"github.com/spf13/cobra"
	"log"
)

var cleanStorageCmd = &cobra.Command{
	Use:   "clean-storage",
	Short: "clean unused StorageClass and PV resource",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := api.NewClient()
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}
		if err := cluster.CleanStorageResources(client); err != nil {
			log.Printf("cleanup failed: %v", err)
		}
	},
}
