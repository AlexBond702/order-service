package httph

import "net/http"

func SendRaw(w http.ResponseWriter, statusCode int, mimeType string, data []byte) {
	if mimeType == "" {
		return
	}
	if len(data) == 0 {
		return
	}
	w.Header().Set("Content-Type", mimeType)
	w.WriteHeader(statusCode)
	_, err := w.Write(data)
	if err != nil {
		return
	}
}
