package grpc

import (
	"context"
	"time"

	pb "github.com/itcmdb/shared/proto/cmdb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// CMDBClient CMDB服务gRPC客户端
type CMDBClient struct {
	conn   *grpc.ClientConn
	client pb.CMDBServiceClient
}

// NewCMDBClient 创建CMDB服务gRPC客户端
func NewCMDBClient(address string) (*CMDBClient, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	client := pb.NewCMDBServiceClient(conn)
	return &CMDBClient{
		conn:   conn,
		client: client,
	}, nil
}

// GetCIInstance 获取CI实例
func (c *CMDBClient) GetCIInstance(ctx context.Context, id uint64) (*pb.GetCIInstanceResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.GetCIInstance(ctx, &pb.GetCIInstanceRequest{
		Id: id,
	})
}

// GetCIInstances 获取CI实例列表
func (c *CMDBClient) GetCIInstances(ctx context.Context, ciTypeID uint64, page, pageSize int32) (*pb.GetCIInstancesResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.GetCIInstances(ctx, &pb.GetCIInstancesRequest{
		CiTypeId: ciTypeID,
		Page:     page,
		PageSize: pageSize,
	})
}

// GetCIType 获取CI类型
func (c *CMDBClient) GetCIType(ctx context.Context, id uint64) (*pb.GetCITypeResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.GetCIType(ctx, &pb.GetCITypeRequest{
		Id: id,
	})
}

// GetCITypes 获取CI类型列表
func (c *CMDBClient) GetCITypes(ctx context.Context) (*pb.GetCITypesResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.GetCITypes(ctx, &pb.GetCITypesRequest{})
}

// GetCIRelations 获取CI关系
func (c *CMDBClient) GetCIRelations(ctx context.Context, ciID uint64) (*pb.GetCIRelationsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.client.GetCIRelations(ctx, &pb.GetCIRelationsRequest{
		CiId: ciID,
	})
}

// Close 关闭连接
func (c *CMDBClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
