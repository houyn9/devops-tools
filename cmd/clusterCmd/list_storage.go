package clusterCmd

import (
	"devops_tools/internal/api"
	"devops_tools/internal/cluster"
	"github.com/spf13/cobra"
	"log"
)

var getStorageClassCmd = &cobra.Command{
	Use:   "get-sc",
	Short: "Get storageclass resource",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := api.NewClient()
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}
		err = cluster.GetStorageClassInfo(client, fileinfo)
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}
	},
}
var getPVCmd = &cobra.Command{
	Use:   "get-pv",
	Short: "Get pv resource",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := api.NewClient()
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}
		err = cluster.GetPersistentVolumeInfo(client, fileinfo)
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}
	},
}
