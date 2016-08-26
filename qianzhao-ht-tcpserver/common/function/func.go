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

// 获取当前日期时间戳
func GetDateUnix() string {
	date, _ := time.ParseInLocation("2006-01-02", time.Now().Format("2006-01-02"), time.Local)
	return convert.ToString(date.Unix())
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
func GetPost(ctx echo.Context, name string) string {
	if ctx.QueryParam(name) == "" {
		return ctx.FormValue(name)
	}

	return ctx.QueryParam(name)
}

// 获取随机数
func GetRand(b int, e int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(e - b) + b
}

//  ip
func GetIP(ctx echo.Context) string {
	return strings.Split(ctx.Request().RemoteAddress(), ":")[0]
}

// aes加密 调用aes.jar
func AESEncrypt(str string) string {
	cmd := exec.Command("java", "-jar", "aes.jar", "-a", "en", "-v", str)
	res, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(res))
}

// aes解密 调用aes.jar
func AESDecrypt(str string) string {
	cmd := exec.Command("java", "-jar", "aes.jar", "-a", "de", "-v", str)
	res, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(res))
}
