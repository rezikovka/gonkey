package checker

import "github.com/rezikovka/gonkey/models"

type CheckerInterface interface {
	Check(models.TestInterface, *models.Result) ([]error, error)
}
