package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DB_Port     string
	DB_User     string
	DB_Password string
	DB_Host     string
	DB_Name     string
	Storage     string
}

func MustloadConfig() (Config, error) {

	err := godotenv.Load("/home/user/work/testwork/test1/posts/.env")
	if err != nil {
		log.Fatal("Error Load .env file", err)
	}

	Cfg := Config{}

	Cfg.DB_Port = os.Getenv("PORT")
	if Cfg.DB_Port == "" {
		Cfg.DB_Port = "8080"
	}
	Cfg.Storage = os.Getenv("STORAGE")
	if Cfg.Storage == "IN_MEMORY" {
		return Cfg, nil
	}
	Cfg.DB_Host = os.Getenv("DB_HOST")
	if Cfg.DB_Host == "" {
		return Cfg, fmt.Errorf("%s", "Error load Db_Host env")
	}
	Cfg.DB_Password = os.Getenv("DB_PASSWORD")
	if Cfg.DB_Password == "" {
		return Cfg, fmt.Errorf("%s", "Error load Db_Password env")
	}
	Cfg.DB_User = os.Getenv("DB_USER")
	if Cfg.DB_User == "" {
		return Cfg, fmt.Errorf("%s", "Error load Db_User env")
	}

	Cfg.DB_Name = os.Getenv("DB_NAME")
	if Cfg.DB_Name == "" {
		return Cfg, fmt.Errorf("%s", "Error load Db_Name env")
	}

	return Cfg, nil

}
