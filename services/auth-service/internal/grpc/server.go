package grpcserver

import (
	"context"

	"github.com/itcmdb/auth-service/internal/service"
	pb "github.com/itcmdb/shared/proto/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	authService service.AuthService
	userService service.UserService
}

func NewAuthServer(authService service.AuthService, userService service.UserService) *AuthServer {
	return &AuthServer{
		authService: authService,
		userService: userService,
	}
}

// ValidateToken 验证JWT令牌
func (s *AuthServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	claims, err := s.authService.ValidateToken(req.Token)
	if err != nil {
		return &pb.ValidateTokenResponse{
			Valid: false,
			Error: err.Error(),
		}, nil
	}

	return &pb.ValidateTokenResponse{
		Valid:    true,
		UserId:   uint64(claims.UserID),
		Username: claims.Username,
	}, nil
}

// GetUser 获取用户信息
func (s *AuthServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	user, err := s.userService.GetUserByID(uint(req.UserId))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &pb.GetUserResponse{
		Id:       uint64(user.ID),
		Username: user.Username,
		Email:    user.Email,
		FullName: user.FullName,
		Status:   user.Status,
	}, nil
}

// GetUserPermissions 获取用户权限
func (s *AuthServer) GetUserPermissions(ctx context.Context, req *pb.GetUserPermissionsRequest) (*pb.GetUserPermissionsResponse, error) {
	permissions, err := s.userService.GetUserPermissions(uint(req.UserId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get permissions: %v", err)
	}

	return &pb.GetUserPermissionsResponse{
		Permissions: permissions,
	}, nil
}

// CheckPermission 检查用户权限
func (s *AuthServer) CheckPermission(ctx context.Context, req *pb.CheckPermissionRequest) (*pb.CheckPermissionResponse, error) {
	allowed, err := s.userService.CheckPermission(uint(req.UserId), req.Resource, req.Action)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check permission: %v", err)
	}

	return &pb.CheckPermissionResponse{
		Allowed: allowed,
	}, nil
}
