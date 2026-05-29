package binding

import "net/http"

var Validator StructValidator = &defaultValidator{}

var (
	bJSON  = jsonBinding{}
	bQuery = queryBinding{}
)

type Binding interface {
	Name() string
	Bind(r *http.Request, obj any) error
}
type BindingUri interface {
	Name() string
	BindUri(params map[string][]string, obj any) error
}
type StructValidator interface {
	ValidateStruct(obj any) error
	Engine() any
}

func validate(obj any) error {
	if Validator == nil {
		return nil
	}
	return Validator.ValidateStruct(obj)
}
