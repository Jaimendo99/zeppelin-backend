package services

import (
	"strconv"
	"zeppelin/internal/domain"
)

func RepresentativesInputToDb(representative *domain.RepresentativeInput) domain.RepresentativeDb {
	return domain.RepresentativeDb{
		Name:        representative.Name,
		Lastname:    representative.Lastname,
		Email:       representative.Email,
		PhoneNumber: representative.PhoneNumber,
	}
}

func ParamToId(representativeId string) (int, error) {
	id, err := strconv.ParseInt(representativeId, 10, 10)
	if err != nil {
		return -1, err
	}
	return int(id), nil
}
