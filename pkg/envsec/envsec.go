package envsec

import (
	"io"
)

type Envsec struct {
	APIHost    string
	Auth       AuthConfig
	IsDev      bool
	Stderr     io.Writer
	WorkingDir string
}

type AuthConfig struct {
	Issuer   string
	ClientID string
	// TODO AUdiences and Scopes
}
