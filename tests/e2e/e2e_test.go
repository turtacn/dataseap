package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/turtacn/dataseap/pkg/logger"
	// Generated gRPC client
	// apiv1 "github.com/turtacn/dataseap/api/v1"
	// "google.golang.org/grpc"
	// "google.golang.org/grpc/credentials/insecure"
)

const (
	httpBaseURL = "http://localhost:8080/api/v1" // Default HTTP API port
	grpcAddress = "localhost:50051"              // Default gRPC API port
)

// TestMain would ideally start the dataseap-server and its dependencies (e.g., StarRocks via Docker Compose).
// For this skeleton, we assume the server is already running or started externally.
func TestMain(m *testing.M) {
	logger.L().Info("E2E Test Setup: Assuming DataSeaP server and dependencies (StarRocks) are running externally.")

	// TODO: Programmatic server start/stop for isolated E2E tests
	// For example, using docker-compose up -d and down, and starting the app server.
	// This requires significant setup.

	// Check if server is accessible (basic health check)
	if os.Getenv("SKIP_E2E_TESTS") != "true" {
		err := waitForServerReady(httpBaseURL+"/../health", 10*time.Second) // Adjust health endpoint
		if err != nil {
			logger.L().Fatalf("E2E Test Setup: Server not ready or accessible: %v. Ensure server is running.", err)
			// os.Exit(1) // Or skip tests
		}
		logger.L().Info("E2E Test Setup: Server is accessible.")
	}

	exitCode := m.Run()
	os.Exit(exitCode)
}

func waitForServerReady(healthURL string, timeout time.Duration) error {
	startTime := time.Now()
	for {
		if time.Since(startTime) > timeout {
			return fmt.Errorf("server not ready after %v", timeout)
		}
		resp, err := http.Get(healthURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		logger.L().Debugf("Waiting for server at %s (Error: %v, Status: %s)", healthURL, err, resp.Status)
		time.Sleep(500 * time.Millisecond)
	}
}

// TestBasicIngestAndQueryE2E is a placeholder for an end-to-end test.
// TestBasicIngestAndQueryE2E 是一个端到端测试的占位符。
func TestBasicIngestAndQueryE2E_Placeholder(t *testing.T) {
	if os.Getenv("SKIP_E2E_TESTS") == "true" {
		t.Skip("Skipping E2E tests as SKIP_E2E_TESTS is set.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// --- Phase 1: Ingest Data via HTTP API ---
	t.Run("IngestDataViaHTTP", func(t *testing.T) {
		ingestURL := httpBaseURL + "/ingest/events"

		// Using a structure similar to apiv1.IngestDataRequest for the body
		// Note: Ensure your actual DTOs for HTTP match what the server expects.
		// Here, we are using a map that can be marshalled to JSON matching apiv1.IngestDataRequest.
		eventTimestamp := time.Now().UTC().Add(-1 * time.Minute) // Event slightly in the past

		// Constructing a body that matches apiv1.IngestDataRequest structure
		ingestReqBody := map[string]interface{}{
			"records": []map[string]interface{}{
				{
					"dataSourceId": "e2e-test-source",
					"dataType":     "e2e_test_logs", // This should map to a table in StarRocks
					"timestamp":    eventTimestamp.Format(time.RFC3339Nano),
					"data": map[string]interface{}{
						"message":   "E2E test log event for query",
						"level":     "INFO",
						"e2e_tag":   "verify_me_" + fmt.Sprintf("%d", time.Now().UnixNano()),
						"unique_id": fmt.Sprintf("e2e-%d", time.Now().UnixNano()), // Unique ID for query
					},
					"tags": map[string]string{"env": "e2e"},
				},
			},
			"requestId": "e2e-ingest-" + fmt.Sprintf("%d", time.Now().UnixNano()),
		}

		reqBodyBytes, err := json.Marshal(ingestReqBody)
		if err != nil {
			t.Fatalf("Failed to marshal ingest request body: %v", err)
		}

		logger.L().Infof("E2E Test: Sending ingest request to %s with body: %s", ingestURL, string(reqBodyBytes))
		resp, err := http.Post(ingestURL, "application/json", bytes.NewBuffer(reqBodyBytes))
		if err != nil {
			t.Fatalf("Failed to send ingest request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var body []byte
			body, _ = io.ReadAll(resp.Body)
			t.Fatalf("Ingest request failed: status %s, body: %s", resp.Status, string(body))
		}

		var ingestResp map[string]interface{} // Or a proper struct matching your APIResponse
		if err := json.NewDecoder(resp.Body).Decode(&ingestResp); err != nil {
			t.Fatalf("Failed to decode ingest response: %v", err)
		}
		logger.L().Infof("E2E Test: Ingest response: %+v", ingestResp)
		// Add assertions for ingestResp if needed, e.g., success=true, ingestedCount=1
		if ingestedCount, ok := ingestResp["data"].(map[string]interface{})["ingestedCount"].(float64); !ok || ingestedCount < 1 {
			t.Errorf("Expected at least 1 event ingested, response: %+v", ingestResp)
		}

		// Give some time for data to be queryable in StarRocks (Stream Load latency)
		time.Sleep(3 * time.Second) // Adjust as needed
	})

	// --- Phase 2: Query Data via HTTP API (or gRPC) ---
	t.Run("QueryDataViaHTTP", func(t *testing.T) {
		// This requires the data from Phase 1 to be queryable.
		// The SQL query needs to target the table where "e2e_test_logs" data was ingested.
		// The unique_id or e2e_tag can be used for specific verification.

		// queryURL := httpBaseURL + "/query/sql"
		// queryReqBody := map[string]interface{}{
		//	 "sql": "SELECT message, level, e2e_tag FROM e2e_test_logs WHERE e2e_tag = '... specific tag from above ...' LIMIT 1",
		// }
		// ... (similar HTTP POST logic as ingestion) ...
		// ... (decode response and assert data) ...

		logger.L().Warn("E2E Test: Query verification phase is a placeholder.")
		t.Log("Placeholder: Query and verification of ingested data would go here.")
		// Example of what might be checked:
		// - Correct number of rows returned
		// - Content of the rows matches the ingested data
	})

	// --- Phase 3: (Optional) Test gRPC Endpoints ---
	// t.Run("QueryDataViaGRPC", func(t *testing.T) {
	//  conn, err := grpc.DialContext(ctx, grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	//	if err != nil {
	//		t.Fatalf("Failed to connect to gRPC server: %v", err)
	//	}
	//	defer conn.Close()
	//
	//	queryClient := apiv1.NewQueryServiceClient(conn)
	//	grpcReq := &apiv1.ExecuteSQLQueryRequest{
	//		SqlQuery: "SELECT COUNT(*) FROM e2e_test_logs WHERE e2e_tag = '...'",
	//	}
	//  grpcResp, err := queryClient.ExecuteSQLQuery(ctx, grpcReq)
	//  // ... assertions ...
	// })
}

// Add more E2E tests for:
// - Full-text search functionality
// - Management APIs (if applicable and safe for E2E)
// - Error responses from APIs
// - Pagination and filtering
