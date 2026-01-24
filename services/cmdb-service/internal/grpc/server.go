package grpcserver

import (
	"context"

	"github.com/itcmdb/cmdb-service/internal/models"
	"github.com/itcmdb/cmdb-service/internal/service"
	pb "github.com/itcmdb/shared/proto/cmdb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type CMDBServer struct {
	pb.UnimplementedCMDBServiceServer
	ciService service.CIService
}

func NewCMDBServer(ciService service.CIService) *CMDBServer {
	return &CMDBServer{
		ciService: ciService,
	}
}

// GetCIInstance 获取CI实例
func (s *CMDBServer) GetCIInstance(ctx context.Context, req *pb.GetCIInstanceRequest) (*pb.GetCIInstanceResponse, error) {
	instance, err := s.ciService.GetCIInstanceByID(uint(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "CI instance not found: %v", err)
	}

	// 转换为protobuf格式
	pbInstance, err := convertCIInstanceToPB(instance)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to convert instance: %v", err)
	}

	return &pb.GetCIInstanceResponse{
		Instance: pbInstance,
	}, nil
}

// GetCIInstances 获取CI实例列表
func (s *CMDBServer) GetCIInstances(ctx context.Context, req *pb.GetCIInstancesRequest) (*pb.GetCIInstancesResponse, error) {
	filters := make(map[string]interface{})
	for k, v := range req.Filters {
		filters[k] = v
	}

	instances, total, err := s.ciService.GetCIInstances(uint(req.CiTypeId), filters, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get instances: %v", err)
	}

	// 转换为protobuf格式
	pbInstances := make([]*pb.CIInstance, 0, len(instances))
	for _, inst := range instances {
		pbInst, err := convertCIInstanceToPB(&inst)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to convert instance: %v", err)
		}
		pbInstances = append(pbInstances, pbInst)
	}

	return &pb.GetCIInstancesResponse{
		Instances: pbInstances,
		Total:     total,
	}, nil
}

// GetCIType 获取CI类型
func (s *CMDBServer) GetCIType(ctx context.Context, req *pb.GetCITypeRequest) (*pb.GetCITypeResponse, error) {
	// TODO: 实现获取CI类型逻辑
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// GetCITypes 获取CI类型列表
func (s *CMDBServer) GetCITypes(ctx context.Context, req *pb.GetCITypesRequest) (*pb.GetCITypesResponse, error) {
	types, err := s.ciService.GetCITypes()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get CI types: %v", err)
	}

	// 转换为protobuf格式
	pbTypes := make([]*pb.CIType, 0, len(types))
	for _, t := range types {
		pbTypes = append(pbTypes, &pb.CIType{
			Id:          uint32(t.ID),
			Name:        t.Name,
			DisplayName: t.DisplayName,
			Icon:        t.Icon,
			Description: t.Description,
			IsActive:    t.IsActive,
		})
	}

	return &pb.GetCITypesResponse{
		CiTypes: pbTypes,
	}, nil
}

// GetCIRelations 获取CI关系
func (s *CMDBServer) GetCIRelations(ctx context.Context, req *pb.GetCIRelationsRequest) (*pb.GetCIRelationsResponse, error) {
	relations, err := s.ciService.GetCIRelations(uint(req.CiId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get relations: %v", err)
	}

	// 转换为protobuf格式
	pbRelations := make([]*pb.CIRelation, 0, len(relations))
	for _, rel := range relations {
		pbRel := &pb.CIRelation{
			Id:           uint32(rel.ID),
			ParentId:     uint32(rel.ParentID),
			ChildId:      uint32(rel.ChildID),
			RelationType: rel.RelationType,
			Description:  rel.Description,
		}

		// 转换parent和child
		if rel.Parent != nil {
			pbParent, err := convertCIInstanceToPB(rel.Parent)
			if err == nil {
				pbRel.Parent = pbParent
			}
		}
		if rel.Child != nil {
			pbChild, err := convertCIInstanceToPB(rel.Child)
			if err == nil {
				pbRel.Child = pbChild
			}
		}

		pbRelations = append(pbRelations, pbRel)
	}

	return &pb.GetCIRelationsResponse{
		Relations: pbRelations,
	}, nil
}

// 辅助函数：转换CI实例为protobuf格式
func convertCIInstanceToPB(instance *models.CIInstance) (*pb.CIInstance, error) {
	// 转换attributes为protobuf Struct
	var attrs *structpb.Struct
	if instance.Attributes != nil {
		var err error
		attrs, err = structpb.NewStruct(instance.Attributes)
		if err != nil {
			return nil, err
		}
	}

	pbInst := &pb.CIInstance{
		Id:         uint32(instance.ID),
		CiTypeId:   uint32(instance.CITypeID),
		Name:       instance.Name,
		Status:     instance.Status,
		Attributes: attrs,
		CreatedAt:  instance.CreatedAt.String(),
		UpdatedAt:  instance.UpdatedAt.String(),
	}

	// 转换CI类型
	if instance.CIType != nil {
		pbInst.CiType = &pb.CIType{
			Id:          uint32(instance.CIType.ID),
			Name:        instance.CIType.Name,
			DisplayName: instance.CIType.DisplayName,
			Icon:        instance.CIType.Icon,
			Description: instance.CIType.Description,
			IsActive:    instance.CIType.IsActive,
		}
	}

	return pbInst, nil
}
