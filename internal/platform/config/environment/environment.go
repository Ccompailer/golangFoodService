package environment

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"

	"golangFoodService/internal/platform/constants"
)

type Environment string

var (
	Development = Environment(constants.Dev)
	Production  = Environment(constants.Production)
	Test        = Environment(constants.Test)
)

func ConfigureAppEnv(environments ...Environment) Environment {
	environment := Development
	if len(environments) > 0 {
		environment = environments[0]
	}

	viper.AutomaticEnv()

	if err := loadEnvFilesRecursive(); err != nil {
		log.Printf("Error loading env files: %v", err)
	}

	setRootWorkingDirectory()
	FixProjectRootWorkDirectoryPath()

	if manualEnv := os.Getenv(constants.AppEnv); manualEnv != "" {
		environment = Environment(manualEnv)
	}

	return environment
}

func IsDevelopment(env Environment) bool { return env == Development }
func IsTest(env Environment) bool        { return env == Test }
func IsProduction(env Environment) bool  { return env == Production }
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
		if err := godotenv.Load(envFilePath); err == nil {
			return nil
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			break
		}
		dir = parentDir
	}

	return errors.New("unable to load environment files in this directory hierarchy")
}

func setRootWorkingDirectory() {
	viper.Set(constants.AppRootPath, GetProjectRootDirectory())
}
