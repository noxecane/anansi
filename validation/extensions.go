package validation

import (
	"regexp"

	ozzo "github.com/go-ozzo/ozzo-validation/v4"
)

var (
	// validator regexes
	isPhone = regexp.MustCompile("0[789][01][0-9]{8,8}")

	// validation errors

	// Error for nigerian phone number validator
	ErrPhone = ozzo.NewError("validation_is_phone", "must be a valid phone number(080xxxxxxxx)")

	// actual validators

	// Phone is a nigerian phone number(start with 080 and the likes)
	Phone = ozzo.NewStringRuleWithError(
		func(p string) bool { return isPhone.MatchString(p) }, ErrPhone,
	)
)
