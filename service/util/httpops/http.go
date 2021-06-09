package httpops

import (
	"encoding/json"
	"net/http"
)

// HttpResponse contains a status code, message, and body to return to the client
// (Content-Type: application/json)
type HttpResponse struct {
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message"`
	Body       interface{} `json:"body"`
}

// ErrResponse writes an http response with the given message, body, and HTTP status code.
func ErrResponse(w http.ResponseWriter, message string, body interface{}, httpStatusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)
	resp := HttpResponse{
		StatusCode: httpStatusCode,
		Message:    message,
		Body:       body,
	}
	jsonResp, _ := json.Marshal(resp)
	w.Write(jsonResp)
}

// RegisterRoutes registers the HTTP handler func of each route
func RegisterRoutes(route string, rootHandler func(http.ResponseWriter, *http.Request)) {
	http.Handle(route, h(rootHandler))
}

// h wraps a http.HandlerFunc and adds common headers.
func h(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf8")
		next.ServeHTTP(w, r)
	})
}
