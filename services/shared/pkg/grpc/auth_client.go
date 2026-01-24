package grpc

import (
	"context"
	"time"

	pb "github.com/itcmdb/shared/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AuthClient Auth服务gRPC客户端
type AuthClient struct {
	conn   *grpc.ClientConn
	client pb.AuthServiceClient
}

// NewAuthClient 创建Auth服务gRPC客户端
func NewAuthClient(address string) (*AuthClient, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	client := pb.NewAuthServiceClient(conn)
	return &AuthClient{
		conn:   conn,
		client: client,
	}, nil
}

// ValidateToken 验证token
func (c *AuthClient) ValidateToken(ctx context.Context, token string) (*pb.ValidateTokenResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.ValidateToken(ctx, &pb.ValidateTokenRequest{
		Token: token,
	})
}

// GetUser 获取用户信息
func (c *AuthClient) GetUser(ctx context.Context, userID uint64) (*pb.GetUserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.GetUser(ctx, &pb.GetUserRequest{
		UserId: userID,
	})
}

// GetUserPermissions 获取用户权限
func (c *AuthClient) GetUserPermissions(ctx context.Context, userID uint64) (*pb.GetUserPermissionsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.GetUserPermissions(ctx, &pb.GetUserPermissionsRequest{
		UserId: userID,
	})
}

// CheckPermission 检查用户权限
func (c *AuthClient) CheckPermission(ctx context.Context, userID uint64, resource, action string) (*pb.CheckPermissionResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.CheckPermission(ctx, &pb.CheckPermissionRequest{
		UserId:   userID,
		Resource: resource,
		Action:   action,
	})
}

// Close 关闭连接
func (c *AuthClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
