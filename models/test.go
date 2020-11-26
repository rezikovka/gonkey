package models

type DataType string

const DataTypePath = DataType("path")
const DataTypeJson = DataType("json")
const DataTypePlainText = DataType("plain")

// TODO: add support for form fields
type Form struct {
	Files map[string]string `yaml:"files"`
}

type DataBody struct {
	Type  DataType `yaml:"type"`
	Value string   `yaml:"value"`
}

type Summary struct {
	Success bool
	Failed  int
	Total   int
}

// Common Test interface
type TestInterface interface {
	ToQuery() string
	GetRequest() string
	ToJSON() ([]byte, error)
	GetMethod() string
	Path() string
	GetResponses() map[int]*DataBody
	GetResponse(code int) (*DataBody, bool)
	GetResponseHeaders(code int) (map[string]string, bool)
	GetName() string
	Fixtures() []string
	Pause() int
	Cookies() map[string]string
	Headers() map[string]string
	ContentType() string
	GetForm() *Form
	DbQueryString() string
	DbResponseJson() []string
	GetVariables() map[string]string
	GetVariablesToSet() map[int]map[string]string

	// setters
	SetQuery(string)
	SetMethod(string)
	SetPath(string)
	SetRequest(string)
	SetForm(form *Form)
	SetResponses(map[int]*DataBody)
	SetHeaders(map[string]string)

	// comparison properties
	NeedsCheckingValues() bool
	IgnoreArraysOrdering() bool
	DisallowExtraFields() bool

	// Clone returns copy of current object
	Clone() TestInterface
}
