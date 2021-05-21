package mongodb

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/1412335/grpc-rest-microservice/pkg/configs"
	"github.com/1412335/grpc-rest-microservice/pkg/dal/errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// https://github.com/mongodb/mongo-go-driver/blob/master/examples/documentation_examples/examples.go
// https://github.com/simagix/mongo-go-examples/tree/master/examples

//Used to execute client creation procedure only once.
var mongoOnce sync.Once

type DataAccessLayer struct {
	dbConfig *configs.MongoDB
	/* Used to create a singleton object of MongoDB client.
	   Initialized and exposed through  GetMongoClient().*/
	clientInstance *mongo.Client
	// Used during creation of singleton client object in GetMongoClient().
	dbInstance *mongo.Database
}

func NewDataAccessLayer(ctx context.Context, config *configs.MongoDB) (*DataAccessLayer, error) {
	dal := &DataAccessLayer{
		dbConfig: config,
	}
	if _, err := dal.GetClient(ctx); err != nil {
		return nil, err
	}

	return dal, nil
}

func (dal *DataAccessLayer) buildConnectionURI() (string, error) {
	return dal.dbConfig.ConnectionURI, nil
}

//GetClient - Return mongodb connection to work with
func (dal *DataAccessLayer) GetClient(ctx context.Context) (interface{}, error) {
	// Perform connection creation operation only once.
	var err error
	mongoOnce.Do(func() {
		// get uri
		uri, e := dal.buildConnectionURI()
		if e != nil {
			err = e
			return
		}
		// Set client options
		clientOptions := options.Client().ApplyURI(uri).
			SetMaxPoolSize(dal.dbConfig.PoolSize).
			SetMaxConnIdleTime(100 * time.Millisecond).
			SetConnectTimeout(10 * time.Second)
		if dal.dbConfig.Auth != nil {
			clientOptions.SetAuth(options.Credential{
				AuthSource: dal.dbConfig.Auth.Source,
				Username:   dal.dbConfig.Auth.User,
				Password:   dal.dbConfig.Auth.Password,
			})
		}
		// Connect to MongoDB
		if client, e := mongo.Connect(ctx, clientOptions); e != nil {
			err = e
		} else {
			// Check the connection
			err = client.Ping(ctx, nil)
			if err != nil {
				return
			}
			dal.clientInstance = client
			// get database instance
			err = dal.GetDatabase()
		}
	})
	return dal.clientInstance, err
}

func (dal *DataAccessLayer) Disconnect(ctx context.Context) error {
	return dal.clientInstance.Disconnect(ctx)
}

func (dal *DataAccessLayer) GetDatabase() error {
	// Get MongoDB connection using connectionhelper.
	if dal.clientInstance == nil {
		return errors.ErrConnectDB
	}
	dal.dbInstance = dal.clientInstance.Database(dal.dbConfig.Database)
	return nil
}

func (dal *DataAccessLayer) GetCollection(collectionName string) (*mongo.Collection, error) {
	return dal.dbInstance.Collection(collectionName), nil
}

// Insert a new document in the collection.
func (dal *DataAccessLayer) Insert(ctx context.Context, collectionName string, data interface{}) (interface{}, error) {
	//Create a handle to the respective collection in the database.
	collection := dal.dbInstance.Collection(collectionName)
	//Perform InsertOne operation & validate against the error.
	result, err := collection.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	//Return success without any error.
	return result.InsertedID, nil
}

// InsertMany multiple documents at once in the collection.
func (dal *DataAccessLayer) InsertMany(ctx context.Context, collectionName string, data []interface{}) ([]interface{}, error) {
	//Create a handle to the respective collection in the database.
	collection := dal.dbInstance.Collection(collectionName)
	//Perform InsertMany operation & validate against the error.
	result, err := collection.InsertMany(ctx, data)
	if err != nil {
		return nil, err
	}
	//Return success without any error.
	return result.InsertedIDs, nil
}

// pre update
func (dal *DataAccessLayer) update(data interface{}, upsert bool) (interface{}, *options.UpdateOptions, error) {
	// options
	opts := options.Update()
	opts.SetUpsert(upsert)
	// updatedAt
	dataMd, ok := data.(bson.M)
	if !ok {
		bytes, err := json.Marshal(data)
		if err != nil {
			return nil, nil, err
		}
		if err := json.Unmarshal(bytes, &dataMd); err != nil {
			return nil, nil, err
		}
	}
	dataMd["$currentDate"] = bson.M{
		"updatedAt": true,
	}
	return dataMd, opts, nil
}

// Update a document in the collection.
func (dal *DataAccessLayer) Update(ctx context.Context, collectionName string, filter interface{}, data interface{}, upsert bool) error {
	//Create a handle to the respective collection in the database.
	collection := dal.dbInstance.Collection(collectionName)
	dataMd, opts, err := dal.update(data, upsert)
	if err != nil {
		return err
	}
	//Perform InsertOne operation & validate against the error.
	_, err = collection.UpdateOne(ctx, filter, dataMd, opts)
	if err != nil {
		return err
	}
	//Return success without any error.
	return nil
}

// Update a document in the collection.
func (dal *DataAccessLayer) UpdateMany(ctx context.Context, collectionName string, filter interface{}, data interface{}, upsert bool) error {
	//Create a handle to the respective collection in the database.
	collection := dal.dbInstance.Collection(collectionName)
	dataMd, opts, err := dal.update(data, upsert)
	if err != nil {
		return err
	}
	// Perform InsertOne operation & validate against the error.
	_, err = collection.UpdateMany(ctx, filter, dataMd, opts)
	if err != nil {
		return err
	}
	//Return success without any error.
	return nil
}

// Delete a document in the collection.
func (dal *DataAccessLayer) Delete(ctx context.Context, collectionName string, filter interface{}) error {
	// Create a handle to the respective collection in the database.
	collection := dal.dbInstance.Collection(collectionName)
	// Perform DeleteOne operation & validate against the error.
	_, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	//Return success without any error.
	return nil
}

// Delete documents in the collection.
func (dal *DataAccessLayer) DeleteMany(ctx context.Context, collectionName string, filter interface{}) error {
	// Create a handle to the respective collection in the database.
	collection := dal.dbInstance.Collection(collectionName)
	// Perform DeleteMany operation & validate against the error.
	_, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}
	//Return success without any error.
	return nil
}

// Find documents in the collection.
func (dal *DataAccessLayer) Find(ctx context.Context, collectionName string, filter, projection, sort interface{}, offset, limit int64) ([]interface{}, error) {
	// Create a handle to the respective collection in the database.
	collection := dal.dbInstance.Collection(collectionName)
	// options
	opts := options.Find()
	opts.SetSkip(offset)
	opts.SetLimit(limit)
	if projection != nil {
		opts.SetProjection(projection)
	}
	if sort != nil {
		opts.SetSort(sort)
	}
	// Perform find operation & validate against the error.
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	// Map result to slice
	var data []interface{}
	for cursor.Next(ctx) {
		item := map[string]interface{}{}
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		// For some reason Decode doesn't work for the _id field. Extract separately.
		// item.ID = cur.Current.Lookup("_id").ObjectID()
		data = append(data, item)
	}
	// once exhausted, close the cursor
	if err := cursor.Close(ctx); err != nil {
		return nil, err
	}
	//Return result without any error.
	return data, nil
}

// Find all documents in the collection.
func (dal *DataAccessLayer) FindAll(ctx context.Context, collectionName string, projection, sort interface{}) ([]interface{}, error) {
	// Create a handle to the respective collection in the database.
	collection := dal.dbInstance.Collection(collectionName)
	// options
	opts := options.Find()
	if projection != nil {
		opts.SetProjection(projection)
	}
	if sort != nil {
		opts.SetSort(sort)
	}
	// filter
	filter := bson.D{}
	// Perform find operation & validate against the error.
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	// Map result to slice
	var data []interface{}
	for cursor.Next(ctx) {
		item := map[string]interface{}{}
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		data = append(data, item)
	}
	// once exhausted, close the cursor
	if err := cursor.Close(ctx); err != nil {
		return nil, err
	}
	//Return result without any error.
	return data, nil
}

// Find a document in the collection.
func (dal *DataAccessLayer) FindOne(ctx context.Context, collectionName string, filter, projection interface{}) (interface{}, error) {
	// Create a handle to the respective collection in the database.
	collection := dal.dbInstance.Collection(collectionName)
	// options
	opts := options.FindOne()
	if projection != nil {
		opts.SetProjection(projection)
	}
	data := map[string]interface{}{}
	// Perform find operation & validate against the error.
	err := collection.FindOne(ctx, filter, opts).Decode(&data)
	if err != nil {
		return nil, err
	}
	//Return result without any error.
	return data, nil
}

// Distinct documents in the collection.
func (dal *DataAccessLayer) Distinct(ctx context.Context, collectionName string, field string, filter interface{}) ([]interface{}, error) {
	// Create a handle to the respective collection in the database.
	collection := dal.dbInstance.Collection(collectionName)
	// Perform find operation & validate against the error.
	return collection.Distinct(ctx, field, filter)
}

// Find document by ObjectID in the collection.
func (dal *DataAccessLayer) FindByID(ctx context.Context, collectionName string, id interface{}, projection interface{}) (interface{}, error) {
	// filter
	objectID, ok := id.(primitive.ObjectID)
	if !ok {
		var err error
		objectID, err = primitive.ObjectIDFromHex(fmt.Sprintf("%v", id))
		if err != nil {
			return nil, err
		}
	}
	filter := bson.D{
		primitive.E{Key: "_id", Value: objectID},
	}
	// Perform find operation & validate against the error.
	return dal.FindOne(ctx, collectionName, filter, projection)
}

// Aggregate
func (dal *DataAccessLayer) Aggregate(ctx context.Context, collectionName string, filter, projection, sort interface{}, offset, limit int64) ([]interface{}, error) {
	// Create a handle to the respective collection in the database.
	collection := dal.dbInstance.Collection(collectionName)
	// create pipeline
	pipeline := mongo.Pipeline{}
	if filter != nil {
		pipeline = append(pipeline, bson.D{
			primitive.E{Key: "$match", Value: filter},
		})
	}
	if projection != nil {
		pipeline = append(pipeline, bson.D{
			primitive.E{Key: "$project", Value: projection},
		})
	}
	if sort != nil {
		pipeline = append(pipeline, bson.D{
			primitive.E{Key: "$sort", Value: sort},
		})
	}
	pipeline = append(pipeline, bson.D{
		primitive.E{Key: "$skip", Value: offset},
	})
	pipeline = append(pipeline, bson.D{
		primitive.E{Key: "$limit", Value: limit},
	})
	// Perform find operation & validate against the error.
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	// Map result to slice
	var data []interface{}
	for cursor.Next(ctx) {
		item := map[string]interface{}{}
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		data = append(data, item)
	}
	// once exhausted, close the cursor
	if err := cursor.Close(ctx); err != nil {
		return nil, err
	}
	//Return result without any error.
	return data, nil
}

// MongoPipeline gets aggregation pipeline from a string
func (dal *DataAccessLayer) parsePipelineJSON(str string) (mongo.Pipeline, error) {
	var pipeline = []bson.D{}
	str = strings.TrimSpace(str)
	if strings.Index(str, "[") != 0 {
		var doc bson.D
		if err := bson.UnmarshalExtJSON([]byte(str), false, &doc); err != nil {
			return nil, err
		}
		pipeline = append(pipeline, doc)
	} else {
		if err := bson.UnmarshalExtJSON([]byte(str), false, &pipeline); err != nil {
			return nil, err
		}
	}
	return pipeline, nil
}

// Aggregate common
func (dal *DataAccessLayer) AggregateCommon(ctx context.Context, collectionName string, pipeline interface{}) ([]interface{}, error) {
	// Create a handle to the respective collection in the database.
	collection := dal.dbInstance.Collection(collectionName)
	// transform pipeline JSON string
	if _, ok := pipeline.(string); ok {
		var err error
		pipeline, err = dal.parsePipelineJSON(pipeline.(string))
		if err != nil {
			return nil, err
		}
	}
	// Perform find operation & validate against the error.
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	// Map result to slice
	var data []interface{}
	for cursor.Next(ctx) {
		item := map[string]interface{}{}
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		data = append(data, item)
	}
	// once exhausted, close the cursor
	if err := cursor.Close(ctx); err != nil {
		return nil, err
	}
	// Return result without any error.
	return data, nil
}

// Count documents in the collection.
func (dal *DataAccessLayer) Counts(ctx context.Context, collectionName string, filter interface{}) (int64, error) {
	// Create a handle to the respective collection in the database.
	collection := dal.dbInstance.Collection(collectionName)
	// Perform find operation & validate against the error.
	return collection.CountDocuments(ctx, filter)
}
