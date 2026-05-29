package binding

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/AlexBond702/order-service/internal/pkg/http/httph"
)

func ScanAndValidateJSON(r *http.Request, to any) error {
	return scanAndValidate(r, to, bJSON)
}

func ScanAndValidateQuery(r *http.Request, to any) error {
	return scanAndValidate(r, to, bQuery)
}

func scanAndValidate(r *http.Request, to any, b Binding) error {
	err := b.Bind(r, to)
	if err == nil {
		return nil
	}
	var validationErr validator.ValidationErrors
	if errors.As(err, &validationErr) {
		return &validationFailedError{validationErr}
	}
	httph.ErrorApplyDetail(r, "Malformed HTTP request "+b.Name()+" source")
	return ErrMalformedSource
}
