package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

const (
	writeWait = 10 * time.Second
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type UserConn struct {
	Conn     *websocket.Conn
	Platform string
}

type ConnectionManager struct {
	// Key is userID:courseId
	userConnections map[string][]UserConn
	mutex           sync.Mutex
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		userConnections: make(map[string][]UserConn),
	}
}

// GetConnectionCount returns the number of connections for a specific user and course.
func (cm *ConnectionManager) GetConnectionCount(userID, courseId string) int {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	connKey := userID + ":" + courseId
	return len(cm.userConnections[connKey])
}

// GetUserPlatforms returns the platform distribution for connections for a specific user and course.
func (cm *ConnectionManager) GetUserPlatforms(userID, courseId string) map[string]int {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	connKey := userID + ":" + courseId
	platforms := make(map[string]int)
	if conns, exists := cm.userConnections[connKey]; exists {
		for _, uc := range conns {
			platforms[uc.Platform]++
		}
	}
	return platforms
}

// sendStatusUpdate sends a status update message to all connections for a specific user and course.
func (cm *ConnectionManager) sendStatusUpdate(logger echo.Logger, userID, courseId string) {
	connKey := userID + ":" + courseId

	cm.mutex.Lock()
	conns := cm.userConnections[connKey]
	cm.mutex.Unlock()

	if len(conns) == 0 {
		return
	}

	platforms := make(map[string]int)
	for _, uc := range conns {
		platforms[uc.Platform]++
	}

	status := map[string]any{
		"type":        "status_update",
		"user_id":     userID,
		"connections": len(conns),
		"platforms":   platforms,
		"course_id":   courseId,
	}

	jsonData, err := json.Marshal(status)
	if err != nil {
		logger.Errorf("User %s (Course %s): Failed to marshal status update: %v", userID, courseId, err)
		return
	}

	cm.mutex.Lock()
	connsToSend := cm.userConnections[connKey]
	cm.mutex.Unlock()

	for _, uc := range connsToSend {
		if err := uc.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
			logger.Warnf("User %s (Course %s, Platform %s): Failed to set write deadline: %v", userID, courseId, uc.Platform, err)
		}

		if err := uc.Conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
			logger.Warnf("User %s (Course %s, Platform %s): Failed to write status update: %v", userID, courseId, uc.Platform, err)
		}

		_ = uc.Conn.SetWriteDeadline(time.Time{})
	}
}

// removeConnection removes a specific connection for a given user and course.
// Returns true if other connections remain for this user/course group, false otherwise.
func (cm *ConnectionManager) removeConnection(logger echo.Logger, userID, courseId string, connToRemove *websocket.Conn) bool {
	connKey := userID + ":" + courseId // Added courseId and created key

	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	conns, exists := cm.userConnections[connKey] // Used connKey
	if !exists {
		logger.Warnf("User %s (Course %s): Attempted to remove connection, but user/course entry already gone.", userID, courseId)
		return false
	}

	found := false
	var updatedConns []UserConn
	for _, uc := range conns {
		if uc.Conn != connToRemove {
			updatedConns = append(updatedConns, uc)
		} else {
			found = true
		}
	}

	if !found {
		logger.Warnf("User %s (Course %s): Attempted to remove connection, but specific connection not found in list.", userID, courseId)
		return len(conns) > 0 // Still return if the original list had connections
	}

	if len(updatedConns) == 0 {
		delete(cm.userConnections, connKey) // Used connKey
		logger.Infof("User %s (Course %s): Removed last connection.", userID, courseId)
		return false
	}

	cm.userConnections[connKey] = updatedConns // Used connKey
	logger.Infof("User %s (Course %s): Removed connection. Remaining: %d.", userID, courseId, len(updatedConns))
	return true
}

func (cm *ConnectionManager) WebSocketHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		logger := c.Logger()

		platform := c.QueryParam("platform")
		if platform == "" {
			platform = "unknown" // Keep "unknown" as default if empty, although validation below overrides
		}
		if platform != "web" && platform != "mobile" {
			logger.Errorf("User: Invalid platform: %s", platform)
			return c.String(http.StatusBadRequest, "Invalid platform")
		}

		courseId := c.QueryParam("courseId")
		if courseId == "" {
			logger.Errorf("User: Invalid course ID")
			return c.String(http.StatusBadRequest, "Invalid course ID")
		}

		userIDRaw := c.Get("user_id")
		userID, ok := userIDRaw.(string)
		if !ok || userID == "" {
			logger.Errorf("Invalid user ID type or empty. Value: %v", userIDRaw)
			return c.String(http.StatusUnauthorized, "Invalid user ID")
		}

		connKey := userID + ":" + courseId // Created connKey

		cm.mutex.Lock()
		conns := cm.userConnections[connKey] // Used connKey
		webConnected := false
		webCount := 0
		mobileCount := 0
		for _, uc := range conns {
			if uc.Platform == "web" {
				webConnected = true
				webCount++
			} else if uc.Platform == "mobile" {
				mobileCount++
			}
		}

		// Check connection limits per user/course group
		if platform == "web" && webCount > 0 {
			cm.mutex.Unlock()
			logger.Errorf("User %s (Course %s): Web connection already exists", userID, courseId)
			return c.String(http.StatusForbidden, "Only one web connection allowed per course")
		}
		if platform == "mobile" && mobileCount > 0 {
			cm.mutex.Unlock()
			logger.Errorf("User %s (Course %s): Mobile connection already exists", userID, courseId)
			return c.String(http.StatusForbidden, "Only one mobile connection allowed per course")
		}

		if platform == "mobile" && !webConnected {
			cm.mutex.Unlock()
			logger.Errorf("User %s (Course %s): Mobile connection rejected, no web connection active for this course", userID, courseId) // Adjusted log message
			return c.String(http.StatusForbidden, "Web connection required for this course first")                                       // Adjusted error message
		}
		cm.mutex.Unlock()

		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			logger.Errorf("User %s (Course %s): Failed WebSocket upgrade: %v", userID, courseId, err) // Updated log
			return nil
		}
		logger.Infof("User %s (Course %s): WebSocket connection established from %s (Platform: %s)", userID, courseId, conn.RemoteAddr(), platform) // Updated log

		userConn := UserConn{Conn: conn, Platform: platform}

		cm.mutex.Lock()
		cm.userConnections[connKey] = append(cm.userConnections[connKey], userConn) // Used connKey
		cm.mutex.Unlock()

		cm.sendStatusUpdate(logger, userID, courseId) // Passed courseId

		go cm.readPump(logger, userID, courseId, platform, conn) // Passed courseId

		return nil
	}
}

func (cm *ConnectionManager) readPump(logger echo.Logger, userID, courseId, platform string, conn *websocket.Conn) { // Added courseId
	connKey := userID + ":" + courseId // Created connKey

	defer func() {
		logger.Infof("User %s (Course %s, Platform %s): Cleaning up connection.", userID, courseId, platform) // Updated log
		othersRemain := cm.removeConnection(logger, userID, courseId, conn)                                   // Passed courseId
		if othersRemain {
			cm.sendStatusUpdate(logger, userID, courseId) // Passed courseId
		}
		_ = conn.Close()
		logger.Infof("User %s (Course %s, Platform %s): Connection cleanup finished.", userID, courseId, platform) // Updated log
	}()

	for {
		msgType, message, err := conn.ReadMessage()
		if err != nil {
			var e *websocket.CloseError
			if errors.As(err, &e) && (e.Code == websocket.CloseNormalClosure || e.Code == websocket.CloseGoingAway || e.Code == websocket.CloseAbnormalClosure) {
				logger.Infof("User %s (Course %s, Platform %s): WebSocket closed normally (code %d).", userID, courseId, platform, e.Code) // Updated log
			} else {
				logger.Warnf("User %s (Course %s, Platform %s): Error reading message: %v", userID, courseId, platform, err) // Updated log
			}
			break
		}

		logger.Infof("User %s (Course %s, Platform %s): Received message (type %d): %s", userID, courseId, platform, msgType, string(message)) // Updated log

		cm.mutex.Lock()
		conns, exists := cm.userConnections[connKey] // Used connKey
		cm.mutex.Unlock()

		if !exists {
			continue
		}

		for _, uc := range conns {
			if uc.Conn != conn {
				if err := uc.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
					logger.Warnf("User %s (Course %s, Platform %s): Broadcast - Failed to set write deadline for target %s: %v", userID, courseId, platform, uc.Platform, err) // Updated log
				}
				if err := uc.Conn.WriteMessage(msgType, message); err != nil {
					logger.Warnf("User %s (Course %s, Platform %s): Failed to broadcast message to target %s: %v", userID, courseId, platform, uc.Platform, err) // Updated log
				}
				_ = uc.Conn.SetWriteDeadline(time.Time{})
			}
		}
	}
}
