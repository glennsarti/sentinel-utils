package linting

import (
	"fmt"
	"io/fs"

	slint "github.com/glennsarti/sentinel-lint/lint"
	"github.com/glennsarti/sentinel-parser/diagnostics"
	"github.com/glennsarti/sentinel-parser/filetypes"
	"github.com/glennsarti/sentinel-utils/lib/filesystem"
	"github.com/glennsarti/sentinel-utils/lib/parsing"

	scast "github.com/glennsarti/sentinel-parser/sentinel_config/ast"
	scparser "github.com/glennsarti/sentinel-parser/sentinel_config/parser"
	cwalker "github.com/glennsarti/sentinel-utils/lib/walkers/sentinel_config"
)

type lintFileVisitor func(file *filesystem.File, lintFile slint.File, parsingIssues slint.Issues) (bool, error)

type lintFileSystemWalker interface {
	Walk(visitor lintFileVisitor) error
	FileSystem() filesystem.FS
	Root() string
}

func newLintWalker(w cwalker.Walker, pf parsing.Factory) lintFileSystemWalker {
	return &lintWalker{
		rootWalker:   w,
		parseFactory: pf,
	}
}

type lintWalker struct {
	rootWalker      cwalker.Walker
	parseFactory    parsing.Factory
	visitedPrimary  bool
	primaryFile     *filesystem.File
	primaryLintFile *slint.ConfigPrimaryFile
	primaryIssues   slint.Issues
}

func (w *lintWalker) Walk(visitor lintFileVisitor) error {
	w.visitedPrimary = false
	w.primaryLintFile = nil

	err := w.rootWalker.Walk(func(file *filesystem.File) (bool, error) {
		return w.visit(file, visitor)
	})

	if err != nil {
		return err
	}

	// Just incase we never actually visited the primary config ....
	if !w.visitedPrimary && w.primaryLintFile != nil {
		if _, err := visitor(w.primaryFile, *w.primaryLintFile, w.primaryIssues); err != nil {
			return err
		}
	}

	return nil
}

func (w *lintWalker) FileSystem() filesystem.FS {
	return w.rootWalker.FileSystem()
}

func (w *lintWalker) Root() string {
	return w.rootWalker.Root()
}

func (w *lintWalker) visit(file *filesystem.File, visitor lintFileVisitor) (bool, error) {
	if file.Content == nil {
		// Read it
		content, err := fs.ReadFile(w.FileSystem(), file.Path)
		if err != nil {
			return false, err
		}
		file.Content = &content
	}

	if file.Type == filetypes.ConfigPrimaryFileType {
		w.primaryFile = file
		w.visitedPrimary = false
		cfg, d, err := w.parseFactory.ParseSentinelConfigFile(file, w.rootWalker.SentinelVersion())
		if err != nil {
			return false, err
		}

		w.primaryLintFile = &slint.ConfigPrimaryFile{
			ConfigFile:         cfg,
			ResolvedConfigFile: scast.CloneFile(cfg),
			FilePath:           file.Path,
		}
		w.primaryIssues = diagsToIssues(d)

		return !d.HasErrors(), nil
	}

	if file.Type == filetypes.ConfigOverrideFileType {
		if w.primaryLintFile == nil || w.primaryLintFile.ResolvedConfigFile == nil {
			return false, fmt.Errorf("the override file %q has no primary file to override", file.Path)
		}
		if w.visitedPrimary {
			return false, fmt.Errorf("the override file %q is in the wrong directory", file.Path)
		}

		cfg, d, err := w.parseFactory.ParseSentinelConfigFile(file, w.rootWalker.SentinelVersion())
		if err != nil {
			return false, err
		}

		f := slint.ConfigOverrideFile{
			ConfigFile:  cfg,
			PrimaryFile: w.primaryLintFile.ResolvedConfigFile,
			FilePath:    file.Path,
		}

		if cont, err := visitor(file, f, diagsToIssues(d)); err != nil {
			return cont, err
		}

		diags := scparser.OverrideFileWith(w.primaryLintFile.ResolvedConfigFile, cfg, w.rootWalker.SentinelVersion())
		if diags.HasErrors() {
			return false, diags
		}
		return true, nil
	}

	// Visit the primary file, as all the overrides have been processed
	if !w.visitedPrimary && w.primaryLintFile != nil {
		w.visitedPrimary = true
		return visitor(w.primaryFile, *w.primaryLintFile, w.primaryIssues)
	}

	// Visit everything else
	switch file.Type {
	case filetypes.PolicyFileType:
		parsed, d, err := w.parseFactory.ParseSentinelFile(file, w.rootWalker.SentinelVersion())
		if err != nil {
			return false, err
		}

		f := slint.PolicyFile{
			File:     parsed,
			FilePath: file.Path,
		}
		if w.primaryLintFile != nil {
			f.ConfigFile = w.primaryLintFile.ResolvedConfigFile
		}
		return visitor(file, f, diagsToIssues(d))

	case filetypes.ModuleFileType:
		parsed, d, err := w.parseFactory.ParseSentinelFile(file, w.rootWalker.SentinelVersion())
		if err != nil {
			return false, err
		}

		return visitor(file, slint.ModuleFile{
			File:     parsed,
			FilePath: file.Path,
		}, diagsToIssues(d))

	case filetypes.ConfigTestFileType:
		cfg, d, err := w.parseFactory.ParseSentinelConfigFile(file, w.rootWalker.SentinelVersion())
		if err != nil {
			return false, err
		}

		return visitor(file, slint.ConfigTestFile{
			ConfigFile: cfg,
			FilePath:   file.Path,
		}, diagsToIssues(d))

	default:
		return true, fmt.Errorf("unknown file %q", file.Path)
	}
}

func diagsToIssues(diags diagnostics.Diagnostics) slint.Issues {
	list := make(slint.Issues, 0)
	for _, diag := range diags {
		if diag != nil && diag.Severity == diagnostics.Error {
			list = append(list, &slint.Issue{
				RuleId:   slint.SyntaxErrorRuleID,
				Summary:  diag.Summary,
				Detail:   diag.Detail,
				Range:    diag.Range,
				Severity: diagSeverityToIssueSeverity(diag.Severity),
			})
		}
	}
	return list
}

func diagSeverityToIssueSeverity(sev diagnostics.SeverityLevel) slint.SeverityLevel {
	switch sev {
	case diagnostics.Error:
		return slint.Error
	case diagnostics.Warning:
		return slint.Warning
	default:
		return slint.Unknown
	}
}
