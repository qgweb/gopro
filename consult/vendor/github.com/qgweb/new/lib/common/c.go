package common

import (
	"os"
	"os/exec"
	"path/filepath"
	"encoding/binary"
	"bytes"
)

// 获取程序执行目录
func GetBasePath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	return filepath.Dir(path)
}

// 获取bigendian binary 的int值
func GetBinaryInt(buf []byte) (v int64) {
	by:=bytes.NewReader(buf)
	binary.Read(by, binary.BigEndian,&v)
	return
}
