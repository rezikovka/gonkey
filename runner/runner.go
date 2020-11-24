package runner

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/rezikovka/gonkey/checker"
	"github.com/rezikovka/gonkey/fixtures"
	"github.com/rezikovka/gonkey/models"
	"github.com/rezikovka/gonkey/output"
	"github.com/rezikovka/gonkey/testloader"
	"github.com/rezikovka/gonkey/variables"
)

type Config struct {
	Host           string
	FixturesLoader fixtures.Loader
	Variables      *variables.Variables
}

type Runner struct {
	loader   testloader.LoaderInterface
	output   []output.OutputInterface
	checkers []checker.CheckerInterface

	config *Config
}

func New(config *Config, loader testloader.LoaderInterface) *Runner {
	return &Runner{
		config: config,
		loader: loader,
	}
}

func (r *Runner) AddOutput(o ...output.OutputInterface) {
	r.output = append(r.output, o...)
}

func (r *Runner) AddCheckers(c ...checker.CheckerInterface) {
	r.checkers = append(r.checkers, c...)
}

func (r *Runner) Run() (*models.Summary, error) {
	if r.loader == nil {
		s := &models.Summary{
			Success: true,
			Failed:  0,
			Total:   0,
		}
		return s, nil
	}

	loader, err := r.loader.Load()
	if err != nil {
		return nil, err
	}

	client, err := newClient()
	if err != nil {
		return nil, err
	}

	totalTests := 0
	failedTests := 0

	for v := range loader {
		testResult, err := r.executeTest(v, client)
		if err != nil {
			// todo: populate error with test name. Currently it is not possible here to get test name.
			return nil, err
		}
		totalTests++
		if len(testResult.Errors) > 0 {
			failedTests++
		}
		for _, o := range r.output {
			if err := o.Process(v, testResult); err != nil {
				return nil, err
			}
		}
	}

	s := &models.Summary{
		Success: failedTests == 0,
		Failed:  failedTests,
		Total:   totalTests,
	}

	return s, nil
}

func (r *Runner) executeTest(v models.TestInterface, client *http.Client) (*models.Result, error) {

	r.config.Variables.Load(v.GetVariables())
	v = r.config.Variables.Apply(v)

	// load fixtures
	if r.config.FixturesLoader != nil && v.Fixtures() != nil {
		if err := r.config.FixturesLoader.Load(v.Fixtures()); err != nil {
			return nil, fmt.Errorf("unable to load fixtures [%s], error:\n%s", strings.Join(v.Fixtures(), ", "), err)
		}
	}

	// make pause
	pause := v.Pause()
	if pause > 0 {
		time.Sleep(time.Duration(pause) * time.Second)
		fmt.Printf("Sleep %ds before requests\n", pause)
	}

	req, err := newRequest(r.config.Host, v)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)

	_ = resp.Body.Close()

	if err != nil {
		return nil, err
	}

	bodyStr := string(body)

	result := models.Result{
		Path:                req.URL.Path,
		Query:               req.URL.RawQuery,
		RequestBody:         actualRequestBody(req),
		ResponseBody:        bodyStr,
		ResponseContentType: resp.Header.Get("Content-Type"),
		ResponseStatusCode:  resp.StatusCode,
		ResponseStatus:      resp.Status,
		ResponseHeaders:     resp.Header,
		Test:                v,
	}

	for _, c := range r.checkers {
		errs, err := c.Check(v, &result)
		if err != nil {
			return nil, err
		}
		result.Errors = append(result.Errors, errs...)
	}

	if err := r.setVariablesFromResponse(v, result.ResponseContentType, bodyStr, resp.StatusCode); err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *Runner) setVariablesFromResponse(t models.TestInterface, contentType, body string, statusCode int) error {

	varTemplates := t.GetVariablesToSet()
	if varTemplates == nil {
		return nil
	}

	isJson := strings.Contains(contentType, "json") && body != ""

	vars, err := variables.FromResponse(varTemplates[statusCode], body, isJson)
	if err != nil {
		return err
	}

	if vars == nil {
		return nil
	}

	r.config.Variables.Merge(vars)

	return nil
}
