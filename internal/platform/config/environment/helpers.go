package environment

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"emperror.dev/errors"
	"github.com/spf13/viper"

	"golangFoodService/internal/platform/constants"
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
			log.Printf("go.mod not found from %s, using working directory", wd)
			return wd
		}
		rootWorkingDirectory = dir
	}

	absoluteRootDirectory, _ := filepath.Abs(rootWorkingDirectory)
	return absoluteRootDirectory
}

func FixProjectRootWorkDirectoryPath() {
	currWd, _ := os.Getwd()
	log.Printf("Current working directory is: %s", currWd)

	rootDir := GetProjectRootDirectory()
	if err := os.Chdir(rootDir); err != nil {
		log.Printf("unable to chdir to project root %s: %v", rootDir, err)
		return
	}
	newWd, _ := os.Getwd()
	log.Printf("New fixed working directory is: %s", newWd)
}

func getProjectRootDirectoryFromProjectName(pn string) string {
	wd, _ := os.Getwd()
	for !strings.HasSuffix(wd, pn) {
		parent := filepath.Dir(wd)
		if parent == wd {
			return wd
		}
		wd = parent
	}
	return wd
}

func searchRootDirectory(dir string) (string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", errors.WrapIf(err, "Error reading directory")
	}

	for _, file := range files {
		if !file.IsDir() && strings.EqualFold(file.Name(), "go.mod") {
			return dir, nil
		}
	}

	parentDir := filepath.Dir(dir)
	if parentDir == dir {
		return "", errors.New("No go.mod file found")
	}
	return searchRootDirectory(parentDir)
}
