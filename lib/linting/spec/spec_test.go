package spec

import (
	"fmt"
	"os"
	"path"
	"slices"
	"strings"
	"testing"

	slint "github.com/glennsarti/sentinel-lint/lint"
	"github.com/glennsarti/sentinel-parser/position"
	"github.com/google/go-cmp/cmp"

	"github.com/glennsarti/sentinel-utils/lib/internal/helpers"
	"github.com/glennsarti/sentinel-utils/lib/internal/txtar_fs"
	subject "github.com/glennsarti/sentinel-utils/lib/linting"
	parsing "github.com/glennsarti/sentinel-utils/lib/parsing/default"
	cwalker "github.com/glennsarti/sentinel-utils/lib/walkers/sentinel_config"
)

func TestLibLintSpecs(t *testing.T) {
	fixturesDir := path.Join("test-fixtures")

	items, err := os.ReadDir(fixturesDir)
	if err != nil {
		t.Error(err)
		return
	}
	for _, item := range items {
		if item.IsDir() {
			t.Run(item.Name(), func(t *testing.T) {
				processTestFixturesDir(item.Name(), fixturesDir, item.Name(), t)
			})
		}
	}
}

func processTestFixturesDir(relPath, srcDir, sentinelVersion string, t *testing.T) {
	dirPath := path.Join(srcDir, relPath)

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".txtar") {
			t.Run(entry.Name(), func(t *testing.T) {
				if err := testSpecFile(entry.Name(), dirPath, sentinelVersion, t); err != nil {
					t.Error(err)
				}
			})
		}
	}
}

func testSpecFile(filename, parentPath, sentinelVersion string, t *testing.T) error {
	filePath := path.Join(parentPath, filename)

	arc, err := parseTxtarArchive(filePath)
	if err != nil {
		return err
	}

	arcfs := txtar_fs.NewTxtarFileSystem(arc.raw)
	pf := parsing.NewDefaultParsingFactory(arcfs)
	w := cwalker.NewSentinelConfigWalker(arcfs, "/", sentinelVersion, pf)
	if w == nil {
		return fmt.Errorf("Failed to create walker")
	}

	visited := make(map[string]slint.Issues, 0)

	err = subject.Lint(w, pf, func(lintFile slint.File, parsingIssues slint.Issues) {
		if val, ok := visited[lintFile.Path()]; !ok {
			visited[lintFile.Path()] = parsingIssues
		} else {
			visited[lintFile.Path()] = append(val, parsingIssues...)
		}
	})
	if err != nil {
		return err
	}

	inspectedStrings := make([]string, 0)
	for _, key := range helpers.SortedKeys(visited) {
		val := visited[key]
		inspectedStrings = append(inspectedStrings, inspectIssues(key, val)...)
	}
	slices.Sort(inspectedStrings)

	expectedString := string(arc.DiagnosticFile.Data)
	actualString := strings.Join(inspectedStrings, "\n") + "\n"
	if diff := cmp.Diff(expectedString, actualString); diff != "" {
		t.Fatal(diff)
	}

	return nil
}

func inspectIssues(filepath string, issues slint.Issues) []string {
	result := make([]string, len(issues))

	if len(issues) == 0 {
		result = append(result, fmt.Sprintf("Path:%s No issues found", filepath))
		return result
	}

	for idx, issue := range issues {
		result[idx] = fmt.Sprintf("Path:%s Issue: [%s] (%s) %s",
			filepath,
			rangeToString(issue.Range),
			issue.RuleId,
			issue.Summary,
		)
	}

	return result
}

func rangeToString(r *position.SourceRange) string {
	if r == nil {
		return "NIL"
	}
	return fmt.Sprintf("%d:%d-%d:%d", r.Start.Line, r.Start.Column, r.End.Line, r.End.Column)
}
