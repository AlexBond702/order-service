package util

import "errors"

func ReplaceErr1(what, from, to error) error {
	return ReplaceErr2(what, from, to, nil, nil)
}

func ReplaceErr2(what, from1, to1, from2, to2 error) error {
	switch {
	case errors.Is(what, from1):
		return to1
	case errors.Is(what, from2):
		return to2
	default:
		return what
	}
}
