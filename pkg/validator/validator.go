package validator

import (
	"fmt"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
)

type ErrorField struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Validator struct {
	V     *validator.Validate
	Trans *ut.Translator
}

func InitValidator() *Validator {
	eng := en.New()
	uni := ut.New(eng, eng)
	trans, _ := uni.GetTranslator("en")

	val := &Validator{
		V:     validator.New(),
		Trans: &trans,
	}
	_ = enTranslations.RegisterDefaultTranslations(val.V, trans)

	return val
}

func (validation *Validator) Validate(object interface{}) (errs []ErrorField) {

	err := validation.V.Struct(object)

	if err != nil {
		return validation.validationError(err)
	}

	return nil
}

func (validation *Validator) validationError(err error) (errs []ErrorField) {

	// check for validation errors
	if _, ok := err.(*validator.InvalidValidationError); ok {
		return append(errs, ErrorField{
			Field:   "validation",
			Message: err.Error(),
		})
	}

	return validation.translateError(err)
}

func (validation *Validator) translateError(err error) (errs []ErrorField) {
	if err == nil {
		return nil
	}
	validatorErrs := err.(validator.ValidationErrors)
	for _, e := range validatorErrs {
		translatedErr := fmt.Errorf(e.Translate(*validation.Trans))
		errs = append(errs, ErrorField{
			Field:   e.Field(),
			Message: translatedErr.Error(),
		})
	}
	return errs
}
