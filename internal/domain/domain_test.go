package domain_test

import (
	"testing"
	"zeppelin/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestRepresentativeGetTableName(t *testing.T) {
	name := domain.RepresentativeDb{}.TableName()
	assert.Equal(t, "representatives", name)
}

func TestRepresentativeInputGetTableName(t *testing.T) {
	name := domain.RepresentativeInput{}.TableName()
	assert.Equal(t, "representatives", name)
}
