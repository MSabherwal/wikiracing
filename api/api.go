package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"goji.io"

	"goji.io/pat"
)

func wikirace(w http.ResponseWriter, r *http.Request) {
	// always return json
	urlVals := r.URL.Query()
	start := urlVals.Get("start")
	end := urlVals.Get("end")

	if start == "" {
		apiError(w, missingParamError("start"))
		return
	}

	if end == "" {
		apiError(w, missingParamError("end"))
		return
	}

	timeoutStr := urlVals.Get("timeout")
	timeout := 10 * time.Second
	if timeoutStr != "" {
		timeOutInt, err := strconv.Atoi(timeoutStr)
		if err != nil {
			apiError(w, invalidParamError("timeout", "int"))
			return
		}
		timeout = time.Duration(timeOutInt) * time.Second

	}
	fmt.Println(timeout)

}

func Run() {
	mux := goji.NewMux()
	mux.HandleFunc(pat.Get("/wikirace"), wikirace)

	http.ListenAndServe("localhost:8000", mux)
}

func missingParamError(paramName string) error {
	return &httpError{
		Status:  http.StatusUnprocessableEntity,
		Message: fmt.Sprintf("missing param: %s", paramName),
	}
}

func invalidParamError(paramName string, expecting string) error {
	return &httpError{
		Status:  http.StatusUnprocessableEntity,
		Message: fmt.Sprintf("invalid parameter for %s: expecting %s", paramName, expecting),
	}
}

//template for other http errors
type httpError struct {
	Status  int    `json:"status"`
	Message string `json:"error"`
}

func (he httpError) Error() string {
	return fmt.Sprintf("Error: %s", he.Message)
}

func apiError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Println("we in apiError!")
	switch newErr := err.(type) {
	case *httpError:
		fmt.Println("we in httpError!")
		w.WriteHeader(newErr.Status)
		json.NewEncoder(w).Encode(&newErr)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}
