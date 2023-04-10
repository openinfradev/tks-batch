package main

/*
import (
	"fmt"
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/openinfradev/tks-common/pkg/helper"
	"github.com/openinfradev/tks-common/pkg/log"

	"github.com/openinfradev/tks-batch/internal/application"
	"github.com/openinfradev/tks-batch/internal/cluster"
)

var (
	db *gorm.DB
)

func init() {
	log.Disable()
}

func TestMain(m *testing.M) {
	pool, resource, err := helper.CreatePostgres()
	if err != nil {
		fmt.Printf("Could not create postgres: %s", err)
		os.Exit(-1)
	}
	testDBHost, testDBPort := helper.GetHostAndPort(resource)

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Seoul",
		testDBHost, "postgres", "password", "tks", testDBPort)
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		os.Exit(-1)
	}

	db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`)

	if err := db.AutoMigrate(&application.ApplicationGroup{}); err != nil {
		os.Exit(-1)
	}
	if err := db.AutoMigrate(&cluster.Cluster{}); err != nil {
		os.Exit(-1)
	}

	clusterAccessor = cluster.New(db)
	applicationAccessor = application.New(db)

	code := m.Run()

	if err := helper.RemovePostgres(pool, resource); err != nil {
		fmt.Printf("Could not remove postgres: %s", err)
		os.Exit(-1)
	}

	os.Exit(code)
}
*/
