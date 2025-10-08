package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	Validator  *validator.Validate = validator.New()
	PhoneRegex string              = `^(0|\+84)(3|5|7|8|9)\d{8}$`
)

func HandlerValidation(err error) string {
	errValidator := ""
	if err == nil {
		return errValidator
	}
	if errVa, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errVa {
			switch e.Tag() {
			case "required":
				errValidator += fmt.Sprintf("%s không được trống, ", strings.ToLower(e.Field()))
			case "email":
				errValidator += fmt.Sprintf("%s không phải là một email hợp lệ, ", strings.ToLower(e.Field()))
			case "phoneVn":
				errValidator += fmt.Sprintf("%s phải theo định dạng số phone Việt Nam, ", strings.ToLower(e.Field()))
			}
		}
		errValidator = strings.TrimSuffix(errValidator, ", ")
	}
	return errValidator
}

func init() {
	// Custom validator cho số điện thoại VN
	_ = Validator.RegisterValidation("phoneVn", func(fl validator.FieldLevel) bool {
		phone := fl.Field().String()
		matched, _ := regexp.MatchString(PhoneRegex, phone)
		return matched
	})
}
