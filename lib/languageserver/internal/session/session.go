package session

import (
	"github.com/glennsarti/sentinel-parser/features"
	lsp "github.com/glennsarti/sentinel-utils/lib/languageserver/internal/protocol"
)

type Session struct {
	ClientCapabilities *lsp.ClientCapabilities
	Ready              bool
	RootDir            *string
	SentinelVersion    *string
}

func NewSession() *Session {
	ver := features.SentinelVersions[0]
	rootDir := ""
	return &Session{
		ClientCapabilities: &lsp.ClientCapabilities{},
		Ready:              false,
		RootDir:            &rootDir,
		SentinelVersion:    &ver,
	}
}
