package spec

import (
	"io"
	"os"

	"golang.org/x/tools/txtar"
)

const archiveDiagOutput = "diagOut.txt"

type parsedArchive struct {
	DiagnosticFile txtar.File
	raw            *txtar.Archive
}

func parseTxtarArchive(filePath string) (*parsedArchive, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close() //nolint:errcheck

	contents, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	f.Close() //nolint:errcheck

	arc := &parsedArchive{}
	arc.raw = txtar.Parse(contents)

	for _, f := range arc.raw.Files {
		switch f.Name {
		case archiveDiagOutput:
			arc.DiagnosticFile = f
		}
	}

	return arc, nil
}
