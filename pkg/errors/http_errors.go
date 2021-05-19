package errors

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/1412335/grpc-rest-microservice/pkg/log"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

type errorBody struct {
	Err  string            `json:"error,omitempty"`
	Data map[string]string `json:"data,omitempty"`
}

func CustomHTTPError(ctx context.Context, _ *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
	const fallback = `{"error": "failed to marshal error message"}`

	w.Header().Set("Content-type", marshaler.ContentType())
	w.WriteHeader(runtime.HTTPStatusFromCode(status.Code(err)))

	st := status.Convert(err)

	errBd := errorBody{
		Err: st.Message(),
	}

	for _, detail := range st.Details() {
		switch t := detail.(type) {
		case *errdetails.BadRequest:
			errBd.Data = make(map[string]string, len(t.GetFieldViolations()))
			for _, violation := range t.GetFieldViolations() {
				errBd.Data[violation.GetField()] = violation.GetDescription()
			}
		case *errdetails.PreconditionFailure:
			errBd.Data = make(map[string]string, len(t.GetViolations()))
			for _, violation := range t.GetViolations() {
				errBd.Data["type"] = violation.GetType()
				errBd.Data["subject"] = violation.GetSubject()
				errBd.Data["msg"] = violation.GetDescription()
			}
		case *errdetails.ErrorInfo:
			errBd.Data = make(map[string]string, 1)
			errBd.Data["domain"] = t.GetDomain()
			errBd.Data["msg"] = t.GetReason()
		}
	}

	jErr := json.NewEncoder(w).Encode(errBd)
	if jErr != nil {
		if _, err := w.Write([]byte(fallback)); err != nil {
			log.Error("write data HTTP failed", zap.Error(err))
		}
	}
}
