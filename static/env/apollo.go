package env

import (
  "os"
  
  "github.com/alomerry/go-tools/static/cons"
  "github.com/alomerry/go-tools/static/cons/apollo"
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
    namespace = apollo.DefaultApplication
	}
	return namespace
}

func ApolloCluster() string {
	cluster := os.Getenv(cons.ApolloCluster)
	if cluster == "" {
    cluster = apollo.DefaultCluster
	}
	return cluster
}

func ApolloEnv() string {
  env := os.Getenv(cons.ApolloEnv)
  if env == "" {
    env = apollo.DefaultEnv
  }
  return env
}

func ApolloOpenapiToken() string {
  return os.Getenv(cons.ApolloOpenapiToken)
}