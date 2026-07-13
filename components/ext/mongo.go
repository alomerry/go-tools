package ext

import (
	"context"
	"log"
	"sync"

	"github.com/alomerry/go-tools/components/apollo"
	"github.com/alomerry/go-tools/components/mongo"
	"github.com/alomerry/go-tools/static/cons"
	apollo2 "github.com/alomerry/go-tools/static/cons/apollo"
	"github.com/alomerry/go-tools/static/env"
	"go.mongodb.org/mongo-driver/v2/bson"
	mongodriver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type CollectionDBResolver func(collectionName string) string

var (
	defaultCollectionDBResolver CollectionDBResolver = func(collectionName string) string {
		return collectionName
	}

	MgoExt   *MongoExt
	mongOnce sync.Once
)

// SetCollectionDBResolver overrides the collection->db resolver used by MongoExt.
func SetCollectionDBResolver(r CollectionDBResolver) {
	if r != nil {
		defaultCollectionDBResolver = r
	}
}

func init() {
	Register(cons.ExtMongoDB, NewMongoExt)
}

type MongoExt struct {
	cfg *MongoExtConfig
	cli *mongo.Mongo
}

type MongoExtConfig struct {
	Uri string `json:"uri"`
}

func NewMongoExt() Ext {
	mongOnce.Do(func() {
		MgoExt = &MongoExt{
			cfg: &MongoExtConfig{},
		}
	})
	return MgoExt
}

func (m *MongoExt) Init(ctx context.Context) error {
	err := apollo.GetJson(apollo2.ApolloKeyMongoCfg, m.cfg)
	if err != nil {
		log.Panicf("init mongodb failed %v", err.Error())
	}
	m.cli, err = mongo.NewMongoClient(ctx, env.GetMongoDSN(m.cfg.Uri))
	if err != nil {
		log.Panicf("create mongodb client failed %v", err.Error())
	}
	return nil
}

func (m *MongoExt) dbOf(collectionName string) string {
	return defaultCollectionDBResolver(collectionName)
}

func (m *MongoExt) EnsureIndex(ctx context.Context, collectionName string, models []mongodriver.IndexModel) error {
	col := m.cli.Client().Database(m.dbOf(collectionName)).Collection(collectionName)
	_, err := col.Indexes().CreateMany(ctx, models)
	return err
}

type DatabaseRepository interface {
	FindAll(ctx context.Context, collectionName string, selector bson.M, result any) error
	FindOne(ctx context.Context, collectionName string, selector bson.M, result any) error
	UpsertOne(ctx context.Context, collectionName string, selector bson.M, update bson.M) error
	UpdateMany(ctx context.Context, collectionName string, selector bson.M, update bson.M) error
	DeleteMany(ctx context.Context, collectionName string, selector bson.M) error
	Insert(ctx context.Context, collectionName string, data any) error
}

func (m *MongoExt) FindAll(ctx context.Context, collectionName string, selector bson.M, result any) error {
	cursor, err := m.cli.Client().Database(m.dbOf(collectionName)).Collection(collectionName).Find(ctx, selector)
	if err != nil {
		return err
	}

	return cursor.All(ctx, result)
}

func (m *MongoExt) FindPage(ctx context.Context, collectionName string, selector bson.M, result any, page, limit int64) (int64, error) {
	databaseName := m.dbOf(collectionName)

	if page == 0 {
		page = 1
	}

	if limit == 0 {
		limit = 10
	}

	// sort to keep paging order stable
	findOptions := options.Find().SetSkip((page - 1) * limit).SetLimit(limit).SetSort(bson.D{{Key: "_id", Value: 1}})
	cursor, err := m.cli.Client().Database(databaseName).Collection(collectionName).Find(ctx, selector, findOptions)
	if err != nil {
		return 0, err
	}
	total, err := m.Count(ctx, collectionName, selector)
	if err != nil {
		return 0, err
	}

	err = cursor.All(ctx, result)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (m *MongoExt) Count(ctx context.Context, collectionName string, selector bson.M) (int64, error) {
	total, err := m.cli.Client().Database(m.dbOf(collectionName)).Collection(collectionName).CountDocuments(ctx, selector)
	if err != nil {
		return 0, err
	}
	return total, err
}

// Aggregate runs an aggregation pipeline against the collection and decodes all
// results into out (a slice). Mirrors the FindAll decode contract.
func (m *MongoExt) Aggregate(ctx context.Context, collectionName string, pipeline any, out any) error {
	cursor, err := m.cli.Client().Database(m.dbOf(collectionName)).Collection(collectionName).Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	return cursor.All(ctx, out)
}

func (m *MongoExt) FindOne(ctx context.Context, collectionName string, selector bson.M, result any) error {
	return m.cli.Client().Database(m.dbOf(collectionName)).Collection(collectionName).FindOne(ctx, selector).Decode(result)
}

func (m *MongoExt) UpsertOne(ctx context.Context, collectionName string, selector, updater bson.M) error {
	databaseName := m.dbOf(collectionName)
	_, err := m.cli.Client().Database(databaseName).Collection(collectionName).UpdateOne(ctx, selector, updater, options.UpdateOne().SetUpsert(true))
	return err
}

func (m *MongoExt) UpdateOne(ctx context.Context, collectionName string, selector, updater bson.M) error {
	_, err := m.cli.Client().Database(m.dbOf(collectionName)).Collection(collectionName).UpdateOne(ctx, selector, updater, options.UpdateOne().SetUpsert(false))
	return err
}

func (m *MongoExt) UpdateMany(ctx context.Context, collectionName string, selector, updater bson.M) error {
	databaseName := m.dbOf(collectionName)
	_, err := m.cli.Client().Database(databaseName).Collection(collectionName).UpdateMany(ctx, selector, updater)
	return err
}

func (m *MongoExt) DeleteMany(ctx context.Context, collectionName string, selector bson.M) error {
	databaseName := m.dbOf(collectionName)
	_, err := m.cli.Client().Database(databaseName).Collection(collectionName).DeleteMany(ctx, selector)
	return err
}

func (m *MongoExt) Insert(ctx context.Context, collectionName string, data any) error {
	databaseName := m.dbOf(collectionName)
	res, err := m.cli.Client().Database(databaseName).Collection(collectionName).InsertOne(ctx, data)
	if err != nil {
		return err
	}
	log.Printf("insert id %v", res.InsertedID)
	return nil
}
