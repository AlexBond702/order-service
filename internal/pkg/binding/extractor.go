package binding

import (
	"errors"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/ru"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	entr "github.com/go-playground/validator/v10/translations/en"
	rutr "github.com/go-playground/validator/v10/translations/ru"
	"github.com/rs/zerolog/log"

	"github.com/AlexBond702/order-service/internal/pkg/http/respondent"
)

var (
	defaultEnTranslator ut.Translator
	defaultRuTranslator ut.Translator
)

var (
	ErrMalformedSource = errors.New("malformed request source")
	ErrValidatorFailed = (*validationFailedError)(nil)
)

type validationFailedError struct {
	originalErr validator.ValidationErrors
}

func (e *validationFailedError) Error() string {
	return "Validation failed"
}

func (e *validationFailedError) Is(other error) bool {
	return errors.As(other, &ErrValidatorFailed)
}

func NewRespondentManifestExtractor(status, errorCode int, message string) respondent.ManifestExtractor {
	return func(err error) *respondent.Manifest {
		manifest := respondent.Manifest{
			Status:    status,
			ErrorCode: errorCode,
			Error:     message,
		}
		var errList validator.ValidationErrors
		var typedErr *validationFailedError

		if errors.As(err, &errList) {
		} else {
			if errors.As(err, &typedErr) {
				errList = typedErr.originalErr
			} else {
				return nil
			}
		}
		manifest.ErrorDetails = make([]string, 0, len(errList))
		for i := 0; i < len(errList); i++ {
			switch {
			case defaultRuTranslator != nil:
				manifest.ErrorDetails = append(manifest.ErrorDetails, errList[i].Translate(defaultRuTranslator))
			case defaultEnTranslator != nil:
				manifest.ErrorDetails = append(manifest.ErrorDetails, errList[i].Translate(defaultEnTranslator))
			default:
				manifest.ErrorDetails = append(manifest.ErrorDetails, errList[i].Error())
			}
		}
		return &manifest
	}
}

func init() {
	v, _ := Validator.Engine().(*validator.Validate)
	ruLocale := ru.New()
	enLocale := en.New()
	uni := ut.New(enLocale, enLocale, ruLocale)

	var found bool
	defaultEnTranslator, found = uni.GetTranslator("en")
	if !found {
		panic("EN translator not found")
	}

	if err := entr.RegisterDefaultTranslations(v, defaultEnTranslator); err != nil {
		panic("Failed to register EN translations: " + err.Error())
	}
	defaultRuTranslator, found := uni.GetTranslator("ru")
	if !found {
		panic("RU translator not found")
	}
	if err := rutr.RegisterDefaultTranslations(v, defaultRuTranslator); err != nil {
		panic("Failed to register RU translations: " + err.Error())
	}
	registerCustomRussianTranslations(v, defaultRuTranslator)
}

func registerCustomRussianTranslations(v *validator.Validate, trans ut.Translator) {
	if err := v.RegisterTranslation("lte", trans,
		func(ut ut.Translator) error {
			if err := ut.Add("lte", "Поле {0} должно быть {1} или меньше", true); err != nil {
				return err
			}
			return nil
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, err := ut.T("lte", fe.Field(), fe.Param())
			if err != nil {
				return fe.Error()
			}
			return t
		},
	); err != nil {
		log.Printf("Failed 'lte' translation: %v", err)
	}
	if err := v.RegisterTranslation("gte", trans,
		func(ut ut.Translator) error {
			if err := ut.Add("gte", "Поле {0} должно быть больше или равно {1}", true); err != nil {
				return err
			}
			return nil
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, err := ut.T("gte", fe.Field(), fe.Param())
			if err != nil {
				return fe.Error()
			}
			return t
		},
	); err != nil {
		log.Printf("Failed 'gte' translation: %v", err)
	}
	if err := v.RegisterTranslation("oneof", trans,
		func(ut ut.Translator) error {
			if err := ut.Add("oneof", "Поле {0} должно быть одним из: {1}", true); err != nil {
				return err
			}
			return nil
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, err := ut.T("oneof", fe.Field(), fe.Param())
			if err != nil {
				return fe.Error()
			}
			return t
		},
	); err != nil {
		log.Printf("Failed 'oneof' translation: %v", err)
	}
	if err := v.RegisterTranslation("uuid", trans,
		func(ut ut.Translator) error {
			if err := ut.Add("uuid", "Поле {0} должно быть валидным UUID", true); err != nil {
				return err
			}
			return nil
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, err := ut.T("uuid", fe.Field(), fe.Param())
			if err != nil {
				return fe.Error()
			}
			return t
		},
	); err != nil {
		log.Printf("Failed 'uuid' translation: %v", err)
	}
	if err := v.RegisterTranslation("email", trans,
		func(ut ut.Translator) error {
			if err := ut.Add("email", "Поле {0} должно быть корректным email адресом", true); err != nil {
				return err
			}
			return nil
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, err := ut.T("email", fe.Field(), fe.Param())
			if err != nil {
				return fe.Error()
			}
			return t
		},
	); err != nil {
		log.Printf("Failed 'email' translation: %v", err)
	}
}
