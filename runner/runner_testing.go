package runner

import (
	"database/sql"
	"net/http/httptest"
	"testing"

	"github.com/joho/godotenv"

	"github.com/rezikovka/gonkey/checker/response_body"
	"github.com/rezikovka/gonkey/checker/response_db"
	"github.com/rezikovka/gonkey/checker/response_header"
	"github.com/rezikovka/gonkey/fixtures"
	"github.com/rezikovka/gonkey/output"
	testingOutput "github.com/rezikovka/gonkey/output/testing"
	"github.com/rezikovka/gonkey/testloader/yaml_file"
	"github.com/rezikovka/gonkey/variables"
)

type RunWithTestingParams struct {
	Server      *httptest.Server
	TestsDir    string
	FixturesDir string
	DB          *sql.DB
	EnvFilePath string
	OutputFunc  output.OutputInterface
	DebugMode   bool   // режим отладки
	TestFilter  string // подстрока для фильтрации тестов по имени файла. Будут запущены только тесты с вхождением подстроки
}

// RunWithTesting is a helper function the wraps the common Run and provides simple way
// to configure Gonkey by filling the params structure.
func RunWithTesting(t *testing.T, params *RunWithTestingParams) {
	if params.EnvFilePath != "" {
		if err := godotenv.Load(params.EnvFilePath); err != nil {
			t.Fatal(err)
		}
	}

	var fixturesLoader fixtures.Loader
	if params.DB != nil {
		fixturesLoader = fixtures.NewLoader(&fixtures.Config{
			Location: params.FixturesDir,
			DB:       params.DB,
			Debug:    params.DebugMode,
		})
	}

	yamlLoader := yaml_file.NewLoader(params.TestsDir)
	yamlLoader.SetFileFilter(params.TestFilter)

	r := New(
		&Config{
			Host:           params.Server.URL,
			FixturesLoader: fixturesLoader,
			Variables:      variables.New(),
		},
		yamlLoader,
	)

	if params.OutputFunc != nil {
		r.AddOutput(params.OutputFunc)
	} else {
		r.AddOutput(testingOutput.NewOutput(t))
	}

	r.AddCheckers(response_body.NewChecker())
	r.AddCheckers(response_header.NewChecker())

	if params.DB != nil {
		r.AddCheckers(response_db.NewChecker(params.DB))
	}

	_, err := r.Run()
	if err != nil {
		t.Fatal(err)
	}
}
