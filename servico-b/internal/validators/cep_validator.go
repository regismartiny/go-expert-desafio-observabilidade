package validators

import (
	"github.com/paemuri/brdoc"
)

type CepValidator struct {
}

func NewCepValidator() *CepValidator {
	return &CepValidator{}
}

func (v *CepValidator) IsCEP(cep string) bool {
	return brdoc.IsCEP(cep)
}
