package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/listou/listou/apps/api/internal/platform/logger"
)

const maxBodyBytes = 1 << 20

type errorBody struct {
	Error errorPayload `json:"error"`
}

type errorPayload struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	RequestID string            `json:"requestId"`
	Fields    map[string]string `json:"fields,omitempty"`
}

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(v)
}

// Fail writes the standard error envelope. Internal errors are logged with
// their cause; the client only ever sees the stable code and message.
func Fail(w http.ResponseWriter, r *http.Request, err error) {
	appErr := AsError(err)
	if appErr.Status >= 500 {
		logger.From(r.Context()).Error("request failed", "error", err.Error())
	}
	JSON(w, appErr.Status, errorBody{Error: errorPayload{
		Code:      appErr.Code,
		Message:   appErr.Message,
		RequestID: RequestIDFrom(r.Context()),
		Fields:    appErr.Fields,
	}})
}

// Decode reads a JSON body into dst, rejecting unknown fields and oversized bodies.
func Decode(r *http.Request, dst any) error {
	if ct := r.Header.Get("Content-Type"); ct != "" && !strings.HasPrefix(ct, "application/json") {
		return NewError(http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE", "Envie os dados em JSON.")
	}
	dec := json.NewDecoder(io.LimitReader(r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return BadRequest("EMPTY_BODY", "O corpo da requisição está vazio.")
		}
		return BadRequest("INVALID_JSON", "Não foi possível ler os dados enviados.").WithCause(err)
	}
	return nil
}
