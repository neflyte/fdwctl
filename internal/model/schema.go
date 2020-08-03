package model

type Schema struct {
	ServerName     string
	LocalSchema    string
	RemoteSchema   string
	ImportENUMs    bool
	ENUMConnection string
}
