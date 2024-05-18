package validators

type CepValidator struct {
}

func NewCepValidator() *CepValidator {
	return &CepValidator{}
}

func (v *CepValidator) IsCEP(cep string) bool {
	return len(cep) == 8
}
