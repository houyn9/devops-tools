package clusterCmd

import (
	"github.com/spf13/cobra"
)

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "cluster commands",
}
var fileinfo string

func ClusterCmd() *cobra.Command {
	return clusterCmd
}
func init() {
	clusterCmd.AddCommand(getStorageClassCmd)
	getStorageClassCmd.Flags().StringVarP(&fileinfo, "file", "f", "", "file path")
	clusterCmd.AddCommand(getPVCmd)
	getPVCmd.Flags().StringVarP(&fileinfo, "file", "f", "", "file path")
	//clusterCmd.AddCommand(cleanStorageCmd)
}
