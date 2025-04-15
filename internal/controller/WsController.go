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
	userConnections map[string][]UserConn
	mutex           sync.Mutex
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		userConnections: make(map[string][]UserConn),
	}
}

func (cm *ConnectionManager) GetConnectionCount(userID string) int {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	return len(cm.userConnections[userID])
}

func (cm *ConnectionManager) GetUserPlatforms(userID string) map[string]int {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	platforms := make(map[string]int)
	if conns, exists := cm.userConnections[userID]; exists {
		for _, uc := range conns {
			platforms[uc.Platform]++
		}
	}
	return platforms
}

func (cm *ConnectionManager) sendStatusUpdate(logger echo.Logger, userID string) {
	cm.mutex.Lock()
	conns := cm.userConnections[userID]
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
	}

	jsonData, err := json.Marshal(status)
	if err != nil {
		logger.Errorf("User %s: Failed to marshal status update: %v", userID, err)
		return
	}

	cm.mutex.Lock()
	connsToSend := cm.userConnections[userID]
	cm.mutex.Unlock()

	for _, uc := range connsToSend {
		if err := uc.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
			logger.Warnf("User %s (Platform %s): Failed to set write deadline: %v", userID, uc.Platform, err)
		}

		if err := uc.Conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
			logger.Warnf("User %s (Platform %s): Failed to write status update: %v", userID, uc.Platform, err)
		}

		_ = uc.Conn.SetWriteDeadline(time.Time{})
	}
}

func (cm *ConnectionManager) removeConnection(logger echo.Logger, userID string, connToRemove *websocket.Conn) bool {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	conns, exists := cm.userConnections[userID]
	if !exists {
		logger.Warnf("User %s: Attempted to remove connection, but user entry already gone.", userID)
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
		logger.Warnf("User %s: Attempted to remove connection, but specific connection not found in list.", userID)
		return len(conns) > 0
	}

	if len(updatedConns) == 0 {
		delete(cm.userConnections, userID)
		logger.Infof("User %s: Removed last connection.", userID)
		return false
	}

	cm.userConnections[userID] = updatedConns
	logger.Infof("User %s: Removed connection. Remaining: %d.", userID, len(updatedConns))
	return true
}

func (cm *ConnectionManager) WebSocketHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		logger := c.Logger()

		platform := c.QueryParam("platform")
		if platform == "" {
			platform = "unknown"
		}

		userIDRaw := c.Get("user_id")
		userID, ok := userIDRaw.(string)
		if !ok || userID == "" {
			logger.Errorf("Invalid user ID type or empty. Value: %v", userIDRaw)
			return c.String(http.StatusUnauthorized, "Invalid user ID")
		}

		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			logger.Errorf("User %s: Failed WebSocket upgrade: %v", userID, err)
			return nil
		}
		logger.Infof("User %s: WebSocket connection established from %s (Platform: %s)", userID, conn.RemoteAddr(), platform)

		userConn := UserConn{Conn: conn, Platform: platform}

		cm.mutex.Lock()
		cm.userConnections[userID] = append(cm.userConnections[userID], userConn)
		cm.mutex.Unlock()

		cm.sendStatusUpdate(logger, userID)

		go cm.readPump(logger, userID, platform, conn)

		return nil
	}
}

func (cm *ConnectionManager) readPump(logger echo.Logger, userID, platform string, conn *websocket.Conn) {
	defer func() {
		logger.Infof("User %s (Platform %s): Cleaning up connection.", userID, platform)
		othersRemain := cm.removeConnection(logger, userID, conn)
		if othersRemain {
			cm.sendStatusUpdate(logger, userID)
		}
		_ = conn.Close()
		logger.Infof("User %s (Platform %s): Connection cleanup finished.", userID, platform)
	}()

	for {
		msgType, message, err := conn.ReadMessage()
		if err != nil {
			var e *websocket.CloseError
			if errors.As(err, &e) && (e.Code == websocket.CloseNormalClosure || e.Code == websocket.CloseGoingAway || e.Code == websocket.CloseAbnormalClosure) {
				logger.Infof("User %s (Platform %s): WebSocket closed normally (code %d).", userID, platform, e.Code)
			} else {
				logger.Warnf("User %s (Platform %s): Error reading message: %v", userID, platform, err)
			}
			break
		}

		logger.Infof("User %s (Platform %s): Received message (type %d): %s", userID, platform, msgType, string(message))

		cm.mutex.Lock()
		conns, exists := cm.userConnections[userID]
		cm.mutex.Unlock()

		if !exists {
			continue
		}

		for _, uc := range conns {
			if uc.Conn != conn {
				if err := uc.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
					logger.Warnf("User %s (Platform %s): Broadcast - Failed to set write deadline for target %s: %v", userID, platform, uc.Platform, err)
				}
				if err := uc.Conn.WriteMessage(msgType, message); err != nil {
					logger.Warnf("User %s (Platform %s): Failed to broadcast message to target %s: %v", userID, platform, uc.Platform, err)
				}
				_ = uc.Conn.SetWriteDeadline(time.Time{})
			}
		}
	}
}
