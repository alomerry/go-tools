package model

// Config 存储配置
type Config struct {
  Type       ClientType // 客户端类型
  Endpoint   string     // 服务端点
  AccessKey  string     // 访问密钥
  SecretKey  string     // 秘密密钥
  Region     string     // 区域（S3 需要）
  AccountId  string     // 账户 ID（R2 需要）
  SSL        bool       // 是否使用 SSL
  BucketName string     // 默认存储桶名称
}

// ClientType 存储类型枚举
type ClientType string

const (
  ClientTypeMinio  ClientType = "minio"
  ClientTypeR2     ClientType = "r2"
  ClientTypeKodo   ClientType = "kodo"
  ClientTypeRustFs ClientType = "rust-fs"
  ClientTypeS3     ClientType = "s3"
)

func ToConfig(clientType ClientType) Config {
  return Config{
    Type: clientType,
  }
}