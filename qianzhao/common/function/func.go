package function

import (
	"golang.org/x/crypto/bcrypt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/goweb/gopro/lib/convert"
)

// 获取程序执行目录
func GetBasePath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	return filepath.Dir(path)
}

// 获取当前时间戳
func GetTimeUnix() string {
	return convert.ToString(time.Now().Unix())
}

// 获取机密
func GetBcrypt(data []byte) string {
	pwd, _ := bcrypt.GenerateFromPassword(data, bcrypt.DefaultCost)
	return string(pwd)
}

// 严重是否相同
func CheckBcrypt(data []byte, pwd []byte) bool {
	err := bcrypt.CompareHashAndPassword(data, pwd)
	if err != nil {
		return false
	}
	return true
}
