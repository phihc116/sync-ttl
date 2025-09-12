package migrations

import (
	"github.com/phihc116/sync-ttl/internals/infrastructure"
	"github.com/phihc116/sync-ttl/internals/models"
)

func RunMigrations() error {
	db := infrastructure.GetSQLClient()

	return db.AutoMigrate(
		&models.User{},
	)
}
