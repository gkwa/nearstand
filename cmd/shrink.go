package cmd

import (
	"github.com/gkwa/nearstand/core"
	"github.com/spf13/cobra"
)

var reshrink bool

var shrinkCmd = &cobra.Command{
	Use:   "shrink file_or_directory [file_or_directory...]",
	Short: "Shrink image file(s)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := LoggerFrom(cmd.Context())
		shrinker := core.NewImageMagickShrinker(core.DefaultFileManager{})
		processor := core.NewImageProcessor(shrinker, logger, cmd.OutOrStdout())
		return processor.ProcessFilesAndDirectories(args, reshrink)
	},
}

func init() {
	rootCmd.AddCommand(shrinkCmd)
	shrinkCmd.Flags().BoolVar(&reshrink, "reshrink", false, "Allow reshrinking of already shrunk images")
}
