package controller

import (
	"github.com/labstack/echo/v4"
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

func (c *CourseController) GetCoursesByStudent() echo.HandlerFunc {
	return func(e echo.Context) error {
		role := e.Get("user_role").(string)
		userID := e.Get("user_id").(string)

		if role != "org:student" {
			return ReturnReadResponse(e, echo.NewHTTPError(403, "Solo los estudiantes pueden ver sus cursos"), nil)
		}

		courses, err := c.Repo.GetCoursesByStudent(userID)
		return ReturnReadResponse(e, err, courses)
	}
}
