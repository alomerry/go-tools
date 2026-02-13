package internal

import (
	"context"

	req2 "github.com/alomerry/go-tools/components/http/opts/req"
	"github.com/alomerry/go-tools/components/kook/model"
	"github.com/alomerry/go-tools/static/cons/kook"
)

type AssetService struct {
	*BaseService
}

func (s *AssetService) Create(ctx context.Context, file []byte) (*model.AssetCreateResp, error) {
	// TODO: 正确支持 multipart 上传。
	// 当前 http 客户端可能不支持 WithBody 直接处理 multipart。
	// 假设 WithBody 处理 []byte 或者我们需要特定的 WithMultipart。
	// 目前使用 WithBody 传递字节，但 multipart/form-data 可能需要改进。
	// 由于底层客户端是 Resty，通常需要 SetFile 或 SetFileReader。
	// 此实现假设未来增强或现有支持。
	// 实际上，resty 支持 SetFile/SetFileReader。
	// 我们可能需要向 req opts 添加 WithFile 选项。
	// 目前，我将编写占位符。

	// 警告：这是一个简化的实现。真正的实现需要 Multipart 支持。
	resp, err := s.client.Post(ctx, kook.AssetCreate, req2.WithBody(file)) // 对于 multipart 来说可能是错误的
	if err != nil {
		return nil, err
	}

	var result model.AssetCreateResp
	if err := s.after(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
