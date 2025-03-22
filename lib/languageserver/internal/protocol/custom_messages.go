package protocol

// Version Notification ( --> Client )
// Version Request ( Client --> Server )
const SentinelVersionCommand = "sentinel/version"

type SentinelVersionParams struct {
	SentinelVersion   string   `json:"sentinelVersion"`
	AvailableVersions []string `json:"availableVersions"`
}

// Set Version Request ( Client --> Server )
const SetSentinelVersionCommand = "sentinel/setVersion"

type SentinelSetVersionRequest struct {
	Version string `json:"version"`
}
type SetSentinelVersionResponse SentinelVersionParams
