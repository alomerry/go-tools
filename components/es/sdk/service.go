package sdk

import (
	"github.com/alomerry/go-tools/components/es"
	"github.com/elastic/go-elasticsearch/v8"
)

type Service struct {
	client *elasticsearch.TypedClient
}

func NewService(esClient *es.Client) *Service {
	return &Service{
		client: esClient.GetEs(),
	}
}
