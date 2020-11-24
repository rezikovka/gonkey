package fixtures

import (
	"database/sql"
	"strings"

	_ "github.com/lib/pq"

	"github.com/lamoda/gonkey/fixtures/postgres"
)

type Config struct {
	DB       *sql.DB
	Location string
	Debug    bool
}

type Loader interface {
	Load(names []string) error
}

func NewLoader(cfg *Config) Loader {

	var loader Loader

	location := strings.TrimRight(cfg.Location, "/")

	loader = postgres.New(
		cfg.DB,
		location,
		cfg.Debug,
	)

	return loader
}
