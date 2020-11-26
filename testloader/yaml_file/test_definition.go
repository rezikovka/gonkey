package yaml_file

import "github.com/rezikovka/gonkey/models"

type TestDefinition struct {
	fileLocatedDir   string
	Name             string                    `yaml:"name"`
	Variables        map[string]string         `yaml:"variables"`
	VariablesToSet   VariablesToSet            `yaml:"variables_to_set"`
	Form             *models.Form              `yaml:"form"`
	Method           string                    `yaml:"method"`
	RequestURL       string                    `yaml:"path"`
	QueryParams      string                    `yaml:"query"`
	Request          *models.DataBody          `yaml:"request"`
	Responses        map[int]*models.DataBody  `yaml:"response"`
	ResponseHeaders  map[int]map[string]string `yaml:"responseHeaders"`
	HeadersVal       map[string]string         `yaml:"headers"`
	CookiesVal       map[string]string         `yaml:"cookies"`
	Cases            []CaseData                `yaml:"cases"`
	ComparisonParams comparisonParams          `yaml:"comparisonParams"`
	FixtureFiles     []string                  `yaml:"fixtures"`
	PauseValue       int                       `yaml:"pause"`
	DbQueryTmpl      string                    `yaml:"dbQuery"`
	DbResponseTmpl   []string                  `yaml:"dbResponse"`
}

type CaseData struct {
	RequestArgs    map[string]interface{}         `yaml:"requestArgs"`
	ResponseArgs   map[int]map[string]interface{} `yaml:"responseArgs"`
	DbQueryArgs    map[string]interface{}         `yaml:"dbQueryArgs"`
	DbResponseArgs map[string]interface{}         `yaml:"dbResponseArgs"`
	DbResponse     []string                       `yaml:"dbResponse"`
}

type comparisonParams struct {
	IgnoreValues         bool `yaml:"ignoreValues"`
	IgnoreArraysOrdering bool `yaml:"ignoreArraysOrdering"`
	DisallowExtraFields  bool `yaml:"disallowExtraFields"`
}

type VariablesToSet map[int]map[string]string

/*
There can be two types of data in yaml-file:
1) JSON-paths:
	VariablesToSet:
		<code1>:
			<varName1>: <JSON_Path1>
			<varName2>: <JSON_Path2>
2) Plain text:
	 VariablesToSet:
		<code1>: <varName1>
		<code2>: <varName2>
		...
   In this case we unmarshall values to format similar to JSON-paths format with empty paths:
	 VariablesToSet:
		<code1>:
			<varName1>: ""
		<code2>:
			<varName2>: ""
*/
//func (v *VariablesToSet) UnmarshalYAML(unmarshal func(interface{}) error) error {
//
//	res := make(map[int]map[string]string)
//
//	// try to unmarshall as plaint text
//	var plain map[int]string
//	if err := unmarshal(&plain); err == nil {
//
//		for code, varName := range plain {
//			res[code] = map[string]string{
//				varName: "",
//			}
//		}
//
//		*v = res
//		return nil
//	}
//
//	// json-paths
//	if err := unmarshal(&res); err != nil {
//		return err
//	}
//
//	*v = res
//	return nil
//}
