package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"html/template"

	"github.com/qgweb/gopro/lib/convert"

	"os/exec"
	"os/signal"

	"regexp"

	"code.google.com/p/go.crypto/ssh"
	"gopkg.in/ini.v1"
)

var (
	iniFile   *ini.File
	conf      = flag.String("conf", "conf.ini", "config file")
	sshClient *ssh.Client
	err       error
	exePath   string
)

type ProgramHealth struct {
	Name   string //名称
	Status string //状态
	Count  int    //运行数量
	Host   string //服务器地址
	PName  string //进程名称
	Mem    string //内存
	CPU    string //cpu
}

func init() {
	flag.Parse()

	if *conf == "" {
		log.Fatalln("配置文件不能为空")
	}

	//加载配置文件内容
	readConfData()

	//初始化ssh链接
	initSSH()

	//
	exePath, _ = exec.LookPath(os.Args[0])
	exePath = filepath.Dir(exePath)

}

func main() {
	var (
		hhost = iniFile.Section("http").Key("host").String()
		hport = iniFile.Section("http").Key("port").String()
	)

	if hhost == "" || hport == "" {
		log.Fatalln("http节信息确实,程序启动失败")
	}

	sch := make(chan os.Signal)
	signal.Notify(sch, syscall.SIGHUP, syscall.SIGKILL, syscall.SIGTERM)

	go func() {
		for {
			ch := <-sch
			switch ch {
			case syscall.SIGHUP:
				log.Println("接收到热重启信号..")
				readConfData()
				log.Println("重新加载配置文件成功")
				break
			case syscall.SIGKILL, syscall.SIGTERM:
				log.Fatalln("程序退出")
				break
			}
		}
	}()

	http.HandleFunc("/health", proHealth)
	http.HandleFunc("/cron", CronJob)
	http.HandleFunc("/rcron", ReceiveCronJob) //接收任务计划

	log.Println("程序启动中")
	log.Println("程序运行在", hhost+":"+hport)
	log.Fatalln(http.ListenAndServe(fmt.Sprintf("%s:%s", hhost, hport), nil))

}

//读取配置文件
func readConfData() {
	confData, err := ioutil.ReadFile(*conf)
	if err != nil {
		log.Fatalln("打开配置文件失败,错误信息:", err)
	}

	iniFile, err = ini.Load(confData)
	if err != nil {
		log.Fatalln("加载配置文件内容出错,错误信息为:", err)
	}
}

//初始化ssh
func initSSH() {
	var (
		user = iniFile.Section("ssh").Key("user").String()
		pwd  = iniFile.Section("ssh").Key("pwd").String()
		host = iniFile.Section("ssh").Key("host").String()
	)

	if user == "" || pwd == "" || host == "" {
		log.Fatalln("读取ssh信息出错")
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pwd),
		},
	}
	sshClient, err = ssh.Dial("tcp", fmt.Sprintf("%s:%s", host, "22"), config)
	if err != nil {
		log.Fatalln("连接主机出错,错误信息为:" + err.Error())
	}
}

//执行ssh
func runSSH(ps string) (string, error) {
	sess, err := sshClient.NewSession()
	if err != nil {
		log.Println("创建sshSession失败,错误信息为:", err.Error())
		return "", err
	}
	defer sess.Close()

	var b bytes.Buffer
	sess.Stdout = &b
	if err := sess.Run(ps); err != nil {
		return "", err
	}
	return b.String(), nil
}

//程序心跳检测
func proHealth(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if a := recover(); a != nil {
			log.Println("心跳检测程序出错,", a)
		}
	}()

	tp, err := template.ParseFiles(exePath + "/tpl/health.html")
	if err != nil {
		io.WriteString(w, "读取模板文件失败")
	}

	joblist := iniFile.Section("healthjob").KeysHash()
	phs := make([]ProgramHealth, 0, 10)

	for k, v := range joblist {
		ph := ProgramHealth{}
		vs := strings.Split(v, "|")
		count := "1"
		ph.Name = k
		ph.Status = "runing"
		ph.Host = vs[1]
		ph.PName = vs[0]

		if len(vs) > 1 {
			count = vs[2]
		}

		cmd := fmt.Sprintf(`ssh %s 'ps aux | grep "%s"'`, ph.Host, ph.PName)
		res, err := runSSH(cmd)

		if err != nil {
			ph.Status = "error"
		} else {

			resary := strings.Split(strings.TrimSpace(res), "\n")
			tcount := len(resary)
			ph.Count = tcount - 2

			if ph.Count <= 0 {
				ph.Status = "stop"
			} else if ph.Count < convert.ToInt(count) {
				ph.Status = "part running"
			}

			if ph.Count == 1 {
				reg, _ := regexp.Compile(`(\w+)\s+(\d+)\s+(\d+\.\d+)\s+(\d+\.\d+)\s+(\d+)\s+(\d+)\s+[\w\/\?]+\s+(\w+)\s+([\w\:]+)\s+(\d+\:\d+)\s+(.+)`)
				for _, vv := range resary {
					vvs := reg.FindStringSubmatch(vv)
					if len(vvs) > 1 && !strings.Contains(vvs[10], "grep") && !strings.Contains(vvs[10], "bash -c") {
						ph.CPU = vvs[3]
						ph.Mem = vvs[4]
					}
				}
			}

		}

		phs = append(phs, ph)
	}

	tp.Execute(w, phs)
}
