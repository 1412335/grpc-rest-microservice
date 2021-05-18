package v1

import (
	"context"
	"database/sql"
	"time"

	api_v1 "github.com/1412335/grpc-rest-microservice/pkg/api/v1"
	"github.com/1412335/grpc-rest-microservice/pkg/log"

	"go.uber.org/zap"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// is implementation of api_v1.ToDoServiceServer proto
type toDoServiceServer struct {
	version string
	db      *sql.DB
	logger  log.Factory
}

func NewToDoServiceServer(version string, db *sql.DB, logger log.Factory) api_v1.ToDoServiceServer {
	return &toDoServiceServer{
		version: version,
		db:      db,
		logger:  logger,
	}
}

func (s *toDoServiceServer) checkAPI(api string) error {
	if len(api) > 0 {
		if s.version != api {
			return status.Errorf(codes.Unimplemented, "unsupported Api version: service implements API version %s, but asked for %s", s.version, api)
		}
	}
	return nil
}

func (s *toDoServiceServer) connect(ctx context.Context) (*sql.Conn, error) {
	c, err := s.db.Conn(ctx)
	if err != nil {
		s.logger.For(ctx).Error("Connect db failed", zap.Error(err))
		return nil, status.Errorf(codes.Unknown, "failed to connect to database: "+err.Error())
	}
	return c, nil
}

func (s *toDoServiceServer) Create(ctx context.Context, req *api_v1.CreateRequest) (*api_v1.CreateResponse, error) {
	s.logger.For(ctx).Info("Create.Req", zap.String("data", req.String()))
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	reminder := req.ToDo.Reminder.AsTime()

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

	s.logger.For(ctx).Info("Create.Resp", zap.Int64("id", id))
	return &api_v1.CreateResponse{
		Api: s.version,
		Id:  id,
	}, nil
}

// Read todo
func (s *toDoServiceServer) Read(ctx context.Context, req *api_v1.ReadRequest) (*api_v1.ReadResponse, error) {
	s.logger.For(ctx).Info("Read.Req", zap.String("data", req.String()))
	if err := s.checkAPI(req.Api); err != nil {
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
		if err = rows.Err(); err != nil {
			return nil, status.Errorf(codes.Unknown, "failed to retrieve data from todo: %v", err.Error())
		}
		return nil, status.Errorf(codes.NotFound, "todo with ID=%v is not found", req.Id)
	}

	// get todo
	var td api_v1.ToDo
	var reminder time.Time
	if err = rows.Scan(&td.Id, &td.Title, &td.Description, &reminder); err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to retrieve filed values from todo: %v", err.Error())
	}
	td.Reminder = timestamppb.New(reminder)
	if rows.Next() {
		return nil, status.Errorf(codes.Unknown, "found multiple Todos with Id=%v:", req.Id)
	}

	return &api_v1.ReadResponse{
		Api:  s.version,
		ToDo: &td,
	}, nil
}

// Update todo Task
func (s *toDoServiceServer) Update(ctx context.Context, req *api_v1.UpdateRequest) (*api_v1.UpdateResponse, error) {
	if err := s.checkAPI(req.Api); err != nil {
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
	if err := s.checkAPI(req.Api); err != nil {
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
	if err := s.checkAPI(req.Api); err != nil {
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
		if err = rows.Err(); err != nil {
			return nil, status.Error(codes.Unknown, "failed to retrieve data from todo:"+err.Error())
		}
		return nil, status.Errorf(codes.NotFound, "todo is not found")
	}

	var reminder time.Time
	list := []*api_v1.ToDo{}
	for rows.Next() {
		td := new(api_v1.ToDo)
		if err = rows.Scan(&td.Id, &td.Title, &td.Description, &reminder); err != nil {
			return nil, status.Error(codes.Unknown, "failed to retrieve filed values from todo:"+err.Error())
		}
		td.Reminder = timestamppb.New(reminder)
		list = append(list, td)
	}

	return &api_v1.ReadAllResponse{
		Api:   s.version,
		ToDos: list,
	}, nil
}
