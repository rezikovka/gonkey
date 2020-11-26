package yaml_file

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/rezikovka/gonkey/models"
	"gopkg.in/yaml.v2"
)

func parseTestDefinitionFile(absPath string) ([]Test, error) {
	data, err := ioutil.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s:\n%s", absPath, err)
	}

	var testDefinitions []TestDefinition

	// reading the test source file
	if err := yaml.Unmarshal(data, &testDefinitions); err != nil {
		return nil, fmt.Errorf("failed to unmarshall %s:\n%s", absPath, err)
	}

	fileLocatedDir := getFileDirectory(absPath)

	var tests []Test

	for _, definition := range testDefinitions {
		definition.fileLocatedDir = fileLocatedDir

		if testCases, err := makeTestFromDefinition(definition); err != nil {
			return nil, err
		} else {
			tests = append(tests, testCases...)
		}
	}

	return tests, nil
}

// getFileDirectory возвращает путь к директории, в которой находится указанный файл.
func getFileDirectory(absPath string) string {
	parts := strings.Split(absPath, "/")

	if len(parts) < 2 {
		return ""
	}

	return strings.Join(parts[:len(parts)-1], "/")
}

func substituteArgs(tmpl string, args map[string]interface{}) (string, error) {
	compiledTmpl, err := template.New("").Parse(tmpl)
	if err != nil {
		return "", err
	}

	buf := &bytes.Buffer{}

	if err := compiledTmpl.Execute(buf, args); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func substituteArgsToMap(tmpl map[string]string, args map[string]interface{}) (map[string]string, error) {
	res := make(map[string]string)
	for key, value := range tmpl {
		var err error
		res[key], err = substituteArgs(value, args)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

// Make tests from the given test definition.
func makeTestFromDefinition(testDefinition TestDefinition) ([]Test, error) {
	var tests []Test

	request, err := resolveRequestBody(testDefinition)
	if err != nil {
		return nil, err
	}

	responses, err := resolveResponses(testDefinition)
	if err != nil {
		return nil, err
	}

	// test definition has no cases, so using request/response as is
	if len(testDefinition.Cases) == 0 {
		test := Test{TestDefinition: testDefinition}
		test.Request = request.Value
		test.Responses = responses
		test.ResponseHeaders = testDefinition.ResponseHeaders
		test.DbQuery = testDefinition.DbQueryTmpl
		test.DbResponse = testDefinition.DbResponseTmpl
		return append(tests, test), nil
	}

	// produce as many tests as cases defined
	for caseIdx, testCase := range testDefinition.Cases {
		test := Test{TestDefinition: testDefinition}
		test.Name = fmt.Sprintf("%s #%d", test.Name, caseIdx)

		// substitute RequestArgs to different parts of request
		test.RequestURL, err = substituteArgs(testDefinition.RequestURL, testCase.RequestArgs)
		if err != nil {
			return nil, err
		}

		test.Request, err = substituteArgs(request.Value, testCase.RequestArgs)
		if err != nil {
			return nil, err
		}

		test.QueryParams, err = substituteArgs(testDefinition.QueryParams, testCase.RequestArgs)
		if err != nil {
			return nil, err
		}

		test.HeadersVal, err = substituteArgsToMap(testDefinition.HeadersVal, testCase.RequestArgs)
		if err != nil {
			return nil, err
		}

		test.CookiesVal, err = substituteArgsToMap(testDefinition.CookiesVal, testCase.RequestArgs)
		if err != nil {
			return nil, err
		}

		// substitute ResponseArgs to different parts of response
		test.Responses = make(map[int]*models.DataBody)
		for status, tpl := range responses {
			args, ok := testCase.ResponseArgs[status]
			if ok {
				// found args for response status

				tpl.Value, err = substituteArgs(tpl.Value, args)
				if err != nil {
					return nil, err
				}
				test.Responses[status] = tpl
			} else {
				// not found args, using response as is
				test.Responses[status] = tpl
			}
		}

		test.ResponseHeaders = make(map[int]map[string]string)
		for status, respHeaders := range testDefinition.ResponseHeaders {
			args, ok := testCase.ResponseArgs[status]
			if ok {
				// found args for response status
				test.ResponseHeaders[status], err = substituteArgsToMap(respHeaders, args)
				if err != nil {
					return nil, err
				}
			} else {
				// not found args, using response as is
				test.ResponseHeaders[status] = respHeaders
			}
		}

		test.DbQuery, err = substituteArgs(testDefinition.DbQueryTmpl, testCase.DbQueryArgs)
		if err != nil {
			return nil, err
		}

		// compile DbResponse
		if testCase.DbResponse != nil {
			// DbResponse from test case has top priority
			test.DbResponse = testCase.DbResponse
		} else {
			if len(testDefinition.DbResponseTmpl) != 0 {
				// compile DbResponse string by string
				for _, tpl := range testDefinition.DbResponseTmpl {
					dbResponseString, err := substituteArgs(tpl, testCase.DbResponseArgs)
					if err != nil {
						return nil, err
					}
					test.DbResponse = append(test.DbResponse, dbResponseString)
				}
			} else {
				test.DbResponse = testDefinition.DbResponseTmpl
			}
		}
		tests = append(tests, test)
	}

	return tests, nil
}

func getDataTypeByFileName(fileName string) models.DataType {
	parts := strings.Split(fileName, ".")

	ext := parts[len(parts)-1]

	switch ext {
	case "json":
		return models.DataTypeJson
	default:
		return models.DataTypePlainText
	}
}

// readDataFile считывает файл и валидирует его согласно типу данных
func readDataFile(directory, fileName string) (*models.DataBody, error) {
	path := strings.TrimRight(directory, "/") + "/" + strings.TrimLeft(fileName, "/")
	dataType := getDataTypeByFileName(fileName)

	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s:\n%s", path, err)
	}

	if dataType == models.DataTypeJson {
		err = validateJson(data)
		if err != nil {
			return nil, fmt.Errorf("invalid json in file %s:\n%s", path, err)
		}
	}

	return &models.DataBody{Type: dataType, Value: string(data)}, nil
}

func validateJson(data []byte) error {
	var j interface{}
	return json.Unmarshal(data, &j)
}

// resolveRequestBody возвращает тело запроса.
func resolveRequestBody(definition TestDefinition) (*models.DataBody, error) {
	request, err := resolveDataBody(definition.fileLocatedDir, definition.Request)
	if err != nil {
		return nil, err
	}

	if request == nil {
		request = &models.DataBody{
			Type:  models.DataTypePlainText,
			Value: "",
		}
	}
	return request, err
}

// resolveResponses формирует мап с ожидаемыми ответами.
func resolveResponses(definition TestDefinition) (map[int]*models.DataBody, error) {
	responses := make(map[int]*models.DataBody)
	if definition.Responses == nil {
		return responses, nil
	}

	var err error
	for key, _ := range responses {
		responses[key], err = resolveDataBody(definition.fileLocatedDir, definition.Responses[key])
		if err != nil {
			return nil, err
		}
	}

	return responses, nil
}

func resolveDataBody(testCaseFileLocationDir string, data *models.DataBody) (*models.DataBody, error) {
	if data == nil {
		return nil, nil
	}

	var err error
	switch data.Type {
	case models.DataTypePath:
		data, err = readDataFile(testCaseFileLocationDir, data.Value)
		if err != nil {
			return nil, err
		}
		return data, nil
	case models.DataTypeJson:
		if err = validateJson([]byte(data.Value)); err != nil {
			return nil, fmt.Errorf("invalid json:\n%s", err)
		}
		return data, nil
	case models.DataTypePlainText:
		return data, nil
	}

	return nil, fmt.Errorf("unexpected data type %s", data.Type)
}
