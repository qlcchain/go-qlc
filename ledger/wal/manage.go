package wal

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/log"
)

type WalLog struct {
	dirFd  *os.File
	logger *zap.SugaredLogger
}

func NewWalLog(dirName string) (*WalLog, error) {
	dirFd, err := OpenOrCreateFd(dirName)
	if err != nil {
		return nil, err
	}
	return &WalLog{
		dirFd:  dirFd,
		logger: log.NewLogger("wallog"),
	}, nil
}

func (w *WalLog) NewBlock(block *types.StateBlock) error {
	b, _ := block.Serialize()
	bytes := makeWriteBytes(block_wal, b)
	if _, err := w.dirFd.Write(bytes); err != nil {
		w.logger.Error(err)
		return err
	}
	//fmt.Println("makeWriteBytes, ", len(bytes), binary.BigEndian.Uint32(bytes[:4]))
	return nil
}

func makeWriteBytes(dataType byte, data []byte) []byte {
	buf := make([]byte, 5)
	binary.BigEndian.PutUint32(buf, uint32(len(data)+1))
	buf[4] = dataType
	buf = append(buf, data...)
	return buf
}

func (w *WalLog) ReadLogs() ([]*Location, error) {
	//fileinfo, err := w.dirFd.Stat()
	//if err != nil {
	//	w.logger.Error(err)
	//	return nil, err
	//}
	//
	//fileSize := fileinfo.Size()
	//buf := make([]byte, fileSize)
	//fmt.Println("======buf", fileSize)
	//n, err := w.dirFd.Read(buf)
	//if err != nil {
	//	w.logger.Error(err)
	//	return nil, err
	//}
	//if n == 0 {
	//	return make([]*Location, 0), nil
	//}
	//fmt.Println(buf)
	content, err := ioutil.ReadAll(w.dirFd)
	if err != nil {
		w.logger.Error(err)
		return nil, err
	}
	fmt.Println(content)
	index := 0
	lengthByte := make([]byte, 0)
	var length int
	typ := make([]byte, 0)
	obj := make([]byte, 0)
	for {
		lengthByte = content[index : index+4]
		length = int(binary.BigEndian.Uint32(lengthByte))
		typ = content[index+4 : index+5]
		obj = content[index+5 : index+length]
		index = index + length + 4
		fmt.Println("typ: ", typ)
		fmt.Println("obj: ", obj)
		if index >= len(content) {
			break
		}
	}

	return nil, nil
}

func (w *WalLog) Flush() error {
	return w.dirFd.Sync()
}

func (w *WalLog) Close() error {
	return w.dirFd.Close()
}
