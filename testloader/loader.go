package testloader

import (
	"github.com/rezikovka/gonkey/models"
)

type LoaderInterface interface {
	Load() (chan models.TestInterface, error)
}
