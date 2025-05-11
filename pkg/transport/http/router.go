package http

import (
	"github.com/turtacn/dataseap/pkg/common/constants"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	apiv1 "github.com/turtacn/dataseap/api/v1" // For request/response DTOs if not mapping directly to domain
	commontypes "github.com/turtacn/dataseap/pkg/common/types"
	// Domain services (already passed via ServiceRegistry)
	// "github.com/turtacn/dataseap/pkg/domain/ingestion"
	// "github.com/turtacn/dataseap/pkg/domain/query"
	// ...
	// Domain models (for request/response bodies if not using DTOs)
	ingestionmodel "github.com/turtacn/dataseap/pkg/domain/ingestion/model"
	lifecyclemodel "github.com/turtacn/dataseap/pkg/domain/management/lifecycle/model"
	querymodel "github.com/turtacn/dataseap/pkg/domain/query/model"
	// ... other domain models
)

// SetupRouter configures the Gin router with all application routes.
// SetupRouter 使用所有应用程序路由配置Gin路由器。
func SetupRouter(engine *gin.Engine, services ServiceRegistry) {
	// Health check endpoint
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "UP",
			"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		})
	})

	// API v1 group
	v1 := engine.Group("/api/v1")
	{
		// --- Ingestion Routes ---
		if services.IngestionSvc != nil {
			ingestionRouter := v1.Group("/ingest")
			{
				// Using a struct that maps to apiv1.IngestDataRequest for request body
				// Or define a specific DTO here.
				ingestionRouter.POST("/events", func(c *gin.Context) {
					var req apiv1.IngestDataRequest // Using proto struct as DTO for HTTP too
					if err := c.ShouldBindJSON(&req); err != nil {
						c.JSON(http.StatusBadRequest, commontypes.NewErrorAPIResponse(
							&commonerrors.AppError{Code: commonerrors.InvalidArgument, Message: "Invalid request body: " + err.Error()},
						))
						return
					}
					// Map to domain model
					domainEvents := make([]*ingestionmodel.RawEvent, len(req.Records))
					for i, r := range req.Records {
						domainEvents[i] = &ingestionmodel.RawEvent{
							DataSourceID: r.GetDataSourceId(), DataType: r.GetDataType(), Timestamp: r.GetTimestamp().AsTime(), Data: r.GetData().AsMap(), Tags: r.GetTags(),
						}
					}
					ingested, persistFailed, validationFailed, err := services.IngestionSvc.IngestEvents(c.Request.Context(), domainEvents)
					if err != nil {
						c.JSON(http.StatusInternalServerError, commontypes.NewErrorAPIResponse(
							&commonerrors.AppError{Code: commonerrors.InternalError, Message: "Ingestion failed: " + err.Error()},
						))
						return
					}
					c.JSON(http.StatusOK, commontypes.NewSuccessAPIResponse(gin.H{
						"ingestedCount": ingested, "persistFailedCount": persistFailed, "validationFailedCount": validationFailed,
					}))
				})
			}
		}

		// --- Query Routes ---
		if services.QuerySvc != nil {
			queryRouter := v1.Group("/query")
			{
				// Using querymodel.SQLQueryRequest directly as it's a Go struct
				queryRouter.POST("/sql", func(c *gin.Context) {
					var req querymodel.SQLQueryRequest
					if err := c.ShouldBindJSON(&req); err != nil {
						c.JSON(http.StatusBadRequest, commontypes.NewErrorAPIResponse(&commonerrors.AppError{Code: commonerrors.InvalidArgument, Message: "Invalid SQL query request: " + err.Error()}))
						return
					}
					result, err := services.QuerySvc.ExecuteSQL(c.Request.Context(), &req)
					if err != nil {
						// TODO: Map domain error to appropriate HTTP status code
						c.JSON(http.StatusInternalServerError, commontypes.NewErrorAPIResponse(&commonerrors.AppError{Code: commonerrors.DatabaseError, Message: "SQL query execution failed: " + err.Error()}))
						return
					}
					c.JSON(http.StatusOK, commontypes.NewSuccessAPIResponse(result))
				})

				queryRouter.POST("/search/fulltext", func(c *gin.Context) {
					var req querymodel.FullTextSearchRequest
					if err := c.ShouldBindJSON(&req); err != nil {
						c.JSON(http.StatusBadRequest, commontypes.NewErrorAPIResponse(&commonerrors.AppError{Code: commonerrors.InvalidArgument, Message: "Invalid full-text search request: " + err.Error()}))
						return
					}
					result, err := services.QuerySvc.SearchFullText(c.Request.Context(), &req)
					if err != nil {
						c.JSON(http.StatusInternalServerError, commontypes.NewErrorAPIResponse(&commonerrors.AppError{Code: commonerrors.InternalError, Message: "Full-text search failed: " + err.Error()}))
						return
					}
					c.JSON(http.StatusOK, commontypes.NewSuccessAPIResponse(result))
				})
			}
		}

		// --- Management Routes ---
		if services.WorkloadSvc != nil || services.MetadataSvc != nil || services.LifecycleSvc != nil {
			mgmtRouter := v1.Group("/management")
			{
				// Example: Workload Group
				if services.WorkloadSvc != nil {
					wgRouter := mgmtRouter.Group("/workload-groups")
					{
						wgRouter.POST("", func(c *gin.Context) { /* ... create workload group ... */
							c.JSON(http.StatusNotImplemented, "Not Implemented")
						})
						wgRouter.GET("/:name", func(c *gin.Context) { /* ... get workload group ... */
							c.JSON(http.StatusNotImplemented, "Not Implemented")
						})
						// ... other workload routes
					}
				}
				// Example: Metadata - Table Schema
				if services.MetadataSvc != nil {
					schemaRouter := mgmtRouter.Group("/databases/:dbName/tables/:tableName/schema")
					{
						schemaRouter.GET("", func(c *gin.Context) {
							dbName := c.Param("dbName")
							tableName := c.Param("tableName")
							schema, err := services.MetadataSvc.GetTableSchema(c.Request.Context(), dbName, tableName)
							if err != nil {
								// TODO: Map error
								c.JSON(http.StatusNotFound, commontypes.NewErrorAPIResponse(&commonerrors.AppError{Code: commonerrors.NotFoundError, Message: err.Error()}))
								return
							}
							c.JSON(http.StatusOK, commontypes.NewSuccessAPIResponse(schema))
						})
					}
				}
				// Example: Lifecycle - Component Status
				if services.LifecycleSvc != nil {
					statusRouter := mgmtRouter.Group("/status/components")
					{
						statusRouter.GET("", func(c *gin.Context) {
							compName := c.Query("componentName")
							compType := c.Query("componentType") // This will be string, need to map to model.ComponentType

							statuses, err := services.LifecycleSvc.GetComponentStatus(c.Request.Context(), compName, lifecyclemodel.ComponentType(compType))
							if err != nil {
								c.JSON(http.StatusInternalServerError, commontypes.NewErrorAPIResponse(&commonerrors.AppError{Code: commonerrors.InternalError, Message: err.Error()}))
								return
							}
							c.JSON(http.StatusOK, commontypes.NewSuccessAPIResponse(statuses))
						})
					}
				}
			}
		}

	}

	// Fallback for unmatched routes
	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, commontypes.NewErrorAPIResponse(
			&commonerrors.AppError{Code: commonerrors.NotFoundError, Message: "Endpoint not found"},
		))
	})
}

// Helper for binding and validating pagination from query parameters
func bindPagination(c *gin.Context) *commontypes.PaginationRequest {
	var page, pageSize int
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10") // Default to 10, adjust as needed

	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	} else {
		page = 1 // Default if invalid
	}

	if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
		pageSize = ps
		// Optional: Cap pageSize to a maximum
		// if pageSize > constants.MaxPageSize { pageSize = constants.MaxPageSize }
	} else {
		pageSize = 10 // Default if invalid
	}
	return &commontypes.PaginationRequest{Page: page, PageSize: pageSize}
}

// Helper for request ID middleware (example, can be more sophisticated)
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(constants.HeaderRequestID)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set(string(constants.ContextKeyRequestID), requestID)     // Set for handlers
		c.Writer.Header().Set(constants.HeaderRequestID, requestID) // Set in response
		c.Next()
	}
}
