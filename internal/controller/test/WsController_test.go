package controller_test

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

	"zeppelin/internal/controller"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServer(t *testing.T, userID string) (*httptest.Server, string, *controller.ConnectionManager) {
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

	e.GET("/ws", connManager.WebSocketHandler(), authMiddleware)

	server := httptest.NewServer(e)
	t.Logf("DEBUG: Test server started at URL: %s", server.URL)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	t.Logf("DEBUG: WebSocket URL for test: %s", wsURL)

	return server, wsURL, connManager
}

func connectWebSocketNoCleanup(t *testing.T, wsURL string, platform string) *websocket.Conn {
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

func readJSONMessage(t *testing.T, conn *websocket.Conn) map[string]interface{} {
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

func TestWebSocketHandler_ConnectionAndStatus(t *testing.T) {
	userID := "user-123"
	platform := "web"
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

	time.Sleep(50 * time.Millisecond)

	t.Logf("DEBUG: Test: Checking state BEFORE reading message for %s.", userID)
	connectionCount := connManager.GetConnectionCount(userID)
	platformCounts := connManager.GetUserPlatforms(userID)
	t.Logf("DEBUG: Test: Checked state BEFORE reading. Count: %d, Platforms: %v", connectionCount, platformCounts)

	require.Equal(t, 1, connectionCount, "Should have 1 connection for the user shortly after connect")
	require.Contains(t, platformCounts, platform, "Platform map should contain the connected platform")
	assert.Equal(t, 1, platformCounts[platform], "Platform count for '%s' should be 1", platform)
	assert.Len(t, platformCounts, 1, "Platform map should only contain one entry")

	t.Logf("DEBUG: Test: Attempting to read initial status message for %s.", userID)
	statusMsg := readJSONMessage(t, conn)
	t.Logf("DEBUG: Test: Successfully read initial status message for %s.", userID)

	assert.Equal(t, "status_update", statusMsg["type"])
	assert.Equal(t, userID, statusMsg["user_id"])
	assert.Equal(t, float64(1), statusMsg["connections"])
	statusPlatformsMap, ok := statusMsg["platforms"].(map[string]interface{})
	require.True(t, ok, "Status message 'platforms' should be a map")
	assert.Equal(t, float64(1), statusPlatformsMap[platform], "Status message platform count mismatch")
	assert.Len(t, statusPlatformsMap, 1, "Status message platform map size mismatch")
}

func TestWebSocketHandler_WebAndMobileConnections(t *testing.T) {
	userID := "user-multi"
	platformWeb := "web"
	platformMobile := "mobile"
	server, wsURL, connManager := setupTestServer(t, userID)
	defer server.Close()

	// Connect web client
	conn1 := connectWebSocketNoCleanup(t, wsURL, platformWeb)
	defer conn1.Close()

	// Read initial status for web client
	statusMsg1 := readJSONMessage(t, conn1)
	assert.Equal(t, float64(1), statusMsg1["connections"])
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, 1, connManager.GetConnectionCount(userID), "Count should be 1 after web connect")
	assert.Equal(t, 1, connManager.GetUserPlatforms(userID)[platformWeb], "Web platform count should be 1")

	// Connect mobile client
	conn2 := connectWebSocketNoCleanup(t, wsURL, platformMobile)
	defer conn2.Close()

	// Read updated status for web client (should show 2 connections)
	statusMsg1_updated := readJSONMessage(t, conn1)
	assert.Equal(t, float64(2), statusMsg1_updated["connections"])
	platformsMap1_updated, _ := statusMsg1_updated["platforms"].(map[string]interface{})
	assert.Equal(t, float64(1), platformsMap1_updated[platformWeb])
	assert.Equal(t, float64(1), platformsMap1_updated[platformMobile])

	// Read initial status for mobile client (should show 2 connections)
	statusMsg2 := readJSONMessage(t, conn2)
	assert.Equal(t, float64(2), statusMsg2["connections"])
	platformsMap2, _ := statusMsg2["platforms"].(map[string]interface{})
	assert.Equal(t, float64(1), platformsMap2[platformWeb])
	assert.Equal(t, float64(1), platformsMap2[platformMobile])

	// Verify internal state
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, 2, connManager.GetConnectionCount(userID), "Count should be 2 after mobile connect")
	platforms := connManager.GetUserPlatforms(userID)
	assert.Len(t, platforms, 2, "Should have 2 platforms registered")
	assert.Equal(t, 1, platforms[platformWeb], "Web platform count should be 1")
	assert.Equal(t, 1, platforms[platformMobile], "Mobile platform count should be 1")
}

func TestWebSocketHandler_SecondWebConnectionRejected(t *testing.T) {
	userID := "user-second-web"
	platform := "web"
	server, wsURL, connManager := setupTestServer(t, userID)
	defer server.Close()

	// Connect first web client
	conn1 := connectWebSocketNoCleanup(t, wsURL, platform)
	defer conn1.Close()
	_ = readJSONMessage(t, conn1) // Read initial status
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, 1, connManager.GetConnectionCount(userID), "Count should be 1 after first web connect")

	// Attempt to connect second web client
	dialer := websocket.Dialer{}
	query := url.Values{}
	query.Set("platform", platform)
	t.Logf("DEBUG: Test SecondWeb: Attempting to dial second web client: %s?%s", wsURL, query.Encode())
	_, resp, err := dialer.Dial(wsURL+"?"+query.Encode(), nil)
	require.Error(t, err, "Expected an error due to second web connection")
	assert.Contains(t, err.Error(), "bad handshake", "Error should indicate bad handshake")
	require.NotNil(t, resp, "Response should not be nil")
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Expected Forbidden status")
	t.Logf("DEBUG: Test SecondWeb: Received expected status code %d", resp.StatusCode)

	// Verify state remains unchanged
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, 1, connManager.GetConnectionCount(userID), "Count should still be 1")
	platforms := connManager.GetUserPlatforms(userID)
	assert.Equal(t, 1, platforms[platform], "Web platform count should still be 1")
}

func TestWebSocketHandler_SecondMobileConnectionRejected(t *testing.T) {
	userID := "user-second-mobile"
	platformWeb := "web"
	platformMobile := "mobile"
	server, wsURL, connManager := setupTestServer(t, userID)
	defer server.Close()

	// Connect web client
	conn1 := connectWebSocketNoCleanup(t, wsURL, platformWeb)
	defer conn1.Close()
	_ = readJSONMessage(t, conn1) // Read initial status

	// Connect first mobile client
	conn2 := connectWebSocketNoCleanup(t, wsURL, platformMobile)
	defer conn2.Close()
	_ = readJSONMessage(t, conn2) // Read initial status
	_ = readJSONMessage(t, conn1) // Read updated status
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, 2, connManager.GetConnectionCount(userID), "Count should be 2 after mobile connect")

	// Attempt to connect second mobile client
	dialer := websocket.Dialer{}
	query := url.Values{}
	query.Set("platform", platformMobile)
	t.Logf("DEBUG: Test SecondMobile: Attempting to dial second mobile client: %s?%s", wsURL, query.Encode())
	_, resp, err := dialer.Dial(wsURL+"?"+query.Encode(), nil)
	require.Error(t, err, "Expected an error due to second mobile connection")
	assert.Contains(t, err.Error(), "bad handshake", "Error should indicate bad handshake")
	require.NotNil(t, resp, "Response should not be nil")
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Expected Forbidden status")
	t.Logf("DEBUG: Test SecondMobile: Received expected status code %d", resp.StatusCode)

	// Verify state remains unchanged
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, 2, connManager.GetConnectionCount(userID), "Count should still be 2")
	platforms := connManager.GetUserPlatforms(userID)
	assert.Equal(t, 1, platforms[platformWeb], "Web platform count should be 1")
	assert.Equal(t, 1, platforms[platformMobile], "Mobile platform count should be 1")
}

func TestWebSocketHandler_MobileWithoutWebRejectedtheater(t *testing.T) {
	userID := "user-no-web"
	platformMobile := "mobile"
	server, wsURL, connManager := setupTestServer(t, userID)
	defer server.Close()

	// Attempt to connect mobile client without web
	dialer := websocket.Dialer{}
	query := url.Values{}
	query.Set("platform", platformMobile)
	t.Logf("DEBUG: Test MobileWithoutWeb: Attempting to dial mobile client: %s?%s", wsURL, query.Encode())
	_, resp, err := dialer.Dial(wsURL+"?"+query.Encode(), nil)
	require.Error(t, err, "Expected an error due to mobile connection without web")
	assert.Contains(t, err.Error(), "bad handshake", "Error should indicate bad handshake")
	require.NotNil(t, resp, "Response should not be nil")
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Expected Forbidden status")
	t.Logf("DEBUG: Test MobileWithoutWeb: Received expected status code %d", resp.StatusCode)

	// Verify no connections were established
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, 0, connManager.GetConnectionCount(userID), "Count should be 0")
}

func TestWebSocketHandler_InvalidPlatform(t *testing.T) {
	userID := "user-invalid-platform"
	platform := "invalid"
	server, wsURL, connManager := setupTestServer(t, userID)
	defer server.Close()

	// Attempt to connect with invalid platform
	dialer := websocket.Dialer{}
	query := url.Values{}
	query.Set("platform", platform)
	t.Logf("DEBUG: Test InvalidPlatform: Attempting to dial with invalid platform: %s?%s", wsURL, query.Encode())
	_, resp, err := dialer.Dial(wsURL+"?"+query.Encode(), nil)
	require.Error(t, err, "Expected an error due to invalid platform")
	assert.Contains(t, err.Error(), "bad handshake", "Error should indicate bad handshake")
	require.NotNil(t, resp, "Response should not be nil")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected Bad Request status")
	t.Logf("DEBUG: Test InvalidPlatform: Received expected status code %d", resp.StatusCode)

	// Verify no connections were established
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, 0, connManager.GetConnectionCount(userID), "Count should be 0")
}

func TestWebSocketHandler_MessageBroadcasting(t *testing.T) {
	userID := "user-broadcast"
	platformWeb := "web"
	platformMobile := "mobile"
	server, wsURL, _ := setupTestServer(t, userID)
	defer server.Close()

	conn1 := connectWebSocketNoCleanup(t, wsURL, platformWeb)
	defer conn1.Close()
	_ = readJSONMessage(t, conn1) // Read initial status (1 conn)

	conn2 := connectWebSocketNoCleanup(t, wsURL, platformMobile)
	defer conn2.Close()
	_ = readJSONMessage(t, conn1) // Read updated status on conn1 (2 conns)
	_ = readJSONMessage(t, conn2) // Read initial status on conn2 (2 conns)

	// Send message from web client
	testMessage := []byte("hello from web")
	t.Logf("DEBUG: Test Broadcast: Sending message from conn1")
	err := conn1.WriteMessage(websocket.TextMessage, testMessage)
	require.NoError(t, err)
	t.Logf("DEBUG: Test Broadcast: Message sent from conn1")

	// Mobile client should receive the message
	t.Logf("DEBUG: Test Broadcast: Attempting read on conn2")
	err = conn2.SetReadDeadline(time.Now().Add(1 * time.Second))
	require.NoError(t, err)
	msgType, msgBytes, err := conn2.ReadMessage()
	require.NoError(t, err, "Mobile client failed to read broadcast message")
	assert.Equal(t, websocket.TextMessage, msgType)
	assert.Equal(t, testMessage, msgBytes)
	t.Logf("DEBUG: Test Broadcast: Received broadcast on conn2")

	// Web client should NOT receive its own message
	t.Logf("DEBUG: Test Broadcast: Attempting read on conn1 (should timeout)")
	err = conn1.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	require.NoError(t, err)
	_, _, err = conn1.ReadMessage()
	assert.Error(t, err, "Web client should not have received its own message")
	netErr, ok := err.(net.Error)
	assert.True(t, ok && netErr.Timeout(), "Expected a timeout error on conn1, got: %v", err)
	t.Logf("DEBUG: Test Broadcast: Correctly timed out reading on conn1")

	// Reset deadlines
	err = conn1.SetReadDeadline(time.Time{})
	require.NoError(t, err)
	err = conn2.SetReadDeadline(time.Time{})
	require.NoError(t, err)
}

func TestWebSocketHandler_Disconnect(t *testing.T) {
	userID := "user-disconnect"
	platformWeb := "web"
	platformMobile := "mobile"
	server, wsURL, connManager := setupTestServer(t, userID)
	defer server.Close()

	conn1 := connectWebSocketNoCleanup(t, wsURL, platformWeb)
	conn2 := connectWebSocketNoCleanup(t, wsURL, platformMobile)
	defer conn2.Close()

	// Read initial messages to clear buffers
	_ = readJSONMessage(t, conn1) // conn1 status (1 conn)
	_ = readJSONMessage(t, conn1) // conn1 status update (2 conns)
	_ = readJSONMessage(t, conn2) // conn2 status (2 conns)

	// Verify initial state
	time.Sleep(20 * time.Millisecond)
	require.Equal(t, 2, connManager.GetConnectionCount(userID), "Should have 2 connections initially")

	// Disconnect web client
	t.Logf("DEBUG: Test Disconnect: Closing conn1")
	_ = conn1.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	time.Sleep(10 * time.Millisecond)
	err := conn1.Close()
	require.NoError(t, err)
	t.Logf("DEBUG: Test Disconnect: conn1 closed")

	// Mobile client should receive a status update indicating only 1 connection left
	t.Logf("DEBUG: Test Disconnect: Attempting read on conn2 for status update")
	err = conn2.SetReadDeadline(time.Now().Add(2 * time.Second))
	require.NoError(t, err)
	statusMsg := readJSONMessage(t, conn2)
	t.Logf("DEBUG: Test Disconnect: Received message on conn2 after conn1 close")

	assert.Equal(t, "status_update", statusMsg["type"])
	assert.Equal(t, userID, statusMsg["user_id"])
	assert.Equal(t, float64(1), statusMsg["connections"])
	platformsMap, ok := statusMsg["platforms"].(map[string]interface{})
	require.True(t, ok)
	_, exists1 := platformsMap[platformWeb]
	assert.False(t, exists1, "Web platform should be removed from status message")
	assert.Equal(t, float64(1), platformsMap[platformMobile])

	// Verify internal state
	time.Sleep(50 * time.Millisecond)
	t.Logf("DEBUG: Test Disconnect: Checking final state via ConnectionManager")
	finalCount := connManager.GetConnectionCount(userID)
	finalPlatforms := connManager.GetUserPlatforms(userID)
	t.Logf("DEBUG: Test Disconnect: Final state - Count: %d, Platforms: %v", finalCount, finalPlatforms)

	assert.Equal(t, 1, finalCount, "Should only have one connection left in manager")
	assert.Len(t, finalPlatforms, 1, "Should only have one platform left in manager")
	assert.Equal(t, 1, finalPlatforms[platformMobile], "Remaining platform should be mobile")
}

func TestWebSocketHandler_NoUserID(t *testing.T) {
	server, wsURL, _ := setupTestServer(t, "")
	defer server.Close()

	dialer := websocket.Dialer{}
	query := url.Values{}
	query.Set("platform", "web")

	t.Logf("DEBUG: Test NoUserID: Attempting to dial %s?%s", wsURL, query.Encode())
	conn, resp, err := dialer.Dial(wsURL+"?"+query.Encode(), nil)

	require.Error(t, err, "Expected an error during dial due to failed upgrade")
	assert.Contains(t, err.Error(), "bad handshake", "Error should indicate bad handshake")
	require.NotNil(t, resp, "Response should not be nil")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Expected Unauthorized status")
	t.Logf("DEBUG: Test NoUserID: Received expected status code %d", resp.StatusCode)

	assert.Nil(t, conn, "Connection should be nil on failed dial")
}
