package v1

import (
	"context"
	"database/sql"
	"time"

	api_v1 "github.com/1412335/grpc-rest-microservice/pkg/api/v1"

	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	apiVersion = "v1"
)

// is implementation of api_v1.ToDoServiceServer proto
type toDoServiceServer struct {
	db *sql.DB
}

func NewToDoServiceServer(db *sql.DB) api_v1.ToDoServiceServer {
	return &toDoServiceServer{
		db: db,
	}
}

func (s *toDoServiceServer) checkApi(api string) error {
	if len(api) > 0 {
		if apiVersion != api {
			return status.Errorf(codes.Unimplemented, "unsupported Api version: service implements API version %s, but asked for %s", apiVersion, api)
		}
	}
	return nil
}

func (s *toDoServiceServer) connect(ctx context.Context) (*sql.Conn, error) {
	c, err := s.db.Conn(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to connect to database: "+err.Error())
	}
	return c, nil
}

func (s *toDoServiceServer) Create(ctx context.Context, req *api_v1.CreateRequest) (*api_v1.CreateResponse, error) {
	if err := s.checkApi(req.Api); err != nil {
		return nil, err
	}

	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	reminder, err := ptypes.Timestamp(req.ToDo.Reminder)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "reminder field has invalid format:"+err.Error())
	}

	// insert todo entity data
	res, err := c.ExecContext(ctx, "INSERT INTO ToDo(`Title`, `Description`, `Reminder`) VALUES (?, ?, ?)", req.ToDo.Title, req.ToDo.Description, reminder)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to insert into ToDo:"+err.Error())
	}

	// get ID of creates ToDo
	id, err := res.LastInsertId()
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve id for created Todo:"+err.Error())
	}

	return &api_v1.CreateResponse{
		Api: apiVersion,
		Id:  id,
	}, nil
}

// Read todo
func (s *toDoServiceServer) Read(ctx context.Context, req *api_v1.ReadRequest) (*api_v1.ReadResponse, error) {
	if err := s.checkApi(req.Api); err != nil {
		return nil, err
	}

	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	// query todo by id
	rows, err := c.QueryContext(ctx, "SELECT `ID`, `Title`, `Description`, `Reminder` FROM ToDo WHERE `ID`=?", req.Id)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to select todo:"+err.Error())
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, status.Error(codes.Unknown, "failed to retrieve data from todo:"+err.Error())
		}
		return nil, status.Errorf(codes.NotFound, "todo with ID=%v is not found", req.Id)
	}

	// get todo
	var td api_v1.ToDo
	var reminder time.Time
	if err := rows.Scan(&td.Id, &td.Title, &td.Description, &reminder); err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve filed values from todo:"+err.Error())
	}
	td.Reminder, err = ptypes.TimestampProto(reminder)
	if err != nil {
		return nil, status.Error(codes.Unknown, "reminder has invalid format:"+err.Error())
	}

	if rows.Next() {
		return nil, status.Errorf(codes.Unknown, "found multiple Todos with Id=%v:", req.Id)
	}

	return &api_v1.ReadResponse{
		Api:  apiVersion,
		ToDo: &td,
	}, nil
}

// Update todo Task
func (s *toDoServiceServer) Update(ctx context.Context, req *api_v1.UpdateRequest) (*api_v1.UpdateResponse, error) {
	if err := s.checkApi(req.Api); err != nil {
		return nil, err
	}

	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	return nil, nil
}

// Delete todo Task
func (s *toDoServiceServer) Delete(ctx context.Context, req *api_v1.DeleteRequest) (*api_v1.DeleteResponse, error) {
	if err := s.checkApi(req.Api); err != nil {
		return nil, err
	}

	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	return nil, nil
}

// ReadAll todo Tasks
func (s *toDoServiceServer) ReadAll(ctx context.Context, req *api_v1.ReadAllRequest) (*api_v1.ReadAllResponse, error) {
	if err := s.checkApi(req.Api); err != nil {
		return nil, err
	}

	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	// query todo by id
	rows, err := c.QueryContext(ctx, "SELECT `ID`, `Title`, `Description`, `Reminder` FROM ToDo")
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to select todo:"+err.Error())
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, status.Error(codes.Unknown, "failed to retrieve data from todo:"+err.Error())
		}
		return nil, status.Errorf(codes.NotFound, "todo is not found")
	}

	var reminder time.Time
	list := []*api_v1.ToDo{}
	for rows.Next() {
		td := new(api_v1.ToDo)
		if err := rows.Scan(&td.Id, &td.Title, &td.Description, &reminder); err != nil {
			return nil, status.Error(codes.Unknown, "failed to retrieve filed values from todo:"+err.Error())
		}
		td.Reminder, err = ptypes.TimestampProto(reminder)
		if err != nil {
			return nil, status.Error(codes.Unknown, "reminder has invalid format:"+err.Error())
		}
		list = append(list, td)
	}

	return &api_v1.ReadAllResponse{
		Api:   apiVersion,
		ToDos: list,
	}, nil
}
