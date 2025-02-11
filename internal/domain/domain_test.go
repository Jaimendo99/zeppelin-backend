package domain_test

import (
	"testing"
	"zeppelin/internal/domain"
)

func TestRepresentativeGetTableName(t *testing.T) {
	name := domain.RepresentativeDb{}.TableName()
	if name != "representatives" {
		t.Errorf("Expected representatives, got %s", name)
	}
}
