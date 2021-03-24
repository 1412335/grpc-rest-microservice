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

func InternalServerError(msg string) error {
	return status.Errorf(codes.Internal, msg)
}
