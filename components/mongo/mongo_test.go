package mongo

import (
	"context"
	"testing"

	"github.com/alomerry/go-tools/static/env"
	"github.com/alomerry/go-tools/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
  "go.mongodb.org/mongo-driver/v2/bson"
  "go.mongodb.org/mongo-driver/v2/mongo"
  "go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	db         = "homelab"
	collection = "test"
)

func TestMongoSuit(t *testing.T) {
	memberSuite := new(MongoSuit)

	suite.Run(t, memberSuite)
}

type MongoSuit struct {
	test.BaseSuite
	client *Mongo
}

func (s *MongoSuit) TearDownTest() {
  s.BaseSuite.TearDownTest()
}


func (s *MongoSuit) SetupSuite() {
  s.Setup()
  
	s.client = s.newClient()

	err := s.client.Client().Database(db).Collection(collection).Drop(context.Background())
	assert.NoError(s.T(), err)

	err = s.client.Client().Database(db).CreateCollection(context.Background(), collection)
	assert.NoError(s.T(), err)
}

func (s *MongoSuit) TearDownSuite() {
  s.TearDown()
  err := s.client.Client().Database(db).Collection("testCount").Drop(context.Background())
  assert.NoError(s.T(), err)
  
  err = s.client.Client().Database(db).Collection("test").Drop(context.Background())
  assert.NoError(s.T(), err)
  
	err = s.client.Close(context.Background())
	assert.NoError(s.T(), err, "close mongo client failed")
}

func (s *MongoSuit) newClient() *Mongo {
	cli, err := NewMongoClient(context.TODO(), env.GetMongoDSN())
	if err != nil {
		panic(err)
	}

	cli.Client()
	return cli
}

func (s *MongoSuit) TestCount() {
  ctx := context.Background()
  id := bson.NewObjectID()
  total, err := s.client.Client().Database(db).Collection(collection).CountDocuments(ctx, bson.M{"_id": id})
  assert.NoError(s.T(), err)
  assert.Equal(s.T(), int64(0), total)
  
  res, err := s.client.Client().Database(db).Collection(collection).InsertOne(ctx, bson.M{"name": "test", "_id": id})
  assert.NoError(s.T(), err)
  assert.Equal(s.T(), id.String(), res.InsertedID.(bson.ObjectID).String())
  
  total, err = s.client.Client().Database(db).Collection(collection).CountDocuments(ctx, bson.M{"_id": res.InsertedID})
  assert.NoError(s.T(), err)
  assert.Equal(s.T(), int64(1), total)
}

func (s *MongoSuit) TestFindByPage() {
    var (
      ctx = context.Background()
      id1 = bson.NewObjectID()
      id2 = bson.NewObjectID()
    )
    
    res, err := s.client.Client().Database(db).Collection(collection).InsertOne(ctx, bson.M{"name": "FindByPage", "_id": id1})
    assert.NoError(s.T(), err)
    assert.Equal(s.T(), id1.String(), res.InsertedID.(bson.ObjectID).String())
    res, err = s.client.Client().Database(db).Collection(collection).InsertOne(ctx, bson.M{"name": "FindByPage", "_id": id2})
    assert.NoError(s.T(), err)
    assert.Equal(s.T(), id2.String(), res.InsertedID.(bson.ObjectID).String())
    
    findOptions := options.Find().SetSkip(1).SetLimit(1).SetSort(bson.M{"_id": -1})
    selector :=bson.M{"name": "FindByPage"}
    cursor, err := s.client.Client().Database(db).Collection(collection).Find(ctx, selector, findOptions)
    assert.NoError(s.T(), err)
    
    results := []bson.M{}
    
    err = cursor.All(ctx, &results)
    assert.NoError(s.T(), err)
    assert.Len(s.T(), results, 1)
    assert.Equal(s.T(), id1.String(), results[0]["_id"].(bson.ObjectID).String())
}

func (s *MongoSuit) TestFindOne() {
  var (
    ctx = context.Background()
    id1 = bson.NewObjectID()
  )
  
  
  selector :=bson.M{"name": "FindOne","_id": id1}
  res := s.client.Client().Database(db).Collection(collection).FindOne(ctx, selector)
  assert.EqualError(s.T(), res.Err(), mongo.ErrNoDocuments.Error())
}