package ui

import (
	slint "github.com/glennsarti/sentinel-lint/lint"
	"github.com/glennsarti/sentinel-utils/lib/filesystem"
)

type Ui interface {
	// Normal string based output
	Output(string)
	Info(string)
	Error(string)
	Warn(string)

	// Outputs Linting issues
	OutputLintIssues(lintFile slint.File, issues slint.Issues, fsys filesystem.FS)
}
