package grpc

import (
	"context"
	"sync"
	"time"

	pb "github.com/itcmdb/shared/proto/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	"github.com/itcmdb/shared/pkg/logger"
)

// AuthClient Auth服务gRPC客户端
type AuthClient struct {
	mu       sync.RWMutex
	address  string
	conn     *grpc.ClientConn
	client   pb.AuthServiceClient
	stopCh   chan struct{}
}

// NewAuthClient 创建Auth服务gRPC客户端
func NewAuthClient(address string) (*AuthClient, error) {
	c := &AuthClient{
		address: address,
		stopCh:  make(chan struct{}),
	}

	if err := c.connect(); err != nil {
		return nil, err
	}

	// 启动后台健康检查和自动重连
	go c.keepAlive()

	return c, nil
}

// connect 建立gRPC连接
func (c *AuthClient) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, err := grpc.NewClient(
		c.address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	if err != nil {
		return err
	}

	c.conn = conn
	c.client = pb.NewAuthServiceClient(conn)
	return nil
}

// keepAlive 后台定期检查连接健康状态，断线时自动重连
func (c *AuthClient) keepAlive() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.mu.RLock()
			conn := c.conn
			c.mu.RUnlock()

			if conn == nil {
				logger.Warn("Auth gRPC client connection is nil, attempting reconnect")
				if err := c.reconnect(); err != nil {
					logger.Error("Failed to reconnect to Auth service", zap.Error(err))
				}
				continue
			}

			// 使用短暂超时的健康检查
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			_, err := c.client.ValidateToken(ctx, &pb.ValidateTokenRequest{Token: "health-check"})
			cancel()

			if err != nil {
				// 预期的错误（token无效）不算连接问题
				if !isConnectionError(err) {
					continue
				}
				logger.Warn("Auth gRPC connection lost, attempting reconnect", zap.Error(err))
				if err := c.reconnect(); err != nil {
					logger.Error("Failed to reconnect to Auth service", zap.Error(err))
				}
			}
		}
	}
}

// reconnect 重新建立连接
func (c *AuthClient) reconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 关闭旧连接
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
		c.client = nil
	}

	// 带重试的连接
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		conn, err := grpc.NewClient(
			c.address,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                10 * time.Second,
				Timeout:             3 * time.Second,
				PermitWithoutStream: true,
			}),
		)
		if err != nil {
			lastErr = err
			continue
		}

		c.conn = conn
		c.client = pb.NewAuthServiceClient(conn)
		logger.Info("Successfully reconnected to Auth service", zap.String("address", c.address))
		return nil
	}

	return lastErr
}

// isConnectionError 判断是否为连接级别的错误
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	// 连接关闭、超时、EOF等视为连接错误
	errStr := err.Error()
	return contains(errStr, "connection") || contains(errStr, "timeout") ||
		contains(errStr, "EOF") || contains(errStr, "broken pipe") ||
		contains(errStr, "unavailable")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ValidateToken 验证token
func (c *AuthClient) ValidateToken(ctx context.Context, token string) (*pb.ValidateTokenResponse, error) {
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()

	if client == nil {
		return nil, grpc.ErrServerStopped
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return client.ValidateToken(ctx, &pb.ValidateTokenRequest{
		Token: token,
	})
}

// GetUser 获取用户信息
func (c *AuthClient) GetUser(ctx context.Context, userID uint64) (*pb.GetUserResponse, error) {
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()

	if client == nil {
		return nil, grpc.ErrServerStopped
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return client.GetUser(ctx, &pb.GetUserRequest{
		UserId: userID,
	})
}

// GetUserPermissions 获取用户权限
func (c *AuthClient) GetUserPermissions(ctx context.Context, userID uint64) (*pb.GetUserPermissionsResponse, error) {
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()

	if client == nil {
		return nil, grpc.ErrServerStopped
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return client.GetUserPermissions(ctx, &pb.GetUserPermissionsRequest{
		UserId: userID,
	})
}

// CheckPermission 检查用户权限
func (c *AuthClient) CheckPermission(ctx context.Context, userID uint64, resource, action string) (*pb.CheckPermissionResponse, error) {
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()

	if client == nil {
		return nil, grpc.ErrServerStopped
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return client.CheckPermission(ctx, &pb.CheckPermissionRequest{
		UserId:   userID,
		Resource: resource,
		Action:   action,
	})
}

// Close 关闭连接
func (c *AuthClient) Close() error {
	close(c.stopCh)

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		c.client = nil
		return err
	}
	return nil
}
