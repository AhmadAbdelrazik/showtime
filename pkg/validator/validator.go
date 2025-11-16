package validator

type Validator struct {
	Errors map[string]string
}

func New() *Validator {
	return &Validator{
		Errors: make(map[string]string),
	}
}

func (v *Validator) AddError(key, value string) {
	if _, ok := v.Errors[key]; !ok {
		v.Errors[key] = value
	}
}

// Check if condition is true. If false add an error
func (v *Validator) Check(condition bool, key, value string) {
	if !condition {
		v.AddError(key, value)
	}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}
