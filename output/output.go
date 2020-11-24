package output

import (
	"github.com/rezikovka/gonkey/models"
)

type OutputInterface interface {
	Process(models.TestInterface, *models.Result) error
}
