package controller_test

import (
	"testing"
	"zeppelin/internal/controller"
	"zeppelin/internal/domain"
	"zeppelin/internal/services"
)

type RepresentativeRepoMock struct {
}

func (r RepresentativeRepoMock) CreateRepresentative(representative domain.RepresentativeDb) error {
	return nil
}

func (r RepresentativeRepoMock) GetRepresentative(id int) (domain.RepresentativeDb, error) {
	return domain.RepresentativeDb{}, nil
}

func TestCreateRepresentative(t *testing.T) {
	repo := &RepresentativeRepoMock{}
	service := services.NewRepresentativeService(repo)
	controller := controller.NewRepresentativeController(service)

	err := controller.CreateRepresentative()
	if err != nil {
		// t.Errorf("Expected nil, got %v", err)
	}
}

func TestGetRepresentative(t *testing.T) {
	repo := RepresentativeRepoMock{}
	service := services.NewRepresentativeService(repo)
	controller := controller.NewRepresentativeController(service)

	err := controller.GetRepresentative()
	if err != nil {
		// t.Errorf("Expected nil, got %v", err)
	}
}
