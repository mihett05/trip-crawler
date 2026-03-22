package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func ParseAndRespond[Req any](w http.ResponseWriter, r *http.Request) (Req, bool) {
	var req Req

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Write(w, Error{
			Code:    INVALIDINPUT,
			Message: err.Error(),
		}, http.StatusUnprocessableEntity)
		return req, false
	}

	return req, true
}

func Write[Resp any](w http.ResponseWriter, resp Resp, code int) {
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("json.Marshal: failed to marshal body: %s", err), http.StatusInternalServerError)
	}

	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonResp)
	if err != nil {
		http.Error(w, fmt.Sprintf("writer.Wrtie: failed to write body: %s", err), http.StatusInternalServerError)
	}
}
