// https://cloud.google.com/apis/design/errors
package errors

import (
	"github.com/gogo/googleapis/google/rpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func BadRequest(msg, field, description string) error {
	st := status.New(codes.InvalidArgument, msg)
	des, err := st.WithDetails(&rpc.BadRequest{
		FieldViolations: []*rpc.BadRequest_FieldViolation{
			{
				Field:       field,
				Description: description,
			},
		},
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

func NotFound(msg, field, description string) error {
	st := status.New(codes.NotFound, msg)
	des, err := st.WithDetails(&rpc.PreconditionFailure{
		Violations: []*rpc.PreconditionFailure_Violation{
			{
				Type:        "NotFound",
				Subject:     field,
				Description: description,
			},
		},
	})
	if err != nil {
		return des.Err()
	}
	return st.Err()
}
