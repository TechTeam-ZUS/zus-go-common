package postgres_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TechTeam-ZUS/zus-go-common/postgres"
)

// Init's success path requires a live PostgreSQL server and isn't covered
// here; only the config-validation error path is testable in isolation.
func TestInit_RequiresDatabaseName(t *testing.T) {
	tests := []struct {
		name        string
		database    string
		expectedErr bool
	}{
		{name: "missing database name errors", database: "", expectedErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("POSTGRES_DATABASE", tt.database)
			_, err := postgres.Init()
			if tt.expectedErr {
				assert.Error(t, err)
			}
		})
	}
}
