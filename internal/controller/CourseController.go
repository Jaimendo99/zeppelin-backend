package controller

import (
	"crypto/rand"
	"encoding/base32"
	"github.com/labstack/echo/v4"
	"strings"
	"zeppelin/internal/domain"
)

type CourseController struct {
	Repo domain.CourseRepo
}

func (c *CourseController) CreateCourse() echo.HandlerFunc {
	return func(e echo.Context) error {
		role := e.Get("user_role").(string)
		userID := e.Get("user_id").(string)

		if role != "org:teacher" {
			return ReturnWriteResponse(e, echo.NewHTTPError(403, "Solo los profesores pueden crear cursos"), nil)
		}

		var input domain.CourseInput
		if err := ValidateAndBind(e, &input); err != nil {
			return err
		}

		course := domain.CourseDB{
			TeacherID:   userID,
			StartDate:   input.StartDate,
			Title:       input.Title,
			Description: input.Description,
			QRCode:      generateQRCode(),
		}

		err := c.Repo.CreateCourse(course)
		return ReturnWriteResponse(e, err, struct {
			Message string `json:"message"`
		}{Message: "Curso creado con éxito"})
	}
}

func (c *CourseController) GetCoursesByTeacher() echo.HandlerFunc {
	return func(e echo.Context) error {
		role := e.Get("user_role").(string)
		userID := e.Get("user_id").(string)

		if role != "org:teacher" {
			return ReturnReadResponse(e, echo.NewHTTPError(403, "Solo los profesores pueden ver sus cursos"), nil)
		}

		courses, err := c.Repo.GetCoursesByTeacher(userID)
		return ReturnReadResponse(e, err, courses)
	}
}

func generateQRCode() string {
	bytes := make([]byte, 5)
	_, err := rand.Read(bytes)
	if err != nil {
		return ""
	}
	code := base32.StdEncoding.EncodeToString(bytes)
	code = strings.TrimRight(code, "=")
	code = strings.ToLower(code[:4]) + code[4:6]
	return code
}
