package domain

import "errors"

var ErrValidationFailed = errors.New("placeholder value validation failed")
var ErrEffectiveDateMissing = errors.New("effective_date missing")
