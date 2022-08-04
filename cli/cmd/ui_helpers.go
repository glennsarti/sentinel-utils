package cmd

import (
	"github.com/spf13/cobra"

	"github.com/glennsarti/sentinel-utils/cli/ui"
)

func NewCommandUi(cmd *cobra.Command) ui.Ui {
	return &ui.BasicUi{
		Writer:      cmd.OutOrStdout(),
		ErrorWriter: cmd.ErrOrStderr(),
	}
}
