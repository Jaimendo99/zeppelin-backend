package controller

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"zeppelin/internal/domain"
)

type AssignmentController struct {
	Repo domain.AssignmentRepo
}

func (c *AssignmentController) GetAssignmentsByStudent() echo.HandlerFunc {
	return func(e echo.Context) error {
		role := e.Get("user_role").(string)
		userID := e.Get("user_id").(string)

		if role != "org:student" {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusForbidden, "Solo los estudiantes pueden ver sus asignaciones"), nil)
		}

		assignments, err := c.Repo.GetAssignmentsByStudent(userID)
		return ReturnReadResponse(e, err, assignments)
	}
}

func (c *AssignmentController) GetStudentsByCourse() echo.HandlerFunc {
	return func(e echo.Context) error {
		role := e.Get("user_role").(string)

		if role != "org:teacher" {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusForbidden, "Solo los profesores pueden ver los estudiantes de sus cursos"), nil)
		}

		courseID, err := strconv.Atoi(e.Param("course_id"))
		if err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusBadRequest, "ID de curso inválido"), nil)
		}

		assignments, err := c.Repo.GetStudentsByCourse(courseID)
		return ReturnReadResponse(e, err, assignments)
	}
}

func (c *AssignmentController) CreateAssignment() echo.HandlerFunc {
	return func(e echo.Context) error {
		role := e.Get("user_role").(string)
		userID := e.Get("user_id").(string)

		if role != "org:student" {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusForbidden, "Solo los estudiantes pueden inscribirse a cursos"), nil)
		}

		var input struct {
			QRCode string `json:"qr_code"`
		}
		if err := e.Bind(&input); err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusBadRequest, "Datos inválidos"), nil)
		}

		courseID, err := c.Repo.GetCourseIDByQRCode(input.QRCode)
		if err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusNotFound, "Curso no encontrado con ese código QR"), nil)
		}

		err = c.Repo.CreateAssignment(userID, courseID)
		return ReturnWriteResponse(e, err, struct {
			Message string `json:"message"`
		}{Message: "Inscripción exitosa"})
	}
}

func (c *AssignmentController) VerifyAssignment() echo.HandlerFunc {
	return func(e echo.Context) error {
		// Solo un profesor o administrador puede verificar
		role := e.Get("user_role").(string)
		if role != "org:teacher" && role != "org:admin" {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusForbidden, "No tienes permiso para verificar"), nil)
		}

		// Obtener el ID de la asignación desde los parámetros
		assignmentID, err := strconv.Atoi(e.Param("assignment_id"))
		if err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusBadRequest, "ID de asignación inválido"), nil)
		}

		err = c.Repo.VerifyAssignment(assignmentID)
		return ReturnWriteResponse(e, err, struct {
			Message string `json:"message"`
		}{Message: "Asignación verificada exitosamente"})
	}
}
