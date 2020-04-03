package migration

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v2"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

func MigrationTo20(dir string) error {
	toolPath := filepath.Join(filepath.Dir(dir), "gqlc_migration_data_tool")
	if _, err := os.Stat(toolPath); err != nil {
		if err := downloadTool(toolPath); err != nil {
			return err
		}
		if err := md5Check(toolPath); err != nil {
			return err
		}
	}

	fmt.Println("migrate badger to v2.0")
	backpath := filepath.Join(filepath.Dir(dir), "ledger16.backup")
	os.Remove(backpath)
	command := []string{"", "backup", "--dir", dir, "-f", backpath}
	cmd := exec.Command(toolPath, command...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execute command failed: %s", err.Error())
	}
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("remove ledger error: %s", err)
	}
	os.RemoveAll(filepath.Join(filepath.Dir(dir), "wallet"))
	if err := doRestore(dir, dir, backpath); err != nil {
		return fmt.Errorf("doRestore error: %s", err)
	}
	return nil
}

func doRestore(sstDir, vlogDir, restoreFile string) error {
	maxPendingWrites := 256
	// Check if the DB already exists
	manifestFile := path.Join(sstDir, badger.ManifestFilename)
	if _, err := os.Stat(manifestFile); err == nil { // No error. File already exists.
		return errors.New("Cannot restore to an already existing database")
	} else if os.IsNotExist(err) {
		// pass
	} else { // Return an error if anything other than the error above
		return err
	}

	// Open DB
	db, err := badger.Open(badger.DefaultOptions(sstDir).WithValueDir(vlogDir))
	if err != nil {
		return err
	}
	defer db.Close()

	// Open File
	f, err := os.Open(restoreFile)
	if err != nil {
		return err
	}
	defer f.Close()

	// Run restore
	return db.Load(f, maxPendingWrites)
}

func downloadTool(dir string) error {
	url := "https://github.com/qlcchain/qlc-go-sdk/releases/download/v1.3.5/gqlc_migration_data_tool"
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dir, data, 0777)
}

func md5Check(dir string) error {
	f, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer f.Close()
	md5 := md5.New()
	if _, err := io.Copy(md5, f); err != nil {
		return err
	}
	md5Str := hex.EncodeToString(md5.Sum(nil))
	m := "7c0c8b4bc225c52bf9c38d70a23a5dbe"
	if md5Str != m {
		return fmt.Errorf("md5 check fail: %s, %s", md5Str, m)
	}
	return nil
}
