package model

// DBSyncRequest db sync request
type DBSyncRequest struct {
	ProjectID string                 `json:"projectId"`
	From      string                 `json:"from"` // dbAlias of existing database
	To        map[string]interface{} `json:"to"`
}

// SpecObject describes the basic structure of config specifications
type SpecObject struct {
	API  string            `json:"api" yaml:"api"`
	Type string            `json:"type" yaml:"type"`
	Meta map[string]string `json:"meta" yaml:"meta"`
	Spec interface{}       `json:"spec" yaml:"spec,omitempty"`
}

// SpecObjects array of spec object
type SpecObjects []*SpecObject

// Response is the object returned by every handler to client
type Response struct {
	Error  string        `json:"error,omitempty"`
	Result []interface{} `json:"result,omitempty"`
}
