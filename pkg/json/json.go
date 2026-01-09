package json

import (
	"github.com/goccy/go-json"
)

var (
	Marshal       = json.Marshal
	Unmarshal     = json.Unmarshal
	MarshalIndent = json.MarshalIndent
	NewDecoder    = json.NewDecoder
	NewEncoder    = json.NewEncoder
)

// RawMessage represents a raw JSON message, compatible with json.RawMessage
type RawMessage = json.RawMessage
