package services_test

import (
	"testing"
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

func (r RepresentativeRepoMock) GetAllRepresentatives() ([]domain.Representative, error) {
	return nil, nil
}

func (r RepresentativeRepoMock) UpdateRepresentative(id int, representative domain.RepresentativeInput) error {
	return nil
}

func TestRepresentativeService(t *testing.T) {

	service := services.NewRepresentativeService(RepresentativeRepoMock{})

	err := service.CreateRepresentative(domain.RepresentativeInput{})
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}

	_, err = service.GetRepresentative("1")
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}

}

func TestRepresentativeServiceCreateRepresentative(t *testing.T) {
	service := services.NewRepresentativeService(RepresentativeRepoMock{})

	err := service.CreateRepresentative(domain.RepresentativeInput{})
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}

func TestRepresentativeServiceGetRepresentative(t *testing.T) {
	service := services.NewRepresentativeService(RepresentativeRepoMock{})

	_, err := service.GetRepresentative("1")
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}

func TestRepresentativeServiceGetAllRepresentatives(t *testing.T) {
	service := services.NewRepresentativeService(RepresentativeRepoMock{})

	_, err := service.GetAllRepresentatives()
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}
