package common

import (
	"bufio"
	"github.com/ngaut/log"
	"io"
	"os"
	//	"os/exec"
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
