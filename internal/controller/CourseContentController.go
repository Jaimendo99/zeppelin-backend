package controller

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"zeppelin/internal/config"
	"zeppelin/internal/domain"

	"github.com/labstack/echo/v4"
)

type CourseContentController struct {
	Repo          domain.CourseContentRepo
	RepoAssigment domain.AssignmentRepo
	RepoCourse    domain.CourseRepo
}

func (c *CourseContentController) GetCourseContentTeacher() echo.HandlerFunc {
	return func(e echo.Context) error {
		userID := e.Get("user_id").(string)
		courseID, err := strconv.Atoi(e.QueryParam("course_id"))
		if err != nil {
			return ReturnReadResponse(e, echo.NewHTTPError(http.StatusBadRequest, "course_id inválido"), nil)
		}

		_, err = c.RepoCourse.GetCourseByTeacherAndCourseID(userID, courseID)
		if err != nil {
			return ReturnReadResponse(e, echo.NewHTTPError(http.StatusForbidden, "Este curso no le pertenece al profesor"), nil)
		}

		data, err := c.Repo.GetContentByCourse(courseID, false)
		if err != nil {
			return ReturnReadResponse(e, err, nil)
		}

		for i, content := range data {
			for j, detail := range content.Details {
				if detail.ContentID == "" {
					continue
				}

				var key string
				switch detail.ContentTypeID {
				case 1: // Video
					key = fmt.Sprintf("focused/%d/video/teacher/%s.json", courseID, detail.ContentID)
				case 2: // Quiz
					key = fmt.Sprintf("focused/%d/quiz/teacher/%s.json", courseID, detail.ContentID)
				case 3: // Text
					key = fmt.Sprintf("focused/%d/text/teacher/%s.json", courseID, detail.ContentID)
				default:
					continue
				}

				signedURL, err := config.GeneratePresignedURL("zeppelin", key)
				if err != nil {
					return ReturnReadResponse(e, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error al generar URL firmada para content_type_id %d", detail.ContentTypeID)), nil)
				}

				data[i].Details[j].Url = signedURL
			}
		}

		return ReturnReadResponse(e, nil, data)
	}
}

func (c *CourseContentController) GetCourseContentForStudent() echo.HandlerFunc {
	return func(e echo.Context) error {
		role := e.Get("user_role").(string)
		userID := e.Get("user_id").(string)

		bourbon := os.Getenv("BOURBON")
		if bourbon == "1" {
			if role != "org:student" && role != "org:teacher" {
				return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusForbidden, "Solo los estudiantes o profesores pueden ver el contenido de los cursos"), nil)
			}
		} else {
			if role != "org:student" {
				return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusForbidden, "Solo los estudiantes pueden ver el contenido de los cursos"), nil)
			}
		}

		courseID, err := strconv.Atoi(e.QueryParam("course_id"))
		if err != nil {
			return ReturnReadResponse(e, echo.NewHTTPError(http.StatusBadRequest, "course_id inválido"), nil)
		}

		_, err = c.RepoAssigment.GetAssignmentsByStudentAndCourse(userID, courseID)
		if err != nil {
			return ReturnReadResponse(e, echo.NewHTTPError(http.StatusForbidden, "Este estudiante no está asignado a este curso"), nil)
		}

		data, err := c.Repo.GetContentByCourseForStudent(courseID, true, userID)
		if err != nil {
			return ReturnReadResponse(e, err, nil)
		}

		for i, content := range data {
			for j, detail := range content.Details {
				if detail.ContentID == "" {
					continue
				}

				var key string
				switch detail.ContentTypeID {
				case 1: // Video
					key = fmt.Sprintf("focused/%d/video/teacher/%s.json", courseID, detail.ContentID)
				case 2: // Quiz
					key = fmt.Sprintf("focused/%d/quiz/student/%s.json", courseID, detail.ContentID)
				case 3: // Text
					key = fmt.Sprintf("focused/%d/text/teacher/%s.json", courseID, detail.ContentID)
				default:
					continue
				}

				signedURL, err := config.GeneratePresignedURL("zeppelin", key)
				if err != nil {
					return ReturnReadResponse(e, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error al generar URL firmada para content_type_id %d", detail.ContentTypeID)), nil)
				}

				data[i].Details[j].Url = signedURL
			}
		}

		return ReturnReadResponse(e, nil, data)
	}
}

func (c *CourseContentController) AddModule() echo.HandlerFunc {
	return func(e echo.Context) error {

		userID := e.Get("user_id").(string)
		var input domain.AddModuleInput
		if err := ValidateAndBind(e, &input); err != nil {
			return err
		}

		courseContentID, err := c.Repo.AddModule(input.CourseID, input.Module, userID)
		if err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusForbidden, err.Error()), nil)
		}

		return ReturnWriteResponse(e, nil, map[string]interface{}{
			"message":           "Módulo creado",
			"course_content_id": courseContentID,
			"module":            input.Module,
		})
	}
}

func (c *CourseContentController) AddSection() echo.HandlerFunc {
	return func(e echo.Context) error {
		userID := e.Get("user_id").(string)
		var input domain.AddSectionInput
		if err := ValidateAndBind(e, &input); err != nil {
			return err
		}

		contentID, err := c.Repo.AddSection(input, userID)
		if err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusForbidden, err.Error()), nil)
		}

		//Comprueba el tipo de contenido segun input.ContentTypeID
		var contentType int
		switch input.ContentTypeID {
		case 1:
			contentType = 1
		case 2:
			contentType = 2
		case 3:
			contentType = 3
		default:
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusBadRequest, "content_type_id inválido"), nil)

		}

		return ReturnWriteResponse(e, nil, map[string]interface{}{
			"message":           "Sección agregada",
			"content_id":        contentID,
			"course_content_id": input.CourseContentID,
			"content_type_id":   input.ContentTypeID,
		})
	}
}

func (c *CourseContentController) UpdateContent() echo.HandlerFunc {
	return func(e echo.Context) error {
		var input domain.UpdateContentInput
		if err := ValidateAndBind(e, &input); err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusBadRequest, "datos inválidos"), nil)
		}

		if input.ContentID == "" {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusBadRequest, "content_id requerido"), nil)
		}

		err := c.Repo.UpdateContent(input)
		if err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, "error al actualizar el contenido"), nil)
		}

		return ReturnWriteResponse(e, nil, map[string]string{"message": "Contenido actualizado"})
	}
}

func (c *CourseContentController) UpdateContentStatus() echo.HandlerFunc {
	return func(e echo.Context) error {
		var input domain.UpdateContentStatusInput
		if err := ValidateAndBind(e, &input); err != nil {
			return err
		}

		err := c.Repo.UpdateContentStatus(input.CourseContentID, input.IsActive)
		return ReturnWriteResponse(e, err, map[string]string{"message": "Estado del contenido actualizado"})
	}
}

func (c *CourseContentController) UpdateModuleTitle() echo.HandlerFunc {
	return func(e echo.Context) error {
		var input domain.UpdateModuleTitleInput
		if err := ValidateAndBind(e, &input); err != nil {
			return err
		}

		err := c.Repo.UpdateModuleTitle(input.CourseContentID, input.ModuleTitle)
		if err != nil {
			return ReturnWriteResponse(e, err, nil)
		}

		return ReturnWriteResponse(e, nil, map[string]string{"message": "Título del módulo actualizado"})
	}
}

func (c *CourseContentController) UpdateUserContentStatus(statusID int) echo.HandlerFunc {
	return func(e echo.Context) error {
		userID := e.Get("user_id").(string)

		var input domain.UpdateUserContentStatusInput
		if err := ValidateAndBind(e, &input); err != nil {
			return err
		}

		err := c.Repo.UpdateUserContentStatus(userID, input.ContentID, statusID)
		if err != nil {
			return ReturnWriteResponse(e, err, nil)
		}

		var msg string
		switch statusID {
		case 2:
			msg = "Contenido marcado como 'en progreso'"
		case 3:
			msg = "Contenido marcado como 'completado'"
		default:
			msg = "Estado actualizado"
		}

		return ReturnWriteResponse(e, nil, map[string]string{"message": msg})
	}
}
