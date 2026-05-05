package db_service

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type DbService[DocType interface{}] interface {
	CreateDocument(ctx context.Context, id string, document *DocType) error
	FindDocument(ctx context.Context, id string) (*DocType, error)
	ListDocuments(ctx context.Context) ([]DocType, error)
	UpdateDocument(ctx context.Context, id string, document *DocType) error
	DeleteDocument(ctx context.Context, id string) error
	Disconnect(ctx context.Context) error
}

var ErrNotFound = fmt.Errorf("document not found")
var ErrConflict = fmt.Errorf("conflict: document already exists")

type MongoServiceConfig struct {
	ServerHost string
	ServerPort int
	UserName   string
	Password   string
	DbName     string
	Collection string
	Timeout    time.Duration
}

type mongoSvc[DocType interface{}] struct {
	MongoServiceConfig
	client     atomic.Pointer[mongo.Client]
	clientLock sync.Mutex
}

func NewMongoService[DocType interface{}](config MongoServiceConfig) DbService[DocType] {
	enviro := func(name string, defaultValue string) string {
		if value, ok := os.LookupEnv(name); ok {
			return value
		}
		return defaultValue
	}

	svc := &mongoSvc[DocType]{}
	svc.MongoServiceConfig = config

	if svc.ServerHost == "" {
		svc.ServerHost = enviro("MEDEDU_API_MONGODB_HOST", "localhost")
	}

	if svc.ServerPort == 0 {
		port := enviro("MEDEDU_API_MONGODB_PORT", "27018")
		if port, err := strconv.Atoi(port); err == nil {
			svc.ServerPort = port
		} else {
			log.Printf("Invalid MongoDB port value: %v", port)
			svc.ServerPort = 27018
		}
	}

	if svc.UserName == "" {
		svc.UserName = enviro("MEDEDU_API_MONGODB_USERNAME", "")
	}

	if svc.Password == "" {
		svc.Password = enviro("MEDEDU_API_MONGODB_PASSWORD", "")
	}

	if svc.DbName == "" {
		svc.DbName = enviro("MEDEDU_API_MONGODB_DATABASE", "kcrp-mededu")
	}

	if svc.Collection == "" {
		svc.Collection = enviro("MEDEDU_API_MONGODB_COLLECTION", "trainings")
	}

	if svc.Timeout == 0 {
		seconds := enviro("MEDEDU_API_MONGODB_TIMEOUT_SECONDS", "10")
		if seconds, err := strconv.Atoi(seconds); err == nil {
			svc.Timeout = time.Duration(seconds) * time.Second
		} else {
			log.Printf("Invalid MongoDB timeout value: %v", seconds)
			svc.Timeout = 10 * time.Second
		}
	}

	log.Printf(
		"MongoDB config: //%v@%v:%v/%v/%v",
		svc.UserName,
		svc.ServerHost,
		svc.ServerPort,
		svc.DbName,
		svc.Collection,
	)
	return svc
}

func (m *mongoSvc[DocType]) connect(ctx context.Context) (*mongo.Client, error) {
	client := m.client.Load()
	if client != nil {
		return client, nil
	}

	m.clientLock.Lock()
	defer m.clientLock.Unlock()

	client = m.client.Load()
	if client != nil {
		return client, nil
	}

	ctx, contextCancel := context.WithTimeout(ctx, m.Timeout)
	defer contextCancel()

	uri := fmt.Sprintf("mongodb://%v:%v", m.ServerHost, m.ServerPort)
	clientOptions := options.Client().
		ApplyURI(uri).
		SetConnectTimeout(m.Timeout).
		SetServerSelectionTimeout(m.Timeout)

	if m.UserName != "" {
		clientOptions.SetAuth(options.Credential{
			AuthSource: "admin",
			Username:   m.UserName,
			Password:   m.Password,
		})
	}

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}

	m.client.Store(client)
	return client, nil
}

func (m *mongoSvc[DocType]) Disconnect(ctx context.Context) error {
	client := m.client.Load()
	if client == nil {
		return nil
	}

	m.clientLock.Lock()
	defer m.clientLock.Unlock()

	client = m.client.Load()
	defer m.client.Store(nil)
	if client != nil {
		return client.Disconnect(ctx)
	}
	return nil
}

func (m *mongoSvc[DocType]) collection(ctx context.Context) (*mongo.Collection, error) {
	client, err := m.connect(ctx)
	if err != nil {
		return nil, err
	}
	return client.Database(m.DbName).Collection(m.Collection), nil
}

func (m *mongoSvc[DocType]) CreateDocument(ctx context.Context, id string, document *DocType) error {
	ctx, contextCancel := context.WithTimeout(ctx, m.Timeout)
	defer contextCancel()

	collection, err := m.collection(ctx)
	if err != nil {
		return err
	}

	result := collection.FindOne(ctx, bson.D{{Key: "id", Value: id}})
	switch result.Err() {
	case nil:
		return ErrConflict
	case mongo.ErrNoDocuments:
	default:
		return result.Err()
	}

	_, err = collection.InsertOne(ctx, document)
	return err
}

func (m *mongoSvc[DocType]) FindDocument(ctx context.Context, id string) (*DocType, error) {
	ctx, contextCancel := context.WithTimeout(ctx, m.Timeout)
	defer contextCancel()

	collection, err := m.collection(ctx)
	if err != nil {
		return nil, err
	}

	result := collection.FindOne(ctx, bson.D{{Key: "id", Value: id}})
	switch result.Err() {
	case nil:
	case mongo.ErrNoDocuments:
		return nil, ErrNotFound
	default:
		return nil, result.Err()
	}

	var document DocType
	if err := result.Decode(&document); err != nil {
		return nil, err
	}
	return &document, nil
}

func (m *mongoSvc[DocType]) ListDocuments(ctx context.Context) ([]DocType, error) {
	ctx, contextCancel := context.WithTimeout(ctx, m.Timeout)
	defer contextCancel()

	collection, err := m.collection(ctx)
	if err != nil {
		return nil, err
	}

	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var documents []DocType
	if err := cursor.All(ctx, &documents); err != nil {
		return nil, err
	}
	if documents == nil {
		documents = []DocType{}
	}
	return documents, nil
}

func (m *mongoSvc[DocType]) UpdateDocument(ctx context.Context, id string, document *DocType) error {
	ctx, contextCancel := context.WithTimeout(ctx, m.Timeout)
	defer contextCancel()

	collection, err := m.collection(ctx)
	if err != nil {
		return err
	}

	result, err := collection.ReplaceOne(ctx, bson.D{{Key: "id", Value: id}}, document)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (m *mongoSvc[DocType]) DeleteDocument(ctx context.Context, id string) error {
	ctx, contextCancel := context.WithTimeout(ctx, m.Timeout)
	defer contextCancel()

	collection, err := m.collection(ctx)
	if err != nil {
		return err
	}

	result, err := collection.DeleteOne(ctx, bson.D{{Key: "id", Value: id}})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}
