package cmd

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/glennsarti/sentinel-utils/lib/languageserver"
	"github.com/glennsarti/sentinel-utils/lib/languageserver/transports"
)

var lspCmd = &cobra.Command{
	Use:   "language-server",
	Short: "Start the Sentinel Language Server",
	Long:  "Start the Sentinel Language Server",
	Run: func(cmd *cobra.Command, args []string) {
		cmdUi := NewCommandUi(cmd)

		var t transports.Transport

		// Setup logging
		var logger *log.Logger
		if lspVerboseLogging {
			logger = log.New(newLogWriter(cmdUi, lspStdioBinding), "sentinel-lsp ", log.LstdFlags)
		} else {
			logger = log.New(io.Discard, "", log.LstdFlags)
		}

		if lspTcpBinding != "" {
			t = transports.NewTCPTransport(lspTcpBinding, logger)
		} else if lspStdioBinding {
			t = transports.NewSTDIOTransport(logger)
		} else {
			cmdUi.Error("The Language Server requires a transport to be selected.")
			os.Exit(1)
		}

		ctx := context.Background()
		srv := languageserver.NewLangServer(logger, ctx)

		if err := srv.StartAndWait(t); err != nil {
			cmdUi.Error(err.Error())
			os.Exit(1)
		}

		os.Exit(0)
	},
}

var lspTcpBinding string
var lspStdioBinding bool
var lspVerboseLogging bool

func init() {
	lspCmd.Flags().StringVarP(&lspTcpBinding, "tcp", "t",
		"",
		"The Language Server will use the TCP transport on the specified port and host. 'host:port'",
	)
	lspCmd.Flags().BoolVarP(&lspStdioBinding, "stdio", "s",
		false,
		"The Language Server will use the STDIO transport",
	)
	lspCmd.Flags().BoolVarP(&lspVerboseLogging, "verbose", "v",
		false,
		"Enable verbose logging",
	)
	rootCmd.AddCommand(lspCmd)
}
