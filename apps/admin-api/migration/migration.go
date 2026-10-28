package migration

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"runtime"

	"github.com/illuminarean/energy-shares-backend-repo/pkg/config"

	database "github.com/illuminarean/energy-shares-backend-repo/db"
)

func MigrateDatabase(env string) {

	var c config.Config

	if err := c.Read(env); err != nil {
		logrus.Error(errors.Wrap(err, "fail read config"))
		panic(err)
	}

	c.Env = env
	if err := dbMigration(c); err != nil { // db migration using goose
		logrus.Error(errors.Wrap(err, "database migration error: fail db migration"))
		os.Exit(1)
	}
}
func dbMigration(cfg config.Config) error {
	//env 플래그가 local(도커 사용)이 아니면서 로컬 환경(MacOS)이면 마이그레이션을 실행하지 않는다.
	if cfg.Env != "local" && runtime.GOOS == "darwin" {
		log.Printf("env: %s, runtime.GOOS: %s. do not run db migrations.\n", cfg.Env, runtime.GOOS)
		return nil // skip
	}

	fmt.Println(cfg.Database.Driver)

	// 데이터베이스 연결 설정
	db, err := sql.Open(cfg.Database.Driver, cfg.Database.Connection)
	if err != nil {
		return err
	}
	defer db.Close()

	// goose에 embed.FS 사용 설정
	goose.SetDialect(cfg.Database.Driver)
	goose.SetBaseFS(database.MigrationsFS)

	log.Println("Migration status:")
	if err := goose.Status(db, "migrations"); err != nil {
		return err
	}

	log.Println("Migrating...")
	// 마이그레이션 실행
	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}

	log.Println("Migration status:")
	if err := goose.Status(db, "migrations"); err != nil {
		return err
	}
	log.Println("Migration successful")
	return nil
}
