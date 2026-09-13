package mongo

import (
  "context"
  "fmt"
  "time"

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
		// 返回 error 而非 panic；带具体 err（原实现 panic 且 Ping 失败时丢 err）
		return nil, fmt.Errorf("init mongo client failed: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	if err = client.Ping(pingCtx, readpref.Primary()); err != nil {
		// Ping 失败时释放底层连接，避免泄漏
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("mongo ping failed: %w", err)
	}

	return &Mongo{
		client: client,
	}, nil
}