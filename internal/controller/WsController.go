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
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// UserConn representa una conexión WebSocket + plataforma (web o mobile)
type UserConn struct {
	Conn     *websocket.Conn
	Platform string
}

// ConnectionManager gestiona sockets por usuario:curso y repositorios
type ConnectionManager struct {
	userConnections     map[string][]UserConn // key = userID:courseId
	mutex               sync.Mutex
	SessionRepo         domain.SessionRepo
	ParentalConsentRepo domain.ParentalConsentRepo
}

// NewConnectionManager inyecta SessionRepo y ParentalConsentRepo
func NewConnectionManager(
	sessionRepo domain.SessionRepo,
	consentRepo domain.ParentalConsentRepo,
) *ConnectionManager {
	return &ConnectionManager{
		userConnections:     make(map[string][]UserConn),
		SessionRepo:         sessionRepo,
		ParentalConsentRepo: consentRepo,
	}
}

// GetConnectionCount devuelve el n.º de conexiones activas para userID:courseId
func (cm *ConnectionManager) GetConnectionCount(userID, courseId string) int {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	return len(cm.userConnections[userID+":"+courseId])
}

// GetUserPlatforms devuelve cuántas conexiones hay por plataforma
func (cm *ConnectionManager) GetUserPlatforms(userID, courseId string) map[string]int {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	platforms := map[string]int{"web": 0, "mobile": 0}
	for _, uc := range cm.userConnections[userID+":"+courseId] {
		platforms[uc.Platform]++
	}
	return platforms
}

// sendStatusUpdate envía a todas las conexiones un JSON con el estado actual,
// incluyendo el parental_consent_status.
func (cm *ConnectionManager) sendStatusUpdate(logger echo.Logger, userID, courseId string) {
	key := userID + ":" + courseId

	cm.mutex.Lock()
	conns := append([]UserConn(nil), cm.userConnections[key]...)
	cm.mutex.Unlock()

	if len(conns) == 0 {
		return
	}

	// Obtener estado de consentimiento
	consent, err := cm.ParentalConsentRepo.GetConsentByUserID(userID)
	consentStatus := "UNKNOWN"
	if err == nil && consent != nil {
		consentStatus = consent.Status
	}

	status := map[string]any{
		"type":                    "status_update",
		"user_id":                 userID,
		"course_id":               courseId,
		"connections":             len(conns),
		"platforms":               cm.GetUserPlatforms(userID, courseId),
		"parental_consent_status": consentStatus,
	}

	jsonData, err := json.Marshal(status)
	if err != nil {
		logger.Errorf("User %s (Course %s): Failed to marshal status update: %v", userID, courseId, err)
		return
	}

	for _, uc := range conns {
		_ = uc.Conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err := uc.Conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
			logger.Warnf("User %s (Course %s, Platform %s): Failed to write status update: %v",
				userID, courseId, uc.Platform, err)
		}
		_ = uc.Conn.SetWriteDeadline(time.Time{})
	}
}

// removeConnection quita una conexión y devuelve si quedan otras
func (cm *ConnectionManager) removeConnection(logger echo.Logger, userID, courseId string, connToRemove *websocket.Conn) bool {
	key := userID + ":" + courseId

	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	conns := cm.userConnections[key]
	var updated []UserConn
	for _, uc := range conns {
		if uc.Conn != connToRemove {
			updated = append(updated, uc)
		}
	}

	if len(updated) == 0 {
		delete(cm.userConnections, key)
		logger.Infof("User %s (Course %s): Removed last connection.", userID, courseId)
		return false
	}

	cm.userConnections[key] = updated
	logger.Infof("User %s (Course %s): Removed connection. Remaining: %d.", userID, courseId, len(updated))
	return true
}

// WebSocketHandler gestiona nuevas conexiones y valida plataforma + consentimiento
func (cm *ConnectionManager) WebSocketHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		logger := c.Logger()

		platform := c.QueryParam("platform")
		if platform != "web" && platform != "mobile" {
			return c.String(http.StatusBadRequest, "Invalid platform")
		}
		courseId := c.QueryParam("courseId")
		if courseId == "" {
			return c.String(http.StatusBadRequest, "Invalid course ID")
		}
		userRaw := c.Get("user_id")
		userID, ok := userRaw.(string)
		if !ok || userID == "" {
			return c.String(http.StatusUnauthorized, "Invalid user ID")
		}

		// ————— Validación de consentimiento para mobile —————
		if platform == "mobile" {
			consent, err := cm.ParentalConsentRepo.GetConsentByUserID(userID)
			if err != nil || consent == nil || consent.Status != "ACCEPTED" {
				logger.Warnf("User %s: mobile connection rejected, consent=%v, err=%v", userID, consent, err)
				return c.String(http.StatusForbidden, "Se requiere consentimiento parental ACCEPTED para móvil")
			}
		}

		key := userID + ":" + courseId

		// ————— Verificar límite de conexiones por plataforma —————
		cm.mutex.Lock()
		conns := cm.userConnections[key]
		webCount, mobileCount := 0, 0
		for _, uc := range conns {
			if uc.Platform == "web" {
				webCount++
			} else {
				mobileCount++
			}
		}
		if platform == "web" && webCount > 0 {
			cm.mutex.Unlock()
			return c.String(http.StatusForbidden, "Only one web connection allowed per course")
		}
		if platform == "mobile" && mobileCount > 0 {
			cm.mutex.Unlock()
			return c.String(http.StatusForbidden, "Only one mobile connection allowed per course")
		}
		if platform == "mobile" && webCount == 0 {
			cm.mutex.Unlock()
			return c.String(http.StatusForbidden, "Web connection required for this course first")
		}
		cm.mutex.Unlock()

		// ————— Upgrade a WebSocket y registro de la conexión —————
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			logger.Errorf("User %s (Course %s): upgrade failed: %v", userID, courseId, err)
			return nil
		}
		logger.Infof("User %s (Course %s): Connected (%s)", userID, courseId, platform)

		cm.mutex.Lock()
		cm.userConnections[key] = append(cm.userConnections[key], UserConn{Conn: conn, Platform: platform})
		cm.mutex.Unlock()

		// Enviar primer status_update (incluye parental_consent_status)
		cm.sendStatusUpdate(logger, userID, courseId)

		// Leer/broadcast en background
		go cm.readPump(logger, userID, courseId, platform, conn)
		return nil
	}
}

// readPump lee mensajes del socket, maneja pomodoro y broadcast
func (cm *ConnectionManager) readPump(logger echo.Logger, userID, courseId, platform string, conn *websocket.Conn) {
	key := userID + ":" + courseId
	defer func() {
		logger.Infof("User %s (Course %s, %s): Cleaning up", userID, courseId, platform)
		if cm.removeConnection(logger, userID, courseId, conn) {
			cm.sendStatusUpdate(logger, userID, courseId)
		}
		_ = conn.Close()
	}()

	for {
		msgType, message, err := conn.ReadMessage()
		if err != nil {
			var e *websocket.CloseError
			if errors.As(err, &e) {
				logger.Infof("User %s (Course %s, %s): Closed (code %d)", userID, courseId, platform, e.Code)
			} else {
				logger.Warnf("User %s (Course %s, %s): Read error: %v", userID, courseId, platform, err)
			}
			break
		}

		if msgType != websocket.TextMessage {
			continue
		}

		// Estructura básica del mensaje entrante
		var msg struct {
			Type      string                 `json:"type"`
			Config    map[string]interface{} `json:"config"`
			SenderId  string                 `json:"senderId"`
			StartedAt int64                  `json:"startedAt"`
			SessionId int                    `json:"session_id"`
		}
		if err := json.Unmarshal(message, &msg); err != nil {
			logger.Warnf("User %s (Course %s, %s): JSON parse error: %v", userID, courseId, platform, err)
			continue
		}

		broadcast := message

		switch msg.Type {
		case "pomodoro_start":
			// Crear sesión en la DB
			sessionID, err := cm.SessionRepo.StartSession(userID)
			if err != nil {
				errMsg := map[string]string{"type": "error", "error": "Failed to start session"}
				data, _ := json.Marshal(errMsg)
				_ = conn.WriteMessage(websocket.TextMessage, data)
			} else {
				// Responder ACK al emisor
				ack := map[string]any{"type": "session_started", "session_id": sessionID}
				ackData, _ := json.Marshal(ack)
				_ = conn.WriteMessage(websocket.TextMessage, ackData)
				// Preparar broadcast con session_id
				bMsg := map[string]any{
					"type":       msg.Type,
					"config":     msg.Config,
					"senderId":   msg.SenderId,
					"startedAt":  msg.StartedAt,
					"session_id": sessionID,
				}
				broadcast, _ = json.Marshal(bMsg)
			}

		case "pomodoro_session_end":
			if msg.SessionId == 0 {
				errMsg := map[string]string{"type": "error", "error": "No session_id provided"}
				d, _ := json.Marshal(errMsg)
				_ = conn.WriteMessage(websocket.TextMessage, d)
			} else if err := cm.SessionRepo.EndSession(msg.SessionId); err != nil {
				errMsg := map[string]string{"type": "error", "error": "Failed to end session"}
				d, _ := json.Marshal(errMsg)
				_ = conn.WriteMessage(websocket.TextMessage, d)
			} else {
				ack := map[string]any{"type": "session_ended", "session_id": msg.SessionId}
				ackData, _ := json.Marshal(ack)
				_ = conn.WriteMessage(websocket.TextMessage, ackData)
			}
			// En cualquier caso, usamos el mensaje original o el modificado para broadcast

		default:
			// Otros tipos: no modificamos broadcast
		}

		// Broadcast a las demás conexiones (o todas si pomodoro_start)
		cm.mutex.Lock()
		conns := append([]UserConn(nil), cm.userConnections[key]...)
		cm.mutex.Unlock()

		for _, uc := range conns {
			// Excluir al emisor salvo en pomodoro_start
			if msg.Type != "pomodoro_start" && uc.Conn == conn {
				continue
			}
			_ = uc.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			_ = uc.Conn.WriteMessage(websocket.TextMessage, broadcast)
			_ = uc.Conn.SetWriteDeadline(time.Time{})
		}
	}
}
