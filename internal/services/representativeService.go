package services

import (
	"database/sql"
	"strconv"
	"zeppelin/internal/domain"
)

func RepresetativeInputToDb(representative *domain.RepresentativeInput) domain.RepresentativeDb {
	return domain.RepresentativeDb{
		Name:     representative.Name,
		Lastname: representative.Lastname,
		Email: sql.NullString{
			String: representative.Email,
			Valid:  representative.Email != "",
		},
		PhoneNumber: sql.NullString{
			String: representative.PhoneNumber,
			Valid:  representative.PhoneNumber != "",
		},
	}
}

func ParamToId(representativeId string) (int, error) {
	id, err := strconv.ParseInt(representativeId, 10, 10)
	if err != nil {
		return -1, err
	}
	return int(id), nil
}
