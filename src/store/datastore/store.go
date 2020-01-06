package datastore

import (
	"os"

	"github.com/coderun-top/coderun/src/store"
	"github.com/coderun-top/coderun/src/model"

	"github.com/sirupsen/logrus"
	"github.com/jinzhu/gorm"
)

var db *gorm.DB

// datastore is an implementation of a model.Store built on top
// of the sql/database driver with a relational database backend.
type datastore struct {
	*gorm.DB

	driver string
	config string
}

// var db *gorm.DB

// New creates a database connection for the given driver and datasource
// and returns a new Store.
func New(driver, config string) store.Store {
	return &datastore{
		DB:     open(driver, config),
		driver: driver,
		config: config,
	}
}

// open opens a new database connection with the specified
// driver and connection string and returns a store.
func open(driver, config string) *gorm.DB {
	db,err := gorm.Open("mysql", os.Getenv("DRONE_DATABASE_DATASOURCE"))
		if err != nil {
			logrus.Errorln(err)
			logrus.Fatalln("database ping attempts failed")
		}
		logrus.Infof("database connected")
		if err := setupDatabase(driver, db); err != nil {
			logrus.Errorln(err)
			logrus.Fatalln("migration failed")
		}
	return db
}

// automated database migration steps.
func setupDatabase(driver string, db *gorm.DB) error {
	db.Set("gorm:table_options", "charset=utf8")
	return db.AutoMigrate(
			&model.Build{},
			&model.Config{},
			&model.File{},
			&model.K8sCluster{},
			&model.K8sDeploy{},
			&model.LogData{},
			&model.Migration{},
			&model.Perm{},
			&model.PipelineEnv{},
			&model.Proc{},
			&model.Star{},
			&model.ProjectEnv{},
			&model.Project{},
			&model.Helm{},
			&model.Registry{},
			&model.Repo{},
			&model.Secret{},
			&model.Sender{},
			&model.Task{},
			&model.User{},
			&model.WebHook{},
			&model.WebHookProject{},
			&model.Email{},
			&model.EmailProject{},
			&model.AgentKey{},
			&model.Agent{},
			).Error
}
