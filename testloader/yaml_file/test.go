package yaml_file

import (
	"github.com/rezikovka/gonkey/models"
)

type Test struct {
	models.TestInterface

	TestDefinition

	Request         string
	Responses       map[int]*models.DataBody
	ResponseHeaders map[int]map[string]string
	DbQuery         string
	DbResponse      []string
}

func (t *Test) ToQuery() string {
	return t.QueryParams
}

func (t *Test) GetMethod() string {
	return t.Method
}

func (t *Test) Path() string {
	return t.RequestURL
}

func (t *Test) GetRequest() string {
	return t.Request
}

func (t *Test) ToJSON() ([]byte, error) {
	return []byte(t.Request), nil
}

func (t *Test) GetResponses() map[int]models.DataBody {
	return t.Responses
}

func (t *Test) GetResponse(code int) (models.DataBody, bool) {
	val, ok := t.Responses[code]
	return val, ok
}

func (t *Test) GetResponseHeaders(code int) (map[string]string, bool) {
	val, ok := t.ResponseHeaders[code]
	return val, ok
}

func (t *Test) NeedsCheckingValues() bool {
	return !t.ComparisonParams.IgnoreValues
}

func (t *Test) GetName() string {
	return t.Name
}

func (t *Test) IgnoreArraysOrdering() bool {
	return t.ComparisonParams.IgnoreArraysOrdering
}

func (t *Test) DisallowExtraFields() bool {
	return t.ComparisonParams.DisallowExtraFields
}

func (t *Test) Fixtures() []string {
	return t.FixtureFiles
}

func (t *Test) Pause() int {
	return t.PauseValue
}

func (t *Test) Cookies() map[string]string {
	return t.CookiesVal
}

func (t *Test) Headers() map[string]string {
	return t.HeadersVal
}

// TODO: it might make sense to do support of case-insensitive checking
func (t *Test) ContentType() string {
	ct, _ := t.HeadersVal["Content-Type"]
	return ct
}

func (t *Test) DbQueryString() string {
	return t.DbQuery
}

func (t *Test) DbResponseJson() []string {
	return t.DbResponse
}

func (t *Test) GetVariables() map[string]string {
	return t.Variables
}

func (t *Test) GetForm() *models.Form {
	return t.Form
}

func (t *Test) GetVariablesToSet() map[int]map[string]string {
	return t.VariablesToSet
}

func (t *Test) Clone() models.TestInterface {
	res := *t

	return &res
}

func (t *Test) SetQuery(val string) {
	t.QueryParams = val
}
func (t *Test) SetMethod(val string) {
	t.Method = val
}
func (t *Test) SetPath(val string) {
	t.RequestURL = val
}

func (t *Test) SetRequest(val string) {
	t.Request = val
}

func (t *Test) SetForm(val *models.Form) {
	t.Form = val
}

func (t *Test) SetResponses(val map[int]models.DataBody) {
	t.Responses = val
}

func (t *Test) SetHeaders(val map[string]string) {
	t.HeadersVal = val
}
