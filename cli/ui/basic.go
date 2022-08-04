package ui

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	slint "github.com/glennsarti/sentinel-lint/lint"
	"github.com/glennsarti/sentinel-utils/lib/filesystem"
)

var _ Ui = &BasicUi{}

// BasicUI is a simple implementation of a user-facing stream UI (like a terminal)
type BasicUi struct {
	Writer      io.Writer
	ErrorWriter io.Writer
}

func (u *BasicUi) Error(message string) {
	w := u.Writer
	if u.ErrorWriter != nil {
		w = u.ErrorWriter
	}

	fmt.Fprint(w, message)
	fmt.Fprint(w, "\n")
}

func (u *BasicUi) Info(message string) {
	u.Output(message)
}

func (u *BasicUi) Output(message string) {
	fmt.Fprint(u.Writer, message)
	fmt.Fprint(u.Writer, "\n")
}

func (u *BasicUi) Warn(message string) {
	u.Error(message)
}

func (u *BasicUi) OutputLintIssues(lintFile slint.File, issues slint.Issues, fsys filesystem.FS) {
	if len(issues) == 0 {
		u.Output(fmt.Sprintf("✅ %s: No issues\n", lintFile.Path()))
	}

	for _, i := range issues {
		prefix := "❓ Unknown"
		switch i.Severity {
		case slint.Error:
			prefix = "❌ Error"
		case slint.Information:
			prefix = "ℹ  Info"
		case slint.Warning:
			prefix = "⚠  Warning"
		}

		u.Output(fmt.Sprintf("%s: %s (%s)\n\n  on %s line %d:",
			prefix,
			i.Summary,
			i.RuleId,
			lintFile.Path(),
			i.Range.Start.Line+1,
		))

		content, _ := fsys.ReadFile(lintFile.Path())

		for idx, l := range u.getLines(content, i.Range.Start.Line, i.Range.End.Line) {
			line := l

			// TODO This only copes with single lines
			line = u.underline(line, i.Range.Start.Column, i.Range.End.Column)

			u.Output(fmt.Sprintf("  %d: %s",
				idx+i.Range.Start.Line+1,
				line,
			))
		}

		if i.Detail != "" {
			u.Output("\n  " + i.Detail)
		}

		if i.Related != nil && len(*i.Related) > 0 {
			for _, related := range *i.Related {
				u.Output(fmt.Sprintf(
					"\n  %s", related.Summary))
				u.Output(fmt.Sprintf(
					"    on %s line %d:",
					related.Range.Filename,
					related.Range.Start.Line+1,
				))

				for idx, l := range u.getLines(content, related.Range.Start.Line, related.Range.End.Line) {
					line := l

					// TODO This only copes with single lines
					line = u.underline(line, related.Range.Start.Column, related.Range.End.Column)

					u.Output(fmt.Sprintf("    %d: %s",
						idx+related.Range.Start.Line+1,
						line,
					))
				}
			}
		}

		u.Output("\n")
	}
}

// Base 0 line numbers
func (u *BasicUi) getLines(content []byte, startLine, endLine int) []string {
	bytesReader := bytes.NewReader(content)
	bufReader := bufio.NewReader(bytesReader)

	lines := make([]string, endLine-startLine+1)
	for i := 0; i < startLine; i++ {
		if _, _, err := bufReader.ReadLine(); err != nil {
			return []string{}
		}
	}

	for i := 0; i <= endLine-startLine; i++ {
		line, _, _ := bufReader.ReadLine()
		lines[i] = string(line)
	}

	return lines
}

// Base 0 from,to columns
func (u *BasicUi) underline(line string, from, to int) string {
	return line[:from] +
		"\x1B[4m" +
		line[from:to] +
		"\x1B[24m" +
		line[to:]
}
