package grpcapi

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// mdValueFromContext returns value of field from provided context.
func mdValueFromContext(ctx context.Context, field string) (string, bool) {
	var none string

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if values := md.Get(field); len(values) > 0 {
			return values[0], true
		}
	}

	return none, false
}
