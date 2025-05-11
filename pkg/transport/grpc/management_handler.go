package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"

	apiv1 "github.com/turtacn/dataseap/api/v1"
	commontypes "github.com/turtacn/dataseap/pkg/common/types"
	commonenum "github.com/turtacn/dataseap/pkg/common/types/enum"
	"github.com/turtacn/dataseap/pkg/domain/management/lifecycle"
	lifecyclemodel "github.com/turtacn/dataseap/pkg/domain/management/lifecycle/model"
	"github.com/turtacn/dataseap/pkg/domain/management/metadata"
	metadatamodel "github.com/turtacn/dataseap/pkg/domain/management/metadata/model"
	"github.com/turtacn/dataseap/pkg/domain/management/workload"
	workloadmodel "github.com/turtacn/dataseap/pkg/domain/management/workload/model"
	"github.com/turtacn/dataseap/pkg/logger"
)

// managementHandler implements the apiv1.ManagementServiceServer interface.
// managementHandler 实现 apiv1.ManagementServiceServer 接口。
type managementHandler struct {
	apiv1.UnimplementedManagementServiceServer // For forward compatibility
	workloadSvc                                workload.Service
	metadataSvc                                metadata.Service
	lifecycleSvc                               lifecycle.Service
}

// NewManagementHandler creates a new gRPC handler for management services.
// NewManagementHandler 为管理服务创建一个新的gRPC处理器。
func NewManagementHandler(
	workloadSvc workload.Service,
	metadataSvc metadata.Service,
	lifecycleSvc lifecycle.Service,
) apiv1.ManagementServiceServer {
	return &managementHandler{
		workloadSvc:  workloadSvc,
		metadataSvc:  metadataSvc,
		lifecycleSvc: lifecycleSvc,
	}
}

// --- Workload Management ---

func (h *managementHandler) CreateWorkloadGroup(ctx context.Context, req *apiv1.CreateWorkloadGroupRequest) (*apiv1.WorkloadGroupResponse, error) {
	l := logger.L().Ctx(ctx).With("handler", "CreateWorkloadGroup")
	protoWg := req.GetWorkloadGroup()
	if protoWg == nil {
		return nil, status.Error(codes.InvalidArgument, "workload group data is required")
	}
	l = l.With("group_name", protoWg.GetName())
	l.Info("Received CreateWorkloadGroup request")

	domainWg := &workloadmodel.WorkloadGroup{
		Name:             protoWg.GetName(),
		CPUShare:         protoWg.GetCpuShare(),
		MemoryLimit:      protoWg.GetMemoryLimit(),
		ConcurrencyLimit: protoWg.GetConcurrencyLimit(),
		MaxQueueSize:     protoWg.GetMaxQueueSize(),
		Properties:       protoWg.GetProperties(),
	}

	err := h.workloadSvc.CreateWorkloadGroup(ctx, domainWg)
	if err != nil {
		l.Errorw("Failed to create workload group", "error", err)
		// TODO: Map domain error to gRPC status
		return &apiv1.WorkloadGroupResponse{Error: toProtoErrorDetail("CREATE_WG_ERROR", err.Error())}, status.Error(codes.Internal, err.Error())
	}

	l.Info("Workload group created successfully")
	// Return the created group (or fetch it again if creation doesn't return it)
	return &apiv1.WorkloadGroupResponse{WorkloadGroup: protoWg}, nil
}

func (h *managementHandler) GetWorkloadGroup(ctx context.Context, req *apiv1.GetWorkloadGroupRequest) (*apiv1.WorkloadGroupResponse, error) {
	l := logger.L().Ctx(ctx).With("handler", "GetWorkloadGroup", "group_name", req.GetName())
	l.Info("Received GetWorkloadGroup request")

	domainWg, err := h.workloadSvc.GetWorkloadGroup(ctx, req.GetName())
	if err != nil {
		l.Errorw("Failed to get workload group", "error", err)
		return &apiv1.WorkloadGroupResponse{Error: toProtoErrorDetail("GET_WG_ERROR", err.Error())}, status.Error(codes.NotFound, err.Error()) // Or Internal
	}

	protoWg := &apiv1.WorkloadGroup{
		Name:             domainWg.Name,
		CpuShare:         domainWg.CPUShare,
		MemoryLimit:      domainWg.MemoryLimit,
		ConcurrencyLimit: domainWg.ConcurrencyLimit,
		MaxQueueSize:     domainWg.MaxQueueSize,
		Properties:       domainWg.Properties,
	}

	l.Info("Workload group retrieved successfully")
	return &apiv1.WorkloadGroupResponse{WorkloadGroup: protoWg}, nil
}

func (h *managementHandler) ListWorkloadGroups(ctx context.Context, req *apiv1.ListWorkloadGroupsRequest) (*apiv1.ListWorkloadGroupsResponse, error) {
	l := logger.L().Ctx(ctx).With("handler", "ListWorkloadGroups")
	l.Info("Received ListWorkloadGroups request")

	var pReq *commontypes.PaginationRequest
	if req.GetPagination() != nil {
		pReq = &commontypes.PaginationRequest{
			Page:     int(req.GetPagination().GetPage()),
			PageSize: int(req.GetPagination().GetPageSize()),
		}
	}

	domainWgs, total, err := h.workloadSvc.ListWorkloadGroups(ctx, pReq)
	if err != nil {
		l.Errorw("Failed to list workload groups", "error", err)
		return &apiv1.ListWorkloadGroupsResponse{Error: toProtoErrorDetail("LIST_WG_ERROR", err.Error())}, status.Error(codes.Internal, err.Error())
	}

	protoWgs := make([]*apiv1.WorkloadGroup, len(domainWgs))
	for i, dwg := range domainWgs {
		protoWgs[i] = &apiv1.WorkloadGroup{
			Name: dwg.Name, CpuShare: dwg.CPUShare, MemoryLimit: dwg.MemoryLimit, /* map others */
		}
	}

	resp := &apiv1.ListWorkloadGroupsResponse{WorkloadGroups: protoWgs}
	if pReq != nil && total > 0 { // Only include pagination response if requested and items exist
		resp.Pagination = &apiv1.PaginationResponse{
			Page:       req.GetPagination().GetPage(),
			PageSize:   req.GetPagination().GetPageSize(),
			TotalItems: total,
		}
	}

	l.Info("Workload groups listed successfully")
	return resp, nil
}

func (h *managementHandler) UpdateWorkloadGroup(ctx context.Context, req *apiv1.UpdateWorkloadGroupRequest) (*apiv1.WorkloadGroupResponse, error) {
	l := logger.L().Ctx(ctx).With("handler", "UpdateWorkloadGroup")
	protoWg := req.GetWorkloadGroup()
	if protoWg == nil {
		return nil, status.Error(codes.InvalidArgument, "workload group data is required for update")
	}
	l = l.With("group_name", protoWg.GetName())
	l.Info("Received UpdateWorkloadGroup request")

	domainWg := &workloadmodel.WorkloadGroup{ /* map from protoWg */ Name: protoWg.Name, CPUShare: protoWg.CpuShare, MemoryLimit: protoWg.MemoryLimit}
	err := h.workloadSvc.UpdateWorkloadGroup(ctx, domainWg)
	if err != nil {
		l.Errorw("Failed to update workload group", "error", err)
		return &apiv1.WorkloadGroupResponse{Error: toProtoErrorDetail("UPDATE_WG_ERROR", err.Error())}, status.Error(codes.Internal, err.Error())
	}
	l.Info("Workload group updated successfully")
	return &apiv1.WorkloadGroupResponse{WorkloadGroup: protoWg}, nil
}

func (h *managementHandler) DeleteWorkloadGroup(ctx context.Context, req *apiv1.DeleteWorkloadGroupRequest) (*emptypb.Empty, error) {
	l := logger.L().Ctx(ctx).With("handler", "DeleteWorkloadGroup", "group_name", req.GetName())
	l.Info("Received DeleteWorkloadGroup request")

	err := h.workloadSvc.DeleteWorkloadGroup(ctx, req.GetName())
	if err != nil {
		l.Errorw("Failed to delete workload group", "error", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	l.Info("Workload group deleted successfully")
	return &emptypb.Empty{}, nil
}

// --- Metadata Management ---
// Placeholder implementations for metadata methods
func (h *managementHandler) GetTableSchema(ctx context.Context, req *apiv1.GetTableSchemaRequest) (*apiv1.GetTableSchemaResponse, error) {
	l := logger.L().Ctx(ctx).With("handler", "GetTableSchema", "db", req.GetDatabaseName(), "table", req.GetTableName())
	l.Info("Received GetTableSchema request")

	domainSchema, err := h.metadataSvc.GetTableSchema(ctx, req.GetDatabaseName(), req.GetTableName())
	if err != nil {
		l.Errorw("Failed to get table schema", "error", err)
		return &apiv1.GetTableSchemaResponse{Error: toProtoErrorDetail("GET_SCHEMA_ERROR", err.Error())}, status.Error(codes.NotFound, err.Error())
	}

	protoSchema := &apiv1.TableSchema{
		TableName:    domainSchema.TableName,
		DatabaseName: domainSchema.DatabaseName,
		Fields:       make([]*apiv1.FieldSchema, len(domainSchema.Fields)),
		// Map other fields like TableType, KeysType, etc.
	}
	for i, df := range domainSchema.Fields {
		protoSchema.Fields[i] = &apiv1.FieldSchema{
			Name:            df.Name,
			DataType:        apiv1.DataType(apiv1.DataType_value[string(df.DataType)]), // Assumes enum names match
			TypeString:      df.TypeString,
			IsNullable:      df.IsNullable,
			IsPrimaryKey:    df.IsPrimaryKey,
			DefaultValue:    df.DefaultValue,
			Comment:         df.Comment,
			AggregationType: df.AggregationType,
		}
	}
	l.Info("Table schema retrieved successfully")
	return &apiv1.GetTableSchemaResponse{Schema: protoSchema}, nil
}

func (h *managementHandler) ListTables(ctx context.Context, req *apiv1.ListTablesRequest) (*apiv1.ListTablesResponse, error) {
	l := logger.L().Ctx(ctx).With("handler", "ListTables", "db", req.GetDatabaseName())
	l.Info("Received ListTables request")
	// ... implementation mapping to h.metadataSvc.ListTables ...
	return nil, status.Errorf(codes.Unimplemented, "method ListTables not implemented")
}

func (h *managementHandler) CreateIndex(ctx context.Context, req *apiv1.CreateIndexRequest) (*apiv1.StandardResponse, error) {
	l := logger.L().Ctx(ctx).With("handler", "CreateIndex", "db", req.GetDatabaseName(), "table", req.GetTableName())
	l.Info("Received CreateIndex request")

	idxDef := req.GetIndexDefinition()
	domainIdxDef := &metadatamodel.IndexDefinition{
		IndexName:    idxDef.GetIndexName(),
		TableName:    req.GetTableName(),
		DatabaseName: req.GetDatabaseName(),
		IndexType:    idxDef.GetIndexType(),
		Fields:       idxDef.GetFields(),
		Properties:   idxDef.GetProperties(),
		Comment:      idxDef.GetComment(),
	}

	err := h.metadataSvc.CreateIndex(ctx, domainIdxDef)
	if err != nil {
		l.Errorw("Failed to create index", "error", err)
		return &apiv1.StandardResponse{Success: false, Message: err.Error(), Error: toProtoErrorDetail("CREATE_INDEX_ERROR", err.Error())}, status.Error(codes.Internal, err.Error())
	}

	l.Info("Index created successfully")
	return &apiv1.StandardResponse{Success: true, Message: "Index created successfully"}, nil
}

func (h *managementHandler) GetIndexInfo(ctx context.Context, req *apiv1.GetIndexInfoRequest) (*apiv1.GetIndexInfoResponse, error) {
	l := logger.L().Ctx(ctx).With("handler", "GetIndexInfo", "db", req.GetDatabaseName(), "table", req.GetTableName())
	l.Info("Received GetIndexInfo request")
	// ... implementation ...
	return nil, status.Errorf(codes.Unimplemented, "method GetIndexInfo not implemented")
}

func (h *managementHandler) CreateMaterializedView(ctx context.Context, req *apiv1.CreateMaterializedViewRequest) (*apiv1.StandardResponse, error) {
	l := logger.L().Ctx(ctx).With("handler", "CreateMaterializedView", "db", req.GetDatabaseName(), "view", req.GetViewName())
	l.Info("Received CreateMaterializedView request")
	// ... implementation ...
	return nil, status.Errorf(codes.Unimplemented, "method CreateMaterializedView not implemented")
}

// --- Lifecycle Management ---
func (h *managementHandler) GetComponentStatus(ctx context.Context, req *apiv1.GetComponentStatusRequest) (*apiv1.GetComponentStatusResponse, error) {
	l := logger.L().Ctx(ctx).With("handler", "GetComponentStatus", "component", req.GetComponentName(), "type", req.GetComponentType())
	l.Info("Received GetComponentStatus request")

	domainStatuses, err := h.lifecycleSvc.GetComponentStatus(ctx, req.GetComponentName(), lifecyclemodel.ComponentType(req.GetComponentType()))
	if err != nil {
		l.Errorw("Failed to get component status", "error", err)
		return &apiv1.GetComponentStatusResponse{Error: toProtoErrorDetail("GET_STATUS_ERROR", err.Error())}, status.Error(codes.Internal, err.Error())
	}

	protoStatuses := make([]*apiv1.ComponentStatus, len(domainStatuses))
	for i, ds := range domainStatuses {
		detailsPb, _ := structpb.NewStruct(ds.Details)
		protoStatuses[i] = &apiv1.ComponentStatus{
			ComponentName: ds.ComponentName,
			Status:        ds.Status,
			Message:       ds.Message,
			Details:       detailsPb,
		}
	}
	l.Info("Component statuses retrieved successfully")
	return &apiv1.GetComponentStatusResponse{Statuses: protoStatuses}, nil
}
