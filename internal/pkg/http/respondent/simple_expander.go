package respondent

import (
	"errors"
	"net/http"
)

type SimpleExpander struct {
	rules    []extractorRule
	fallback ManifestExtractor
}
type extractorRule struct {
	Pattern   error
	Extractor ManifestExtractor
}

func (se *SimpleExpander) Expand(err error) *Manifest {
	if err == nil {
		return nil
	}
	for _, rule := range se.rules {
		if errors.Is(err, rule.Pattern) {
			return rule.Extractor(err)
		}
	}
	if fallErr := se.fallback(err); fallErr != nil {
		return fallErr
	}
	return nil
}

func (se *SimpleExpander) ExtractorFor(sentinelErr error, extractor ManifestExtractor) *SimpleExpander {
	if sentinelErr == nil || extractor == nil {
		return nil
	}
	se.rules = append(se.rules, extractorRule{
		Pattern:   sentinelErr,
		Extractor: extractor,
	})
	return se
}

func (se *SimpleExpander) FallbackExtractor(extractor ManifestExtractor) *SimpleExpander {
	return &SimpleExpander{
		rules:    se.rules,
		fallback: extractor,
	}
}

func NewSimpleExpander() *SimpleExpander {
	return &SimpleExpander{
		rules: make([]extractorRule, 0),
		fallback: func(err error) *Manifest {
			return &Manifest{
				Status:      http.StatusInternalServerError,
				Error:       "Internal server error",
				ErrorCode:   50001,
				ErrorDetail: err.Error(),
			}
		},
	}
}

func (se *SimpleExpander) WithoutDetail(err error, status, errorCode int, message string) *SimpleExpander {
	extractor := func(_ error) *Manifest {
		return &Manifest{
			Status:    status,
			Error:     message,
			ErrorCode: errorCode,
		}
	}
	return se.ExtractorFor(err, extractor)
}

func (se *SimpleExpander) WithDetail(err error, status, errorCode int, message, detail string) *SimpleExpander {
	extractor := func(_ error) *Manifest {
		return &Manifest{
			Status:      status,
			Error:       message,
			ErrorCode:   errorCode,
			ErrorDetail: detail,
		}
	}
	return se.ExtractorFor(err, extractor)
}
