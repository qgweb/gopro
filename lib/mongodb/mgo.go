// mongodb 连接池
// 多协程读取数据
package mongodb

import (
	"container/heap"
	"fmt"
	"github.com/ngaut/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"math"
	"runtime"
	"sync"
	"time"
)

// session
type Session struct {
	*mgo.Session
	ref   int
	index int
}

type MulQueryParam struct {
	DbName  string
	ColName string
	Query   bson.M
	Size    int                          // 取多少页
	Fun     func(map[string]interface{}) // 回调函数
}

type MgoConfig struct {
	UserName string
	UserPwd  string
	Host     string
	Port     string
	DBName   string
}

// session heap
type SessionHeap []*Session

func (h SessionHeap) Len() int {
	return len(h)
}

func (h SessionHeap) Less(i, j int) bool {
	return h[i].ref < h[j].ref
}

func (h SessionHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *SessionHeap) Push(s interface{}) {
	s.(*Session).index = len(*h)
	*h = append(*h, s.(*Session))
}

func (h *SessionHeap) Pop() interface{} {
	l := len(*h)
	s := (*h)[l-1]
	s.index = -1
	*h = (*h)[:l-1]
	return s
}

type DialContext struct {
	sync.Mutex
	sessions SessionHeap
	isDebug  bool
}

func GetLinkUrl(p MgoConfig) string {
	if p.UserName == "" && p.UserPwd == "" {
		return fmt.Sprintf("%s:%s/%s", p.Host, p.Port, p.DBName)
	}
	return fmt.Sprintf("%s:%s@%s:%s/%s", p.UserName, p.UserPwd,
		p.Host, p.Port, p.DBName)
}

// goroutine safe
func Dial(url string, sessionNum int) (*DialContext, error) {
	c, err := DialWithTimeout(url, sessionNum, 10*time.Second, 5*time.Minute)
	return c, err
}

func GetCpuSessionNum() int {
	return runtime.NumCPU() * 2
}

// goroutine safe
func DialWithTimeout(url string, sessionNum int, dialTimeout time.Duration, timeout time.Duration) (*DialContext, error) {
	if sessionNum <= 0 {
		sessionNum = 100
		log.Error("invalid sessionNum, reset to %v", sessionNum)
	}

	s, err := mgo.DialWithTimeout(url, dialTimeout)
	if err != nil {
		return nil, err
	}
	s.SetSyncTimeout(timeout)
	s.SetSocketTimeout(timeout)
	s.SetCursorTimeout(0)

	c := new(DialContext)

	// sessions
	c.sessions = make(SessionHeap, sessionNum)
	c.sessions[0] = &Session{s, 0, 0}
	for i := 1; i < sessionNum; i++ {
		c.sessions[i] = &Session{s.New(), 0, i}
	}
	heap.Init(&c.sessions)

	return c, nil
}

// goroutine safe
func (c *DialContext) Close() {
	c.Lock()
	for _, s := range c.sessions {
		s.Close()
		if s.ref != 0 {
			log.Error("session ref = %v", s.ref)
		}
	}
	c.Unlock()
}

// goroutine safe
func (c *DialContext) Ref() *Session {
	c.Lock()
	s := c.sessions[0]
	if s.ref == 0 {
		s.Refresh()
	}
	s.ref++
	heap.Fix(&c.sessions, 0)
	c.Unlock()

	return s
}

// goroutine safe
func (c *DialContext) UnRef(s *Session) {
	c.Lock()
	s.ref--
	heap.Fix(&c.sessions, s.index)
	c.Unlock()
}

func (c *DialContext) Debug() {
	c.isDebug = true
}

func (c *DialContext) Log(msg ...interface{}) {
	if c.isDebug {
		log.Debug(msg...)
	}
}

func (c *DialContext) Query(param MulQueryParam) error {
	var (
		wg       = sync.WaitGroup{}
		count    = 0
		pageSize = 0
		sess     = c.Ref()
		err      error
		cpunum   = runtime.NumCPU()
	)

	defer func() {
		if msg := recover(); msg != nil {
			log.Error(msg)
		}
	}()

	count, err = sess.DB(param.DbName).C(param.ColName).Find(param.Query).Count()
	c.UnRef(sess)
	c.Log(count)
	if err != nil {
		return err
	}
	if param.Size == 0 {
		param.Size = int(math.Ceil(float64(count) / float64(cpunum) / 2))
	}
	pageSize = int(math.Ceil(float64(count) / float64(param.Size)))
	c.Log(pageSize)
	for i := 1; i <= pageSize; i++ {
		wg.Add(1)
		go func(p int) {
			sess := c.Ref()
			defer c.UnRef(sess)
			iter := sess.DB(param.DbName).C(param.ColName).Find(param.Query).
				Limit(param.Size).Skip((p - 1) * param.Size).Iter()
			for {
				var info map[string]interface{}
				if !iter.Next(&info) {
					break
				}
				param.Fun(info)
			}
			iter.Close()
			c.Log(iter.Err())
			wg.Done()
		}(i)
	}
	wg.Wait()
	c.Log("ok")
	return nil
}

func (c *DialContext) Insert(param MulQueryParam, list []interface{}) {
	var (
		wg        = sync.WaitGroup{}
		totalSize = len(list)
		cpunum    = runtime.NumCPU()
		pageSize  = int(math.Ceil(float64(totalSize) / float64(cpunum) / 2))
		pageCount = int(math.Ceil(float64(totalSize) / float64(pageSize)))
	)

	for i := 1; i <= pageCount; i++ {
		bg := (i - 1) * pageSize
		eg := bg + pageSize
		if i == pageCount {
			eg = totalSize
		}
		wg.Add(1)
		go func(b int, e int) {
			sess := c.Ref()
			defer c.UnRef(sess)
			err := sess.DB(param.DbName).C(param.ColName).Insert(list[b:e]...)
			c.Log(err)
			wg.Done()
			c.Log(b, e)
		}(bg, eg)
	}
	wg.Wait()
	c.Log("ok")
}
