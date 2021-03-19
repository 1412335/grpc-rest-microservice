package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/1412335/grpc-rest-microservice/pkg/configs"
)

//Used to execute client creation procedure only once.
var once sync.Once

// errors

type DataAccessLayer struct {
	dbConfig *configs.Database
	//Used during creation of singleton client object in GetMongoClient().
	dbInstance *sql.DB
}

func NewDataAccessLayer(ctx context.Context, cfg *configs.Database) (*DataAccessLayer, error) {
	dal := &DataAccessLayer{
		dbConfig: cfg,
	}
	if _, err := dal.Connect(ctx); err != nil {
		return nil, err
	}
	return dal, nil
}

// Build connection string
func (dal *DataAccessLayer) buildConnectionDSN() (string, error) {
	cfg := dal.dbConfig
	param := "parseTime=true"
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Scheme,
		param,
	)
	return dsn, nil
}

// Connect
func (dal *DataAccessLayer) Connect(ctx context.Context) (*sql.DB, error) {
	//Perform connection creation operation only once.
	var err error
	once.Do(func() {
		// build connection string
		dsn, err := dal.buildConnectionDSN()
		if err != nil {
			return
		}
		// connect db
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			return
		}
		defer db.Close()

		// https://github.com/go-sql-driver/mysql/
		// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
		db.SetMaxIdleConns(10)

		// SetMaxOpenConns sets the maximum number of open connections to the database.
		db.SetMaxOpenConns(100)

		// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
		db.SetConnMaxLifetime(time.Hour)

		dal.dbInstance = db
	})
	return dal.dbInstance, err
}

func (dal *DataAccessLayer) Disconnect() error {
	return dal.dbInstance.Close()
}

func (dal *DataAccessLayer) GetDatabase() *sql.DB {
	return dal.dbInstance
}
