package demo_test

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var pool *dockertest.Pool

func TestMain(m *testing.M) {
	var err error
	pool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
		os.Exit(1)
	}
	code := m.Run()
	os.Exit(code)
}

func db(t *testing.T) *sql.DB {
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "12.3",
		Env:        []string{"POSTGRES_PASSWORD=secret", "POSTGRES_DB=testdb"},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		t.Fatalf("Could not start resource: %s", err)
	}

	pgxURL := fmt.Sprintf("postgres://postgres:secret@localhost:%s/%s?sslmode=disable", resource.GetPort("5432/tcp"), "testdb")
	var db *sql.DB
	db, err = sql.Open("pgx", pgxURL)
	if err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}

	t.Cleanup(func() {
		err := pool.Purge(resource)
		if err != nil {
			t.Logf("Could not purge resource: %s", err)
		}
	})

	return db
}

func TestDemo(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			db := db(t)
			db.Ping()
			time.Sleep(2 * time.Second)
		})
	}
}
