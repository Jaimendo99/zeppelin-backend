package controller

import (
	"github.com/clerkinc/clerk-sdk-go/clerk"
	"net/http"
	"zeppelin/internal/domain"

	"github.com/labstack/echo/v4"
)

type UserController struct {
	AuthService domain.AuthServiceI
	UserRepo    domain.UserRepo
}

func (c *UserController) RegisterUser(role string) echo.HandlerFunc {
	return func(e echo.Context) error {
		var req domain.UserInput
		if err := ValidateAndBind(e, &req); err != nil {
			e.Logger().Errorf("Error al validar y enlazar: %v", err)
			return err
		}

		typeID, err := GetTypeID(role)
		if err != nil {
			return ReturnWriteResponse(e, err, nil)
		}

		organizationID := "org_2tjxBeJV0WLJUFU6Q3AwjzMyXTs"

		user, err := c.AuthService.CreateUser(req, organizationID, role)
		if err != nil {
			return ReturnWriteResponse(e, err, struct {
				Message string `json:"message"`
			}{Message: "Error al crear usuario en Clerk"})
		}

		userDb := domain.UserDb{
			UserID:   user.UserID,
			Name:     req.Name,
			Lastname: req.Lastname,
			Email:    req.Email,
			TypeID:   typeID,
		}

		err = c.UserRepo.CreateUser(userDb)
		return ReturnWriteResponse(e, err, struct {
			Message string `json:"message"`
		}{Message: "Usuario registrado con éxito"})
	}
}

func (c *UserController) GetUser() echo.HandlerFunc {
	return func(e echo.Context) error {
		claims, ok := e.Get("user").(*clerk.SessionClaims)
		if !ok || claims == nil {
			return ReturnReadResponse(e, echo.ErrUnauthorized, nil)
		}

		userID := claims.Subject
		student, err := c.UserRepo.GetUser(userID)
		return ReturnReadResponse(e, err, student)
	}
}

func (c *UserController) GetAllTeachers() echo.HandlerFunc {
	return func(e echo.Context) error {
		teachers, err := c.UserRepo.GetAllTeachers()
		return ReturnReadResponse(e, err, teachers)
	}
}

func (c *UserController) GetAllStudents() echo.HandlerFunc {
	return func(e echo.Context) error {
		students, err := c.UserRepo.GetAllStudents()
		return ReturnReadResponse(e, err, students)
	}
}

func GetTypeID(role string) (int, error) {
	switch role {
	case "org:student":
		return 3, nil
	case "org:teacher":
		return 2, nil
	default:
		return 0, echo.NewHTTPError(http.StatusBadRequest, struct {
			Message string `json:"message"`
		}{Message: "Rol inválido"})
	}
}
