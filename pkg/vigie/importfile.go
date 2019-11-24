package vigie

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/vincoll/vigie/pkg/importing"
	"github.com/vincoll/vigie/pkg/utils"

	"github.com/ghodss/yaml"
)

// loadTestFiles try to import all the testfiles
func (v *Vigie) loadTestFiles() (err error) {

	umt, err := importing.NewUnMarshallTool(v.VarsFiles, v.probes)

	// loop on each TestSuite files
	for i, f := range v.TestsFiles {

		// importingFileToVigie each Tests Files as TestSuite
		ts, err := umt.ImportTestSuite(f)
		if err != nil {
			utils.Log.WithFields(log.Fields{
				"error": err.Error(),
				"file":  f,
			}).Error("Cannot load this file.")
			return fmt.Errorf("%s ", err.Error())

		} else {
			// After Validation : Append this Valid TestSuites to Vigie
			ts.SourceFile = f
			//ts.Name += " [" + filepath.Base(f) + "]"
			ts.ID = int64(i)
			v.TestSuites[int64(i)] = ts
			utils.Log.WithFields(log.Fields{
				"file": f,
			}).Debug("Has been loaded.")

		}
	}

	// Delete Vars
	utils.ALLVARS = make(map[string][]string, 0)

	return nil
}

// Ajout des variables tableau dans une map Var:[valeurs tableau]
func addVariablesFromFiles(varPaths []string) (mapvars map[string][]string, err error) {

	mapvars = map[string][]string{}

	for _, varFile := range varPaths {

		varFileMap := make(map[string][]string)
		bytes, err := ioutil.ReadFile(varFile)
		if err != nil {
			log.Fatal(err)
		}
		switch filepath.Ext(varFile) {
		case ".json":
			err = json.Unmarshal(bytes, &varFileMap)
		case ".yaml", ".yml":
			err = yaml.Unmarshal(bytes, &varFileMap)
		default:
			log.Fatal("unsupported varFile format")
		}
		if err != nil {
			log.Fatal(err)
		}

		for key, value := range varFileMap {
			(mapvars)[key] = value
		}
	}
	return mapvars, nil
}
