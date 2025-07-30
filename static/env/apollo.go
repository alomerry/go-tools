package env

import (
	"github.com/alomerry/go-tools/static/cons"
	"os"
)

func ApolloHost() string {
	host := os.Getenv(cons.ApolloHost)
	if host == "" {
		host = "https://apollo-cfg.alomerry.cn"
	}
	return host
}

func ApolloSK() string {
	return os.Getenv(cons.ApolloSK)
}

func ApolloNamespace() string {
	namespace := os.Getenv(cons.ApolloNamespace)
	if namespace == "" {
		namespace = "application"
	}
	return namespace
}

func ApolloCluster() string {
	cluster := os.Getenv(cons.ApolloCluster)
	if cluster == "" {
		cluster = "default"
	}
	return cluster
}
