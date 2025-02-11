package controller

import (
	"net/http"
	"zeppelin/internal/domain"

	"github.com/labstack/echo/v4"
)

type RepresentativeController struct {
	Service domain.RepresentativeServiceI
}

func NewRepresentativeController(service domain.RepresentativeServiceI) *RepresentativeController {
	return &RepresentativeController{Service: service}
}

func (c *RepresentativeController) CreateRepresentative() echo.HandlerFunc {
	return func(e echo.Context) error {
		representative := domain.RepresentativeInput{}
		if err := e.Bind(&representative); err != nil {
			return e.JSON(http.StatusBadRequest, struct{ Message string }{Message: "Invalid request"})
		}
		err := c.Service.CreateRepresentative(representative)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, struct{ Message string }{Message: "Internal server error"})
		}
		return e.JSON(http.StatusCreated, struct{ Message string }{Message: "Representative created"})
	}
}

func (c *RepresentativeController) GetRepresentative() echo.HandlerFunc {
	return func(e echo.Context) error {
		representativeId := e.Param("representative_id")
		representative, err := c.Service.GetRepresentative(representativeId)
		if err != nil {
			if err.Error() == "record not found" {
				return e.JSON(http.StatusOK, domain.RepresentativeDb{})
			}
			return e.JSON(http.StatusInternalServerError, struct{ Message string }{Message: "Internal server error"})
		}
		return e.JSON(http.StatusOK, representative)
	}
}

func (c *RepresentativeController) GetAllRepresentatives() echo.HandlerFunc {
	return func(e echo.Context) error {
		representatives, err := c.Service.GetAllRepresentatives()
		if err != nil {
			e.Logger().Error(err)
			return e.JSON(http.StatusInternalServerError, struct{ Message string }{Message: "Internal server error"})
		}
		return e.JSON(http.StatusOK, representatives)
	}
}

func (c *RepresentativeController) UpdateRepresentative() echo.HandlerFunc {
	return func(e echo.Context) error {
		representativeId := e.Param("representative_id")
		e.Logger().Error(representativeId)
		representative := domain.RepresentativeInput{}
		e.Logger().Error(representative)
		if err := e.Bind(&representative); err != nil {
			return e.JSON(http.StatusBadRequest, struct{ Message string }{Message: "Invalid request"})
		}
		e.Logger().Error(representative)
		err := c.Service.UpdateRepresentative(representativeId, representative)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, struct{ Message string }{Message: "Internal server error"})
		}
		return e.JSON(http.StatusOK, struct{ Message string }{Message: "Representative updated"})
	}
}
