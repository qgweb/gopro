package common

import (
	"github.com/qgweb/gopro/lib/convert"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// 获取程序执行目录
func GetBasePath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	return filepath.Dir(path)
}

/**
 * 获取任意一天整点时间戳
 */
func GetDayTimestamp(day int) string {
	d := time.Now().AddDate(0, 0, day).Format("20060102")
	a, _ := time.ParseInLocation("20060102", d, time.Local)
	return convert.ToString(a.Unix())
}
