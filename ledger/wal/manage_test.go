package wal

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
)

func TestWalLog_NewBlock(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), uuid.New().String())
	w, err := NewWalLog(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	start := time.Now()
	if err := w.NewBlock(mock.StateBlockWithoutWork()); err != nil {
		t.Fatal(err)
	}
	fmt.Println("time: ", time.Now().Sub(start))

	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	w2, err := NewWalLog(dir)
	if _, err := w2.ReadLogs(); err != nil {
		t.Fatal(err)
	}
	if err := w2.Close(); err != nil {
		t.Fatal(err)
	}
}

//func TestWalLog_NewBlock2(t *testing.T) {
//	dir := filepath.Join(config.QlcTestDataDir(), uuid.New().String())
//	w, err := NewWalLog(dir)
//	if err != nil {
//		t.Fatal(err)
//	}
//	defer func() {
//		w.Close()
//		_ = os.RemoveAll(dir)
//	}()
//	w.ReadLogs()
//}

//func TestWalLog(t *testing.T) {
//	str := "12345456"
//	fd, _ := os.OpenFile("abc.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
//	for i := 0; i < 10; i++ {
//		fd_time := time.Now().Format("2006-01-02 15:04:05")
//		fd_content := strings.Join([]string{"======", fd_time, "=====", str, "\n"}, "")
//		buf := []byte(fd_content)
//		start := time.Now()
//		fd.Write(buf)
//		fmt.Println(time.Now().Sub(start))
//	}
//	return
//	fd.Close()
//}
