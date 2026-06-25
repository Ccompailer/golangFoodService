package environment

import (
	"errors"
	"golangFoodService/internal/pkg/constants"
	"log"
	"os"
	"path/filepath"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/spf13/viper"
)

type Environment string

var (
	Development = Environment(constants.Dev)
	Production  = Environment(constants.Production)
	Test        = Environment(constants.Test)
)

func ConfigureAppEnv(environments ...Environment) Environment {
	environment := Environment("")

	if len(environments) > 0 {
		environment = environments[0]
	} else {
		environment = Development
	}

	viper.AutomaticEnv()

	err := loadEnvFilesRecursive()
	if err != nil {
		log.Printf("Error loading env files: %v", err)
	}

	setRootWorkingDirectory()

	FixProjectRootWorkDirectoryPath()

	manualEnv := os.Getenv(constants.AppEnv)
	if manualEnv != "" {
		environment = Environment(manualEnv)
	}

	return environment
}

func IsDevelopment(env Environment) bool {
	return env == Development
}

func IsTest(env Environment) bool {
	return env == Test
}

func IsProduction(env Environment) bool {
	return env == Production
}

func GetEnvironment(env Environment) string {
	return string(env)
}

func EnvString(key string, fallback string) string {
	if value, ok := syscall.Getenv(key); ok {
		return value
	}

	return fallback
}

func loadEnvFilesRecursive() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	for {
		envFilePath := filepath.Join(dir, ".env")
		err := godotenv.Load(envFilePath)

		if err == nil {
			return nil
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			break
		}

		dir = parentDir
	}

	return errors.New(
		"unable to load environment files in this directory hierarchy")
}

func setRootWorkingDirectory() {
	absoluteRootWorkingDirectory := GetProjectRootDirectory()
	viper.Set(constants.AppRootPath, absoluteRootWorkingDirectory)
}
