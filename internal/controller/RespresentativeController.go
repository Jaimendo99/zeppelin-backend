package controller

import (
	"zeppelin/internal/domain"
	"zeppelin/internal/services"

	"github.com/labstack/echo/v4"
)

type RepresentativeController struct {
	Repo domain.RepresentativeRepo
}

func (c *RepresentativeController) GetRepresentative() echo.HandlerFunc {
	return func(e echo.Context) error {
		representativeId := e.Param("representative_id")
		id, err := services.ParamToId(representativeId)
		if err != nil {
			return ReturnReadResponse(e, err, nil)
		}
		var representative *domain.Representative
		representative, err = c.Repo.GetRepresentative(id)
		return ReturnReadResponse(e, err, representative)
	}
}

func (c *RepresentativeController) GetAllRepresentatives() echo.HandlerFunc {
	return func(e echo.Context) error {
		representatives, err := c.Repo.GetAllRepresentatives()
		return ReturnReadResponse(e, err, representatives)
	}
}

func (c *RepresentativeController) UpdateRepresentative() echo.HandlerFunc {
	return func(e echo.Context) error {
		representativeId := e.Param("representative_id")
		id, err := services.ParamToId(representativeId)
		if err != nil {
			return ReturnWriteResponse(e, err, nil)
		}
		representative := domain.RepresentativeInput{}
		if err := ValidateAndBind(e, &representative); err != nil {
			return err
		}
		err = c.Repo.UpdateRepresentative(id, representative)
		return ReturnWriteResponse(e, err, struct {
			Message string `json:"message"`
		}{Message: "Representative updated"})
	}
}
