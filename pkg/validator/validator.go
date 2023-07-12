package validator

import (
	"fmt"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"strings"
)

type ErrorField struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Validator struct {
	V      *validator.Validate
	Trans  *ut.Translator
	Models *gorm.DB
}

func InitValidator(models *gorm.DB) *Validator {
	eng := en.New()
	uni := ut.New(eng, eng)
	trans, _ := uni.GetTranslator("en")

	val := &Validator{
		V:      validator.New(),
		Trans:  &trans,
		Models: models,
	}
	_ = enTranslations.RegisterDefaultTranslations(val.V, trans)

	// register custom validators
	registerCustomValidators(val)

	return val
}

// registerCustomValidators registers custom validators
func registerCustomValidators(v *Validator) {
	err := v.V.RegisterValidation("unique", v.uniqueValidator)
	if err != nil {
		zap.S().Fatalf("error registering unique validator: %v", err)
	}
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

func (validation *Validator) uniqueValidator(fl validator.FieldLevel) bool {
	/** Translation */
	_ = validation.V.RegisterTranslation("unique", *validation.Trans, func(ut ut.Translator) error {
		return ut.Add("unique", "{0} '{1}' already taken", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("unique", fe.Field(), fe.Value().(string))
		return t
	})
	/** End Translation */

	value := fl.Field().String()
	parameters := strings.Split(fl.Param(), "/")

	zap.S().Debugf("value: %s, parameters: %s", value, parameters)

	table, column := parameters[0], parameters[1]

	var count int64
	validation.Models.Table(table).
		Where(column+" = ?", value).
		Count(&count)

	if count > 0 {
		return false
	}

	return true
}
