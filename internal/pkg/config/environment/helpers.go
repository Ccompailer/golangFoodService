package environment

import (
	"golangFoodService/internal/pkg/constants"
	"log"
	"os"
	"path/filepath"
	"strings"

	"emperror.dev/errors"

	"github.com/spf13/viper"
)

func GetProjectRootDirectory() string {
	var rootWorkingDirectory string

	pn := viper.GetString(constants.ProjectNameEnv)
	if pn != "" {
		rootWorkingDirectory = getProjectRootDirectoryFromProjectName(pn)
	} else {
		wd, _ := os.Getwd()
		dir, err := searchRootDirectory(wd)
		if err != nil {
			log.Fatal(err)
		}

		rootWorkingDirectory = dir
	}

	absolutRootDirectory, _ := filepath.Abs(rootWorkingDirectory)
	return absolutRootDirectory
}

func FixProjectRootWorkDirectoryPath() {
	currWd, _ := os.Getwd()
	log.Printf("Current working directory is: %s", currWd)

	rootDir := GetProjectRootDirectory()
	_ = os.Chdir(rootDir)
	newWd, _ := os.Getwd()
	log.Printf("New fixed working directory is: %s", newWd)
}

func getProjectRootDirectoryFromProjectName(pn string) string {
	wd, _ := os.Getwd()

	for !strings.HasSuffix(wd, pn) {
		wd = filepath.Dir(wd)
	}

	return wd
}

func searchRootDirectory(dir string) (string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", errors.WrapIf(err, "Error reading directory")
	}

	for _, file := range files {
		if !file.IsDir() {
			fileName := file.Name()
			if strings.EqualFold(fileName, "go.mod") {
				return dir, nil
			}
		}
	}

	parentDir := filepath.Dir(dir)
	if parentDir == dir {
		return "", errors.WrapIf(err, "No go.mod file found")
	}

	return searchRootDirectory(parentDir)
}
