//go:build integration

// Integration tests run against a real PostgreSQL started in Docker by
// testcontainers. They are excluded from the normal `go test ./...`
// by the build tag above. Run them with: make test-integration

package postgres

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

// testPool is shared by all tests in this package.
var testPool *pgxpool.Pool

// TestMain runs once before all tests of the package: it starts a
// throwaway PostgreSQL container, applies migrations, runs the tests
// and removes the container.
func TestMain(m *testing.M) {
	os.Exit(runWithDatabase(m))
}

func runWithDatabase(m *testing.M) int {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx, "postgres:17-alpine",
		tcpostgres.WithDatabase("goledger_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.BasicWaitStrategies(), // wait until postgres accepts connections
	)
	if err != nil {
		log.Printf("start postgres container: %v", err)
		return 1
	}
	defer func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			log.Printf("terminate container: %v", err)
		}
	}()

	url, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Printf("connection string: %v", err)
		return 1
	}

	pool, err := Connect(ctx, url)
	if err != nil {
		log.Printf("connect: %v", err)
		return 1
	}
	defer pool.Close()

	if err := Migrate(ctx, pool); err != nil {
		log.Printf("migrate: %v", err)
		return 1
	}

	testPool = pool
	return m.Run()
}

// newUserRepo empties the users table and returns a fresh repository,
// so every test starts from a clean database.
func newUserRepo(t *testing.T) *UserRepo {
	t.Helper()
	if _, err := testPool.Exec(context.Background(), "TRUNCATE users"); err != nil {
		t.Fatalf("truncate users: %v", err)
	}
	return NewUserRepo(testPool)
}
