package migration

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	badger16 "github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/v2"

	"github.com/qlcchain/go-qlc/log"
)

var logBadger = log.NewLogger("badger")

func MigrationTo20(dir string) error {
	fmt.Println(badger16.DefaultIteratorOptions)
	fmt.Println("migrate badger to v2.0, ", toolName)
	toolPath := filepath.Join(filepath.Dir(dir), toolName)
	if _, err := os.Stat(toolPath); err != nil {
		if err := downloadTool(toolPath); err != nil {
			return err
		}
		defer os.Remove(toolPath)
	}
	logBadger.Info("backup ledger")
	backup := filepath.Join(filepath.Dir(dir), "ledger16.backup")
	os.Remove(backup)
	command := []string{"", "backup", "--dir", dir, "-f", backup}
	cmd := exec.Command(toolPath, command...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execute command failed: %s (%s, %s)", err, runtime.GOOS, runtime.GOARCH)
	}
	os.RemoveAll(dir)
	os.RemoveAll(filepath.Join(filepath.Dir(dir), "wallet"))
	if err := doRestore(dir, dir, backup); err != nil {
		return fmt.Errorf("doRestore error: %s", err)
	}
	os.Remove(backup)
	return nil
}

func doRestore(sstDir, vlogDir, restoreFile string) error {
	logBadger.Info("restore ledger")
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
	logBadger.Info("download tool")
	url := "https://github.com/qlcchain/qlc-go-sdk/releases/download/v1.3.5"

	// get tool
	toolUrl := fmt.Sprintf("%s/%s", url, toolName)
	resTool, err := http.Get(toolUrl)
	if err != nil {
		return fmt.Errorf("get tool: %s (%s, %s, %s)", err, runtime.GOOS, runtime.GOARCH, dir)
	}
	defer resTool.Body.Close()
	data, err := ioutil.ReadAll(resTool.Body)
	if err != nil {
		return fmt.Errorf("read data: %s (%s)", err, dir)
	}
	if err := ioutil.WriteFile(dir, data, 0777); err != nil {
		return fmt.Errorf("write file: %s", err)
	}

	logBadger.Info("download tool successfully")
	// get tool hash
	checkUrl := fmt.Sprintf("%s/checksums.txt", url)
	res, err := http.Get(checkUrl)
	if err != nil {
		return fmt.Errorf("get checksum: %s (%s, %s)", err, runtime.GOOS, runtime.GOARCH)
	}
	defer res.Body.Close()
	var toolHash string
	r := bufio.NewReader(res.Body)
	for {
		r, _, err := r.ReadLine()
		if err != nil {
			break
		}
		checkValue := strings.Split(string(r), " ")
		if checkValue[1] == toolName {
			toolHash = checkValue[0]
			break
		}
	}
	if toolHash == "" {
		return fmt.Errorf("get tool failed: %s, %s", runtime.GOOS, runtime.GOARCH)
	}
	logBadger.Info("check tool md5")
	// check tool hash
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
	if md5Str != toolHash {
		return fmt.Errorf("md5 check fail: %s, %s", md5Str, toolHash)
	}
	return nil
}

var (
	toolName string
)

func init() {
	prefix := "gqlc_migration_data_tool"
	toolName = fmt.Sprintf("%s_%s_%s", prefix, runtime.GOOS, runtime.GOARCH)

	if runtime.GOOS == "darwin" {
		if runtime.GOARCH == "amd64" {
			toolName = prefix + "_darwin_x64"
		}
	} else if runtime.GOOS == "windows" {
		if runtime.GOARCH == "amd64" {
			toolName = prefix + "_windows_x64.exe"
		} else {
			toolName = prefix + "_windows_386.exe"
		}
	} else {
		if runtime.GOARCH == "amd64" {
			toolName = prefix + "_linux_x64"
		} else {
			toolName = prefix + "_linux_arm"
		}
	}
}
