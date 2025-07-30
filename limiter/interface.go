package limiter

import "context"

type Limiter interface {
	// 检查是否可以处理
	CanProcess(ctx context.Context, userID, resourceID string) bool
	// 处理请求
	Process(ctx context.Context, userID, resourceID string) bool
	// 完成请求
	Finish(ctx context.Context, userID, resourceID string)
}
