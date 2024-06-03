package limiter

import "context"

type Limiter interface {
	CanProcess(ctx context.Context, userID, resourceID string) bool
	Process(ctx context.Context, userID, resourceID string) bool
	Finish(ctx context.Context, userID, resourceID string)
}
