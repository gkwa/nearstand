package cmd

import (
	"fmt"
	
	"github.com/gkwa/nearstand/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of nearstand",
	Long:  `All software has versions. This is nearstand's`,
	Run: func(cmd *cobra.Command, args []string) {
		buildInfo := version.GetBuildInfo()
		fmt.Println(buildInfo)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
