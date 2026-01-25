package grpcserver

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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
			Id:          uint64(t.ID),
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
			Id:           uint64(rel.ID),
			ParentId:     uint64(rel.ParentID),
			ChildId:      uint64(rel.ChildID),
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
		Id:         uint64(instance.ID),
		CiTypeId:   uint64(instance.CITypeID),
		Name:       instance.Name,
		Status:     instance.Status,
		Attributes: attrs,
		CreatedAt:  instance.CreatedAt.String(),
		UpdatedAt:  instance.UpdatedAt.String(),
	}

	// 转换CI类型
	if instance.CIType != nil {
		pbInst.CiType = &pb.CIType{
			Id:          uint64(instance.CIType.ID),
			Name:        instance.CIType.Name,
			DisplayName: instance.CIType.DisplayName,
			Icon:        instance.CIType.Icon,
			Description: instance.CIType.Description,
			IsActive:    instance.CIType.IsActive,
		}
	}

	return pbInst, nil
}

// ReportHardwareInfo 上报服务器硬件信息
func (s *CMDBServer) ReportHardwareInfo(ctx context.Context, req *pb.HardwareInfoReport) (*pb.HardwareInfoResponse, error) {
	// 转换硬件信息为attributes
	attributes := make(map[string]interface{})

	// 基本信息
	if req.ReportInfo != nil {
		attributes["hostname"] = req.ReportInfo.Hostname
		attributes["system_serial"] = req.ReportInfo.SystemSerial
		attributes["last_hardware_report"] = req.ReportInfo.Timestamp
	}

	// 内存信息（转换为JSON字符串）
	if len(req.Memory) > 0 {
		memoryData := convertMemoryInfoToMap(req.Memory)
		attributes["memory_info"] = memoryData
		attributes["memory_slots"] = len(req.Memory)
	}

	// 存储信息
	if len(req.Storage) > 0 {
		storageData := convertStorageInfoToMap(req.Storage)
		attributes["storage_info"] = storageData
		attributes["storage_count"] = len(req.Storage)

		// 计算总存储容量
		totalStorage := 0.0
		for _, storage := range req.Storage {
			size := parseSize(storage.Size)
			totalStorage += size
		}
		attributes["total_storage_gb"] = totalStorage
	}

	// GPU信息
	if len(req.Gpu) > 0 {
		gpuData := convertGPUInfoToMap(req.Gpu)
		attributes["gpu_info"] = gpuData
		attributes["gpu_count"] = len(req.Gpu)
	}

	// 网卡信息
	if len(req.Network) > 0 {
		networkData := convertNetworkInfoToMap(req.Network)
		attributes["network_info"] = networkData
		attributes["network_count"] = len(req.Network)
	}

	// 电源信息
	if len(req.PowerSupply) > 0 {
		psuData := convertPowerSupplyInfoToMap(req.PowerSupply)
		attributes["power_supply_info"] = psuData
		attributes["power_supply_count"] = len(req.PowerSupply)

		// 计算总电源容量
		totalCapacity := 0.0
		for _, psu := range req.PowerSupply {
			capacity := parseCapacity(psu.Capacity)
			totalCapacity += capacity
		}
		attributes["total_power_capacity_w"] = totalCapacity
	}

	// 光模块信息
	if len(req.OpticalModules) > 0 {
		opticalData := convertOpticalModuleInfoToMap(req.OpticalModules)
		attributes["optical_modules_info"] = opticalData
		attributes["optical_modules_count"] = len(req.OpticalModules)
	}

	// 使用hostname作为服务器名称，如果不存在则使用system_serial
	serverName := req.ReportInfo.Hostname
	if serverName == "" {
		serverName = req.ReportInfo.SystemSerial
	}

	// 查找是否已存在该服务器的CI实例
	// 优先通过system_serial查找，其次通过hostname
	existingInstance, err := s.findServerBySerialOrHost(req.ReportInfo.SystemSerial, req.ReportInfo.Hostname)
	var instanceID uint64

	if err == nil && existingInstance != nil {
		// 更新现有实例
		updateReq := &service.UpdateCIInstanceRequest{
			Name: serverName,
			Attributes: attributes,
			Status: "active",
		}

		_, err = s.ciService.UpdateCIInstance(existingInstance.ID, updateReq, 1) // userID=1 for system
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update server instance: %v", err)
		}
		instanceID = uint64(existingInstance.ID)
	} else {
		// 创建新实例
		createReq := &service.CreateCIInstanceRequest{
			CITypeID: 1, // 服务器类型
			Name: serverName,
			Attributes: attributes,
			Status: "active",
		}

		instance, err := s.ciService.CreateCIInstance(createReq, 1) // userID=1 for system
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to create server instance: %v", err)
		}
		instanceID = uint64(instance.ID)
	}

	return &pb.HardwareInfoResponse{
		Success:     true,
		Message:     "Hardware info reported successfully",
		CiInstanceId: instanceID,
	}, nil
}

// findServerBySerialOrHost 通过序列号或主机名查找服务器
func (s *CMDBServer) findServerBySerialOrHost(serial, host string) (*models.CIInstance, error) {
	// 先通过system_serial查找
	if serial != "" {
		filters := map[string]interface{}{
			"system_serial": serial,
		}
		instances, _, err := s.ciService.GetCIInstances(1, filters, 1, 1)
		if err == nil && len(instances) > 0 {
			return &instances[0], nil
		}
	}

	// 再通过hostname查找
	if host != "" {
		filters := map[string]interface{}{
			"hostname": host,
		}
		instances, _, err := s.ciService.GetCIInstances(1, filters, 1, 1)
		if err == nil && len(instances) > 0 {
			return &instances[0], nil
		}
	}

	return nil, fmt.Errorf("server not found")
}

// 辅助转换函数
func convertMemoryInfoToMap(memories []*pb.MemoryInfo) interface{} {
	result := make([]map[string]string, 0, len(memories))
	for _, m := range memories {
		result = append(result, map[string]string{
			"size":         m.Size,
			"form_factor":  m.FormFactor,
			"locator":      m.Locator,
			"type":         m.Type,
			"speed":        m.Speed,
			"manufacturer":  m.Manufacturer,
			"part_number":   m.PartNumber,
			"serial":        m.Serial,
		})
	}
	return result
}

func convertStorageInfoToMap(storages []*pb.StorageInfo) interface{} {
	result := make([]map[string]string, 0, len(storages))
	for _, s := range storages {
		result = append(result, map[string]string{
			"name":   s.Name,
			"size":   s.Size,
			"model":  s.Model,
			"serial": s.Serial,
		})
	}
	return result
}

func convertGPUInfoToMap(gpus []*pb.GPUInfo) interface{} {
	result := make([]map[string]string, 0, len(gpus))
	for _, g := range gpus {
		result = append(result, map[string]string{
			"gpu_id":       g.GpuId,
			"product_name": g.ProductName,
			"serial":       g.Serial,
		})
	}
	return result
}

func convertNetworkInfoToMap(networks []*pb.NetworkInfo) interface{} {
	result := make([]map[string]string, 0, len(networks))
	for _, n := range networks {
		result = append(result, map[string]string{
			"vendor":   n.Vendor,
			"model":    n.Model,
			"speed":    n.Speed,
			"mac":      n.Mac,
			"firmware": n.Firmware,
			"serial":   n.Serial,
		})
	}
	return result
}

func convertPowerSupplyInfoToMap(psus []*pb.PowerSupplyInfo) interface{} {
	result := make([]map[string]string, 0, len(psus))
	for _, p := range psus {
		result = append(result, map[string]string{
			"location":     p.Location,
			"name":         p.Name,
			"manufacturer": p.Manufacturer,
			"serial":       p.Serial,
			"model":        p.Model,
			"capacity":     p.Capacity,
		})
	}
	return result
}

func convertOpticalModuleInfoToMap(modules []*pb.OpticalModuleInfo) interface{} {
	result := make([]map[string]string, 0, len(modules))
	for _, m := range modules {
		result = append(result, map[string]string{
			"pci_addr":       m.PciAddr,
			"port":           m.Port,
			"ib_ca":          m.IbCa,
			"identifier":     m.Identifier,
			"vendor_name":     m.VendorName,
			"part_number":    m.PartNumber,
			"serial_number":  m.SerialNumber,
			"speed":          m.Speed,
			"temperature":    m.Temperature,
			"wavelength":     m.Wavelength,
		})
	}
	return result
}

// parseSize 解析大小字符串（例如 "894.3G", "3.5T"）
func parseSize(sizeStr string) float64 {
	if sizeStr == "" || sizeStr == "None" {
		return 0
	}

	var value float64
	var unit string
	_, err := fmt.Sscanf(sizeStr, "%f%s", &value, &unit)
	if err != nil {
		return 0
	}

	switch unit {
	case "T", "TB":
		return value * 1024
	case "G", "GB":
		return value
	case "M", "MB":
		return value / 1024
	case "K", "KB":
		return value / 1024 / 1024
	default:
		// 假设默认单位是GB
		return value
	}
}

// parseCapacity 解析电源容量（例如 "3000 W"）
func parseCapacity(capStr string) float64 {
	if capStr == "" {
		return 0
	}

	var value float64
	_, err := fmt.Sscanf(capStr, "%f", &value)
	if err != nil {
		return 0
	}
	return value
}
