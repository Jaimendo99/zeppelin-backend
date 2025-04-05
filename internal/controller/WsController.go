package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"zeppelin/internal/middleware"
	"zeppelin/internal/services"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type UserConn struct {
	Conn     *websocket.Conn
	Platform string
}

var userConnections = make(map[string][]UserConn)
var mutex sync.Mutex

// Envía la actualización de estado a todas las conexiones activas de un usuario.
func sendStatusUpdate(userID string) {
	mutex.Lock()
	defer mutex.Unlock()

	conns := userConnections[userID]
	platforms := map[string]int{}
	for _, uc := range conns {
		platforms[uc.Platform]++
	}

	status := map[string]interface{}{
		"type":        "status_update",
		"user_id":     userID,
		"connections": len(conns),
		"platforms":   platforms,
	}

	jsonData, err := json.Marshal(status)
	if err != nil {
		fmt.Printf("Error al convertir status a JSON: %v\n", err)
		return
	}

	// Enviar el status a cada conexión activa del usuario
	for _, uc := range conns {
		if err := uc.Conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
			fmt.Printf("Error al enviar status a %s: %v\n", userID, err)
		}
	}
}

func WebSocketHandler(authService *services.AuthService) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.QueryParam("token")
		platform := c.QueryParam("platform")

		if token == "" {
			return c.String(http.StatusUnauthorized, "Token requerido")
		}
		if platform == "" {
			platform = "unknown"
		}

		claims, err := middleware.ValidateTokenAndRole(token, authService, "org:student")
		if err != nil {
			fmt.Printf("⛔ Token inválido: %v\n", err)
			return c.String(http.StatusUnauthorized, "Token inválido")
		}

		userID := claims.Subject
		if userID == "" {
			return c.String(http.StatusUnauthorized, "ID de usuario inválido")
		}

		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			fmt.Printf("❌ Error al hacer upgrade: %v\n", err)
			return err
		}

		fmt.Printf("🔌 WebSocket conectado: %s desde %s\n", userID, platform)

		userConn := UserConn{Conn: conn, Platform: platform}

		mutex.Lock()
		userConnections[userID] = append(userConnections[userID], userConn)
		mutex.Unlock()

		// Envía actualización de estado a todas las conexiones de este usuario
		sendStatusUpdate(userID)

		go func() {
			defer func() {
				mutex.Lock()
				conns := userConnections[userID]
				for i, uc := range conns {
					if uc.Conn == conn {
						userConnections[userID] = append(conns[:i], conns[i+1:]...)
						break
					}
				}
				mutex.Unlock()
				// Envía actualización tras remover la conexión
				sendStatusUpdate(userID)
				conn.Close()
				fmt.Printf("🔌 Conexión cerrada: %s (%s)\n", userID, platform)
			}()

			for {
				msgType, msg, err := conn.ReadMessage()
				if err != nil {
					fmt.Printf("❌ Error al leer mensaje de %s: %v\n", userID, err)
					break
				}

				fmt.Printf("📩 Mensaje recibido de %s: %s\n", userID, string(msg))

				// Si se reciben mensajes de otro tipo, se pueden reenviar a las demás conexiones
				mutex.Lock()
				for _, uc := range userConnections[userID] {
					if uc.Conn != conn {
						uc.Conn.WriteMessage(msgType, msg)
					}
				}
				mutex.Unlock()
			}
		}()

		return nil
	}
}
