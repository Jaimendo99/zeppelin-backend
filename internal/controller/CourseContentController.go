package controller

import (
	"net/http"
	"strconv"
	"zeppelin/internal/domain"

	"github.com/labstack/echo/v4"
)

type CourseContentController struct {
	Repo domain.CourseContentRepo
}

func (c *CourseContentController) GetCourseContent() echo.HandlerFunc {
	return func(e echo.Context) error {
		courseID, err := strconv.Atoi(e.QueryParam("course_id"))
		if err != nil {
			return ReturnReadResponse(e, echo.NewHTTPError(http.StatusBadRequest, "course_id inválido"), nil)
		}

		data, err := c.Repo.GetContentByCourse(courseID)
		return ReturnReadResponse(e, err, data)
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

		contentID, err := c.Repo.CreateQuiz(input.Title, input.Description, nil) // json_content optional on create
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

		err := c.Repo.UpdateQuiz(input.ContentID, input.Title, input.Description, input.JsonContent)
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
