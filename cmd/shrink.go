package cmd

import (
	"github.com/gkwa/nearstand/core"
	"github.com/spf13/cobra"
)

var shrinkCmd = &cobra.Command{
	Use:   "shrink [file_or_directory]",
	Short: "Shrink image file(s)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := LoggerFrom(cmd.Context())
		target := args[0]

		shrinker := core.NewImageMagickShrinker(core.DefaultFileManager{})
		processor := core.NewImageProcessor(shrinker, logger)

		return processor.Process(target)
	},
}

func init() {
	rootCmd.AddCommand(shrinkCmd)
}
