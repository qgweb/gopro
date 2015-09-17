package function

import (
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/echo"

	"golang.org/x/crypto/bcrypt"

	"github.com/qgweb/gopro/lib/convert"
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

// 验证是否相同
func CheckBcrypt(data []byte, pwd []byte) bool {
	err := bcrypt.CompareHashAndPassword(data, pwd)
	if err != nil {
		return false
	}
	return true
}

// get或者post
func GetPost(ctx *echo.Context, name string) string {
	if ctx.Query(name) == "" {
		return ctx.Form(name)
	}

	return ctx.Query(name)
}

// 获取随机数
func GetRand(b int, e int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(e-b) + b
}

//  ip
func GetIP(ctx *echo.Context) string {
	return strings.Split(ctx.Request().RemoteAddr, ":")[0]
}
