//go:build integration

package db

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	db_name string
	url     string
	db      *pgxpool.Pool
)

func init() {
	_, debug := os.LookupEnv("SLOG_DEBUG")
	if debug {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))
	}
	db_name = os.Getenv("DB_TEST_NAME")
	if db_name == "" {
		slog.Error("test require DB_TEST_NAME enviroment variable to be set")
		os.Exit(1)
	}
	url = os.Getenv("DB_TEST_URL")
	if url == "" {
		slog.Error("test require DB_TEST_URL enviroment variable to be set")
		os.Exit(1)
	}

	var err error
	db, err = pgxpool.New(context.Background(), url)
	if err != nil {
		slog.Error("failed to connect to default test database", slog.String("DB_TEST_NAME", db_name), slog.String("DB_TEST_URL", url), slog.Any("err", err))
		os.Exit(1)
	}
}

func testDBName(tb testing.TB) string {
	tb.Helper()
	return strings.ToLower(db_name + "_" + tb.Name())
}

func testDBUrl(tb testing.TB) string {
	tb.Helper()
	return strings.Replace(url, db_name, testDBName(tb), 1)
}

func dropTestDB(tb testing.TB, db *pgxpool.Pool) {
	tb.Helper()

	slog.Debug("dropping test db", slog.String("db", testDBName(tb)))
	sql := "DROP DATABASE IF EXISTS " + testDBName(tb)
	_, err := db.Exec(context.Background(), sql)
	if err != nil {
		tb.Fatal(err)
	}
}

func createTestDB(tb testing.TB, db *pgxpool.Pool) {
	tb.Helper()

	slog.Debug("creating test db", slog.String("db", testDBName(tb)))
	sql := "CREATE DATABASE " + testDBName(tb)
	_, err := db.Exec(context.Background(), sql)
	if err != nil {
		tb.Fatal(err)
	}
}

func prepareTestDB(tb testing.TB) {
	tb.Helper()

	dropTestDB(tb, db)
	createTestDB(tb, db)
}

func getDB(tb testing.TB) *pgxpool.Pool {
	tb.Helper()
	slog.Debug("new db connection", slog.String("url", testDBUrl(tb)))
	pool, err := pgxpool.New(context.Background(), testDBUrl(tb))
	if err != nil {
		tb.Fatal(err)
	}
	tb.Cleanup(func() {
		slog.Debug("closing db connection", slog.String("url", testDBUrl(tb)))
		pool.Close()
	})
	return pool
}

func getMigrate(tb testing.TB) *migrate.Migrate {
	tb.Helper()
	migrationSource := "file://../../db/migrations"
	slog.Debug("new migrate instance", "url", testDBUrl(tb))
	m, err := migrate.New(migrationSource, testDBUrl(tb))
	if err != nil {
		tb.Fatal(err)
	}
	tb.Cleanup(func() {
		sErr, dErr := m.Close()
		if sErr != nil || dErr != nil {
			tb.Fatal(sErr, " or ", dErr)
		}
		slog.Debug("migrate instance closed", slog.String("test", tb.Name()))
	})
	slog.Debug("new migrate instance acquired", slog.String("test", tb.Name()))
	return m
}

func migrateUp(tb testing.TB) DBTX {
	tb.Helper()
	prepareTestDB(tb)
	m := getMigrate(tb)

	err := m.Up()
	if err != nil {
		tb.Fatal(err)
	}

	slog.Debug("migrate up done", slog.String("test", tb.Name()))
	return getDB(tb)
}

func randomData(tb testing.TB) []byte {
	tb.Helper()
	fake, err := gofakeit.JSON(&gofakeit.JSONOptions{
		Type: "object",
		Fields: []gofakeit.Field{
			{Name: "title", Function: "phrase"},
			{Name: "duration", Function: "uint8"},
		},
		Indent: false,
	})
	if err != nil {
		slog.Error("random data generation failed", slog.Any("err", err))
		tb.FailNow()
	}
	str := strings.ReplaceAll(string(fake), ":", ": ")
	str = strings.ReplaceAll(str, ",", ", ")
	return []byte(str)
}
