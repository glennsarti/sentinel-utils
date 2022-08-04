package spec

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/glennsarti/sentinel-utils/lib/filesystem"
	"github.com/glennsarti/sentinel-utils/lib/internal/txtar_fs"

	parsing "github.com/glennsarti/sentinel-utils/lib/parsing/default"
	subject "github.com/glennsarti/sentinel-utils/lib/walkers/sentinel_config"
)

func TestLibSentinelConfigWalkerSpecs(t *testing.T) {
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
	w := subject.NewSentinelConfigWalker(arcfs, "/", sentinelVersion, pf)
	if w == nil {
		return fmt.Errorf("Failed to create walker")
	}

	visited := make([]string, 0)
	err = w.Walk(func(file *filesystem.File) (bool, error) {
		visited = append(visited, inspectFile(file))
		return true, nil
	})
	if err != nil {
		return err
	}

	expectedString := string(arc.WalkerFile.Data)
	actualString := strings.Join(visited, "\n") + "\n"
	if diff := cmp.Diff(expectedString, actualString); diff != "" {
		t.Fatal(diff)
	}

	return nil
}

func inspectFile(file *filesystem.File) string {
	return fmt.Sprintf("Path:%s FileType:%s", file.Path, file.Type)
}
