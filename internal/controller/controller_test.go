package controller_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"zeppelin/internal/controller"
	"zeppelin/internal/domain"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type MockRepresentativeRepo struct {
	CreateRep func(representative domain.RepresentativeDb) error
	GetRep    func(id int) (*domain.RepresentativeInput, error)
	GetAllRep func() ([]domain.Representative, error)
	UpdateRep func(id int, representative domain.RepresentativeInput) error
}

func (m MockRepresentativeRepo) CreateRepresentative(representative domain.RepresentativeDb) error {
	return m.CreateRep(representative)
}

func (m MockRepresentativeRepo) GetRepresentative(id int) (*domain.RepresentativeInput, error) {
	return m.GetRep(id)
}

func (m MockRepresentativeRepo) GetAllRepresentatives() ([]domain.Representative, error) {
	return m.GetAllRep()
}

func (m MockRepresentativeRepo) UpdateRepresentative(id int, representative domain.RepresentativeInput) error {
	return m.UpdateRep(id, representative)
}

func TestCreateRepresentative_Success(t *testing.T) {
	var repeJson = `{"name":"Anthony","lastname":"Cochea","email":"anthony@gmail.com","phone_number":"+593990269309"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(repeJson))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	e.Validator = &controller.CustomValidator{Validator: validator.New()}

	mockRepo := MockRepresentativeRepo{
		CreateRep: func(representative domain.RepresentativeDb) error {
			return nil
		},
	}
	recontroller := controller.RepresentativeController{Repo: mockRepo}
	handler := recontroller.CreateRepresentative()

	if assert.NoError(t, handler(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedMessage := `{"Body":{"message":"Representative created"}}`
		assert.JSONEq(t, expectedMessage, rec.Body.String())
	}
}

func TestCreateRepresentative_BadRequest(t *testing.T) {
	var repeJson = `{"name":"Anthony","lastname":"Cochea","email":"","phone_number":""`
	e := echo.New()
	e.HTTPErrorHandler = testHTTPErrorHandler

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(repeJson))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	e.Validator = &controller.CustomValidator{Validator: validator.New()}

	mockRepo := MockRepresentativeRepo{
		CreateRep: func(representative domain.RepresentativeDb) error {
			return nil
		},
	}
	recontroller := controller.RepresentativeController{Repo: mockRepo}
	handler := recontroller.CreateRepresentative()

	err := handler(c)
	if err != nil {
		e.HTTPErrorHandler(err, c)
	}

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	expectedMessage := `{"message":"Invalid request body"}`
	assert.JSONEq(t, expectedMessage, rec.Body.String())
}

func TestGetRepresentative_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockRepo := MockRepresentativeRepo{
		GetRep: func(id int) (*domain.RepresentativeInput, error) {
			return &domain.RepresentativeInput{
				Name:        "Anthony",
				Lastname:    "Cochea",
				Email:       "anthony@gmail.com",
				PhoneNumber: "+593990269309",
			}, nil
		},
	}
	recontroller := controller.RepresentativeController{Repo: mockRepo}
	handler := recontroller.GetRepresentative()

	c.SetPath("/representative/:representative_id")
	c.SetParamNames("representative_id")
	c.SetParamValues("1")

	if assert.NoError(t, handler(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedMessage := `{"name":"Anthony","lastname":"Cochea","email":"anthony@gmail.com","phone_number":"+593990269309"}`
		assert.JSONEq(t, expectedMessage, rec.Body.String())
	}

}

func TestGetRepresentative_NotFound(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = testHTTPErrorHandler

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockRepo := MockRepresentativeRepo{
		GetRep: func(id int) (*domain.RepresentativeInput, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	recontroller := controller.RepresentativeController{Repo: mockRepo}
	handler := recontroller.GetRepresentative()

	c.SetPath("/representative/:representative_id")
	c.SetParamNames("representative_id")
	c.SetParamValues("1")

	err := handler(c)
	if err != nil {
		e.HTTPErrorHandler(err, c)
	}

	assert.Equal(t, http.StatusNotFound, rec.Code)
	expectedMessage := `{"message":"{Record not found}"}`
	assert.JSONEq(t, expectedMessage, rec.Body.String())
}

func TestGetAllRepresentatives_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockRepo := MockRepresentativeRepo{
		GetAllRep: func() ([]domain.Representative, error) {
			return []domain.Representative{
				{
					RepresentativeId: 1,
					Name:             "Anthony",
					Lastname:         "Cochea",
					Email:            "",
					PhoneNumber:      "",
				},
				{
					RepresentativeId: 2,
					Name:             "Mateo",
					Lastname:         "Mejia",
					Email:            "mateo@gmail.com",
					PhoneNumber:      "+593990269309",
				},
			}, nil
		},
	}

	recontroller := controller.RepresentativeController{Repo: mockRepo}
	handler := recontroller.GetAllRepresentatives()

	if assert.NoError(t, handler(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedMessage := `[{"representative_id":1,"name":"Anthony","lastname":"Cochea","email":"","phone_number":""},{"representative_id":2,"name":"Mateo","lastname":"Mejia","email":"mateo@gmail.com","phone_number":"+593990269309"}]`
		assert.JSONEq(t, expectedMessage, rec.Body.String())
	}
}

func testHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	msg := http.StatusText(code)
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if m, ok := he.Message.(string); ok {
			msg = m
		} else {
			msg = fmt.Sprintf("%v", he.Message)
		}
	}
	if !c.Response().Committed {
		_ = c.JSON(code, map[string]interface{}{"message": msg})
	}
}
