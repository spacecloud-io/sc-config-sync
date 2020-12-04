package model

// DBSyncRequest db sync request
type DBSyncRequest struct {
	ProjectID string          `json:"projectId"`
	From      string          `json:"from"` // dbAlias of existing database
	To        *DatabaseConfig `json:"to"`
}

// DatabaseConfig stores information of database config
type DatabaseConfig struct {
	DbAlias string `json:"dbAlias,omitempty" yaml:"dbAlias" mapstructure:"dbAlias"`
	DBName  string `json:"name,omitempty" yaml:"name" mapstructure:"name"` // name of the logical database or schema name according to the database type
	Conn    string `json:"conn,omitempty" yaml:"conn" mapstructure:"conn"`
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
