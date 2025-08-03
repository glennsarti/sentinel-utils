package linting

import (
	"fmt"

	slint "github.com/glennsarti/sentinel-lint/lint"
	"github.com/glennsarti/sentinel-parser/filetypes"
	"github.com/glennsarti/sentinel-parser/position"
)

var _ slint.File = unknownFile{}

func newUnknownFile(path string) slint.File {
	return unknownFile{path: path}
}

type unknownFile struct {
	path string
}

func (uf unknownFile) Type() filetypes.FileType {
	return filetypes.UnknownFileType
}

func (uf unknownFile) Path() string {
	return uf.path
}

func newFileNotExistIssue(filePath string, src *position.SourceRange) *slint.Issue {
	return &slint.Issue{
		Severity: slint.Error,
		RuleId:   "FileSystem/Error", // TODO: Should be constantised from sentinel-lint
		Summary:  "File does not exist",
		Detail:   fmt.Sprintf("File %q does not exist", filePath),
		Range:    src,
	}
}
