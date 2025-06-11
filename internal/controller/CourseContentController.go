package controller

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"zeppelin/internal/domain"

	"github.com/labstack/echo/v4"
)

type CourseContentController struct {
	Repo                 domain.CourseContentRepo
	RepoAssigment        domain.AssignmentRepo
	RepoCourse           domain.CourseRepo
	GeneratePresignedURL func(bucket, key string) (string, error)
	UploadTextFunc       func(courseID, contentID string, json []byte) error
	UploadQuizFunc       func(courseID, contentID string, jsonBytes []byte) error
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
		// En tu controlador (antes de llamar al repositorio)
		fmt.Printf("Repositorio: %+v\n", err)

		data, err := c.Repo.GetContentByCourse(courseID)
		fmt.Printf("Data recibida: %+v\n", data)        // Agrega esta línea para ver los datos
		fmt.Printf("Error del repositorio: %+v\n", err) // Asegúrate de que este err sea nil

		if err != nil {
			return ReturnReadResponse(e, err, nil)
		}

		for i, content := range data {
			// **Verifica si content.Details no es nil antes de iterar**
			if content.Details != nil {
				for j, detail := range content.Details {
					if detail.ContentID == "" {
						continue
					}

					var key string
					switch detail.ContentTypeID {
					case 1: // Video
						key = fmt.Sprintf("focused/%d/video/teacher/%s.json", courseID, detail.ContentID)
					case 2: // Quiz
						key = fmt.Sprintf("focused/%d/text/teacher/%s.json", courseID, detail.ContentID)
					case 3: // Text
						key = fmt.Sprintf("focused/%d/quiz/teacher/%s.json", courseID, detail.ContentID)
					default:
						continue
					}

					// Use the injected GeneratePresignedURL function
					// Verifica si c.GeneratePresignedURL no es nil
					if c.GeneratePresignedURL == nil {
						return ReturnReadResponse(e, echo.NewHTTPError(http.StatusInternalServerError, "función GeneratePresignedURL no inicializada"), nil)
					}

					signedURL, err := c.GeneratePresignedURL("zeppelin", key)
					if err != nil {
						return ReturnReadResponse(e, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error al generar URL firmada para content_type_id %d", detail.ContentTypeID)), nil)
					}

					if detail.ContentTypeID == 1 {
						data[i].Details[j].Url = detail.Url
					} else {
						data[i].Details[j].Url = signedURL
					}
				}
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

		data, err := c.Repo.GetContentByCourseForStudent(courseID, userID)
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
				case 2: // Quiz
					key = fmt.Sprintf("focused/%d/text/teacher/%s.json", courseID, detail.ContentID)
				case 3: // Text
					key = fmt.Sprintf("focused/%d/quiz/student/%s.json", courseID, detail.ContentID)
				default:
					continue
				}

				signedURL, err := c.GeneratePresignedURL("zeppelin", key)

				if err != nil {
					return ReturnReadResponse(e, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error al generar URL firmada para content_type_id %d", detail.ContentTypeID)), nil)
				}

				if detail.ContentTypeID == 1 {
					data[i].Details[j].Url = detail.Url
				} else {
					data[i].Details[j].Url = signedURL
				}

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

		contentTypeID, err := c.Repo.GetContentTypeID(input.ContentID)
		if err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, err.Error()), nil)
		}

		var unsignedURL string
		// Verificar el tipo de contenido para subir a R2 si es necesario
		switch contentTypeID {
		case 2: // Texto
			if input.JsonData != nil {
				courseIDStr := strconv.Itoa(input.CourseID)
				err = c.UploadTextFunc(courseIDStr, input.ContentID, input.JsonData)
				if err != nil {
					return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, "error al subir texto a R2"), nil)
				}
				// Generate unsigned URL for text
				accountID := os.Getenv("R2_ACCOUNT_ID")
				unsignedURL = fmt.Sprintf("https://%s.r2.cloudflarestorage.com/focused/%s/text/teacher/%s.json",
					accountID, courseIDStr, input.ContentID)
				input.Url = unsignedURL
			}
		case 3: // Quiz
			if input.JsonData != nil {
				courseIDStr := strconv.Itoa(input.CourseID)
				err = c.UploadQuizFunc(courseIDStr, input.ContentID, input.JsonData)
				if err != nil {
					return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, "error al subir quiz a R2"), nil)
				}
				// Generate unsigned URL for quiz (teacher version)
				accountID := os.Getenv("R2_ACCOUNT_ID")
				unsignedURL = fmt.Sprintf("https://%s.r2.cloudflarestorage.com/focused/%s/quiz/teacher/%s.json",
					accountID, courseIDStr, input.ContentID) // Fixed: use ContentID instead of CourseID
				input.Url = unsignedURL
			}
		case 1:
			input.Url = input.VideoID
		default:
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusBadRequest, "content_type_id inválido"), nil)
		}

		err = c.Repo.UpdateContent(input)
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

		err := c.Repo.UpdateContentStatus(input.ContentID, input.IsActive)
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
