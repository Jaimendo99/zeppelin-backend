package controller_test

import (
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"zeppelin/internal/controller"
	"zeppelin/internal/domain"
	"zeppelin/internal/services"
)

// --- Reusing Setup and Dummy Helpers ---

func setupTest(req *http.Request) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	// Add validator if needed by ValidateAndBind
	e.Validator = &controller.CustomValidator{Validator: validator.New()}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func TestRepresentativeController_CreateRepresentative(t *testing.T) {
	mockRepo := new(domain.MockRepresentativeRepo)
	representativeController := controller.RepresentativeController{Repo: mockRepo}

	t.Run("Success", func(t *testing.T) {
		repInput := domain.RepresentativeInput{
			Name:        "John",
			Lastname:    "Doe",
			Email:       "john@doe.com",
			PhoneNumber: "+1234567890",
		}
		repInputJSON, _ := json.Marshal(repInput)
		req := httptest.NewRequest(http.MethodPost, "/representatives", strings.NewReader(string(repInputJSON)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c, rec := setupTest(req)

		// Assume RepresentativesInputToDb works correctly
		expectedDb := services.RepresentativesInputToDb(&repInput)
		mockRepo.On("CreateRepresentative", expectedDb).Return(nil).Once()

		handler := representativeController.CreateRepresentative()
		err := handler(c) // Calls ReturnWriteResponse internally

		assert.NoError(t, err) // Expect nil error from ReturnWriteResponse on success
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedResp := `{"Body":{"message":"Representative created"}}`
		assert.JSONEq(t, expectedResp, rec.Body.String())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure_BindingError", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/representatives", strings.NewReader(`{"invalid`)) // Malformed JSON
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c, rec := setupTest(req)
		handler := representativeController.CreateRepresentative()
		err := handler(c)
		assert.Error(t, err)
		var httpErr *echo.HTTPError
		ok := errors.As(err, &httpErr)
		assert.True(t, ok)
		if ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			assert.Equal(t, "Invalid request body", httpErr.Message)
		}
		assert.Empty(t, rec.Body.String())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure_RepoError", func(t *testing.T) {
		repInput := domain.RepresentativeInput{
			Name:        "Jane",
			Lastname:    "Doe",
			Email:       "asd@asd.com",
			PhoneNumber: "+0987654321",
		}
		repInputJSON, _ := json.Marshal(repInput)
		req := httptest.NewRequest(http.MethodPost, "/representatives", strings.NewReader(string(repInputJSON)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c, rec := setupTest(req)

		expectedDb := services.RepresentativesInputToDb(&repInput)
		repoError := errors.New("db connection error")
		mockRepo.On("CreateRepresentative", expectedDb).Return(repoError).Once()

		handler := representativeController.CreateRepresentative()
		err := handler(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		expectedResp := `{"message":"db connection error"}`
		assert.JSONEq(t, expectedResp, rec.Body.String())
		mockRepo.AssertExpectations(t)
	})
}

func TestRepresentativeController_GetRepresentative(t *testing.T) {
	mockRepo := new(domain.MockRepresentativeRepo)
	representativeController := controller.RepresentativeController{Repo: mockRepo}

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/representatives/123", nil)
		c, rec := setupTest(req)
		c.SetParamNames("representative_id")
		c.SetParamValues("123")

		expectedID := 123
		expectedRep := &domain.RepresentativeInput{ /* Populate with expected data */ }
		mockRepo.On("GetRepresentative", expectedID).Return(expectedRep, nil).Once()

		handler := representativeController.GetRepresentative()
		err := handler(c) // Calls ReturnReadResponse

		// Assert based on ReturnReadResponse success path
		assert.NoError(t, err) // Expect nil from e.JSON
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedJSON, _ := json.Marshal(expectedRep)
		assert.JSONEq(t, string(expectedJSON), rec.Body.String())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure_InvalidID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/representatives/abc", nil)
		c, rec := setupTest(req)
		c.SetParamNames("representative_id")
		c.SetParamValues("abc")

		handler := representativeController.GetRepresentative()
		err := handler(c) // ParamToId fails, error passed to ReturnReadResponse

		assert.Error(t, err) // Expect error from ReturnReadResponse
		var httpErr *echo.HTTPError
		ok := errors.As(err, &httpErr)
		assert.True(t, ok)
		if ok {
			assert.Equal(t, http.StatusInternalServerError, httpErr.Code) // Or 400 if ReturnReadResponse handles it specifically
			expectedMsgStruct := struct {
				Message string `json:"message"`
			}{Message: "Internal server error"} // Default from dummy
			assert.Equal(t, expectedMsgStruct, httpErr.Message)
		}
		assert.Empty(t, rec.Body.String())
		mockRepo.AssertExpectations(t) // Repo method not called
	})

	t.Run("Failure_NotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/representatives/404", nil)
		c, rec := setupTest(req)
		c.SetParamNames("representative_id")
		c.SetParamValues("404")

		expectedID := 404
		// Simulate NotFound: return nil, nil OR nil, gorm.ErrRecordNotFound
		// Let's test with gorm.ErrRecordNotFound as ReturnReadResponse handles it
		mockRepo.On("GetRepresentative", expectedID).Return(nil, gorm.ErrRecordNotFound).Once()

		handler := representativeController.GetRepresentative()
		err := handler(c) // Calls ReturnReadResponse with ErrRecordNotFound

		assert.Error(t, err) // Expect error from ReturnReadResponse
		var httpErr *echo.HTTPError
		ok := errors.As(err, &httpErr)
		assert.True(t, ok)
		if ok {
			assert.Equal(t, http.StatusNotFound, httpErr.Code)
			expectedMsgStruct := struct {
				Message string `json:"message"`
			}{Message: "Record not found"}
			assert.Equal(t, expectedMsgStruct, httpErr.Message)
		}
		assert.Empty(t, rec.Body.String())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure_NotFound_NilNil", func(t *testing.T) {
		// Test the case where repo returns (nil, nil) which ReturnReadResponse maps to 404 error
		req := httptest.NewRequest(http.MethodGet, "/representatives/404", nil)
		c, rec := setupTest(req)
		c.SetParamNames("representative_id")
		c.SetParamValues("404")

		expectedID := 404
		mockRepo.On("GetRepresentative", expectedID).Return(nil, nil).Once()

		handler := representativeController.GetRepresentative()
		err := handler(c) // Calls ReturnReadResponse with (nil, nil)

		// --- Assertions should now pass with the corrected ReturnReadResponse ---
		assert.Error(t, err) // Expect error because ReturnReadResponse returns HTTPError
		var httpErr *echo.HTTPError
		ok := errors.As(err, &httpErr)
		assert.True(t, ok, "Expected error to be *echo.HTTPError")

		if ok {
			assert.Equal(t, http.StatusNotFound, httpErr.Code) // Expect 404
			expectedMsgStruct := struct {
				Message string `json:"message"`
			}{Message: "Resource not found"} // Expect the specific message
			assert.Equal(t, expectedMsgStruct, httpErr.Message)
		}
		assert.Empty(t, rec.Body.String())

		assert.Empty(t, rec.Body.String(), "Response body should be empty")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure_RepoError", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/representatives/500", nil)
		c, rec := setupTest(req)
		c.SetParamNames("representative_id")
		c.SetParamValues("500")

		expectedID := 500
		repoError := errors.New("db query failed")
		mockRepo.On("GetRepresentative", expectedID).Return(nil, repoError).Once()

		handler := representativeController.GetRepresentative()
		err := handler(c) // Calls ReturnReadResponse with repoError

		assert.Error(t, err) // Expect error from ReturnReadResponse
		var httpErr *echo.HTTPError
		ok := errors.As(err, &httpErr)
		assert.True(t, ok)
		if ok {
			assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
			expectedMsgStruct := struct {
				Message string `json:"message"`
			}{Message: "Internal server error"} // Default from dummy
			assert.Equal(t, expectedMsgStruct, httpErr.Message)
		}
		assert.Empty(t, rec.Body.String())
		mockRepo.AssertExpectations(t)
	})
}
func TestRepresentativeController_GetAllRepresentatives(t *testing.T) {
	mockRepo := new(domain.MockRepresentativeRepo)
	representativeController := controller.RepresentativeController{Repo: mockRepo}

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/representatives", nil)
		c, rec := setupTest(req)

		expectedReps := []domain.Representative{
			{
				Name:        "John",
				Lastname:    "Doe",
				Email:       "j@d.com",
				PhoneNumber: "+1234567890",
			},
			{
				Name:        "Jane",
				Lastname:    "Doe",
				Email:       "j@dd.com",
				PhoneNumber: "+0987654321",
			},
		}
		mockRepo.On("GetAllRepresentatives").Return(expectedReps, nil).Once()

		handler := representativeController.GetAllRepresentatives()
		err := handler(c) // Calls ReturnReadResponse

		assert.NoError(t, err) // Expect nil from e.JSON
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedJSON, _ := json.Marshal(expectedReps)
		assert.JSONEq(t, string(expectedJSON), rec.Body.String())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success_Empty", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/representatives", nil)
		c, rec := setupTest(req)
		expectedReps := make([]domain.Representative, 0)
		mockRepo.On("GetAllRepresentatives").Return(expectedReps, nil).Once()
		handler := representativeController.GetAllRepresentatives()
		err := handler(c) // Calls ReturnReadResponse with the empty slice

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `[]`, rec.Body.String()) // Expect empty JSON array
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success_Empty", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/representatives", nil)
		c, rec := setupTest(req)
		expectedReps := make([]domain.Representative, 0)
		mockRepo.On("GetAllRepresentatives").Return(expectedReps, nil).Once()

		handler := representativeController.GetAllRepresentatives()
		err := handler(c)      // Calls ReturnReadResponse
		assert.NoError(t, err) // Expect nil from e.JSON
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, `[]`, strings.TrimSpace(rec.Body.String()), "Response body should be an empty JSON array")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure_RepoError", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/representatives", nil)
		c, rec := setupTest(req)

		repoError := errors.New("db connection failed")
		mockRepo.On("GetAllRepresentatives").Return(nil, repoError).Once()

		handler := representativeController.GetAllRepresentatives()
		err := handler(c) // Calls ReturnReadResponse with repoError

		assert.Error(t, err) // Expect error from ReturnReadResponse
		var httpErr *echo.HTTPError
		ok := errors.As(err, &httpErr)
		assert.True(t, ok)
		if ok {
			assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
			expectedMsgStruct := struct {
				Message string `json:"message"`
			}{Message: "Internal server error"} // Default from dummy
			assert.Equal(t, expectedMsgStruct, httpErr.Message)
		}
		assert.Empty(t, rec.Body.String())
		mockRepo.AssertExpectations(t)
	})
}

func TestRepresentativeController_UpdateRepresentative(t *testing.T) {
	mockRepo := new(domain.MockRepresentativeRepo)
	representativeController := controller.RepresentativeController{Repo: mockRepo}

	t.Run("Success", func(t *testing.T) {
		repInput := domain.RepresentativeInput{
			Name:        "John",
			Lastname:    "Doe",
			Email:       "j@a.com",
			PhoneNumber: "+1234567890",
		}
		repInputJSON, _ := json.Marshal(repInput)
		req := httptest.NewRequest(http.MethodPut, "/representatives/123", strings.NewReader(string(repInputJSON))) // Or PATCH
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c, rec := setupTest(req)
		c.SetParamNames("representative_id")
		c.SetParamValues("123")

		expectedID := 123
		mockRepo.On("UpdateRepresentative", expectedID, repInput).Return(nil).Once()

		handler := representativeController.UpdateRepresentative()
		err := handler(c) // Calls ReturnWriteResponse

		assert.NoError(t, err) // Expect nil from ReturnWriteResponse
		assert.Equal(t, http.StatusOK, rec.Code)
		expectedResp := `{"Body":{"message":"Representative updated"}}`
		assert.JSONEq(t, expectedResp, rec.Body.String())
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure_InvalidID", func(t *testing.T) {
		repInput := domain.RepresentativeInput{
			Name:        "John",
			Lastname:    "Doe",
			Email:       "j@a.com",
			PhoneNumber: "+1234567890",
		}
		repInputJSON, _ := json.Marshal(repInput)
		req := httptest.NewRequest(http.MethodPut, "/representatives/abc", strings.NewReader(string(repInputJSON)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c, rec := setupTest(req)
		c.SetParamNames("representative_id")
		c.SetParamValues("abc")

		handler := representativeController.UpdateRepresentative()
		err := handler(c)

		assert.NoError(t, err)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		expectedResp := `{"message":"strconv.ParseInt: parsing \"abc\": invalid syntax"}`
		assert.JSONEq(t, expectedResp, rec.Body.String())

		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure_BindingError", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/representatives/123", strings.NewReader(`{"invalid`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c, rec := setupTest(req)
		c.SetParamNames("representative_id")
		c.SetParamValues("123")

		handler := representativeController.UpdateRepresentative()
		err := handler(c) // ValidateAndBind fails

		assert.Error(t, err) // Expect error from ValidateAndBind
		var httpErr *echo.HTTPError
		ok := errors.As(err, &httpErr)
		assert.True(t, ok)
		if ok {
			assert.Equal(t, http.StatusBadRequest, httpErr.Code)
			assert.Equal(t, "Invalid request body", httpErr.Message) // From dummy ValidateAndBind
		}
		assert.Empty(t, rec.Body.String())
		mockRepo.AssertExpectations(t) // Repo not called
	})

	t.Run("Failure_RepoError", func(t *testing.T) {
		repInput := domain.RepresentativeInput{
			Name:        "John",
			Lastname:    "Doe",
			Email:       "j@a.com",
			PhoneNumber: "+1234567890",
		}
		repInputJSON, _ := json.Marshal(repInput)
		req := httptest.NewRequest(http.MethodPut, "/representatives/123", strings.NewReader(string(repInputJSON)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c, rec := setupTest(req)
		c.SetParamNames("representative_id")
		c.SetParamValues("123")
		expectedID := 123
		repoError := errors.New("update failed")
		mockRepo.On("UpdateRepresentative", expectedID, repInput).Return(repoError).Once()
		handler := representativeController.UpdateRepresentative()
		err := handler(c) // Calls ReturnWriteResponse with repoError
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code) // Adjust if your helper uses a different code
		expectedResp := `{"message":"update failed"}`
		assert.JSONEq(t, expectedResp, rec.Body.String())
		mockRepo.AssertExpectations(t)
	})

}
