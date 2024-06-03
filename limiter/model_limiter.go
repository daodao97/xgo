package limiter

import (
	"context"
)

// 模型限制器

type ModelLimiter struct {
	AllowedResources map[string]bool
}

func (l *ModelLimiter) CanProcess(ctx context.Context, userID, resourceID string) bool {
	return l.AllowedResources[resourceID]
}

func (l *ModelLimiter) Process(ctx context.Context, userID, resourceID string) bool {
	return l.CanProcess(ctx, userID, resourceID)
}

func (l *ModelLimiter) Finish(ctx context.Context, userID, resourceID string) {
	// No specific finish action needed for ModelLimiter
}
