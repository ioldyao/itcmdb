package middleware

import (
	"context"
	"strings"

	grpcclient "github.com/itcmdb/shared/pkg/grpc"
	"github.com/itcmdb/shared/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	// AgentTokenKey 用于标识Agent的token
	AgentTokenKey = "agent-token"
)

// UnaryAuthInterceptor gRPC一元拦截器（用于认证）
func UnaryAuthInterceptor(authClient *grpcclient.AuthClient) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// 从metadata获取token
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		// 获取authorization token
		tokens := md.Get("authorization")
		if len(tokens) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization token")
		}

		// 解析Bearer token
		authHeader := tokens[0]
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return nil, status.Error(codes.Unauthenticated, "invalid token format")
		}

		token := parts[1]

		// 验证token
		resp, err := authClient.ValidateToken(ctx, token)
		if err != nil {
			logger.Error("Failed to validate token",
				zap.Error(err),
				zap.String("method", info.FullMethod))
			return nil, status.Error(codes.Unauthenticated, "token validation failed")
		}

		if !resp.Valid {
			logger.Warn("Invalid token",
				zap.String("method", info.FullMethod),
				zap.String("error", resp.Error))
			return nil, status.Error(codes.Unauthenticated, resp.Error)
		}

		// 将用户信息添加到context
		ctx = context.WithValue(ctx, "user_id", resp.UserId)
		ctx = context.WithValue(ctx, "username", resp.Username)

		logger.Debug("gRPC request authenticated",
			zap.String("method", info.FullMethod),
			zap.Uint64("user_id", resp.UserId),
			zap.String("username", resp.Username))

		// 调用实际的handler
		return handler(ctx, req)
	}
}

// StreamAuthInterceptor gRPC流式拦截器（用于认证）
func StreamAuthInterceptor(authClient *grpcclient.AuthClient) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()

		// 从metadata获取token
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return status.Error(codes.Unauthenticated, "missing metadata")
		}

		// 获取authorization token
		tokens := md.Get("authorization")
		if len(tokens) == 0 {
			return status.Error(codes.Unauthenticated, "missing authorization token")
		}

		// 解析Bearer token
		authHeader := tokens[0]
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return status.Error(codes.Unauthenticated, "invalid token format")
		}

		token := parts[1]

		// 验证token
		resp, err := authClient.ValidateToken(ctx, token)
		if err != nil {
			logger.Error("Failed to validate token",
				zap.Error(err),
				zap.String("method", info.FullMethod))
			return status.Error(codes.Unauthenticated, "token validation failed")
		}

		if !resp.Valid {
			logger.Warn("Invalid token",
				zap.String("method", info.FullMethod),
				zap.String("error", resp.Error))
			return status.Error(codes.Unauthenticated, resp.Error)
		}

		// 将用户信息添加到context
		ctx = context.WithValue(ctx, "user_id", resp.UserId)
		ctx = context.WithValue(ctx, "username", resp.Username)

		logger.Debug("gRPC stream authenticated",
			zap.String("method", info.FullMethod),
			zap.Uint64("user_id", resp.UserId),
			zap.String("username", resp.Username))

		// 使用wrapped stream调用实际的handler
		wrapped := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		return handler(srv, wrapped)
	}
}

// wrappedServerStream 包装ServerStream以更新context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// UnaryAgentAuthInterceptor Agent专用的gRPC拦截器（简化认证，使用固定token）
func UnaryAgentAuthInterceptor(validToken string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// 从metadata获取token
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		// 获取authorization token
		tokens := md.Get("authorization")
		if len(tokens) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization token")
		}

		// 解析Bearer token
		authHeader := tokens[0]
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return nil, status.Error(codes.Unauthenticated, "invalid token format")
		}

		token := parts[1]

		// 验证固定token（用于Agent认证）
		if token != validToken {
			logger.Warn("Invalid agent token",
				zap.String("method", info.FullMethod))
			return nil, status.Error(codes.Unauthenticated, "invalid agent token")
		}

		logger.Debug("gRPC agent request authenticated",
			zap.String("method", info.FullMethod))

		// 将系统用户ID添加到context（ID=1表示系统用户）
		ctx = context.WithValue(ctx, "user_id", uint64(1))

		// 调用实际的handler
		return handler(ctx, req)
	}
}

// LoggingInterceptor 日志拦截器（记录所有gRPC调用）
func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		logger.Debug("gRPC call",
			zap.String("method", info.FullMethod))

		// 调用实际的handler
		resp, err := handler(ctx, req)

		if err != nil {
			logger.Error("gRPC call failed",
				zap.String("method", info.FullMethod),
				zap.Error(err))
		}

		return resp, err
	}
}
