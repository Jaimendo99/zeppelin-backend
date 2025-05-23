package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"
	"zeppelin/internal/domain"

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
	SessionRepo     domain.SessionRepo
}

func NewConnectionManager(sessionRepo domain.SessionRepo) *ConnectionManager {
	return &ConnectionManager{
		userConnections: make(map[string][]UserConn),
		SessionRepo:     sessionRepo,
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
	platforms := map[string]int{
		"web":    0,
		"mobile": 0,
	}
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
	platforms := cm.GetUserPlatforms(userID, courseId)
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

func (cm *ConnectionManager) readPump(logger echo.Logger, userID, courseId, platform string, conn *websocket.Conn) {
	connKey := userID + ":" + courseId

	defer func() {
		logger.Infof("User %s (Course %s, Platform %s): Cleaning up connection.", userID, courseId, platform)
		othersRemain := cm.removeConnection(logger, userID, courseId, conn)
		if othersRemain {
			cm.sendStatusUpdate(logger, userID, courseId)
		}
		_ = conn.Close()
		logger.Infof("User %s (Course %s, Platform %s): Connection cleanup finished.", userID, courseId, platform)
	}()

	for {
		msgType, message, err := conn.ReadMessage()
		if err != nil {
			var e *websocket.CloseError
			if errors.As(err, &e) && (e.Code == websocket.CloseNormalClosure || e.Code == websocket.CloseGoingAway || e.Code == websocket.CloseAbnormalClosure) {
				logger.Infof("User %s (Course %s, Platform %s): WebSocket closed normally (code %d).", userID, courseId, platform, e.Code)
			} else {
				logger.Warnf("User %s (Course %s, Platform %s): Error reading message: %v", userID, courseId, platform, err)
			}
			break
		}

		logger.Infof("User %s (Course %s, Platform %s): Received message (type %d): %s", userID, courseId, platform, msgType, string(message))

		// Only process text messages
		if msgType != websocket.TextMessage {
			continue
		}

		// Parse the message to determine action
		var msg struct {
			Type      string                 `json:"type"`
			Config    map[string]interface{} `json:"config"`
			SenderId  string                 `json:"senderId"`
			StartedAt int64                  `json:"startedAt"`
			SessionId int                    `json:"session_id"`
		}
		if err := json.Unmarshal(message, &msg); err != nil {
			logger.Warnf("User %s (Course %s, Platform %s): Failed to parse message: %v", userID, courseId, platform, err)
			continue
		}

		// Message to broadcast (default to original message)
		broadcastMessage := message

		// Handle specific message types
		switch msg.Type {
		case "pomodoro_start":
			logger.Infof("User %s (Course %s, Platform %s): Starting Pomodoro session", userID, courseId, platform)
			// Start a new session
			sessionID, err := cm.SessionRepo.StartSession(userID)
			if err != nil {
				logger.Errorf("User %s (Course %s, Platform %s): Failed to start session: %v", userID, courseId, platform, err)
				// Send error back to client
				errorMsg := map[string]interface{}{
					"type":  "error",
					"error": "Failed to start session",
				}
				errorData, _ := json.Marshal(errorMsg)
				_ = conn.WriteMessage(websocket.TextMessage, errorData)
			} else {
				logger.Infof("User %s (Course %s, Platform %s): Started session with ID %d", userID, courseId, platform, sessionID)
				// Send session ID back to client
				response := map[string]interface{}{
					"type":       "session_started",
					"session_id": sessionID,
				}
				responseData, _ := json.Marshal(response)
				if err := conn.WriteMessage(websocket.TextMessage, responseData); err != nil {
					logger.Warnf("User %s (Course %s, Platform %s): Failed to send session start response: %v", userID, courseId, platform, err)
				}
				// Update broadcast message with session_id
				broadcastMsg := map[string]interface{}{
					"type":       msg.Type,
					"config":     msg.Config,
					"senderId":   msg.SenderId,
					"startedAt":  msg.StartedAt,
					"session_id": sessionID,
				}
				broadcastMessage, _ = json.Marshal(broadcastMsg)
			}

		case "pomodoro_session_end":
			// Use sessionId from the message
			if msg.SessionId == 0 {
				logger.Errorf("User %s (Course %s, Platform %s): No session_id provided in pomodoro_session_end message", userID, courseId, platform)
				errorMsg := map[string]interface{}{
					"type":  "error",
					"error": "No session_id provided",
				}
				errorData, _ := json.Marshal(errorMsg)
				_ = conn.WriteMessage(websocket.TextMessage, errorData)
			} else {
				// End the session
				if err := cm.SessionRepo.EndSession(msg.SessionId); err != nil {
					logger.Errorf("User %s (Course %s, Platform %s): Failed to end session %d: %v", userID, courseId, platform, msg.SessionId, err)
					errorMsg := map[string]interface{}{
						"type":  "error",
						"error": "Failed to end session",
					}
					errorData, _ := json.Marshal(errorMsg)
					_ = conn.WriteMessage(websocket.TextMessage, errorData)
				} else {
					logger.Infof("User %s (Course %s, Platform %s): Ended session %d", userID, courseId, platform, msg.SessionId)
					// Send confirmation back to client
					response := map[string]interface{}{
						"type":       "session_ended",
						"session_id": msg.SessionId,
					}
					responseData, _ := json.Marshal(response)
					if err := conn.WriteMessage(websocket.TextMessage, responseData); err != nil {
						logger.Warnf("User %s (Course %s, Platform %s): Failed to send session end response: %v", userID, courseId, platform, err)
					}
					// Use original message for broadcast (already includes session_id)
					broadcastMessage = message
				}
			}

		default:
			// No specific action for other message types (e.g., pomodoro_phase_end, pomodoro_extend); broadcast original message
		}

		// Broadcast the message to connections
		cm.mutex.Lock()
		conns, exists := cm.userConnections[connKey]
		cm.mutex.Unlock()

		if !exists {
			continue
		}

		for _, uc := range conns {
			// For pomodoro_start, broadcast to all connections (including sender)
			// For other messages, broadcast only to other connections (exclude sender)
			if msg.Type != "pomodoro_start" && uc.Conn == conn {
				logger.Infof("User %s (Course %s, Platform %s): Skipping broadcast to sender (%s) for message type %s", userID, courseId, platform, uc.Platform, msg.Type)
				continue
			}
			logger.Infof("User %s (Course %s, Platform %s): Broadcasting message (type %s) to platform %s", userID, courseId, platform, msg.Type, uc.Platform)
			if err := uc.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				logger.Warnf("User %s (Course %s, Platform %s): Broadcast - Failed to set write deadline for target %s: %v", userID, courseId, platform, uc.Platform, err)
			}
			if err := uc.Conn.WriteMessage(msgType, broadcastMessage); err != nil {
				logger.Warnf("User %s (Course %s, Platform %s): Failed to broadcast message to target %s: %v", userID, courseId, platform, uc.Platform, err)
			}
			_ = uc.Conn.SetWriteDeadline(time.Time{})
		}
	}
}
