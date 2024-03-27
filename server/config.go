package server

import (
	"fmt"
	"os"
)

type Config struct {
	SocketPath         string
	DBType             string
	DBAddress          string
	DbName             string
	DbUserName         string
	DbPassword         string
	DBConnectionString string
}

func Configure() *Config {

	socketPath := os.Getenv("SOCKET_PATH")
	dbType := os.Getenv("DB_TYPE")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432" // Default port
	}
	dbAddress := os.Getenv("DB_ADDRESS")
	dbUserName := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")

	if socketPath == "" || dbName == "" || dbUserName == "" || dbPassword == "" || dbType == "" || dbAddress == "" {
		fmt.Println("SOCKET_PATH, DB_NAME, DB_USERNAME, DB_PASSWORD, DB_TYPE and DB_ADDRESS environment variables are required")
		return nil
	}
	dbConnectionString := ""
	switch dbType {
	case "postgres":
		dbConnectionString = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			dbAddress, dbPort, dbUserName, dbPassword, dbName)
	case "mysql":
		dbConnectionString = fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbUserName, dbPassword, dbAddress, dbName)
	case "sqlite":
		dbConnectionString = dbName
	default:
		fmt.Println("Invalid database type")
		return nil
	}
	return &Config{
		SocketPath:         socketPath,
		DBType:             dbType,
		DBConnectionString: dbConnectionString,
	}
}
