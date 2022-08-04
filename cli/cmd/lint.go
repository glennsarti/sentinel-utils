package cmd

import (
	"fmt"
	"os"

	slint "github.com/glennsarti/sentinel-lint/lint"
	"github.com/glennsarti/sentinel-parser/features"
	defaultfs "github.com/glennsarti/sentinel-utils/lib/filesystem/os"
	"github.com/glennsarti/sentinel-utils/lib/linting"
	parsing "github.com/glennsarti/sentinel-utils/lib/parsing/default"
	cwalker "github.com/glennsarti/sentinel-utils/lib/walkers/sentinel_config"
	"github.com/spf13/cobra"
)

var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Lint one or more sentinel files",
	Long:  `Searches for Sentinel configuration and policy files to lint. It requires the primary configuration file (sentinel.hcl) to be in the root of the directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdUi := NewCommandUi(cmd)

		wd, err := os.Getwd()
		if err != nil {
			cmdUi.Error(err.Error())
			os.Exit(1)
		}

		exitCode := 0

		fsys := defaultfs.NewOSFileSystem(wd)
		pf := parsing.NewDefaultParsingFactory(fsys)
		walker := cwalker.NewSentinelConfigWalker(fsys, wd, sentinelVersion, pf)
		if walker == nil {
			cmdUi.Error("Failed to create walker")
			os.Exit(1)
		}

		err = linting.Lint(walker, pf, func(lintFile slint.File, issues slint.Issues) {
			cmdUi.OutputLintIssues(lintFile, issues, fsys)
			if len(issues) > 0 {
				exitCode = 1
			}
		})
		if err != nil {
			cmdUi.Error(err.Error())
			os.Exit(1)
		}
		os.Exit(exitCode)
	},
}

func init() {
	rootCmd.AddCommand(lintCmd)

	lintCmd.Flags().StringVarP(&sentinelVersion, "sentinel-version", "s",
		features.LatestSentinelVersion,
		fmt.Sprintf("The Sentinel version to use when linting. Default is the latest version (%s)", features.SentinelVersions[0]),
	)

	lintCmd.Flags().StringVarP(&usePath, "path", "p",
		"",
		"The path to search for files to lint. Default is the current working directory",
	)
}
