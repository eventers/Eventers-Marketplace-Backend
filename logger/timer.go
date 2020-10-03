package logger

import (
	"context"
	"fmt"
	"time"
)

func LogExecutionTime(ctx context.Context, startTime time.Time, msg string) {
	Info(ctx, fmt.Sprintf("%s, Took: %vms", msg, time.Since(startTime).Milliseconds()))
}
