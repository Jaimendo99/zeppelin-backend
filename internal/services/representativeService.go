package services

import (
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

func (s *RepresentativeService) GetRepresentative(representative_id string) (domain.RepresentativeDb, error) {
	id, err := strconv.ParseInt(representative_id, 10, 10)
	if err != nil {
		return domain.RepresentativeDb{}, err
	}
	representative, err := s.Repo.GetRepresentative(int(id))
	if err != nil {
		return domain.RepresentativeDb{}, err
	}
	return representative, nil
}
