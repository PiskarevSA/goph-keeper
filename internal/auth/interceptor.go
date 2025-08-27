package auth

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const userIDKey contextKey = "userID"

func UnaryAuthInterceptor(
	jwtManager *JWTManager, exemptMethods map[string]bool,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if exemptMethods[info.FullMethod] {
			// Register и Login идут без проверки
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata missing")
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization header missing")
		}

		token := strings.TrimPrefix(authHeader[0], "Bearer ")
		claims, err := jwtManager.Verify(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		// Кладём userID в context
		ctx = context.WithValue(ctx, userIDKey, claims.UserID)

		return handler(ctx, req)
	}
}

// Достаём userID внутри handler-а
func UserIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(userIDKey).(int64)
	return id, ok
}
