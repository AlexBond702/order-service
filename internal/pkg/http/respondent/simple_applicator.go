package respondent

import "github.com/AlexBond702/order-service/internal/pkg/http/httph"

type SimpleApplicator struct{}

func (sa *SimpleApplicator) Apply(ctx any, manifest *Manifest) {
	if manifest == nil {
		return
	}
	type ManifestJSON struct {
		Status       int      `json:"_"`
		Error        string   `json:"error"`
		ErrorID      string   `json:"error_id,omitempty"`
		ErrorCode    int      `json:"error_code"`
		ErrorDetail  string   `json:"error_detail,omitempty"`
		ErrorDetails []string `json:"error_details,omitempty"`
	}
	httpCtx, ok := ctx.(*HttpContext)
	if !ok {
		return
	}
	w := httpCtx.W
	r := httpCtx.R

	jsonManifest := (*ManifestJSON)(manifest)
	httph.SendEncoded(w, r, jsonManifest.Status, jsonManifest)
}

func NewSimpleApplicator() *SimpleApplicator {
	return &SimpleApplicator{}
}
