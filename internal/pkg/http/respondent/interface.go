package respondent

type Manifest struct {
	Status       int
	Error        string
	ErrorID      string
	ErrorCode    int
	ErrorDetail  string
	ErrorDetails []string
}
type Expander interface {
	Expand(err error) *Manifest
}
type Applicator interface {
	Apply(ctx any, manifest *Manifest)
}
