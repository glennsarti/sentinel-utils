package parsing

import (
	"io/fs"

	"github.com/glennsarti/sentinel-parser/diagnostics"
	sast "github.com/glennsarti/sentinel-parser/sentinel/ast"
	sparser "github.com/glennsarti/sentinel-parser/sentinel/parser"
	scast "github.com/glennsarti/sentinel-parser/sentinel_config/ast"
	scparser "github.com/glennsarti/sentinel-parser/sentinel_config/parser"
	"github.com/glennsarti/sentinel-utils/lib/filesystem"

	"github.com/glennsarti/sentinel-utils/lib/parsing"
)

var _ parsing.Factory = defaultParsingFactory{}

func NewDefaultParsingFactory(fsys filesystem.FS) parsing.Factory {
	return defaultParsingFactory{
		fsys: fsys,
	}
}

type defaultParsingFactory struct {
	fsys filesystem.FS
}

func (dpf defaultParsingFactory) ParseSentinelFile(file *filesystem.File, sentinelVersion string) (*sast.File, diagnostics.Diagnostics, error) {
	if file.Content == nil {
		// Read it
		content, err := fs.ReadFile(dpf.fsys, file.Path)
		if err != nil {
			return nil, nil, err
		}
		file.Content = &content
	}

	// if file.Content == nil {
	// 	panic(fmt.Sprintf("NO CONTENT FOR %s", file.Path))
	// }

	// Parse it
	parsed, _, diags, err := sparser.ParseFile(sentinelVersion, file.Path, *file.Content)
	return parsed, diags, err
}

func (dpf defaultParsingFactory) ParseSentinelConfigFile(file *filesystem.File, sentinelVersion string) (*scast.File, diagnostics.Diagnostics, error) {
	if file.Content == nil {
		// Read it
		content, err := fs.ReadFile(dpf.fsys, file.Path)
		if err != nil {
			return nil, nil, err
		}
		file.Content = &content
	}

	// if file.Content == nil {
	// 	panic(fmt.Sprintf("NO CONTENT FOR %s", file.Path))
	// }

	// // Read it
	// content, err := fs.ReadFile(dpf.fsys, file.Path)
	// if err != nil {
	// 	return nil, nil, err
	// }

	// Parse it
	p, err := scparser.New(sentinelVersion)
	if err != nil {
		return nil, nil, err
	}

	cfg, diags := p.ParseFile(file.Path, *file.Content)
	return cfg, diags, nil
}
