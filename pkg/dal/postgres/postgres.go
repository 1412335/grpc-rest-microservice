package postgres

import (
	"context"
	"fmt"
	"sync"

	"github.com/1412335/grpc-rest-microservice/pkg/configs"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//Used to execute client creation procedure only once.
var once sync.Once

// errors

type DataAccessLayer struct {
	dbConfig *configs.Database
	//Used during creation of singleton client object in GetMongoClient().
	dbInstance *gorm.DB
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
	return fmt.Sprintf("host=%s port=%v user=%s dbname=%s sslmode=disable password=%s", cfg.Host, cfg.Port, cfg.User, cfg.Scheme, cfg.Password), nil
}

// Connect
func (dal *DataAccessLayer) Connect(ctx context.Context) (*gorm.DB, error) {
	//Perform connection creation operation only once.
	var err error
	once.Do(func() {
		// build connection string
		dsn, err := dal.buildConnectionDSN()
		if err != nil {
			return
		}
		// connect db
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return
		}

		if dal.dbConfig.Debug {
			db = db.Debug()
		}

		sqlDB, err := db.DB()
		if err != nil {
			return
		}

		// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
		sqlDB.SetMaxIdleConns(dal.dbConfig.MaxIdleConns)

		// SetMaxOpenConns sets the maximum number of open connections to the database.
		sqlDB.SetMaxOpenConns(dal.dbConfig.MaxOpenConns)

		// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
		sqlDB.SetConnMaxLifetime(dal.dbConfig.ConnectTimeout)

		dal.dbInstance = db
	})
	return dal.dbInstance, err
}

func (dal *DataAccessLayer) Disconnect() error {
	sqlDB, err := dal.dbInstance.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (dal *DataAccessLayer) GetDatabase() *gorm.DB {
	return dal.dbInstance
}
