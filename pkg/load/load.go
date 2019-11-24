package load

import (
	"crypto/tls"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/vincoll/vigie/pkg/utils"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/client"
	githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"net/http"
	"os"
)

// CloneGitRepo clone a git repo containing the tests and vars
func CloneGitRepo(cg ConfGit) error {

	if _, err := os.Stat(cg.Path); !os.IsNotExist(err) {
		if err != nil {
			return fmt.Errorf("%s", err)
		}
		err = os.RemoveAll(cg.Path)
		if err != nil {
			return fmt.Errorf("%s", err)
		}
	}

	var customClient *http.Client

	if cg.Insecure {
		customClient = &http.Client{
			// accept any certificate (might be useful for testing)
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}

		// Override http(s) default protocol to use our custom client
		client.InstallProtocol("https", githttp.NewClient(customClient))
	}

	r, err := git.PlainClone(cg.Path, false, &git.CloneOptions{
		URL:               cg.Repo,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})

	if err != nil {
		return fmt.Errorf("git clone of %s to %s failed : %s", cg.Repo, cg.Path, err)
	}

	// ... retrieving the branch being pointed by HEAD
	ref, err := r.Head()
	if err != nil {
	}
	// ... retrieving the commit object
	commit, err := r.CommitObject(ref.Hash())

	utils.Log.WithFields(log.Fields{}).Infof("Commit %s from %s has been cloned in %s", commit, cg.Repo, cg.Path)

	return nil
}

// importFileandVars Charge la config d'un fichier vigieConf dans une instance Vigie
// La fonction retourne tous les fichiers (test,vars) éligibles contenus dans les répertoires Tests et Vars
func ImportFileandVars(ctf ConfTestfiles, cv ConfVariables) (testsFiles []string, varsFiles []string, err error) {

	// importing Files
	// Path to each individual Step & Var file.

	// Add Step path for each file
	// TODO: Gérer l'erreur ?
	testsFiles, _ = getAllFilesInsideDir(ctf.Included, ctf.Excluded, defaultTestSuitePath)
	if len(testsFiles) == 0 {
		return nil, nil, fmt.Errorf("no files or path to import")
	}

	// Add Var path for each file
	varsFiles, _ = getAllFilesInsideDir(cv.Included, cv.Excluded, defaultVariablePath)

	return testsFiles, varsFiles, nil
}
