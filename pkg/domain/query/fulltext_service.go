package query

import (
	"context"
	"fmt"
	"strings"

	"github.com/turtacn/dataseap/pkg/adapter/starrocks"
	"github.com/turtacn/dataseap/pkg/common/errors"
	commontypes "github.com/turtacn/dataseap/pkg/common/types"
	metadataService "github.com/turtacn/dataseap/pkg/domain/management/metadata"
	"github.com/turtacn/dataseap/pkg/domain/query/model"
	"github.com/turtacn/dataseap/pkg/logger"
)

// fullTextSearchSubServiceImpl implements FullTextSearchSubService.
// fullTextSearchSubServiceImpl 实现 FullTextSearchSubService 接口。
type fullTextSearchSubServiceImpl struct {
	starrocksClient starrocks.Client
	metadataSvc     metadataService.Service // To get info about indexed tables/fields
}

// NewFullTextSearchSubService creates a new instance of the full-text search sub-service.
// NewFullTextSearchSubService 创建一个新的全文检索子服务实例。
func NewFullTextSearchSubService(srClient starrocks.Client, metaSvc metadataService.Service) FullTextSearchSubService {
	return &fullTextSearchSubServiceImpl{
		starrocksClient: srClient,
		metadataSvc:     metaSvc,
	}
}

// Search performs the full-text search logic.
// Search 执行全文检索逻辑。
func (s *fullTextSearchSubServiceImpl) Search(ctx context.Context, req *model.FullTextSearchRequest) (*model.FullTextSearchResult, error) {
	l := logger.L().Ctx(ctx).With("method", "FullTextSearchSubService.Search", "keywords", req.Keywords)
	l.Info("Performing full-text search")

	// 1. Determine target tables and fields
	var tablesToSearch []string
	if len(req.TargetTables) > 0 {
		tablesToSearch = req.TargetTables
	} else {
		// If no tables specified, try to get all tables with relevant inverted indexes
		// This requires metadataSvc to provide such information. Placeholder:
		// allTables, _, err := s.metadataSvc.ListTables(ctx, "default_database" /* or from config */, nil)
		// if err != nil {
		// 	return nil, errors.Wrap(err, errors.InternalError, "failed to list tables for full-text search")
		// }
		// tablesToSearch = allTables // Further filter these by checking for indexed fields
		l.Warn("No target tables specified, and auto-discovery of indexed tables is not yet implemented. Search may fail or be limited.")
		// For now, if no tables specified, we might return an error or search a predefined set.
		// Let's assume an error if no tables are specified and auto-discovery isn't ready.
		return nil, errors.New(errors.InvalidArgument, "TargetTables must be specified for full-text search (auto-discovery not yet implemented)")
	}

	var allHits []*model.SearchHit
	var totalHitsAcrossTables int64 = 0

	// TODO: Implement pagination across multiple table results.
	// This is complex: need to fetch more than page_size from each, sort globally (by score/time), then paginate.
	// For a simpler initial version, pagination might apply per-table or be an estimate.
	// Here, we'll collect all results and then truncate if needed, which isn't true cross-table pagination for large sets.

	for _, tableName := range tablesToSearch {
		// 2. For each table, determine searchable fields (if req.TargetFields is empty)
		var fieldsInTable []string
		if len(req.TargetFields) > 0 {
			fieldsInTable = req.TargetFields // Use specified fields
		} else {
			// Get schema or indexed fields for this table from metadataSvc
			// tableSchema, err := s.metadataSvc.GetTableSchema(ctx, "your_database_name", tableName)
			// if err != nil {
			//    l.Warnw("Failed to get schema for table, skipping", "table", tableName, "error", err)
			//    continue
			// }
			// for _, field := range tableSchema.Fields {
			//    // Check if field has inverted index (metadataSvc should provide this)
			//    // if field.HasInvertedIndex && (field.DataType == enum.DataTypeString || field.DataType == enum.DataTypeVarchar) {
			//    //    fieldsInTable = append(fieldsInTable, field.Name)
			//    // }
			// }
			// if len(fieldsInTable) == 0 {
			//    l.Infow("No text-indexed fields found or specified for table, skipping", "table", tableName)
			//    continue
			// }
			return nil, errors.Newf(errors.InvalidArgument, "TargetFields must be specified for table %s (auto-discovery not yet implemented)", tableName)
		}

		// 3. Construct MATCH query for StarRocks
		// Example: SELECT *, relevance_score() as score FROM table WHERE (field1 MATCH 'keywords' OR field2 MATCH 'keywords')
		// The match_type (MATCH_ANY or MATCH_ALL) depends on req.RecallPriority.
		matchOperator := "MATCH_ANY" // Recall priority
		if !req.RecallPriority {
			matchOperator = "MATCH_ALL"
		}

		var matchClauses []string
		for _, field := range fieldsInTable {
			// Sanitize field name if necessary (though usually not for StarRocks column names)
			// Sanitize keywords: escape special characters for StarRocks MATCH syntax if any.
			// For now, assume keywords are simple.
			matchClauses = append(matchClauses, fmt.Sprintf("%s %s '%s'", field, matchOperator, req.Keywords))
		}
		if len(matchClauses) == 0 {
			l.Infow("No fields to search in table after filtering", "table", tableName)
			continue
		}
		whereClause := strings.Join(matchClauses, " OR ")

		// TODO: Include time range filter and additional filters from req
		// if req.TimeRangeFilter != nil { ... whereClause += AND timestamp_field BETWEEN ... }
		// if len(req.AdditionalFilters) > 0 { ... whereClause += AND key = value ... }

		// Include all columns for `Document` field in SearchHit
		// Also select a pseudo score column if possible (StarRocks MATCH doesn't directly provide a score like ES)
		// We might need to select primary keys or unique IDs for the `ID` field in SearchHit.
		// For simplicity, we select all columns.
		// A pseudo-score might be hard with just MATCH. True relevance scoring is complex.
		sql := fmt.Sprintf("SELECT * FROM %s WHERE %s", tableName, whereClause)

		// Add pagination and sorting to the SQL if applicable per table
		limit := 1000 // Default large limit per table before global pagination
		offset := 0
		if req.Pagination != nil {
			// This per-table pagination isn't true global pagination.
			// limit = req.Pagination.GetLimit()
			// offset = req.Pagination.GetOffset()
		}
		sql += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

		l.Debugw("Executing full-text SQL query for table", "table", tableName, "sql", sql)
		srResult, err := s.starrocksClient.Execute(ctx, sql)
		if err != nil {
			l.Errorw("Failed to execute full-text search query for table", "table", tableName, "sql", sql, "error", err)
			// Collect errors or fail fast? For now, continue and try other tables.
			continue
		}

		// 4. Transform results to []*model.SearchHit
		for _, srRow := range srResult.Rows {
			doc := make(map[string]interface{})
			hitFieldsMap := make(map[string]string) // TODO: Populate with actual snippets if possible
			var docID string

			for j, colName := range srResult.Columns {
				if j < len(srRow) {
					doc[colName] = srRow[j]
					// Try to find a primary key or unique ID
					if strings.ToLower(colName) == "id" || strings.HasSuffix(strings.ToLower(colName), "_id") {
						if idVal, ok := srRow[j].(string); ok {
							docID = idVal
						} else if idValInt, ok := srRow[j].(int64); ok {
							docID = fmt.Sprintf("%d", idValInt)
						}
					}
					// Basic check if this field was part of the match (crude)
					for _, searchedField := range fieldsInTable {
						if colName == searchedField {
							if valStr, ok := srRow[j].(string); ok && strings.Contains(strings.ToLower(valStr), strings.ToLower(req.Keywords)) {
								// This is a very basic way to get a "snippet", real snippet generation is complex
								snippet := valStr
								if len(snippet) > 100 {
									snippet = snippet[:100] + "..."
								}
								hitFieldsMap[colName] = snippet
							}
						}
					}
				}
			}

			hit := &model.SearchHit{
				SourceTable: tableName,
				ID:          docID,
				Document:    doc,
				HitFields:   hitFieldsMap,
				Score:       1.0, // Placeholder score
			}
			allHits = append(allHits, hit)
		}
		// totalHitsAcrossTables += srResult.TotalRowsIfAvailable // Assuming QueryResult provides this for pagination
		if srResult.Stats != nil {
			//粗略统计，不是精确的
			totalHitsAcrossTables += int64(len(srResult.Rows))
		} else {
			totalHitsAcrossTables += int64(len(srResult.Rows))
		}
	}

	// TODO: Implement global sorting (e.g., by score if available, or timestamp) and then apply global pagination.
	// For now, just return collected hits, possibly truncated.

	paginatedHits := allHits
	finalPagination := req.Pagination

	if req.Pagination != nil {
		start := req.Pagination.GetOffset()
		end := start + req.Pagination.GetLimit()
		if start < 0 {
			start = 0
		}
		if start >= len(allHits) {
			paginatedHits = []*model.SearchHit{}
		} else {
			if end > len(allHits) {
				end = len(allHits)
			}
			paginatedHits = allHits[start:end]
		}
		// Update pagination response based on actual total hits
		// totalHitsAcrossTables should be the count *before* this client-side pagination
		finalPaginationResp := &commontypes.PaginationResponse{
			Page:     req.Pagination.Page,
			PageSize: req.Pagination.PageSize,
			Total:    totalHitsAcrossTables, // This is an estimate across tables without proper global count
		}
		// Update the result with this pagination response
		// For now, this part is simplified
		_ = finalPaginationResp // Placeholder usage
	}

	return &model.FullTextSearchResult{
		Hits:       paginatedHits,
		Pagination: finalPagination,       // This should be a PaginationResponse object
		TotalHits:  totalHitsAcrossTables, // This is an estimate
	}, nil
}
