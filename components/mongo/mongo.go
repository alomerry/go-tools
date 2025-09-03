package mongo

import (
	"context"
	"github.com/qiniu/qmgo"
	"log"
)

type Mongo struct {
	client *qmgo.Client
}

func NewMongoDB(ctx context.Context, uri string) (*Mongo, error) {
	client, err := qmgo.NewClient(ctx, &qmgo.Config{
		Uri: uri,
	})
	if err != nil {
		log.Panicf("init mongo client failed, err: %v", err.Error())
	}

	if err := client.Ping(5); err != nil {
		log.Panicf("can't connect mongodb")
	}

	return &Mongo{
		client: client,
	}, nil
}

func (m *Mongo) Client() *qmgo.Client {
	return m.client
}
