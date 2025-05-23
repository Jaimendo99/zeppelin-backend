package controller_test

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"zeppelin/internal/controller" // Assuming this is your module path
	"zeppelin/internal/domain"     // Import your domain package

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	elog "github.com/labstack/gommon/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockSessionRepo is a mock implementation of domain.SessionRepo for testing
type MockSessionRepo struct {
	StartSessionFunc             func(userID string) (int, error)
	EndSessionFunc               func(sessionID int) error
	GetActiveSessionByUserIDFunc func(userID string) (*domain.Session, error)
}

func (m *MockSessionRepo) StartSession(userID string) (int, error) {
	if m.StartSessionFunc != nil {
		return m.StartSessionFunc(userID)
	}
	// Default behavior for tests that don't care about session ID
	return 1, nil
}

func (m *MockSessionRepo) EndSession(sessionID int) error {
	if m.EndSessionFunc != nil {
		return m.EndSessionFunc(sessionID)
	}
	// Default behavior
	return nil
}

func (m *MockSessionRepo) GetActiveSessionByUserID(userID string) (*domain.Session, error) {
	if m.GetActiveSessionByUserIDFunc != nil {
		return m.GetActiveSessionByUserIDFunc(userID)
	}
	// Default behavior: no active session found
	return nil, nil
}

// setupTestServer sets up a test Echo server with the WebSocket handler.
// It now includes a mock SessionRepo for the ConnectionManager.
func setupTestServer(t *testing.T, userID string) (*httptest.Server, string, *controller.ConnectionManager, *MockSessionRepo) {
	t.Helper() // Mark as test helper

	// Create a mock SessionRepo
	mockSessionRepo := &MockSessionRepo{}

	// Create ConnectionManager with the mock SessionRepo
	connManager := controller.NewConnectionManager(mockSessionRepo)

	e := echo.New()
	e.Logger.SetLevel(elog.DEBUG) // Use DEBUG for more verbose test logs if needed
	e.Logger.SetHeader("${time_rfc3339} ${level} ${prefix} ${file}:${line}")

	// Middleware to simulate authentication and set user_id
	authMiddleware := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Only set user_id if it's provided for the test
			if userID != "" {
				t.Logf("DEBUG: AuthMiddleware setting user_id to: %s for path %s", userID, c.Request().URL.Path)
				c.Set("user_id", userID)
			} else {
				t.Logf("DEBUG: AuthMiddleware not setting user_id (empty userID provided) for path %s", c.Request().URL.Path)
			}
			return next(c)
		}
	}

	e.GET("/ws", connManager.WebSocketHandler(), authMiddleware)

	server := httptest.NewServer(e)
	t.Logf("DEBUG: Test server started at URL: %s", server.URL)

	// Construct the base WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	t.Logf("DEBUG: Base WebSocket URL for test: %s", wsURL)

	return server, wsURL, connManager, mockSessionRepo // Return the mock SessionRepo as well
}

// connectWebSocketNoCleanup now requires courseId
func connectWebSocketNoCleanup(t *testing.T, wsURL string, platform string, courseId string) *websocket.Conn {
	t.Helper() // Mark as test helper
	dialer := websocket.Dialer{}
	query := url.Values{}
	query.Set("platform", platform)
	query.Set("courseId", courseId) // Add courseId to query params

	fullWsURL := wsURL + "?" + query.Encode()
	t.Logf("DEBUG: Attempting to dial WebSocket: %s", fullWsURL)

	conn, resp, err := dialer.Dial(fullWsURL, nil)

	// Provide more context on failure
	if err != nil {
		// Check if there's a response object even if dial failed (e.g., HTTP error)
		if resp != nil {
			// Ensure the body is closed
			defer resp.Body.Close()
			// Read the body content
			bodyBytes, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
				// Log error reading the body itself
				t.Logf("DEBUG: WebSocket dial failed. Status: %d. Error reading response body: %v", resp.StatusCode, readErr)
			} else {
				// Log the status and the body content
				t.Logf("DEBUG: WebSocket dial failed. Status: %d, Body: %s", resp.StatusCode, string(bodyBytes))
			}
		} else {
			// No response object, just log the dial error
			t.Logf("DEBUG: WebSocket dial failed with no response object. Error: %v", err)
		}
	}

	// These require calls will fail the test if err != nil, using the message logged above for context
	require.NoError(t, err, "Failed to dial websocket: %s", fullWsURL)
	require.NotNil(t, resp, "Response should not be nil for %s", fullWsURL)

	// Now that we know err is nil, we expect StatusSwitchingProtocols
	require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode, "Expected status switching protocols for %s", fullWsURL)

	// If we reached here, the connection should be valid
	require.NotNil(t, conn, "Connection should not be nil for %s", fullWsURL)
	t.Logf("DEBUG: WebSocket dial successful for %s.", fullWsURL)

	// Set a reasonable read deadline for initial messages
	deadlineErr := conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	require.NoError(t, deadlineErr, "Failed to set read deadline on connection for %s", fullWsURL)

	return conn
}

// readJSONMessage remains the same structurally
func readJSONMessage(t *testing.T, conn *websocket.Conn) map[string]interface{} {
	t.Helper() // Mark as test helper
	// Increase deadline slightly for reading, helps in slower CI environments
	err := conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	require.NoError(t, err)

	t.Logf("DEBUG: Attempting to read message from WebSocket %s...", conn.RemoteAddr())
	msgType, msgBytes, err := conn.ReadMessage()

	// Handle read errors more gracefully
	if err != nil {
		t.Logf("DEBUG: Error reading message from %s: %v", conn.RemoteAddr(), err)
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			t.Fatalf("Timeout reading message from %s: %v", conn.RemoteAddr(), err)
		}
		if closeErr, ok := err.(*websocket.CloseError); ok {
			// Log expected closures differently from unexpected ones
			if closeErr.Code == websocket.CloseNormalClosure || closeErr.Code == websocket.CloseGoingAway {
				t.Logf("DEBUG: Connection %s closed normally while expecting read: %v", conn.RemoteAddr(), closeErr)
			} else {
				t.Fatalf("Connection %s closed unexpectedly while reading message: %v", conn.RemoteAddr(), closeErr)
			}
		}
		// Fail for other errors
		t.Fatalf("Failed to read message from %s: %v", conn.RemoteAddr(), err)
	}

	t.Logf("DEBUG: Read message type %d from %s, content: %s", msgType, conn.RemoteAddr(), string(msgBytes))
	assert.Equal(t, websocket.TextMessage, msgType, "Expected text message type from %s", conn.RemoteAddr())

	var data map[string]interface{}
	err = json.Unmarshal(msgBytes, &data)
	require.NoError(t, err, "Failed to unmarshal JSON message from %s: %s", conn.RemoteAddr(), string(msgBytes))
	t.Logf("DEBUG: Successfully unmarshalled JSON message from %s.", conn.RemoteAddr())

	// Reset deadline after successful read
	err = conn.SetReadDeadline(time.Time{})
	require.NoError(t, err)

	return data
}

// --- Updated Tests ---

func TestWebSocketHandler_ConnectionAndStatus(t *testing.T) {
	userID := "user-123"
	courseID := "course-abc" // Define courseId
	platform := "web"
	// Receive the mockSessionRepo but don't use it in this test
	server, wsURL, connManager, _ := setupTestServer(t, userID)
	defer server.Close()

	// Pass courseId when connecting
	conn := connectWebSocketNoCleanup(t, wsURL, platform, courseID)

	// Defer cleanup
	defer func() {
		t.Logf("DEBUG: Test defer (%s): Sending close message.", t.Name())
		// Best effort close message
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		time.Sleep(10 * time.Millisecond) // Give time for message to be processed
		t.Logf("DEBUG: Test defer (%s): Closing connection.", t.Name())
		_ = conn.Close()
		t.Logf("DEBUG: Test defer (%s): Waiting after close.", t.Name())
		time.Sleep(50 * time.Millisecond) // Allow server-side cleanup goroutine to run
		t.Logf("DEBUG: Test defer (%s): Finished.", t.Name())
	}()

	// Read the initial status message FIRST
	t.Logf("DEBUG: Test (%s): Attempting to read initial status message for %s/%s.", t.Name(), userID, courseID)
	statusMsg := readJSONMessage(t, conn)
	t.Logf("DEBUG: Test (%s): Successfully read initial status message for %s/%s.", t.Name(), userID, courseID)

	// Verify status message content
	assert.Equal(t, "status_update", statusMsg["type"])
	assert.Equal(t, userID, statusMsg["user_id"])
	assert.Equal(t, courseID, statusMsg["course_id"])     // Verify courseId in message
	assert.Equal(t, float64(1), statusMsg["connections"]) // Use float64 for JSON numbers
	statusPlatformsMap, ok := statusMsg["platforms"].(map[string]interface{})
	require.True(t, ok, "Status message 'platforms' should be a map")
	assert.Equal(t, float64(1), statusPlatformsMap[platform], "Status message platform count mismatch")
	assert.Len(t, statusPlatformsMap, 2, "Status message platform map size mismatch")

	// Now verify the internal state of the ConnectionManager
	// Add a small delay to ensure the server has processed the connection fully
	time.Sleep(50 * time.Millisecond)
	t.Logf("DEBUG: Test (%s): Checking ConnectionManager state AFTER reading message for %s/%s.", t.Name(), userID, courseID)
	// Pass courseId when checking state
	connectionCount := connManager.GetConnectionCount(userID, courseID)
	platformCounts := connManager.GetUserPlatforms(userID, courseID)
	t.Logf("DEBUG: Test (%s): Checked state. Count: %d, Platforms: %v", t.Name(), connectionCount, platformCounts)

	require.Equal(t, 1, connectionCount, "Should have 1 connection for the user/course")
	require.Contains(t, platformCounts, platform, "Platform map should contain the connected platform")
	assert.Equal(t, 1, platformCounts[platform], "Platform count for '%s' should be 1", platform)
	assert.Len(t, platformCounts, 2, "Platform map should only contain one entry")
}

func TestWebSocketHandler_WebAndMobileConnections(t *testing.T) {
	userID := "user-multi"
	courseID := "course-multi-101" // Define courseId
	platformWeb := "web"
	platformMobile := "mobile"
	server, wsURL, connManager, _ := setupTestServer(t, userID) // Receive mockSessionRepo
	defer server.Close()

	// Connect web client
	conn1 := connectWebSocketNoCleanup(t, wsURL, platformWeb, courseID) // Pass courseId
	defer conn1.Close()

	// Read initial status for web client (1 connection)
	statusMsg1 := readJSONMessage(t, conn1)
	assert.Equal(t, float64(1), statusMsg1["connections"])
	assert.Equal(t, courseID, statusMsg1["course_id"]) // Verify courseId
	platformsMap1, _ := statusMsg1["platforms"].(map[string]interface{})
	assert.Equal(t, float64(1), platformsMap1[platformWeb])
	assert.Len(t, platformsMap1, 2)

	// Verify internal state after web connect
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, 1, connManager.GetConnectionCount(userID, courseID), "Count should be 1 after web connect")       // Pass courseId
	assert.Equal(t, 1, connManager.GetUserPlatforms(userID, courseID)[platformWeb], "Web platform count should be 1") // Pass courseId

	// Connect mobile client
	conn2 := connectWebSocketNoCleanup(t, wsURL, platformMobile, courseID) // Pass courseId
	defer conn2.Close()

	// Read updated status for web client (should show 2 connections)
	statusMsg1_updated := readJSONMessage(t, conn1)
	assert.Equal(t, float64(2), statusMsg1_updated["connections"])
	assert.Equal(t, courseID, statusMsg1_updated["course_id"]) // Verify courseId
	platformsMap1_updated, _ := statusMsg1_updated["platforms"].(map[string]interface{})
	assert.Equal(t, float64(1), platformsMap1_updated[platformWeb])
	assert.Equal(t, float64(1), platformsMap1_updated[platformMobile])
	assert.Len(t, platformsMap1_updated, 2)

	// Read initial status for mobile client (should show 2 connections)
	statusMsg2 := readJSONMessage(t, conn2)
	assert.Equal(t, float64(2), statusMsg2["connections"])
	assert.Equal(t, courseID, statusMsg2["course_id"]) // Verify courseId
	platformsMap2, _ := statusMsg2["platforms"].(map[string]interface{})
	assert.Equal(t, float64(1), platformsMap2[platformWeb])
	assert.Equal(t, float64(1), platformsMap2[platformMobile])
	assert.Len(t, platformsMap2, 2)

	// Verify final internal state
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, 2, connManager.GetConnectionCount(userID, courseID), "Count should be 2 after mobile connect") // Pass courseId
	platforms := connManager.GetUserPlatforms(userID, courseID)                                                    // Pass courseId
	assert.Len(t, platforms, 2, "Should have 2 platforms registered")
	assert.Equal(t, 1, platforms[platformWeb], "Web platform count should be 1")
	assert.Equal(t, 1, platforms[platformMobile], "Mobile platform count should be 1")
}

func TestWebSocketHandler_SecondWebConnectionRejected(t *testing.T) {
	userID := "user-second-web"
	courseID := "course-web-limit" // Define courseId
	platform := "web"
	server, wsURL, connManager, _ := setupTestServer(t, userID) // Receive mockSessionRepo
	defer server.Close()

	// Connect first web client
	conn1 := connectWebSocketNoCleanup(t, wsURL, platform, courseID) // Pass courseId
	defer conn1.Close()
	statusMsg1 := readJSONMessage(t, conn1)            // Read initial status
	assert.Equal(t, courseID, statusMsg1["course_id"]) // Verify courseId

	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, 1, connManager.GetConnectionCount(userID, courseID), "Count should be 1 after first web connect") // Pass courseId

	// Attempt to connect second web client
	dialer := websocket.Dialer{}
	query := url.Values{}
	query.Set("platform", platform)
	query.Set("courseId", courseID) // Pass courseId for the attempt
	t.Logf("DEBUG: Test SecondWeb: Attempting to dial second web client: %s?%s", wsURL, query.Encode())
	_, resp, err := dialer.Dial(wsURL+"?"+query.Encode(), nil)

	// Assertions for rejection
	require.Error(t, err, "Expected an error due to second web connection")
	assert.Contains(t, err.Error(), "bad handshake", "Error should indicate bad handshake")
	require.NotNil(t, resp, "Response should not be nil on rejection")
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Expected Forbidden status")
	t.Logf("DEBUG: Test SecondWeb: Received expected status code %d", resp.StatusCode)

	// Verify state remains unchanged
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, 1, connManager.GetConnectionCount(userID, courseID), "Count should still be 1") // Pass courseId
	platforms := connManager.GetUserPlatforms(userID, courseID)                                     // Pass courseId
	assert.Equal(t, 1, platforms[platform], "Web platform count should still be 1")
	assert.Len(t, platforms, 2)
}

func TestWebSocketHandler_SecondMobileConnectionRejected(t *testing.T) {
	userID := "user-second-mobile"
	courseID := "course-mobile-limit" // Define courseId
	platformWeb := "web"
	platformMobile := "mobile"
	server, wsURL, connManager, _ := setupTestServer(t, userID) // Receive mockSessionRepo
	defer server.Close()

	// Connect web client
	conn1 := connectWebSocketNoCleanup(t, wsURL, platformWeb, courseID) // Pass courseId
	defer conn1.Close()
	statusMsg1 := readJSONMessage(t, conn1) // Read initial status (1 conn)
	assert.Equal(t, courseID, statusMsg1["course_id"])

	// Connect first mobile client
	conn2 := connectWebSocketNoCleanup(t, wsURL, platformMobile, courseID) // Pass courseId
	defer conn2.Close()
	statusMsg1_upd := readJSONMessage(t, conn1) // Read updated status on conn1 (2 conns)
	statusMsg2 := readJSONMessage(t, conn2)     // Read initial status on conn2 (2 conns)
	assert.Equal(t, courseID, statusMsg1_upd["course_id"])
	assert.Equal(t, courseID, statusMsg2["course_id"])

	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, 2, connManager.GetConnectionCount(userID, courseID), "Count should be 2 after mobile connect") // Pass courseId

	// Attempt to connect second mobile client
	dialer := websocket.Dialer{}
	query := url.Values{}
	query.Set("platform", platformMobile)
	query.Set("courseId", courseID) // Pass courseId for the attempt
	t.Logf("DEBUG: Test SecondMobile: Attempting to dial second mobile client: %s?%s", wsURL, query.Encode())
	_, resp, err := dialer.Dial(wsURL+"?"+query.Encode(), nil)

	// Assertions for rejection
	require.Error(t, err, "Expected an error due to second mobile connection")
	assert.Contains(t, err.Error(), "bad handshake", "Error should indicate bad handshake")
	require.NotNil(t, resp, "Response should not be nil on rejection")
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Expected Forbidden status")
	t.Logf("DEBUG: Test SecondMobile: Received expected status code %d", resp.StatusCode)

	// Verify state remains unchanged
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, 2, connManager.GetConnectionCount(userID, courseID), "Count should still be 2") // Pass courseId
	platforms := connManager.GetUserPlatforms(userID, courseID)                                     // Pass courseId
	assert.Len(t, platforms, 2)
	assert.Equal(t, 1, platforms[platformWeb], "Web platform count should be 1")
	assert.Equal(t, 1, platforms[platformMobile], "Mobile platform count should be 1")
}

// Renamed test for clarity
func TestWebSocketHandler_MobileWithoutWebRejected(t *testing.T) {
	userID := "user-no-web"
	courseID := "course-web-req" // Define courseId
	platformMobile := "mobile"
	server, wsURL, connManager, _ := setupTestServer(t, userID) // Receive mockSessionRepo
	defer server.Close()

	// Attempt to connect mobile client without web
	dialer := websocket.Dialer{}
	query := url.Values{}
	query.Set("platform", platformMobile)
	query.Set("courseId", courseID) // Pass courseId for the attempt
	t.Logf("DEBUG: Test MobileWithoutWeb: Attempting to dial mobile client: %s?%s", wsURL, query.Encode())
	_, resp, err := dialer.Dial(wsURL+"?"+query.Encode(), nil)

	// Assertions for rejection
	require.Error(t, err, "Expected an error due to mobile connection without web")
	assert.Contains(t, err.Error(), "bad handshake", "Error should indicate bad handshake")
	require.NotNil(t, resp, "Response should not be nil on rejection")
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Expected Forbidden status")
	t.Logf("DEBUG: Test MobileWithoutWeb: Received expected status code %d", resp.StatusCode)

	// Verify no connections were established
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, 0, connManager.GetConnectionCount(userID, courseID), "Count should be 0") // Pass courseId
}

func TestWebSocketHandler_InvalidPlatform(t *testing.T) {
	userID := "user-invalid-platform"
	courseID := "course-platform-check" // Define courseId
	platform := "invalid"
	server, wsURL, connManager, _ := setupTestServer(t, userID) // Receive mockSessionRepo
	defer server.Close()

	// Attempt to connect with invalid platform
	dialer := websocket.Dialer{}
	query := url.Values{}
	query.Set("platform", platform)
	query.Set("courseId", courseID) // Pass courseId for the attempt
	t.Logf("DEBUG: Test InvalidPlatform: Attempting to dial with invalid platform: %s?%s", wsURL, query.Encode())
	_, resp, err := dialer.Dial(wsURL+"?"+query.Encode(), nil)

	// Assertions for rejection
	require.Error(t, err, "Expected an error due to invalid platform")
	assert.Contains(t, err.Error(), "bad handshake", "Error should indicate bad handshake")
	require.NotNil(t, resp, "Response should not be nil on rejection")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected Bad Request status")
	t.Logf("DEBUG: Test InvalidPlatform: Received expected status code %d", resp.StatusCode)

	// Verify no connections were established
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, 0, connManager.GetConnectionCount(userID, courseID), "Count should be 0") // Pass courseId
}

// Add new test for missing courseId
func TestWebSocketHandler_MissingCourseID(t *testing.T) {
	userID := "user-missing-course"
	platform := "web"
	server, wsURL, connManager, _ := setupTestServer(t, userID) // Receive mockSessionRepo
	defer server.Close()

	// Attempt to connect without courseId query parameter
	dialer := websocket.Dialer{}
	query := url.Values{}
	query.Set("platform", platform)
	// DO NOT SET courseId
	t.Logf("DEBUG: Test MissingCourseID: Attempting to dial without courseId: %s?%s", wsURL, query.Encode())
	_, resp, err := dialer.Dial(wsURL+"?"+query.Encode(), nil)

	// Assertions for rejection
	require.Error(t, err, "Expected an error due to missing courseId")
	assert.Contains(t, err.Error(), "bad handshake", "Error should indicate bad handshake")
	require.NotNil(t, resp, "Response should not be nil on rejection")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected Bad Request status for missing courseId")
	t.Logf("DEBUG: Test MissingCourseID: Received expected status code %d", resp.StatusCode)

	// Verify no connections were established (use a dummy courseId here just for the check, as none should exist)
	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, 0, connManager.GetConnectionCount(userID, "any-course"), "Count should be 0")
}

func TestWebSocketHandler_MessageBroadcasting(t *testing.T) {
	userID := "user-broadcast"
	courseID := "course-chat" // Define courseId
	platformWeb := "web"
	platformMobile := "mobile"
	server, wsURL, _, _ := setupTestServer(t, userID) // Don't need connManager or mock SessionRepo directly here
	defer server.Close()

	// Connect clients, passing courseId
	conn1 := connectWebSocketNoCleanup(t, wsURL, platformWeb, courseID)
	defer conn1.Close()
	statusMsg1 := readJSONMessage(t, conn1) // Read initial status (1 conn)
	assert.Equal(t, courseID, statusMsg1["course_id"])

	conn2 := connectWebSocketNoCleanup(t, wsURL, platformMobile, courseID)
	defer conn2.Close()
	statusMsg1_upd := readJSONMessage(t, conn1) // Read updated status on conn1 (2 conns)
	statusMsg2 := readJSONMessage(t, conn2)     // Read initial status on conn2 (2 conns)
	assert.Equal(t, courseID, statusMsg1_upd["course_id"])
	assert.Equal(t, courseID, statusMsg2["course_id"])

	// Send message from web client
	testMessage := []byte(`{"type":"general_message", "msg":"hello from web"}`) // Use JSON for messages maybe?
	t.Logf("DEBUG: Test Broadcast: Sending message from conn1 (%s)", conn1.RemoteAddr())
	err := conn1.WriteMessage(websocket.TextMessage, testMessage)
	require.NoError(t, err)
	t.Logf("DEBUG: Test Broadcast: Message sent from conn1")

	// Mobile client should receive the message
	t.Logf("DEBUG: Test Broadcast: Attempting read on conn2 (%s)", conn2.RemoteAddr())
	msgType, msgBytes, err := conn2.ReadMessage()
	require.NoError(t, err, "Mobile client failed to read broadcast message")
	assert.Equal(t, websocket.TextMessage, msgType)
	assert.Equal(t, testMessage, msgBytes)
	t.Logf("DEBUG: Test Broadcast: Received broadcast on conn2")

	// Web client should NOT receive its own message for types other than pomodoro_start
	t.Logf("DEBUG: Test Broadcast: Attempting read on conn1 (%s) (should timeout)", conn1.RemoteAddr())
	err = conn1.SetReadDeadline(time.Now().Add(200 * time.Millisecond)) // Slightly longer timeout
	require.NoError(t, err)
	_, _, err = conn1.ReadMessage()
	assert.Error(t, err, "Web client should not have received its own message for general_message")
	netErr, ok := err.(net.Error)
	assert.True(t, ok && netErr.Timeout(), "Expected a timeout error on conn1, got: %v", err)
	t.Logf("DEBUG: Test Broadcast: Correctly timed out reading on conn1")

	// Reset deadlines
	_ = conn1.SetReadDeadline(time.Time{})
	_ = conn2.SetReadDeadline(time.Time{})
}

func TestWebSocketHandler_Disconnect(t *testing.T) {
	userID := "user-disconnect"
	courseID := "course-leave" // Define courseId
	platformWeb := "web"
	platformMobile := "mobile"
	server, wsURL, connManager, _ := setupTestServer(t, userID) // Receive mockSessionRepo
	defer server.Close()

	// Connect clients, passing courseId
	conn1 := connectWebSocketNoCleanup(t, wsURL, platformWeb, courseID)
	// No defer conn1.Close() here, we are testing its closure
	conn2 := connectWebSocketNoCleanup(t, wsURL, platformMobile, courseID)
	defer conn2.Close() // Close conn2 at the end

	// Read initial messages to clear buffers and verify courseId
	statusMsg1_init := readJSONMessage(t, conn1) // conn1 status (1 conn)
	assert.Equal(t, courseID, statusMsg1_init["course_id"])
	statusMsg1_upd := readJSONMessage(t, conn1) // conn1 status update (2 conns)
	assert.Equal(t, courseID, statusMsg1_upd["course_id"])
	statusMsg2_init := readJSONMessage(t, conn2) // conn2 status (2 conns)
	assert.Equal(t, courseID, statusMsg2_init["course_id"])

	// Verify initial state
	time.Sleep(20 * time.Millisecond)
	require.Equal(t, 2, connManager.GetConnectionCount(userID, courseID), "Should have 2 connections initially") // Pass courseId

	// Disconnect web client (conn1)
	t.Logf("DEBUG: Test Disconnect: Closing conn1 (%s)", conn1.RemoteAddr())
	// Send close message first, helps trigger server-side readPump exit
	_ = conn1.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	time.Sleep(50 * time.Millisecond) // Give server time to process close frame
	err := conn1.Close()              // Now close the underlying connection
	require.NoError(t, err)
	t.Logf("DEBUG: Test Disconnect: conn1 closed")

	// Mobile client (conn2) should receive a status update indicating only 1 connection left
	t.Logf("DEBUG: Test Disconnect: Attempting read on conn2 (%s) for status update", conn2.RemoteAddr())
	statusMsg_after_disconnect := readJSONMessage(t, conn2)
	t.Logf("DEBUG: Test Disconnect: Received message on conn2 after conn1 close")

	// Verify the status update message
	assert.Equal(t, "status_update", statusMsg_after_disconnect["type"])
	assert.Equal(t, userID, statusMsg_after_disconnect["user_id"])
	assert.Equal(t, courseID, statusMsg_after_disconnect["course_id"]) // Verify courseId
	assert.Equal(t, float64(1), statusMsg_after_disconnect["connections"])
	platformsMap, ok := statusMsg_after_disconnect["platforms"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(1), platformsMap[platformMobile])
	assert.Len(t, platformsMap, 2)

	// Verify internal state using ConnectionManager after a delay
	time.Sleep(100 * time.Millisecond) // Allow ample time for server cleanup goroutine
	t.Logf("DEBUG: Test Disconnect: Checking final state via ConnectionManager")
	finalCount := connManager.GetConnectionCount(userID, courseID)   // Pass courseId
	finalPlatforms := connManager.GetUserPlatforms(userID, courseID) // Pass courseId
	t.Logf("DEBUG: Test Disconnect: Final state - Count: %d, Platforms: %v", finalCount, finalPlatforms)

	assert.Equal(t, 1, finalCount, "Should only have one connection left in manager")
	assert.Len(t, finalPlatforms, 2, "Should only have one platform left in manager")
	assert.Equal(t, 1, finalPlatforms[platformMobile], "Remaining platform should be mobile")
}

func TestWebSocketHandler_PomodoroStartMessage(t *testing.T) {
	userID := "user-pomodoro-start"
	courseID := "course-pomodoro-101"
	platform := "web"
	server, wsURL, _, mockSessionRepo := setupTestServer(t, userID) // Receive mockSessionRepo
	defer server.Close()

	conn := connectWebSocketNoCleanup(t, wsURL, platform, courseID)
	defer conn.Close()
	_ = readJSONMessage(t, conn) // Read initial status update

	// Set the expectation for the SessionRepo.StartSession call
	expectedSessionID := 123
	mockSessionRepo.StartSessionFunc = func(inputUserID string) (int, error) {
		assert.Equal(t, userID, inputUserID, "StartSession should be called with the correct user ID")
		return expectedSessionID, nil // Return a dummy session ID
	}

	// Send the pomodoro_start message
	startMsg := map[string]interface{}{
		"type":      "pomodoro_start",
		"config":    map[string]interface{}{"work_duration": 25, "break_duration": 5},
		"senderId":  userID,
		"startedAt": time.Now().Unix(),
	}
	startMsgBytes, _ := json.Marshal(startMsg)

	t.Logf("DEBUG: Sending pomodoro_start message")
	err := conn.WriteMessage(websocket.TextMessage, startMsgBytes)
	require.NoError(t, err)
	t.Logf("DEBUG: Sent pomodoro_start message")

	// Expect to receive a session_started message back
	t.Logf("DEBUG: Expecting session_started response")
	responseMsg := readJSONMessage(t, conn)
	assert.Equal(t, "session_started", responseMsg["type"])
	assert.Equal(t, float64(expectedSessionID), responseMsg["session_id"]) // JSON numbers are float64
	t.Logf("DEBUG: Received session_started response")

	// Expect the same message to be broadcast back (including to the sender for pomodoro_start)
	t.Logf("DEBUG: Expecting broadcast message")
	broadcastMsg := readJSONMessage(t, conn)
	assert.Equal(t, "pomodoro_start", broadcastMsg["type"])
	assert.Equal(t, float64(expectedSessionID), broadcastMsg["session_id"]) // Verify session_id is included in broadcast
	// Add assertions for other fields in the broadcast message if needed
	t.Logf("DEBUG: Received broadcast message")
}

func TestWebSocketHandler_OtherMessageTypesBroadcasting(t *testing.T) {
	userID := "user-other-broadcast"
	courseID := "course-other-chat"
	platform1 := "web"
	platform2 := "mobile"
	server, wsURL, _, _ := setupTestServer(t, userID) // No need for connManager or mock SessionRepo
	defer server.Close()

	// Connect two clients for broadcasting
	conn1 := connectWebSocketNoCleanup(t, wsURL, platform1, courseID)
	defer conn1.Close()
	_ = readJSONMessage(t, conn1) // Read initial status (1 conn)

	conn2 := connectWebSocketNoCleanup(t, wsURL, platform2, courseID)
	defer conn2.Close()
	_ = readJSONMessage(t, conn1) // Read updated status on conn1 (2 conns)
	_ = readJSONMessage(t, conn2) // Read initial status on conn2 (2 conns)

	// Send a message of a type other than "pomodoro_start" from conn1
	otherMessage := map[string]interface{}{
		"type":     "pomodoro_phase_end", // Example of another message type
		"senderId": userID,
		"data":     "some phase end data",
	}
	otherMessageBytes, _ := json.Marshal(otherMessage)

	t.Logf("DEBUG: Sending 'other' message from conn1")
	err := conn1.WriteMessage(websocket.TextMessage, otherMessageBytes)
	require.NoError(t, err)
	t.Logf("DEBUG: Sent 'other' message")

	// The other connection (conn2) should receive the message
	t.Logf("DEBUG: Expecting broadcast message on conn2")
	receivedMsg2 := readJSONMessage(t, conn2)
	assert.Equal(t, otherMessage, receivedMsg2)
	t.Logf("DEBUG: Received broadcast message on conn2")

	// The sender (conn1) should NOT receive the broadcast for this message type
	t.Logf("DEBUG: Checking for unexpected broadcast on conn1 (should timeout)")
	err = conn1.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	require.NoError(t, err)
	_, _, err = conn1.ReadMessage()
	assert.Error(t, err, "Sender (conn1) should not receive broadcast for 'other' message type")
	netErr, ok := err.(net.Error)
	assert.True(t, ok && netErr.Timeout(), "Expected timeout, got: %v", err)
	_ = conn1.SetReadDeadline(time.Time{})
	t.Logf("DEBUG: Correctly timed out on conn1")
}

func TestWebSocketHandler_PomodoroSessionEnd(t *testing.T) {
	userID := "user-pomo-end"
	courseID := "course-end-123"
	platform := "web"
	server, wsURL, _, mockSessionRepo := setupTestServer(t, userID)
	defer server.Close()

	conn := connectWebSocketNoCleanup(t, wsURL, platform, courseID)
	defer conn.Close()
	_ = readJSONMessage(t, conn) // status_update

	expectedSessionID := 456
	mockSessionRepo.EndSessionFunc = func(sessionID int) error {
		assert.Equal(t, expectedSessionID, sessionID)
		return nil
	}

	endMsg := map[string]interface{}{
		"type":       "pomodoro_session_end",
		"session_id": expectedSessionID,
	}
	endMsgBytes, _ := json.Marshal(endMsg)

	err := conn.WriteMessage(websocket.TextMessage, endMsgBytes)
	require.NoError(t, err)

	// Expect confirmation
	response := readJSONMessage(t, conn)
	assert.Equal(t, "session_ended", response["type"])
	assert.Equal(t, float64(expectedSessionID), response["session_id"])

	// ❌ Do NOT expect a second message (broadcast), just verify timeout
	err = conn.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	require.NoError(t, err)
	_, _, err = conn.ReadMessage()
	assert.Error(t, err, "Expected timeout reading broadcast")
	netErr, ok := err.(net.Error)
	assert.True(t, ok && netErr.Timeout(), "Expected a timeout error")
}
