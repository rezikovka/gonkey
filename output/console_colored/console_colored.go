package console_colored

import (
	"bytes"
	"text/template"

	"github.com/fatih/color"
	"github.com/rezikovka/gonkey/models"
	"github.com/rezikovka/gonkey/output"
)

const dotsPerLine = 80

type ConsoleColoredOutput struct {
	output.OutputInterface

	verbose       bool
	dots          int
	coloredPrintf func(format string, a ...interface{})
}

func NewOutput(verbose bool) *ConsoleColoredOutput {
	return &ConsoleColoredOutput{
		verbose:       verbose,
		coloredPrintf: color.New().PrintfFunc(),
	}
}

func (o *ConsoleColoredOutput) Process(t models.TestInterface, result *models.Result) error {
	if !result.Passed() || o.verbose {
		text, err := renderResult(result)
		if err != nil {
			return err
		}
		o.coloredPrintf("%s", text)
	} else {
		o.coloredPrintf(".")
		o.dots++
		if o.dots%dotsPerLine == 0 {
			o.coloredPrintf("\n")
		}
	}
	return nil
}

func renderResult(result *models.Result) (string, error) {
	text := `
       Name: {{ green .Test.GetName }}

Request:
     Method: {{ cyan .Test.GetMethod }}
       Path: {{ cyan .Test.Path }}
      Query: {{ cyan .Test.ToQuery }}
{{- if .Test.Headers }}
    Headers: 
{{- range $key, $value := .Test.Headers }}
      {{ $key }}: {{ $value }}
{{- end }}
{{- end }}
{{- if .Test.Cookies }}
    Cookies: 
{{- range $key, $value := .Test.Cookies }}
      {{ $key }}: {{ $value }}
{{- end }}
{{- end }}
       Body:
{{ if .RequestBody }}{{ cyan .RequestBody }}{{ else }}{{ cyan "<no body>" }}{{ end }}

Response:
     Status: {{ cyan .ResponseStatus }}
       Body:
{{ if .ResponseBody }}{{ yellow .ResponseBody }}{{ else }}{{ yellow "<no body>" }}{{ end }}

{{ if .DbQuery }}
       Db Request:
{{ cyan .DbQuery }}
       Db Response:
{{ range $value := .DbResponse }}
{{ yellow $value }}{{ end }}
{{ end }}

{{ if .Errors }}
     Result: {{ danger "ERRORS!" }}

Errors:
{{ range $i, $e := .Errors }}
{{ inc $i }}) {{ $e.Error }}
{{ end }}
{{ else }}
     Result: {{ success "OK" }}
{{ end }}
`

	var buffer bytes.Buffer
	t := template.Must(template.New("letter").Funcs(templateFuncMap()).Parse(text))
	if err := t.Execute(&buffer, result); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func templateFuncMap() template.FuncMap {
	return template.FuncMap{
		"green":   color.GreenString,
		"cyan":    color.CyanString,
		"yellow":  color.YellowString,
		"danger":  color.New(color.FgHiWhite, color.BgRed).Sprint,
		"success": color.New(color.FgHiWhite, color.BgGreen).Sprint,
		"inc":     func(i int) int { return i + 1 },
	}
}

func (o *ConsoleColoredOutput) ShowSummary(summary *models.Summary) {
	o.coloredPrintf("\nFailed tests: %d/%d\n", summary.Failed, summary.Total)
}
