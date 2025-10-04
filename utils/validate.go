package utils

import (
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
)

var (
	Validator  *validator.Validate = validator.New()
	PhoneRegex string              = `^(0|\+84)(3|5|7|8|9)\d{8}$`
)

func HandlerValidation(err error) map[string]string {
	errValidator := map[string]string{}
	if err == nil {
		return errValidator
	}
	if errVa, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errVa {
			switch e.Tag() {
			case "required":
				errValidator[e.Field()] = fmt.Sprintf("%s is required", e.Field())
			case "email":
				errValidator[e.Field()] = fmt.Sprintf("%s is not a valid email", e.Field())
			case "phoneVn":
				errValidator[e.Field()] = fmt.Sprintf("%s is not a valid phone", e.Field())
			}
		}
	}
	return errValidator
}

func init() {
	// Custom validator cho số điện thoại VN
	Validator.RegisterValidation("phoneVn", func(fl validator.FieldLevel) bool {
		phone := fl.Field().String()
		matched, _ := regexp.MatchString(PhoneRegex, phone)
		return matched
	})
}
