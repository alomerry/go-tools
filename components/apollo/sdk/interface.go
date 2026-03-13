package sdk

type FullQuery interface {
  AppInfo
  ClusterInfo
  NamespaceInfo
}

type NamespaceQuery interface {
  ClusterInfo
  AppInfo
}

type AppInfo interface {
  GetAppId() string
}

type NamespaceInfo interface {
  GetNamespace() string
}

type ClusterInfo interface {
  GetCluster() string
  GetEnv() string
}