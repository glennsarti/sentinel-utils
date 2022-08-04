package parsing

import (
	"github.com/glennsarti/sentinel-parser/diagnostics"
	sast "github.com/glennsarti/sentinel-parser/sentinel/ast"
	scast "github.com/glennsarti/sentinel-parser/sentinel_config/ast"
	"github.com/glennsarti/sentinel-utils/lib/filesystem"
)

type Factory interface {
	ParseSentinelFile(file *filesystem.File, sentinelVersion string) (*sast.File, diagnostics.Diagnostics, error)
	ParseSentinelConfigFile(file *filesystem.File, sentinelVersion string) (*scast.File, diagnostics.Diagnostics, error)
}
