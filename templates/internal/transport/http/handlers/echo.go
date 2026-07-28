// [[ when (modeIs "http") ]]
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"[[.ModulePath]]/internal/domain/echo"
	"[[.ModulePath]]/internal/observability/logging"
)

type EchoHandler struct {
	service      echo.Service
	maxBodyBytes int64
}

func NewEchoHandler(service echo.Service, maxBodyBytes int64) EchoHandler {
	return EchoHandler{service: service, maxBodyBytes: maxBodyBytes}
}

func (h EchoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Cap the body so a large or endless upload cannot exhaust memory.
	r.Body = http.MaxBytesReader(w, r.Body, h.maxBodyBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req echo.Request
	if err := dec.Decode(&req); err != nil {
		// Decode failures echo attacker-controlled input, so log the detail and
		// return a fixed message.
		logging.FromContext(r.Context()).Warn("decode echo request", "error_message", err.Error())

		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	res := h.service.Echo(req)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		// The status is already committed if any bytes were written; just record it.
		logging.FromContext(r.Context()).Error("encode echo response", "error_message", err.Error())
	}
}
