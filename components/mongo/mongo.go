package mongo

import (
  "context"
  "time"
  
  "github.com/alomerry/go-tools/components/log"
  "go.mongodb.org/mongo-driver/v2/mongo"
  "go.mongodb.org/mongo-driver/v2/mongo/options"
  "go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type Client interface {
}

type Mongo struct {
  client *mongo.Client
}

func (m *Mongo) Client() *mongo.Client {
  return m.client
}

func (m *Mongo) Close(ctx context.Context) error {
  err := m.client.Disconnect(ctx)
  if err != nil {
    return err
  }
  return nil
}

func NewMongoClient(ctx context.Context, uri string) (*Mongo, error) {
  client, err := mongo.Connect(options.Client().ApplyURI(uri))
  if err != nil {
    log.Panicf(ctx, "init mongo client failed, err: %v", err.Error())
    return nil, err
  }
  ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
  defer cancel()
  if err = client.Ping(ctx, readpref.Primary()); err != nil {
    log.Panicf(ctx, "can't connect mongodb")
    return nil, err
  }
  
  return &Mongo{
    client: client,
  }, nil
}