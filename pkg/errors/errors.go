// https://cloud.google.com/apis/design/errors
package errors

import (
	"github.com/gogo/googleapis/google/rpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func BadRequest(msg string, fields map[string]string) error {
	st := status.New(codes.InvalidArgument, msg)
	if len(fields) == 0 {
		return st.Err()
	}
	var fieldViolations []*rpc.BadRequest_FieldViolation
	for field, desc := range fields {
		fieldViolations = append(fieldViolations, &rpc.BadRequest_FieldViolation{
			Field:       field,
			Description: desc,
		})
	}
	des, err := st.WithDetails(&rpc.BadRequest{
		FieldViolations: fieldViolations,
	})
	if err != nil {
		return st.Err()
	}
	return des.Err()
}

func InternalServerError(msg, detail string) error {
	st := status.New(codes.Internal, msg)
	des, err := st.WithDetails(&rpc.DebugInfo{
		Detail: detail,
	})
	if err != nil {
		return st.Err()
	}
	return des.Err()
}

func Unauthenticated(msg, field, description string) error {
	st := status.New(codes.Unauthenticated, msg)
	des, err := st.WithDetails(&rpc.ErrorInfo{
		Reason: description,
		Domain: field,
	})
	if err != nil {
		return st.Err()
	}
	return des.Err()
}

func NotFound(msg string, fields map[string]string) error {
	st := status.New(codes.NotFound, msg)
	if len(fields) == 0 {
		return st.Err()
	}
	var fieldViolations []*rpc.PreconditionFailure_Violation
	for field, desc := range fields {
		fieldViolations = append(fieldViolations, &rpc.PreconditionFailure_Violation{
			Type:        "NotFound",
			Subject:     field,
			Description: desc,
		})
	}
	des, err := st.WithDetails(&rpc.PreconditionFailure{
		Violations: fieldViolations,
	})
	if err != nil {
		return st.Err()
	}
	return des.Err()
}
