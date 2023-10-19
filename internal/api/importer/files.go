package load

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// getFilesPath retrieve recursively all json and yml test files in a directory. Except excluded files
func getFilesPath(paths []string, exclude []string) ([]string, error) {

	validExtFile := map[string]bool{".json": true, ".yml": true, ".yaml": true}
	// Get all files
	okPaths, err := uniqueRecursiveFilesPath(paths, validExtFile)
	if err != nil {
		var empty []string
		return empty, err
	}
	// Return if no files have been found.
	if len(okPaths) == 0 {
		var empty []string
		return empty, nil
	}

	// If flag exclude: Fetch and Delete paths from okPaths List
	if len(exclude) > 1 {
		excludedPaths, _ := uniqueRecursiveFilesPath(exclude, validExtFile)

		// Delete excluded paths
		for key := range okPaths {
			if excludedPaths[key] {
				delete(okPaths, key)
			}
		}
	}

	var filePaths = make([]string, 0, len(okPaths))

	// Return a clean list of unique entries
	for k := range okPaths {
		filePaths = append(filePaths, k)
	}
	return filePaths, nil
}

// uniqueRecursiveFilesPath Returns map of unique files into folder by recursion
func uniqueRecursiveFilesPath(SrcPaths []string, extfile map[string]bool) (map[string]bool, error) {

	var filesPath = make(map[string]bool)

	for _, p := range SrcPaths {
		p = strings.TrimSpace(p)
		p = filepath.Clean(p)

		// Check if path exists
		if _, err := os.Stat(p); os.IsNotExist(err) {
			return nil, err
		}

		err := filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			if _, match := extfile[filepath.Ext(info.Name())]; match {
				filesPath[path] = true
			}
			return nil
		})
		if err != nil {
			log.Errorf("Error reading files on path:%s :%s", SrcPaths, err)
			return nil, errors.Wrapf(err, "error reading files on path %q", SrcPaths)
		}
	}
	return filesPath, nil
}

// getAllFilesInsideDir deep search to find
func getAllFilesInsideDir(p []string, pathsExcluded []string, defaultpath string) ([]string, error) {

	/* importing files
	Path to each individual Step & Var file.
	*/

	// Clean list from args ( --test and all except flag)
	// Tests Path
	paths := importArgsPathShort(p)
	if len(paths) == 1 && paths[0] == "" {
		// If paths empty : Remplace [0] = "" by the default folder
		paths[0] = defaultpath
	}

	// Crée une Liste de fichier contenu dans un dossier (yml,yaml,json)
	// Tests Path
	files, err := getFilesPath(paths, pathsExcluded)
	if err != nil {
		return nil, fmt.Errorf("error during file import: %s", err)
	}

	return files, err
}

// importArgsPath: Create a clean list from arguments
func importArgsPathShort(testsPath []string) (uTestPath []string) {

	uTestPath = make([]string, 0, 0)

	if len(testsPath) == 0 {
		// If no Args Files Paths : Add ./tests
		uTestPath[0] = defaultTestSuitePath
		return uTestPath
	} else {
		// Add Unique entry to Map
		return unique(testsPath)
	}
}

func unique(intSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func sliceUniqMap(s []string) []string {
	seen := make(map[string]struct{}, len(s))
	j := 0
	for _, v := range s {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		s[j] = v
		j++
	}
	return s[:j]
}
