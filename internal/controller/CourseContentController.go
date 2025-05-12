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
	RepoCourse    domain.CourseRepo
	RepoAssigment domain.AssignmentRepo
}

func (c *CourseContentController) GetCourseContentTeacher() echo.HandlerFunc {
	return func(e echo.Context) error {
		userID := e.Get("user_id").(string) // Este es el ID del profesor autenticado
		courseID, err := strconv.Atoi(e.QueryParam("course_id"))
		if err != nil {
			return ReturnReadResponse(e, echo.NewHTTPError(http.StatusBadRequest, "course_id inválido"), nil)
		}

		// Verificar si el curso le pertenece al profesor
		_, err = c.RepoCourse.GetCourseByTeacherAndCourseID(userID, courseID)
		if err != nil {
			return ReturnReadResponse(e, echo.NewHTTPError(http.StatusForbidden, "Este curso no le pertenece al profesor"), nil)
		}

		// Para los profesores, no filtramos por IsActive (traemos todos los contenidos)
		data, err := c.Repo.GetContentByCourse(courseID, false)
		if err != nil {
			return ReturnReadResponse(e, err, nil)
		}

		// Generamos las URLs firmadas para los archivos de contenido
		for i, content := range data {
			switch content.ContentType {
			case "quiz":
				// Generar URL firmada para el quiz
				quizContent := content.Details.(domain.QuizContent)
				key := fmt.Sprintf("focused/%d/quiz/teacher/%s.json", courseID, quizContent.ContentID)
				signedURL, err := config.GeneratePresignedURL("zeppelin", key)
				if err != nil {
					return ReturnReadResponse(e, echo.NewHTTPError(http.StatusInternalServerError, "error al generar URL firmada para quiz"), nil)
				}

				// Imprimir la URL firmada generada para el quiz
				fmt.Println("URL firmada para el quiz:", signedURL)

				// Asignar la URL firmada al contenido
				quizContent.Url = signedURL
				data[i].Details = quizContent

			case "text":
				// Generar URL firmada para el texto
				textContent := content.Details.(domain.TextContent)
				key := fmt.Sprintf("focused/%d/text/teacher/%s.json", courseID, textContent.ContentID)
				signedURL, err := config.GeneratePresignedURL("zeppelin", key)
				if err != nil {
					return ReturnReadResponse(e, echo.NewHTTPError(http.StatusInternalServerError, "error al generar URL firmada para text"), nil)
				}

				// Imprimir la URL firmada generada para el texto
				fmt.Println("URL firmada para el texto:", signedURL)

				// Asignar la URL firmada al contenido
				textContent.Url = signedURL
				data[i].Details = textContent

			case "video":
				// Generar URL firmada para el video
				videoContent := content.Details.(domain.VideoContent)
				key := fmt.Sprintf("focused/%d/video/teacher/%s.json", courseID, videoContent.ContentID)
				signedURL, err := config.GeneratePresignedURL("zeppelin", key)
				if err != nil {
					return ReturnReadResponse(e, echo.NewHTTPError(http.StatusInternalServerError, "error al generar URL firmada para video"), nil)
				}

				// Imprimir la URL firmada generada para el video
				fmt.Println("URL firmada para el video:", signedURL)

				// Asignar la URL firmada al contenido
				videoContent.Url = signedURL
				data[i].Details = videoContent
			}
		}

		return ReturnReadResponse(e, nil, data)
	}
}

func (c *CourseContentController) GetCourseContentForStudent() echo.HandlerFunc {
	return func(e echo.Context) error {
		// Obtener el rol del usuario y su ID desde el contexto
		role := e.Get("user_role").(string)
		userID := e.Get("user_id").(string)

		// Verificar si el usuario es un estudiante
		if role != "org:student" {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusForbidden, "Solo los estudiantes pueden ver el contenido de los cursos"), nil)
		}

		// Obtener el `courseID` desde la consulta de la URL
		courseID, err := strconv.Atoi(e.QueryParam("course_id"))
		if err != nil {
			return ReturnReadResponse(e, echo.NewHTTPError(http.StatusBadRequest, "course_id inválido"), nil)
		}

		// Verificar si el estudiante está asignado a este curso
		_, err = c.RepoAssigment.GetAssignmentsByStudentAndCourse(userID, courseID)
		if err != nil {
			return ReturnReadResponse(e, echo.NewHTTPError(http.StatusForbidden, "Este estudiante no está asignado a este curso"), nil)
		}

		// Para los estudiantes, solo se traen los contenidos activos
		data, err := c.Repo.GetContentByCourseForStudent(courseID, true, userID)
		if err != nil {
			return ReturnReadResponse(e, err, nil)
		}

		// Generamos las URLs firmadas para los archivos de contenido
		for i, content := range data {
			switch content.ContentType {
			case "quiz":
				// Generar URL firmada para el quiz
				quizContent := content.Details.(domain.QuizContent)
				key := fmt.Sprintf("focused/%d/quiz/student/%s.json", courseID, quizContent.ContentID)
				signedURL, err := config.GeneratePresignedURL("zeppelin", key)
				if err != nil {
					return ReturnReadResponse(e, echo.NewHTTPError(http.StatusInternalServerError, "error al generar URL firmada para quiz"), nil)
				}

				// Imprimir la URL firmada generada para el quiz
				fmt.Println("URL firmada para el quiz:", signedURL)

				// Asignar la URL firmada al contenido
				quizContent.Url = signedURL
				data[i].Details = quizContent

			case "text":
				// Generar URL firmada para el texto
				textContent := content.Details.(domain.TextContent)
				key := fmt.Sprintf("focused/%d/text/teacher/%s.json", courseID, textContent.ContentID)
				signedURL, err := config.GeneratePresignedURL("zeppelin", key)
				if err != nil {
					return ReturnReadResponse(e, echo.NewHTTPError(http.StatusInternalServerError, "error al generar URL firmada para text"), nil)
				}

				// Imprimir la URL firmada generada para el texto
				fmt.Println("URL firmada para el texto:", signedURL)

				// Asignar la URL firmada al contenido
				textContent.Url = signedURL
				data[i].Details = textContent

			case "video":
				// Generar URL firmada para el video
				videoContent := content.Details.(domain.VideoContent)
				key := fmt.Sprintf("focused/%d/video/teacher/%s.json", courseID, videoContent.ContentID)
				signedURL, err := config.GeneratePresignedURL("zeppelin", key)
				if err != nil {
					return ReturnReadResponse(e, echo.NewHTTPError(http.StatusInternalServerError, "error al generar URL firmada para video"), nil)
				}

				// Imprimir la URL firmada generada para el video
				fmt.Println("URL firmada para el video:", signedURL)

				// Asignar la URL firmada al contenido
				videoContent.Url = signedURL
				data[i].Details = videoContent
			}
		}

		return ReturnReadResponse(e, nil, data)
	}
}

func (c *CourseContentController) AddVideoSection() echo.HandlerFunc {
	return func(e echo.Context) error {
		courseID, err := strconv.Atoi(e.QueryParam("course_id"))
		if err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusBadRequest, "course_id inválido"), nil)
		}

		var input domain.AddVideoSectionInput
		if err := ValidateAndBind(e, &input); err != nil {
			return err
		}
		contentID, err := c.Repo.CreateVideo(input.Url, input.Title, input.Description)
		if err != nil {
			return ReturnWriteResponse(e, err, nil)
		}

		err = c.Repo.AddVideoSection(courseID, contentID, input.Module, input.SectionIndex, input.ModuleIndex)
		return ReturnWriteResponse(e, err, map[string]string{"message": "Sección de video agregada", "content_id": contentID})
	}
}

func (c *CourseContentController) AddQuizSection() echo.HandlerFunc {
	return func(e echo.Context) error {
		courseID, err := strconv.Atoi(e.QueryParam("course_id"))
		if err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusBadRequest, "course_id inválido"), nil)
		}

		var input domain.AddQuizSectionInput
		if err := ValidateAndBind(e, &input); err != nil {
			return err
		}

		contentID, err := c.Repo.CreateQuiz(input.Title, input.Url, input.Description, nil) // json_content optional on create
		if err != nil {
			return ReturnWriteResponse(e, err, nil)
		}

		err = c.Repo.AddQuizSection(courseID, contentID, input.Module, input.SectionIndex, input.ModuleIndex)
		return ReturnWriteResponse(e, err, map[string]string{"message": "Sección de quiz agregada", "content_id": contentID})
	}
}

func (c *CourseContentController) AddTextSection() echo.HandlerFunc {
	return func(e echo.Context) error {
		courseID, err := strconv.Atoi(e.QueryParam("course_id"))
		if err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusBadRequest, "course_id inválido"), nil)
		}

		var input domain.AddTextSectionInput
		if err := ValidateAndBind(e, &input); err != nil {
			return err
		}

		contentID, err := c.Repo.CreateText(input.Title, "", nil) // url and json_content optional on create
		if err != nil {
			return ReturnWriteResponse(e, err, nil)
		}

		err = c.Repo.AddTextSection(courseID, contentID, input.Module, input.SectionIndex, input.ModuleIndex)
		return ReturnWriteResponse(e, err, map[string]string{"message": "Sección de texto agregada", "content_id": contentID})
	}
}

func (c *CourseContentController) UpdateVideoContent() echo.HandlerFunc {
	return func(e echo.Context) error {
		var input domain.UpdateVideoContentInput
		if err := ValidateAndBind(e, &input); err != nil {
			return err
		}

		err := c.Repo.UpdateVideo(input.ContentID, input.Title, input.Url, input.Description)
		return ReturnWriteResponse(e, err, map[string]string{"message": "Video actualizado"})
	}
}

func (c *CourseContentController) UpdateQuizContent() echo.HandlerFunc {
	return func(e echo.Context) error {
		var input domain.UpdateQuizContentInput
		if err := ValidateAndBind(e, &input); err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusBadRequest, "datos inválidos"), nil)
		}

		if input.CourseID == 0 {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusBadRequest, "course_id inválido"), nil)
		}

		if input.ContentID == "" {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusBadRequest, "content_id requerido"), nil)
		}

		// Si hay contenido JSON, subirlo a R2 y generar URL
		if input.JsonContent != nil {
			courseIDStr := strconv.Itoa(input.CourseID)

			err := config.UploadTeacherQuiz(courseIDStr, input.ContentID, input.JsonContent)
			if err != nil {
				return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, "error al subir JSON a R2"), nil)
			}

			accountID := os.Getenv("R2_ACCOUNT_ID")
			input.Url = fmt.Sprintf("https://%s.r2.cloudflarestorage.com/focused/%s/quiz/teacher/%s.json",
				accountID, courseIDStr, input.ContentID)
		}

		// Actualizar en la base de datos
		err := c.Repo.UpdateQuiz(
			input.ContentID,
			input.Title,
			input.Url,
			input.Description,
			input.JsonContent,
		)
		if err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, "error al actualizar el quiz"), nil)
		}

		return ReturnWriteResponse(e, nil, map[string]string{"message": "Quiz actualizado"})
	}
}

func (c *CourseContentController) UpdateTextContent() echo.HandlerFunc {
	return func(e echo.Context) error {
		var input domain.UpdateTextContentInput
		if err := ValidateAndBind(e, &input); err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusBadRequest, "datos inválidos"), nil)
		}

		if input.CourseID == 0 {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusBadRequest, "course_id inválido"), nil)
		}

		if input.ContentID == "" {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusBadRequest, "content_id requerido"), nil)
		}

		// Si hay contenido JSON, subirlo a R2 y generar URL
		if input.JsonContent != nil {
			courseIDStr := strconv.Itoa(input.CourseID)

			// Subir el texto a R2
			err := config.UploadTeacherText(courseIDStr, input.ContentID, input.JsonContent)
			if err != nil {
				return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, "error al subir JSON a R2"), nil)
			}

			// Generar la URL para el archivo subido
			accountID := os.Getenv("R2_ACCOUNT_ID")
			input.Url = fmt.Sprintf("https://%s.r2.cloudflarestorage.com/focused/%s/text/teacher/%s.json",
				accountID, courseIDStr, input.ContentID)
		}

		// Actualizar en la base de datos
		err := c.Repo.UpdateText(input.ContentID, input.Title, input.Url, input.JsonContent)
		if err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, "error al actualizar el texto"), nil)
		}

		return ReturnWriteResponse(e, nil, map[string]string{"message": "Texto actualizado"})
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
