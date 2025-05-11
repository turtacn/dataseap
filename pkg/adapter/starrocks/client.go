package starrocks

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/turtacn/dataseap/pkg/common/errors"
	"github.com/turtacn/dataseap/pkg/config"
	"github.com/turtacn/dataseap/pkg/logger"
)

// starrocksClient implements the Client interface for StarRocks.
// starrocksClient 实现StarRocks的Client接口。
type starrocksClient struct {
	cfg        config.StarRocksConfig
	httpClient *http.Client
	feHosts    []string // FE节点列表 (host:http_port) FE node list (host:http_port)
	loadURLIdx int      // 用于轮询Load URL的索引 Index for round-robin Load URL
	mu         sync.RWMutex
}

// NewClient creates a new StarRocks client.
// NewClient 创建一个新的StarRocks客户端。
func NewClient(cfg config.StarRocksConfig) (Client, error) {
	if len(cfg.Hosts) == 0 {
		return nil, errors.New(errors.ConfigError, "StarRocks FE hosts are not configured")
	}

	// 确保FE Host格式为 host:port
	// Ensure FE Host format is host:port
	parsedHosts := make([]string, 0, len(cfg.Hosts))
	for _, h := range cfg.Hosts {
		if !strings.Contains(h, ":") {
			// 如果没有端口，使用配置中的QueryPort或默认8030
			// If no port, use QueryPort from config or default 8030
			port := cfg.QueryPort
			if port == 0 {
				port = 8030 // Default StarRocks FE HTTP port
			}
			parsedHosts = append(parsedHosts, fmt.Sprintf("%s:%d", h, port))
		} else {
			parsedHosts = append(parsedHosts, h)
		}
	}

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20, // Increased for potential multiple FE nodes
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false, // Enable compression if StarRocks supports it well for API
	}

	timeout := time.Duration(cfg.ConnectTimeout) * time.Second
	if cfg.ConnectTimeout == 0 {
		timeout = 10 * time.Second // Default connect timeout
	}

	return &starrocksClient{
		cfg:     cfg,
		feHosts: parsedHosts,
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		loadURLIdx: rand.Intn(len(parsedHosts)), // 随机起始点 Random starting point
	}, nil
}

// getQueryURL 选择一个FE节点用于查询
// getQueryURL selects an FE node for querying (round-robin).
func (c *starrocksClient) getQueryURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	// Simple round-robin for query FEs
	// For high availability, a more sophisticated health check and selection might be needed.
	idx := rand.Intn(len(c.feHosts))
	return fmt.Sprintf("http://%s/api/%s", c.feHosts[idx], c.cfg.Database)
}

// getStreamLoadURL 选择一个FE节点用于Stream Load
// getStreamLoadURL selects an FE node for Stream Load (round-robin).
func (c *starrocksClient) getStreamLoadURL(database, table string) string {
	c.mu.Lock() // Lock for updating loadURLIdx
	idx := c.loadURLIdx
	c.loadURLIdx = (c.loadURLIdx + 1) % len(c.feHosts)
	c.mu.Unlock()
	// Example: http://fe_host:http_port/api/{db}/{table}/_stream_load
	return fmt.Sprintf("http://%s/api/%s/%s/_stream_load", c.feHosts[idx], database, table)
}

// Execute performs a DQL or DML query.
// Execute 执行 DQL 或 DML 查询。
func (c *starrocksClient) Execute(ctx context.Context, query string, args ...interface{}) (*QueryResult, error) {
	l := logger.L().With("method", "Execute", "query", query) // Basic logging

	// 注意：StarRocks的HTTP API /api/{db}/{table}/_query 或 /api/query/action
	// 不直接支持参数化查询像 database/sql 那样。
	// 如果需要参数化，需要在应用层安全地构建SQL，或使用JDBC/ODBC。
	// 此处假设 query 已经是完整的SQL。
	// Note: StarRocks HTTP API /api/{db}/{table}/_query or /api/query/action
	// does not directly support parameterized queries like database/sql.
	// If parameterization is needed, SQL must be built safely at the application layer, or use JDBC/ODBC.
	// Here, it's assumed 'query' is already the complete SQL.
	if len(args) > 0 {
		// For safety, prevent accidental use of args if not properly handled
		l.Warnw("Execute called with args, but StarRocks HTTP API does not support native parameterization. Ensure query is pre-formatted safely.", "argsCount", len(args))
		// return nil, errors.New(errors.InvalidArgument, "parameterized queries via this basic HTTP client are not directly supported; pre-format your SQL safely")
	}

	// 使用 /api/query/action 端点执行SQL
	// Use /api/query/action endpoint to execute SQL
	// FE host is chosen by getQueryURL which internally uses c.cfg.Database for the path construction.
	// For a generic query API, we might need a different URL construction.
	// Let's assume a generic query URL for now.
	// The path /api/{db} used in getQueryURL() is for _Stream_Load, not generic query.
	// A common endpoint for SQL is often /api/query or similar.
	// StarRocks documentation suggests POST to /api/v1/query for SQL statements.
	// Let's assume we pick one FE and construct /api/v1/query for it.

	feHostOnly := strings.Split(c.feHosts[rand.Intn(len(c.feHosts))], ":")[0]
	queryPort := c.cfg.QueryPort
	if queryPort == 0 {
		queryPort = 8030
	}
	srURL := fmt.Sprintf("http://%s:%d/api/v1/query", feHostOnly, queryPort)

	payload := map[string]string{"sql": query}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		l.Errorw("Failed to marshal query payload", "error", err)
		return nil, errors.Wrap(err, errors.SerializationError, "failed to marshal query payload")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, srURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		l.Errorw("Failed to create HTTP request", "url", srURL, "error", err)
		return nil, errors.Wrap(err, errors.NetworkError, "failed to create HTTP request for StarRocks query")
	}

	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.cfg.User+":"+c.cfg.Password)))
	if c.cfg.Database != "" {
		req.Header.Set("Database", c.cfg.Database) // Set database via header
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		l.Errorw("Failed to execute StarRocks query", "url", srURL, "error", err)
		return nil, errors.Wrap(err, errors.NetworkError, "failed to execute StarRocks query")
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Errorw("Failed to read StarRocks query response body", "url", srURL, "error", err)
		return nil, errors.Wrap(err, errors.NetworkError, "failed to read StarRocks query response body")
	}

	if resp.StatusCode != http.StatusOK {
		l.Errorw("StarRocks query failed", "url", srURL, "status", resp.Status, "response", string(bodyBytes))
		return nil, errors.Newf(errors.DatabaseError, "StarRocks query failed: %s, Response: %s", resp.Status, string(bodyBytes))
	}

	var srResp struct {
		Msg  string `json:"msg"`
		Code int    `json:"code"` // 0 for success
		Data struct {
			Type string `json:"type"` // schema or result_set
			Meta []struct {
				Name string `json:"name"`
				Type string `json:"type"`
			} `json:"meta"`
			Result   [][]interface{}        `json:"result"`
			Property map[string]interface{} `json:"property"` // Contains stats like "Affected Rows", "Time", etc.
		} `json:"data"`
		Count int `json:"count"` // Usually 0 for queries
	}

	if err := json.Unmarshal(bodyBytes, &srResp); err != nil {
		l.Errorw("Failed to unmarshal StarRocks query response JSON", "response", string(bodyBytes), "error", err)
		return nil, errors.Wrapf(err, errors.DeserializationError, "failed to unmarshal StarRocks response: %s", string(bodyBytes))
	}

	if srResp.Code != 0 {
		l.Errorw("StarRocks query returned error code", "code", srResp.Code, "message", srResp.Msg, "response", string(bodyBytes))
		return nil, errors.Newf(errors.DatabaseError, "StarRocks query error: Code %d, Msg: %s", srResp.Code, srResp.Msg)
	}

	queryResult := &QueryResult{
		Rows:  srResp.Data.Result,
		Stats: &QueryStats{},
	}
	for _, m := range srResp.Data.Meta {
		queryResult.Columns = append(queryResult.Columns, m.Name)
	}

	// Extract stats from property map
	if srResp.Data.Property != nil {
		if val, ok := srResp.Data.Property["Affected Rows"].(float64); ok { // JSON numbers are float64
			// This is more for DML, but API might return it
		}
		if val, ok := srResp.Data.Property["Time"].(string); ok { // e.g., "23ms"
			// Parse time string if needed
			queryResult.Stats.Message = fmt.Sprintf("Time: %s", val)
		}
	}

	return queryResult, nil
}

// StreamLoad ingests data using StarRocks Stream Load.
// StreamLoad 使用StarRocks Stream Load导入数据。
func (c *starrocksClient) StreamLoad(ctx context.Context, database, table string, data io.Reader, opts *StreamLoadOptions) (*StreamLoadResponse, error) {
	l := logger.L().With("method", "StreamLoad", "database", database, "table", table)

	if opts == nil {
		opts = &StreamLoadOptions{} // Use default options if nil
	}
	if opts.Format == "" {
		opts.Format = "json" // Default to JSON format
	}
	if opts.TimeoutSeconds == 0 {
		opts.TimeoutSeconds = 300 // Default timeout 5 minutes
	}

	loadURL := c.getStreamLoadURL(database, table)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, loadURL, data)
	if err != nil {
		l.Errorw("Failed to create StreamLoad HTTP request", "url", loadURL, "error", err)
		return nil, errors.Wrap(err, errors.NetworkError, "failed to create StreamLoad HTTP request")
	}

	// Set headers
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.cfg.User+":"+c.cfg.Password)))
	req.Header.Set("Expect", "100-continue")                   // Required for Stream Load
	req.Header.Set("Content-Type", "application/octet-stream") // Or specific if known and required

	// Standard Stream Load headers
	req.Header.Set("format", opts.Format)
	if opts.Format == "csv" && opts.ColumnSeparator != "" {
		req.Header.Set("column_separator", opts.ColumnSeparator)
	}
	if opts.Format == "csv" && opts.RowDelimiter != "" {
		req.Header.Set("row_delimiter", opts.RowDelimiter)
	}
	if opts.Format == "json" && opts.StripOuterArray {
		req.Header.Set("strip_outer_array", "true")
	}
	if opts.MaxFilterRatio > 0 {
		req.Header.Set("max_filter_ratio", fmt.Sprintf("%f", opts.MaxFilterRatio))
	}
	req.Header.Set("timeout", strconv.Itoa(opts.TimeoutSeconds))

	if opts.TwoPhaseCommit && opts.TransactionID != "" {
		req.Header.Set("txn_id", opts.TransactionID)
		req.Header.Set("two_phase_commit", "true")
	}

	// Custom headers (e.g., for CSV columns)
	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}
	if opts.MergeCondition != "" {
		req.Header.Set("merge_condition", opts.MergeCondition)
	}

	l.Infow("Executing StreamLoad", "url", loadURL, "headers", req.Header)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		l.Errorw("StreamLoad request failed", "url", loadURL, "error", err)
		return nil, errors.Wrap(err, errors.NetworkError, "StreamLoad request failed")
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Errorw("Failed to read StreamLoad response body", "url", loadURL, "error", err)
		return nil, errors.Wrap(err, errors.NetworkError, "failed to read StreamLoad response body")
	}

	var srResp StreamLoadResponse
	if err := json.Unmarshal(bodyBytes, &srResp); err != nil {
		l.Errorw("Failed to unmarshal StreamLoad response JSON", "response", string(bodyBytes), "error", err)
		// Try to get basic status if unmarshal fails
		if resp.StatusCode != http.StatusOK {
			return nil, errors.Wrapf(err, errors.DeserializationError, "StreamLoad failed with status %s and unparsable body: %s", resp.Status, string(bodyBytes))
		}
		return nil, errors.Wrapf(err, errors.DeserializationError, "failed to unmarshal StreamLoad response: %s", string(bodyBytes))
	}

	if srResp.Status != "Success" && srResp.Status != "Publish Timeout" { // Publish Timeout can sometimes be treated as a soft failure/retryable
		l.Errorw("StreamLoad operation failed", "status", srResp.Status, "message", srResp.Message, "response", string(bodyBytes))
		// Return the full response even on failure, as it contains useful info like ErrorURL
		return &srResp, errors.Newf(errors.DatabaseError, "StreamLoad failed: Status %s, Message: %s, ErrorURL: %s", srResp.Status, srResp.Message, srResp.ErrorURL)
	}

	l.Infow("StreamLoad completed", "status", srResp.Status, "loadedRows", srResp.NumberLoadedRows, "totalRows", srResp.NumberTotalRows)
	return &srResp, nil
}

// BeginTransaction begins a two-phase commit transaction for Stream Load.
// BeginTransaction 开始一个两阶段提交事务 (用于Stream Load)。
func (c *starrocksClient) BeginTransaction(ctx context.Context, database, table, label string, timeoutSeconds int) (int64, error) {
	l := logger.L().With("method", "BeginTransaction", "database", database, "table", table, "label", label)

	// API: PUT /api/{db}/{table}/_stream_load_2pc?txn_action=begin
	// Requires a unique label for the transaction.
	urlStr := fmt.Sprintf("%s/api/%s/%s/_stream_load_2pc?txn_action=begin", c.getFeBaseURL(), database, table)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, urlStr, nil) // No body for begin
	if err != nil {
		return 0, errors.Wrap(err, errors.NetworkError, "failed to create begin transaction request")
	}

	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.cfg.User+":"+c.cfg.Password)))
	req.Header.Set("label", label)
	req.Header.Set("timeout", strconv.Itoa(timeoutSeconds))
	req.Header.Set("Expect", "100-continue")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, errors.Wrap(err, errors.NetworkError, "begin transaction request failed")
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	var srResp struct {
		Status string `json:"Status"`
		TxnID  int64  `json:"TxnId"`
		Msg    string `json:"msg"`
	}

	if err := json.Unmarshal(bodyBytes, &srResp); err != nil {
		return 0, errors.Wrapf(err, errors.DeserializationError, "failed to unmarshal begin transaction response: %s", string(bodyBytes))
	}

	if srResp.Status != "Success" {
		l.Errorw("Failed to begin transaction", "status", srResp.Status, "msg", srResp.Msg)
		return 0, errors.Newf(errors.DatabaseError, "failed to begin transaction: %s - %s", srResp.Status, srResp.Msg)
	}

	l.Infow("Transaction begun successfully", "txn_id", srResp.TxnID)
	return srResp.TxnID, nil
}

// CommitTransaction commits a two-phase commit transaction.
// CommitTransaction 提交一个两阶段提交事务。
func (c *starrocksClient) CommitTransaction(ctx context.Context, database string, txnID int64) error {
	l := logger.L().With("method", "CommitTransaction", "database", database, "txn_id", txnID)

	// API: PUT /api/{db}/_stream_load_2pc?txn_action=commit&txn_id={txn_id}
	urlStr := fmt.Sprintf("%s/api/%s/_stream_load_2pc?txn_action=commit&txn_id=%d", c.getFeBaseURL(), database, txnID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, urlStr, nil)
	if err != nil {
		return errors.Wrap(err, errors.NetworkError, "failed to create commit transaction request")
	}
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.cfg.User+":"+c.cfg.Password)))
	req.Header.Set("Expect", "100-continue")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, errors.NetworkError, "commit transaction request failed")
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	var srResp struct {
		Status string `json:"Status"`
		Msg    string `json:"msg"`
	}
	if err := json.Unmarshal(bodyBytes, &srResp); err != nil {
		return errors.Wrapf(err, errors.DeserializationError, "failed to unmarshal commit transaction response: %s", string(bodyBytes))
	}

	if srResp.Status != "Success" {
		l.Errorw("Failed to commit transaction", "status", srResp.Status, "msg", srResp.Msg)
		return errors.Newf(errors.DatabaseError, "failed to commit transaction: %s - %s", srResp.Status, srResp.Msg)
	}
	l.Infow("Transaction committed successfully", "txn_id", txnID)
	return nil
}

// AbortTransaction aborts a two-phase commit transaction.
// AbortTransaction 中止一个两阶段提交事务。
func (c *starrocksClient) AbortTransaction(ctx context.Context, database string, txnID int64) error {
	l := logger.L().With("method", "AbortTransaction", "database", database, "txn_id", txnID)

	// API: PUT /api/{db}/_stream_load_2pc?txn_action=abort&txn_id={txn_id}
	urlStr := fmt.Sprintf("%s/api/%s/_stream_load_2pc?txn_action=abort&txn_id=%d", c.getFeBaseURL(), database, txnID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, urlStr, nil)
	if err != nil {
		return errors.Wrap(err, errors.NetworkError, "failed to create abort transaction request")
	}
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.cfg.User+":"+c.cfg.Password)))
	req.Header.Set("Expect", "100-continue")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, errors.NetworkError, "abort transaction request failed")
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	var srResp struct {
		Status string `json:"Status"`
		Msg    string `json:"msg"`
	}
	if err := json.Unmarshal(bodyBytes, &srResp); err != nil {
		return errors.Wrapf(err, errors.DeserializationError, "failed to unmarshal abort transaction response: %s", string(bodyBytes))
	}

	if srResp.Status != "Success" { // Abort success still returns "Success" status
		l.Errorw("Failed to abort transaction cleanly, or already aborted", "status", srResp.Status, "msg", srResp.Msg)
		// It's not necessarily an error if it was already aborted or doesn't exist.
		// Check srResp.Msg for details like "transaction not found" or "transaction already visible".
	}
	l.Infow("Transaction abort request sent", "txn_id", txnID, "status", srResp.Status, "msg", srResp.Msg)
	return nil
}

// getFeBaseURL returns a base URL for one of the FE nodes.
// getFeBaseURL 返回一个FE节点的基础URL。
func (c *starrocksClient) getFeBaseURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	idx := rand.Intn(len(c.feHosts))
	return fmt.Sprintf("http://%s", c.feHosts[idx])
}

// Close cleans up resources used by the client.
// Close 清理客户端使用的资源。
func (c *starrocksClient) Close() error {
	// HTTP client in Go typically doesn't need explicit closing unless custom transports
	// with specific cleanup are used. The idle connections will be managed by the transport.
	// If we were using database/sql, we'd close the *sql.DB here.
	logger.L().Info("StarRocks client closed (HTTP client managed by transport).")
	return nil
}
