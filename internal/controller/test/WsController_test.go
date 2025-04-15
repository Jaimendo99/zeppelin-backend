package controller_test // Correct package name

import (
	"encoding/json"
	elog "github.com/labstack/gommon/log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	// Import the package under test
	"zeppelin/internal/controller" // Adjust the import path to your actual controller package

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestServer now returns the ConnectionManager instance from the controller package
func setupTestServer(t *testing.T, userID string) (*httptest.Server, string, *controller.ConnectionManager) {
	// Create the manager using the imported package's constructor
	connManager := controller.NewConnectionManager()

	e := echo.New()
	e.Logger.SetLevel(elog.DEBUG)
	e.Logger.SetHeader("${time_rfc3339} ${level} ${prefix} ${file}:${line}")

	authMiddleware := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			t.Logf("DEBUG: AuthMiddleware running for request path: %s", c.Request().URL.Path)
			c.Set("user_id", userID)
			t.Logf("DEBUG: AuthMiddleware set user_id to: %s", userID)
			return next(c)
		}
	}

	// Register the METHOD from the specific instance
	e.GET("/ws", connManager.WebSocketHandler(), authMiddleware)

	server := httptest.NewServer(e)
	t.Logf("DEBUG: Test server started at URL: %s", server.URL)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	t.Logf("DEBUG: WebSocket URL for test: %s", wsURL)

	return server, wsURL, connManager
}

// connectWebSocketNoCleanup remains the same
func connectWebSocketNoCleanup(t *testing.T, wsURL string, platform string) *websocket.Conn {
	// ... (implementation as before) ...
	dialer := websocket.Dialer{}
	query := url.Values{}
	query.Set("platform", platform)

	t.Logf("DEBUG: Attempting to dial WebSocket: %s?%s", wsURL, query.Encode())
	conn, resp, err := dialer.Dial(wsURL+"?"+query.Encode(), nil)
	require.NoError(t, err, "Failed to dial websocket")
	require.NotNil(t, resp, "Response should not be nil")
	require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode, "Expected status switching protocols")
	require.NotNil(t, conn, "Connection should not be nil")
	t.Logf("DEBUG: WebSocket dial successful.")

	err = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	require.NoError(t, err)

	return conn
}

// readJSONMessage remains the same
func readJSONMessage(t *testing.T, conn *websocket.Conn) map[string]interface{} {
	// ... (implementation as before) ...
	t.Logf("DEBUG: Attempting to read message from WebSocket...")
	msgType, msgBytes, err := conn.ReadMessage()
	if err != nil {
		t.Logf("DEBUG: Error reading message: %v", err)
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			t.Fatalf("Timeout reading message: %v", err)
		}
		if closeErr, ok := err.(*websocket.CloseError); ok {
			t.Fatalf("Connection closed while reading message: %v", closeErr)
		}
		t.Fatalf("Failed to read message: %v", err)
	}
	t.Logf("DEBUG: Read message type %d, content: %s", msgType, string(msgBytes))
	assert.Equal(t, websocket.TextMessage, msgType, "Expected text message type")

	var data map[string]interface{}
	err = json.Unmarshal(msgBytes, &data)
	require.NoError(t, err, "Failed to unmarshal JSON message: %s", string(msgBytes))
	t.Logf("DEBUG: Successfully unmarshalled JSON message.")
	return data
}

// --- Test Case ---

func TestWebSocketHandler_ConnectionAndStatus(t *testing.T) {
	userID := "user-123"
	platform := "web"
	// Get the connManager instance (type *controller.ConnectionManager)
	server, wsURL, connManager := setupTestServer(t, userID)
	defer server.Close()

	conn := connectWebSocketNoCleanup(t, wsURL, platform)

	defer func() {
		t.Logf("DEBUG: Test defer: Sending close message.")
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		time.Sleep(10 * time.Millisecond)
		t.Logf("DEBUG: Test defer: Closing connection.")
		_ = conn.Close()
		t.Logf("DEBUG: Test defer: Waiting after close.")
		time.Sleep(50 * time.Millisecond)
		t.Logf("DEBUG: Test defer: Finished.")
	}()

	// --- Check State BEFORE Reading Message using Exported Methods ---
	t.Logf("DEBUG: Test: Waiting briefly before checking state...")
	time.Sleep(50 * time.Millisecond) // Allow handler time to process

	t.Logf("DEBUG: Test: Checking state BEFORE reading message for %s.", userID)
	// Use the exported methods - no manual locking needed here
	connectionCount := connManager.GetConnectionCount(userID)
	platformCounts := connManager.GetUserPlatforms(userID)
	t.Logf("DEBUG: Test: Checked state BEFORE reading. Count: %d, Platforms: %v", connectionCount, platformCounts)

	// Assert state BEFORE reading
	require.Equal(t, 1, connectionCount, "Should have 1 connection for the user shortly after connect")
	require.Contains(t, platformCounts, platform, "Platform map should contain the connected platform")
	assert.Equal(t, 1, platformCounts[platform], "Platform count for '%s' should be 1", platform)
	assert.Len(t, platformCounts, 1, "Platform map should only contain one entry")
	// ---------------------------------------------

	// 1. Now, verify the initial status update message
	t.Logf("DEBUG: Test: Attempting to read initial status message for %s.", userID)
	statusMsg := readJSONMessage(t, conn)
	t.Logf("DEBUG: Test: Successfully read initial status message for %s.", userID)

	assert.Equal(t, "status_update", statusMsg["type"])
	assert.Equal(t, userID, statusMsg["user_id"])
	assert.Equal(t, float64(1), statusMsg["connections"]) // Check message content
	statusPlatformsMap, ok := statusMsg["platforms"].(map[string]interface{})
	require.True(t, ok, "Status message 'platforms' should be a map")
	// JSON unmarshals numbers to float64
	assert.Equal(t, float64(1), statusPlatformsMap[platform], "Status message platform count mismatch")
	assert.Len(t, statusPlatformsMap, 1, "Status message platform map size mismatch")

	// 2. Optional: Verify state AGAIN AFTER reading (should still be the same)
	t.Logf("DEBUG: Test: Checking state AGAIN AFTER reading message for %s.", userID)
	connectionCountAfter := connManager.GetConnectionCount(userID)
	platformCountsAfter := connManager.GetUserPlatforms(userID)
	t.Logf("DEBUG: Test: Checked state AFTER reading. Count: %d, Platforms: %v", connectionCountAfter, platformCountsAfter)
	require.Equal(t, 1, connectionCountAfter, "Should still have 1 connection after reading message")
	assert.Equal(t, 1, platformCountsAfter[platform], "Platform count should still be 1 after reading")
	assert.Len(t, platformCountsAfter, 1, "Platform map size should still be 1 after reading")
}

// --- Add other test cases (MultipleConnections, Disconnect, etc.) ---
// Remember to adapt them to use connManager.GetConnectionCount() and
// connManager.GetUserPlatforms() instead of direct map access.

// TestWebSocketHandler_MultipleConnections (Adapted version from previous step)
func TestWebSocketHandler_MultipleConnections(t *testing.T) {
	userID := "user-multi"
	platform1 := "ios"
	platform2 := "android"
	server, wsURL, connManager := setupTestServer(t, userID)
	defer server.Close()

	// Connect first client
	conn1 := connectWebSocketNoCleanup(t, wsURL, platform1)
	defer conn1.Close() // Simple defer for cleanup

	// Read initial status for client 1 & verify state
	statusMsg1 := readJSONMessage(t, conn1)
	assert.Equal(t, float64(1), statusMsg1["connections"])
	time.Sleep(20 * time.Millisecond) // Allow state update
	assert.Equal(t, 1, connManager.GetConnectionCount(userID), "Count should be 1 after first connect")
	assert.Equal(t, 1, connManager.GetUserPlatforms(userID)[platform1], "Platform 1 count should be 1")

	// Connect second client
	conn2 := connectWebSocketNoCleanup(t, wsURL, platform2)
	defer conn2.Close() // Simple defer

	// Read updated status for client 1 (should now show 2 connections)
	statusMsg1_updated := readJSONMessage(t, conn1)
	assert.Equal(t, float64(2), statusMsg1_updated["connections"])
	platformsMap1_updated, _ := statusMsg1_updated["platforms"].(map[string]interface{})
	assert.Equal(t, float64(1), platformsMap1_updated[platform1])
	assert.Equal(t, float64(1), platformsMap1_updated[platform2])

	// Read initial status for client 2 (should show 2 connections)
	statusMsg2 := readJSONMessage(t, conn2)
	assert.Equal(t, float64(2), statusMsg2["connections"])
	platformsMap2, _ := statusMsg2["platforms"].(map[string]interface{})
	assert.Equal(t, float64(1), platformsMap2[platform1])
	assert.Equal(t, float64(1), platformsMap2[platform2])

	// Verify internal state using helpers
	time.Sleep(20 * time.Millisecond) // Allow state update
	assert.Equal(t, 2, connManager.GetConnectionCount(userID), "Count should be 2 after second connect")
	platforms := connManager.GetUserPlatforms(userID)
	assert.Len(t, platforms, 2, "Should have 2 platforms registered")
	assert.Equal(t, 1, platforms[platform1], "Platform 1 count should be 1")
	assert.Equal(t, 1, platforms[platform2], "Platform 2 count should be 1")
}

func TestWebSocketHandler_MessageBroadcasting(t *testing.T) {
	userID := "user-broadcast"
	platform1 := "web"
	platform2 := "desktop"
	// connManager is not strictly needed for assertions here, but setup returns it
	server, wsURL, _ := setupTestServer(t, userID)
	defer server.Close()

	conn1 := connectWebSocketNoCleanup(t, wsURL, platform1)
	defer conn1.Close()           // Ensure cleanup
	_ = readJSONMessage(t, conn1) // Read initial status (1 conn)

	conn2 := connectWebSocketNoCleanup(t, wsURL, platform2)
	defer conn2.Close()           // Ensure cleanup
	_ = readJSONMessage(t, conn1) // Read updated status on conn1 (2 conns)
	_ = readJSONMessage(t, conn2) // Read initial status on conn2 (2 conns)

	// Send message from client 1
	testMessage := []byte("hello from client 1")
	t.Logf("DEBUG: Test Broadcast: Sending message from conn1")
	err := conn1.WriteMessage(websocket.TextMessage, testMessage)
	require.NoError(t, err)
	t.Logf("DEBUG: Test Broadcast: Message sent from conn1")

	// Client 2 should receive the message
	t.Logf("DEBUG: Test Broadcast: Attempting read on conn2")
	// Set a reasonable deadline for receiving the broadcast
	err = conn2.SetReadDeadline(time.Now().Add(1 * time.Second))
	require.NoError(t, err)
	msgType, msgBytes, err := conn2.ReadMessage()
	require.NoError(t, err, "Client 2 failed to read broadcast message")
	assert.Equal(t, websocket.TextMessage, msgType)
	assert.Equal(t, testMessage, msgBytes)
	t.Logf("DEBUG: Test Broadcast: Received broadcast on conn2")

	// Client 1 should NOT receive its own message
	t.Logf("DEBUG: Test Broadcast: Attempting read on conn1 (should timeout)")
	// Set a short deadline - expecting ReadMessage to time out
	err = conn1.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	require.NoError(t, err)
	_, _, err = conn1.ReadMessage()
	assert.Error(t, err, "Client 1 should not have received its own message")
	// Check if the error is a timeout error
	netErr, ok := err.(net.Error)
	assert.True(t, ok && netErr.Timeout(), "Expected a timeout error on conn1, got: %v", err)
	t.Logf("DEBUG: Test Broadcast: Correctly timed out reading on conn1")

	// Reset deadline for cleanup
	err = conn1.SetReadDeadline(time.Time{}) // No deadline
	require.NoError(t, err)
	err = conn2.SetReadDeadline(time.Time{}) // No deadline
	require.NoError(t, err)
}

func TestWebSocketHandler_Disconnect(t *testing.T) {
	userID := "user-disconnect"
	platform1 := "web"
	platform2 := "mobile"
	server, wsURL, connManager := setupTestServer(t, userID)
	defer server.Close()

	conn1 := connectWebSocketNoCleanup(t, wsURL, platform1)
	// No defer for conn1 yet, we close it manually

	conn2 := connectWebSocketNoCleanup(t, wsURL, platform2)
	defer conn2.Close() // conn2 stays open until test end

	// Read initial messages to clear buffers and confirm setup
	_ = readJSONMessage(t, conn1) // conn1 status (1 conn)
	_ = readJSONMessage(t, conn1) // conn1 status update (2 conns)
	_ = readJSONMessage(t, conn2) // conn2 status (2 conns)

	// Verify initial state
	time.Sleep(20 * time.Millisecond) // Allow potential state updates
	require.Equal(t, 2, connManager.GetConnectionCount(userID), "Should have 2 connections initially")

	// Disconnect client 1
	t.Logf("DEBUG: Test Disconnect: Closing conn1")
	// Send close frame first for cleaner server-side handling
	_ = conn1.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	time.Sleep(10 * time.Millisecond)
	err := conn1.Close()
	require.NoError(t, err)
	t.Logf("DEBUG: Test Disconnect: conn1 closed")

	// Client 2 should receive a status update indicating only 1 connection left
	t.Logf("DEBUG: Test Disconnect: Attempting read on conn2 for status update")
	// Set deadline for receiving the update
	err = conn2.SetReadDeadline(time.Now().Add(2 * time.Second))
	require.NoError(t, err)
	statusMsg := readJSONMessage(t, conn2) // This read serves as synchronization
	t.Logf("DEBUG: Test Disconnect: Received message on conn2 after conn1 close")

	assert.Equal(t, "status_update", statusMsg["type"])
	assert.Equal(t, userID, statusMsg["user_id"])
	assert.Equal(t, float64(1), statusMsg["connections"]) // Only conn2 left
	platformsMap, ok := statusMsg["platforms"].(map[string]interface{})
	require.True(t, ok)
	_, exists1 := platformsMap[platform1]
	assert.False(t, exists1, "Platform1 should be removed from status message")
	assert.Equal(t, float64(1), platformsMap[platform2]) // Platform2 should remain

	// Verify internal state using exported methods AFTER the update was received
	// Add a small delay to ensure server goroutine cleanup has likely finished
	time.Sleep(50 * time.Millisecond)
	t.Logf("DEBUG: Test Disconnect: Checking final state via ConnectionManager")
	finalCount := connManager.GetConnectionCount(userID)
	finalPlatforms := connManager.GetUserPlatforms(userID)
	t.Logf("DEBUG: Test Disconnect: Final state - Count: %d, Platforms: %v", finalCount, finalPlatforms)

	assert.Equal(t, 1, finalCount, "Should only have one connection left in manager")
	assert.Len(t, finalPlatforms, 1, "Should only have one platform left in manager")
	assert.Equal(t, 1, finalPlatforms[platform2], "Remaining platform should be platform2")

	// Reset deadline
	err = conn2.SetReadDeadline(time.Time{})
	require.NoError(t, err)
}

func TestWebSocketHandler_NoUserID(t *testing.T) {
	// Setup server BUT pass an empty userID to simulate missing auth info
	// The connManager instance isn't used for assertions here.
	server, wsURL, _ := setupTestServer(t, "") // Pass empty userID
	defer server.Close()

	dialer := websocket.Dialer{}
	query := url.Values{}
	query.Set("platform", "test-no-auth")

	t.Logf("DEBUG: Test NoUserID: Attempting to dial %s?%s", wsURL, query.Encode())
	// Attempt to connect
	conn, resp, err := dialer.Dial(wsURL+"?"+query.Encode(), nil)

	// We expect the *upgrade* itself to fail because the handler returns an error
	// *before* the upgrade happens. Gorilla's dialer might return an error,
	// or the response status code will indicate failure.
	require.Error(t, err, "Expected an error during dial due to failed upgrade")
	if err != nil {
		// Check for the specific "bad handshake" error which gorilla/websocket returns
		// when the server responds with a non-101 status code.
		assert.Contains(t, err.Error(), "bad handshake", "Error should indicate bad handshake")
		t.Logf("DEBUG: Test NoUserID: Received expected dial error: %v", err)
	}
	// The response object might still be non-nil even if err is not nil
	require.NotNil(t, resp, "Response should not be nil even on dial error")
	// The handler returns StatusUnauthorized *before* upgrading.
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Expected Unauthorized status")
	t.Logf("DEBUG: Test NoUserID: Received expected status code %d", resp.StatusCode)

	// Connection should be nil if the dial failed
	assert.Nil(t, conn, "Connection should be nil on failed dial")
}
