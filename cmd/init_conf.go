package cmd

import (
	"os"

	"github.com/genigo/genigo/internal/config"
	"github.com/genigo/goje"
	"gopkg.in/yaml.v3"
)

// init config file for the requested driver (mysql | postgres)
func initConfig(path string, driver string) {
	var conf config.Config

	conf.Tags = []string{"db", "json"}
	conf.Pkg = "models"
	conf.Dir = "./models"
	conf.Replace = true

	port := 3306
	if driver == "postgres" {
		port = 5432
	}

	conf.DB = goje.DBConfig{
		Driver: driver,
		Host:   "127.0.0.1",
		Port:   port,
		User:   "root",
		Schema: "database",
	}

	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	err = yaml.NewEncoder(file).Encode(conf)
	if err != nil {
		panic(err)
	}
}
