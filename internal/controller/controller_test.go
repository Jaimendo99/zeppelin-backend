package controller_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"zeppelin/internal/controller"
	"zeppelin/internal/domain"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type MockRepresentativeService struct {
	// Define functions to simulate behavior.
	CreateFunc func(input domain.RepresentativeInput) error
	GetFunc    func(id string) (domain.RepresentativeDb, error)
	GetAllFunc func() ([]domain.Representative, error)
	UpdateFunc func(id string, input domain.RepresentativeInput) error
}

func (m *MockRepresentativeService) CreateRepresentative(input domain.RepresentativeInput) error {
	return m.CreateFunc(input)
}
func (m *MockRepresentativeService) GetRepresentative(id string) (domain.RepresentativeDb, error) {
	return m.GetFunc(id)
}

func (m *MockRepresentativeService) GetAllRepresentatives() ([]domain.Representative, error) {
	return m.GetAllFunc()
}

func (m *MockRepresentativeService) UpdateRepresentative(id string, input domain.RepresentativeInput) error {
	return m.UpdateFunc(id, input)
}

func TestCreateRepresentative_Success(t *testing.T) {
	e := echo.New()
	requestJSON := `{"name": "John", "lastname": "Doe"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(requestJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockService := &MockRepresentativeService{
		CreateFunc: func(input domain.RepresentativeInput) error {
			if input.Name != "John" || input.Lastname != "Doe" {
				return errors.New("unexpected input")
			}
			return nil
		},
	}

	controller := controller.NewRepresentativeController(mockService)
	handler := controller.CreateRepresentative()
	if assert.NoError(t, handler(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		// Expected response JSON: {"Message":"Representative created"}
		expectedResponse := `{"Message":"Representative created"}`
		assert.JSONEq(t, expectedResponse, rec.Body.String())
	}
}

func TestCreateRepresentative_BadRequest(t *testing.T) {
	// Setup Echo instance and request with invalid JSON.
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("invalid json"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// The service won't be called because binding will fail.
	mockService := &MockRepresentativeService{
		CreateFunc: func(input domain.RepresentativeInput) error {
			return nil
		},
	}

	controller := controller.NewRepresentativeController(mockService)

	handler := controller.CreateRepresentative()
	// Expect a bad request response.
	if assert.NoError(t, handler(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		expectedResponse := `{"Message":"Invalid request"}`
		assert.JSONEq(t, expectedResponse, rec.Body.String())
	}
}

func TestGetRepresentative_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// Set path parameter "representative_id".
	c.SetParamNames("representative_id")
	c.SetParamValues("1")

	// Sample representative that the service should return.
	sampleRepresentative := domain.RepresentativeDb{
		Name:     "John",
		Lastname: "Doe",
	}

	// Create a mock service that returns the sample representative.
	mockService := &MockRepresentativeService{
		GetFunc: func(id string) (domain.RepresentativeDb, error) {
			if id == "1" {
				return sampleRepresentative, nil
			}
			return domain.RepresentativeDb{}, errors.New("not found")
		},
	}

	controller := controller.NewRepresentativeController(mockService)

	handler := controller.GetRepresentative()
	if assert.NoError(t, handler(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		// Convert the sample representative to JSON for comparison.
		expectedJSON, err := json.Marshal(sampleRepresentative)
		assert.NoError(t, err)
		assert.JSONEq(t, string(expectedJSON), rec.Body.String())
	}
}

func TestGetRepresentative_Error(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("representative_id")
	c.SetParamValues("2")

	mockService := &MockRepresentativeService{
		GetFunc: func(id string) (domain.RepresentativeDb, error) {
			return domain.RepresentativeDb{}, errors.New("internal error")
		},
	}

	controller := controller.NewRepresentativeController(mockService)

	handler := controller.GetRepresentative()
	if assert.NoError(t, handler(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		expectedResponse := `{"Message":"Internal server error"}`
		assert.JSONEq(t, expectedResponse, rec.Body.String())
	}
}

func TestGetAllRepresentatives_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Sample representatives that the service should return.
	sampleRepresentatives := []domain.Representative{
		{
			Name:     "John",
			Lastname: "Doe",
		},
		{
			Name:     "Jane",
			Lastname: "Doe",
		},
	}

	// Create a mock service that returns the sample representatives.
	mockService := &MockRepresentativeService{
		GetAllFunc: func() ([]domain.Representative, error) {
			return sampleRepresentatives, nil
		},
	}

	controller := controller.NewRepresentativeController(mockService)

	handler := controller.GetAllRepresentatives()
	if assert.NoError(t, handler(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		// Convert the sample representatives to JSON for comparison.
		expectedJSON, err := json.Marshal(sampleRepresentatives)
		assert.NoError(t, err)
		assert.JSONEq(t, string(expectedJSON), rec.Body.String())
	}
}

func TestGetAllRepresentatives_Error(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockService := &MockRepresentativeService{
		GetAllFunc: func() ([]domain.Representative, error) {
			return nil, errors.New("internal error")
		},
	}

	controller := controller.NewRepresentativeController(mockService)

	handler := controller.GetAllRepresentatives()
	if assert.NoError(t, handler(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		expectedResponse := `{"Message":"Internal server error"}`
		assert.JSONEq(t, expectedResponse, rec.Body.String())
	}
}

func TestUpdateRepresentative_Success(t *testing.T) {
	e := echo.New()
	requestJSON := `{"name": "John", "lastname": "Doe"}`
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(requestJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("representative_id")
	c.SetParamValues("1")

	mockService := &MockRepresentativeService{
		UpdateFunc: func(id string, input domain.RepresentativeInput) error {
			if id != "1" || input.Name != "John" || input.Lastname != "Doe" {
				return errors.New("unexpected input")
			}
			return nil
		},
	}

	controller := controller.NewRepresentativeController(mockService)

	handler := controller.UpdateRepresentative()
	if assert.NoError(t, handler(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedResponse := `{"Message":"Representative updated"}`
		assert.JSONEq(t, expectedResponse, rec.Body.String())
	}
}

func TestUpdateRepresentative_BadRequest(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader("invalid json"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("representative_id")
	c.SetParamValues("1")

	mockService := &MockRepresentativeService{
		UpdateFunc: func(id string, input domain.RepresentativeInput) error {
			return nil
		},
	}

	controller := controller.NewRepresentativeController(mockService)

	handler := controller.UpdateRepresentative()
	if assert.NoError(t, handler(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		expectedResponse := `{"Message":"Invalid request"}`
		assert.JSONEq(t, expectedResponse, rec.Body.String())
	}
}
