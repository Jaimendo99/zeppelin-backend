package services

import (
	"fmt"
	"strconv"
	"zeppelin/internal/domain"
)

type RepresentativeService struct {
	Repo domain.RepresentativeRepo
}

func NewRepresentativeService(repo domain.RepresentativeRepo) *RepresentativeService {
	return &RepresentativeService{Repo: repo}
}

func (s *RepresentativeService) CreateRepresentative(representative domain.RepresentativeInput) error {
	representativeDb := domain.RepresentativeDb{
		Name:     representative.Name,
		Lastname: representative.Lastname,
		Email:    &representative.Email,
		Phone:    &representative.Phone,
	}

	err := s.Repo.CreateRepresentative(representativeDb)
	if err != nil {
		return err
	}
	return nil
}

func (s *RepresentativeService) GetRepresentative(representativeId string) (domain.RepresentativeDb, error) {
	id, err := strconv.ParseInt(representativeId, 10, 10)
	if err != nil {
		return domain.RepresentativeDb{}, err
	}
	representative, err := s.Repo.GetRepresentative(int(id))
	if err != nil {
		return domain.RepresentativeDb{}, err
	}
	return representative, nil
}

func (s *RepresentativeService) GetAllRepresentatives() ([]domain.Representative, error) {
	representatives, err := s.Repo.GetAllRepresentatives()
	if err != nil {
		return nil, err
	}
	return representatives, nil
}

func (s *RepresentativeService) UpdateRepresentative(representativeId string, representative domain.RepresentativeInput) error {
	id, err := strconv.ParseInt(representativeId, 10, 10)
	if err != nil {
		return err
	}
	fmt.Println(id)
	err = s.Repo.UpdateRepresentative(int(id), representative)
	if err != nil {
		fmt.Printf("REPRE_SERVICE: %v \n", representative)
		fmt.Printf("REPRE_SERVICE: error: %v", err)
		return err
	}
	return nil
}
