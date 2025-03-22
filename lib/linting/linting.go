package linting

import (
	slint "github.com/glennsarti/sentinel-lint/lint"
	"github.com/glennsarti/sentinel-lint/rules"
	"github.com/glennsarti/sentinel-lint/runner"

	"github.com/glennsarti/sentinel-utils/lib/filesystem"
	"github.com/glennsarti/sentinel-utils/lib/parsing"
	cwalker "github.com/glennsarti/sentinel-utils/lib/walkers/sentinel_config"
)

type LintIssueYielder func(lintFile slint.File, parsingIssues slint.Issues)

func Lint(walker cwalker.Walker, pf parsing.Factory, yielder LintIssueYielder) error {
	lintRuleSet := rules.NewDefaultRuleSet() // TODO: Parameterise this stuff
	cfg := slint.Config{
		SentinelVersion: walker.SentinelVersion(),
	}

	visitor := func(file *filesystem.File, lintFile slint.File, parsingIssues slint.Issues) (bool, error) {
		allIssues := make(slint.Issues, 0)
		allIssues = append(allIssues, parsingIssues...)

		if parsingIssues.HasErrors() {
			yielder(lintFile, allIssues)
			return true, nil // TODO: Should this be false?
		}

		r, _ := runner.NewRunner(cfg, lintRuleSet, lintFile)
		if issues, err := r.Run(); err != nil {
			return false, err
		} else {
			allIssues = append(allIssues, issues...)
		}

		yielder(lintFile, allIssues)

		return true, nil
	}

	lw := newLintWalker(walker, yielder, pf)
	err := lw.Walk(visitor)

	return err
}
