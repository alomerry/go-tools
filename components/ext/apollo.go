package ext

import (
  "context"
  "sync"
  
  "github.com/alomerry/go-tools/components/apollo"
  "github.com/alomerry/go-tools/components/apollo/mananger"
  "github.com/alomerry/go-tools/static/cons"
  apollo2 "github.com/alomerry/go-tools/static/cons/apollo"
)

type ApolloExtension interface {
  Ext
  mananger.Manager
}

var (
  apolloExt  ApolloExtension
  apolloOnce sync.Once
)

func init() {
  Register(cons.ExtApollo, func() Ext {
    return Apollo()
  })
}

type apolloExtension struct {
  mananger.Manager
}

func Apollo() ApolloExtension {
  apolloOnce.Do(func() {
    apolloExt = &apolloExtension{
      Manager: mananger.ApolloManager(),
    }
  })
  return apolloExt
}

func (*apolloExtension) Init(_ context.Context) error {
  apollo.Init(apollo2.DefaultNamespace, apollo2.DefaultApp)
  return nil
}
