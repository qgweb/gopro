package common

import (
	"bufio"
	"github.com/ngaut/log"
	"github.com/qgweb/new/lib/convert"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

//  例子
//	dw := common.NewDispathWriter()
//	dw.ReStart()
//	dw.Dispath("advert_2016_")
//	for j := 100; j < 102; j++ {
//	for i := 0; i < 100; i++ {
//	ad := "ad" + convert.ToString(i)
//	ua := "ua" + convert.ToString(i)
//	dw.Push(ad, ua, convert.ToString(j))
//	}
//
//	for i := 0; i < 100; i++ {
//	ad := "ad" + convert.ToString(i)
//	ua := "ua" + convert.ToString(i)
//	dw.Push(ad, ua, convert.ToString(j))
//	}
//	}
//
//	dw.CloseReadChan()
//	dw.Wait()
//	fmt.Println(dw.WC())

type advert struct {
	ad string
	ua string
	id string
}

type WriteCloser struct {
	w     *bufio.Writer
	c     io.Closer
	fname string
}

type DispathWriter struct {
	readChan chan advert
	overChan chan bool
	writers  map[string]*WriteCloser
}

func NewDispathWriter() *DispathWriter {
	return &DispathWriter{
		make(chan advert),
		make(chan bool),
		make(map[string]*WriteCloser),
	}
}

func (this *DispathWriter) CloseReadChan() {
	close(this.readChan)
}

func (this *DispathWriter) ReStart() {
	this.readChan = make(chan advert)
	this.writers = make(map[string]*WriteCloser)
}

func (this *DispathWriter) Push(ad string, ua string, id string) {
	this.readChan <- advert{ad, ua, id}
}

func (this *DispathWriter) Dispath(filePrefix string) {
	go func() {
		for {
			v, ok := <-this.readChan
			if !ok {
				this.overChan <- true
				break
			}

			if _, ok := this.writers[v.id]; !ok {
				fn := filePrefix + v.id + ".txt"
				f, err := os.OpenFile(fn, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0664)
				if err != nil {
					log.Error(err)
					this.overChan <- true
					break
				}
				this.writers[v.id] = &WriteCloser{bufio.NewWriter(f), f, fn}
			}

			this.writers[v.id].w.WriteString(v.ad + "_" + v.ua + "\n")
		}
	}()
}

func (this *DispathWriter) uniq(fn string) {
	generator := exec.Command("sort", fn)
	consumer := exec.Command("uniq")

	p, _ := generator.StdoutPipe()
	generator.Start()
	consumer.Stdin = p
	pp, _ := consumer.StdoutPipe()
	consumer.Start()
	f, _ := os.Create(fn + ".bak")
	io.Copy(f, pp)
	f.Close()
	os.Rename(fn+".bak", fn)
}

func (this *DispathWriter) flen(fn string) int {
	f, err := os.Open(fn)
	if err != nil {
		log.Error(err)
		return 0
	}
	defer f.Close()
	bi := bufio.NewReader(f)
	num := 0
	for {
		_, err := bi.ReadString('\n')
		if io.EOF == err || nil != err {
			break
		}
		num += 1
	}
	return num
}

func (this *DispathWriter) Wait() {
	<-this.overChan
	wg := sync.WaitGroup{}
	for _, w := range this.writers {
		w.w.Flush()
		w.c.Close()
		wg.Add(1)
		go func(n string) {
			defer wg.Done()
			this.uniq(n)
		}(w.fname)
	}
	wg.Wait()
}

func (this *DispathWriter) WC() (mp map[string]int) {
	mp = make(map[string]int)
	for _, w := range this.writers {
		mp[strings.TrimSuffix(filepath.Base(w.fname), ".txt")] = this.flen(w.fname)
	}
	return
}

func (this *DispathWriter) Clean() {
	for _, w := range this.writers {
		os.Remove(w.fname))
	}
}

// ===========================================================
//  内容之间用\t分割
//	kf := common.NewKVFile("./xxx.txt")
//	kf.AddFun(func(out chan interface{}, in chan int8) {
//	out <- fmt.Sprintf("%s\t%s\t%s", "ad1", "ua1", "1,2,3")
//	out <- fmt.Sprintf("%s\t%s\t%s", "ad1", "ua1", "4")
//	out <- fmt.Sprintf("%s\t%s\t%s", "ad2", "ua2", "4")
//	in <- 1
//	})
//	kf.AddFun(func(out chan interface{}, in chan int8) {
//	out <- fmt.Sprintf("%s\t%s\t%s", "ad3", "ua3", "1,2,3")
//	out <- fmt.Sprintf("%s\t%s\t%s", "ad4", "ua4", "4")
//	out <- fmt.Sprintf("%s\t%s\t%s", "ad3", "ua3", "4,3")
//	in <- 1
//	})
//	fmt.Println(kf.WriteFile())
//
//	kf.AdSet(func(as string) {
//		fmt.Println("ad", as)
//	})
//
//	kf.AdUaIdsSet(func(ad string, ua string, aid map[string]int8) {
//	fmt.Println("adua", ad, ua, aid)
//	})
//
//	kf.IDAdUaSet("advert_2016_", func(m map[string]int) {
//	fmt.Println(m)
//	})

// 第一个参数输出， 第二个完成时写入值
type ReadFun func(chan interface{}, chan int8)
type AdFun func(string)
type AdUaIdsFun func(string, string, map[string]int8)
type OriginFun func([]string)

type AdUaAdverts struct {
	Ad  string
	UA  string
	AId map[string]int8
}

type KVFile struct {
	fname    string
	rchan    chan interface{}
	overChan chan int8
	funAry   []ReadFun
}

func NewKVFile(fname string) *KVFile {
	return &KVFile{
		fname:    fname,
		rchan:    make(chan interface{}, 2),
		overChan: make(chan int8),
		funAry:   make([]ReadFun, 0),
	}
}

func (this *KVFile) AddFun(rf ReadFun) {
	this.funAry = append(this.funAry, rf)
}

func (this *KVFile) Clean() {
	os.Remove(this.fname)
}

func (this *KVFile) WriteFile() error {
	wf, err := this.createFile()
	if err != nil {
		return err
	}
	defer wf.Close()

	this.overChan = make(chan int8, len(this.funAry))
	tmpOverChan := make(chan int8)
	for _, f := range this.funAry {
		go func(fn ReadFun) {
			fn(this.rchan, this.overChan)
		}(f)
	}

	go func() {
		for {
			v, ok := <-this.rchan
			if !ok {
				tmpOverChan <- 1
				break
			}
			wf.WriteString(convert.ToString(v) + "\n")
		}
	}()

	for i := 0; i < cap(this.overChan); i++ {
		<-this.overChan
	}
	close(this.rchan)
	<-tmpOverChan
	this.sortuniqm()
	return nil
}

func (this *KVFile) createFile() (*os.File, error) {
	f, err := os.Create(this.fname)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (this *KVFile) sortuniq() {
	fn := this.fname
	if _, err := os.Stat(fn); !(err == nil || os.IsExist(err)) {
		return
	}
	generator := exec.Command("sort", fn)
	consumer := exec.Command("uniq")

	p, _ := generator.StdoutPipe()
	generator.Start()
	consumer.Stdin = p
	pp, _ := consumer.StdoutPipe()
	consumer.Start()
	f, _ := os.Create(fn + ".bak")
	io.Copy(f, pp)
	f.Close()
	os.Rename(fn+".bak", fn)
}

func (this *KVFile) sortuniqm() {
	cmd:=exec.Command("bash", GetBasePath() + "/sort.sh", this.fname)
	cmd.Run()
}


// 原始数据
func (this *KVFile) Origin(fun OriginFun) error {
	f, err := os.Open(this.fname)
	if err != nil {
		return err
	}
	defer f.Close()

	bi := bufio.NewReader(f)
	for {
		line, err := bi.ReadString('\n')
		if err == io.EOF || err != nil {
			break
		}

		infos := strings.Split(strings.TrimSpace(line), "\t")
		fun(infos)
	}
	return nil
}

// 返回ad集合
func (this *KVFile) AdSet(fun AdFun) error {
	f, err := os.Open(this.fname)
	if err != nil {
		return err
	}
	defer f.Close()

	bi := bufio.NewReader(f)
	ad := ""
	for {
		line, err := bi.ReadString('\n')
		if err == io.EOF || err != nil {
			fun(ad)
			break
		}

		line = strings.TrimSpace(line)
		infos := strings.Split(line, "\t")

		if len(infos) < 3 {
			continue
		}

		if infos[0] != ad && ad != "" {
			fun(ad)
		}
		ad = infos[0]
	}
	return nil
}

func (this *KVFile) AdUaIdsSet(fun AdUaIdsFun) error {
	f, err := os.Open(this.fname)
	if err != nil {
		return err
	}
	defer f.Close()

	bi := bufio.NewReader(f)
	ad := AdUaAdverts{AId: make(map[string]int8)}
	for {
		line, err := bi.ReadString('\n')
		if err == io.EOF || err != nil {
			fun(ad.Ad, ad.UA, ad.AId)
			break
		}

		line = strings.TrimSpace(line)
		infos := strings.Split(line, "\t")

		if len(infos) < 3 {
			continue
		}

		if (ad.Ad != infos[0] && ad.UA != infos[1]) && (ad.Ad != "" && ad.UA != "") {
			fun(ad.Ad, ad.UA, ad.AId)
			ad.AId = make(map[string]int8)
			ad.Ad = ""
			ad.UA = ""
		}

		ad.Ad = infos[0]
		ad.UA = infos[1]
		for _, id := range strings.Split(infos[2], ",") {
			ad.AId[id] = 1
		}
	}
	return nil
}

func (this *KVFile) IDAdUaSet(prifix string, fun func(map[string]int)) error {
	dw := NewDispathWriter()
	dw.ReStart()
	dw.Dispath(prifix)

	f, err := os.Open(this.fname)
	if err != nil {
		return err
	}
	defer f.Close()

	bi := bufio.NewReader(f)
	for {
		line, err := bi.ReadString('\n')
		if err == io.EOF || err != nil {
			break
		}

		line = strings.TrimSpace(line)
		infos := strings.Split(line, "\t")

		if len(infos) < 3 {
			continue
		}

		for _, v := range strings.Split(infos[2], ",") {
			dw.Push(infos[0], infos[1], v)
		}
	}
	dw.CloseReadChan()
	dw.Wait()
	fun(dw.WC())
	return nil
}
