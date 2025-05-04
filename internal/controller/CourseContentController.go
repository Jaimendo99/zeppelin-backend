package controller

import (
	"net/http"
	"strconv"
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
			return err
		}

		err := c.Repo.UpdateQuiz(input.ContentID, input.Title, input.Url, input.Description, input.JsonContent)
		return ReturnWriteResponse(e, err, map[string]string{"message": "Quiz actualizado"})
	}
}

func (c *CourseContentController) UpdateTextContent() echo.HandlerFunc {
	return func(e echo.Context) error {
		var input domain.UpdateTextContentInput
		if err := ValidateAndBind(e, &input); err != nil {
			return err
		}

		err := c.Repo.UpdateText(input.ContentID, input.Title, input.Url, input.JsonContent)
		return ReturnWriteResponse(e, err, map[string]string{"message": "Texto actualizado"})
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
