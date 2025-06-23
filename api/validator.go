package api

import (
	"github.com/avfirsov/golang-backend-masterclass/util"
	"github.com/go-playground/validator/v10"
)

var ValidCurrency validator.Func = func(fl validator.FieldLevel) bool {
	currency, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}
	return util.IsSupportedCurrency(currency)
}
