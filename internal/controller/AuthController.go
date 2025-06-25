package controller

import (
	"errors"
	"fmt"
	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/labstack/echo/v4"
	"net/http"
	"zeppelin/internal/domain"
)

type AuthController struct {
	Clerk domain.ClerkInterface
}

func (auth *AuthController) GetTokenFromSession() echo.HandlerFunc {
	return func(c echo.Context) error {
		sessionId := c.QueryParam("sessionId")
		template := c.QueryParam("template")

		if sessionId == "" || template == "" {
			return ReturnWriteResponse(c, domain.ErrRequiredParamsMissing, nil)
		}

		urlPath := fmt.Sprintf("sessions/%s/tokens/%s", sessionId, template)

		req, err := auth.Clerk.NewRequest("POST", urlPath, nil)
		if err != nil {
			c.Logger().Errorf("Error creating Clerk request: %v", err)
			return ReturnWriteResponse(c, errors.New("error creating auth request"), nil)
		}

		var tokenResponse = clerk.SessionToken{}

		_, err = auth.Clerk.Do(req, &tokenResponse)
		if err != nil {
			c.Logger().Errorf("Error executing Clerk request or processing response: %v", err)
			return ReturnWriteResponse(c, errors.New("error processing auth response"+err.Error()), nil)
		}

		if tokenResponse.JWT == "" {
			c.Logger().Warnf("Clerk token response successful but JWT was empty for session %s, template %s", sessionId, template)
			return ReturnWriteResponse(c, errors.New("empty JWT in token response"), nil)
		}

		return c.JSON(http.StatusOK, tokenResponse)
	}
}

//inMW.RoleMiddleware(auth, "org:admin", "org:teacher", "org:student"))
