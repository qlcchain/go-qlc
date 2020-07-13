package migration

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	badger16 "github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/y"
	"github.com/verybluebot/tarinator-go"

	"github.com/qlcchain/go-qlc/log"
)

var logBadger = log.NewLogger("badger")

func MigrationTo20(dir string) error {
	logBadger.Info("WARN: Migration Data. It will take a long time, maybe from 10 minutes to 1 hour. Please wait patiently and do not end the program")
	backup := filepath.Join(filepath.Dir(dir), "ledger16.backup")
	os.Remove(backup)

	tarFile := fmt.Sprintf("%s/ledger_backup_v1.6.tar.gz", filepath.Dir(dir))
	tarTemp := fmt.Sprintf("%s/tar.Temp", filepath.Dir(dir))
	if _, err := os.Stat(tarFile); err != nil {
		if err := tar(dir, tarFile, tarTemp); err != nil {
			return err
		}
	} else {
		if _, err := os.Stat(tarTemp); err != nil { // interrupt when tar ledger, so tar file isn't integrity
			os.RemoveAll(tarFile)
			if err := tar(dir, tarFile, tarTemp); err != nil {
				return err
			}
		} else { // interrupt when badger v1 backup, so .backup isn't integrity
			os.RemoveAll(backup)
			//if err := untar(dir, tarFile); err != nil {
			//	return err
			//}
		}
	}

	if err := doBackup(dir, dir, backup); err != nil {
		return fmt.Errorf("doBackup error: %s", err)
	}
	os.RemoveAll(dir)
	os.RemoveAll(filepath.Join(filepath.Dir(dir), "wallet"))
	if err := doRestore(dir, dir, backup); err != nil {
		return fmt.Errorf("doRestore error: %s", err)
	}
	os.Remove(backup)
	return nil
}

func tar(dir, tarPath, tarTemp string) error {
	err := tarinator.Tarinate([]string{dir}, tarPath)
	if err != nil {
		return fmt.Errorf("pack ledger error: %s", err)
	}
	file, err := os.Create(tarTemp)
	if err != nil {
		return fmt.Errorf("create file error: %s", err)
	}
	defer file.Close()
	return nil
}

func untar(dir, tarPath string) error {
	err := tarinator.UnTarinate(filepath.Dir(dir), tarPath)
	if err != nil {
		return fmt.Errorf("unpack ledger error: %s", err)
	}
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

func doBackup(sstDir, vlogDir, backupFile string) error {
	// Open DB
	logBadger.Info("backup ledger")
	db, err := badger16.Open(badger16.DefaultOptions(sstDir).
		WithValueDir(vlogDir).
		WithTruncate(false))
	if err != nil {
		return err
	}
	defer db.Close()

	// Create File
	f, err := os.Create(backupFile)
	if err != nil {
		return err
	}

	bw := bufio.NewWriterSize(f, 64<<20)
	if _, err = db.Backup(bw, 0); err != nil {
		return err
	}

	if err = bw.Flush(); err != nil {
		return err
	}

	if err = y.FileSync(f); err != nil {
		return err
	}

	return f.Close()
}
