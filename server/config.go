package server

import (
	"fmt"
	"os"
	"strings"
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
	loadEnvFromFile()
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

func loadEnvFromFile() {
	if len(os.Args) < 2 {
		fmt.Println("ENV_FILE environment variable is required")
		return
	}
	filePath := os.Args[1]
	if filePath == "" {
		fmt.Println("ENV_FILE environment variable is required")
		return
	}
	text, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Failed to read the file")
		return
	}
	lines := strings.Split(string(text), "\n")
	for _, line := range lines {
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			continue
		}
		if strings.HasPrefix(parts[0], "#") {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		os.Setenv(key, value)
	}
}
