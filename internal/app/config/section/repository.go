package section

import (
	"fmt"
	"time"
)

type (
	Repository struct {
		Postgres RepositoryPostgres
		// Service1 RepositoryService1
		// Service2 RepositoryService2
	}

	RepositoryPostgres struct {
		Host         string        `envconfig:"APP_REPOSITORY_POSTGRES_HOST"`
		Port         string        `envconfig:"APP_REPOSITORY_POSTGRES_PORT"`
		Username     string        `envconfig:"APP_REPOSITORY_POSTGRES_USERNAME"`
		Password     string        `envconfig:"APP_REPOSITORY_POSTGRES_PASSWORD"`
		Name         string        `envconfig:"APP_REPOSITORY_POSTGRES_DB"`
		ReadTimeout  time.Duration `envconfig:"APP_REPOSITORY_POSTGRES_READ_TIMEOUT"      default:"30s"`
		WriteTimeout time.Duration `envconfig:"APP_REPOSITORY_POSTGRES_WRITE_TIMEOUT"     default:"30s"`
	}
)

func (p *RepositoryPostgres) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		p.Host,
		p.Port,
		p.Username,
		p.Password,
		p.Name,
	)
}
