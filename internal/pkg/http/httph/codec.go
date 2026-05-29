package httph

import (
	"encoding/json"
	"net/http"
)

func EncodeJSON(w http.ResponseWriter, data interface{}) error {
	if err := json.NewEncoder(w).Encode(data); err != nil {
		return err
	}
	return nil
}

func DecodeJSON(r *http.Request, v interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return err
	}
	return nil
}
