package domain

import "strings"

func trimOptionalString(v *string) *string {
	if v == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*v)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func trimOptionalAreaCode(v *AreaCode) *AreaCode {
	if v == nil {
		return nil
	}
	trimmed := AreaCode(strings.TrimSpace(string(*v)))
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
