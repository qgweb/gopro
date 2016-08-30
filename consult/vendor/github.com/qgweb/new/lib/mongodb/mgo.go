package mongodb

import (
	"fmt"
	"sync"
	"time"

	"github.com/juju/errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MM bson.M

type MongodbConf struct {
	Host          string
	Port          string
	UName         string
	Upwd          string
	Db            string
	DialTimeout   time.Duration
	SyncTimeout   time.Duration
	SocketTimeout time.Duration
}

type MongodbQueryConf struct {
	Db     string
	Table  string
	Query  MM
	Select MM
	Insert []interface{}
	Index  []string
	Update MM
	Delete MM
}

type Mongodb struct {
	sync.RWMutex
	conn *mgo.Session
	conf MongodbConf
}

func GetLinkUrl(p MongodbConf) string {
	if p.UName == "" && p.Upwd == "" {
		return fmt.Sprintf("%s:%s/%s", p.Host, p.Port, p.Db)
	}
	return fmt.Sprintf("%s:%s@%s:%s/%s", p.UName, p.Upwd,
		p.Host, p.Port, p.Db)
}

func GetObjectId() string {
	return bson.NewObjectId().Hex()
}

func IsObjectId(v string) bool {
	return bson.IsObjectIdHex(v)
}

func ObjectId(v string) bson.ObjectId {
	if bson.IsObjectIdHex(v) {
		return bson.ObjectIdHex(v)
	}
	return bson.NewObjectId()
}

func NewMongodb(conf MongodbConf) (*Mongodb, error) {
	if conf.DialTimeout == 0 {
		conf.DialTimeout = time.Second * 30
	}
	if conf.SyncTimeout == 0 {
		conf.SyncTimeout = time.Minute * 30
	}
	if conf.SocketTimeout == 0 {
		conf.SocketTimeout = time.Minute * 30
	}
	sess, err := mgo.DialWithTimeout(GetLinkUrl(conf), conf.DialTimeout)
	if err == nil {
		sess.SetSocketTimeout(conf.SocketTimeout)
		sess.SetSyncTimeout(conf.SyncTimeout)
	}
	return &Mongodb{sync.RWMutex{}, sess, conf}, err
}

func (this *Mongodb) Get() (*Mongodb, error) {
	this.Lock()
	defer this.Unlock()
	if err := this.conn.Ping(); err != nil {
		return nil, errors.Trace(err)
	}

	return &Mongodb{sync.RWMutex{}, this.conn.Copy(), this.conf}, nil
}

//获取多个实例
func (this *Mongodb) GetMul(size int) ([]*Mongodb, error) {
	this.Lock()
	defer this.Unlock()
	if err := this.conn.Ping(); err != nil {
		return nil, errors.Trace(err)
	}

	if size <= 0 {
		return nil, errors.New("参数不能小于0")
	}

	db := make([]*Mongodb, 0, size)

	for i := 0; i < size; i++ {
		db = append(db, &Mongodb{sync.RWMutex{}, this.conn.Copy(), this.conf})
	}

	return db, nil
}

func (this *Mongodb) GetConf() MongodbConf {
	return this.conf
}

func (this *Mongodb) Count(qconf MongodbQueryConf) (int, error) {
	c, err := this.conn.DB(qconf.Db).C(qconf.Table).Find(qconf.Query).Count()
	return c, errors.Trace(err)
}

func (this *Mongodb) One(qconf MongodbQueryConf) (info map[string]interface{}, err error) {
	err = this.conn.DB(qconf.Db).C(qconf.Table).Find(qconf.Query).Select(qconf.Select).One(&info)
	return
}

func (this *Mongodb) Query(qconf MongodbQueryConf, fun func(map[string]interface{})) error {
	iter := this.conn.DB(qconf.Db).C(qconf.Table).Find(qconf.Query).Select(qconf.Select).Iter()
	for {
		var info map[string]interface{}
		if !iter.Next(&info) {
			break
		}

		fun(info)
	}
	return errors.Trace(iter.Close())
}

func (this *Mongodb) Insert(qconf MongodbQueryConf) error {
	return errors.Trace(this.conn.DB(qconf.Db).C(qconf.Table).Insert(qconf.Insert...))
}

func (this *Mongodb) Create(qconf MongodbQueryConf) error {
	return errors.Trace(this.conn.DB(qconf.Db).C(qconf.Table).Create(&mgo.CollectionInfo{}))
}

func (this *Mongodb) Delete(qconf MongodbQueryConf) error {
	return errors.Trace(this.conn.DB(qconf.Db).C(qconf.Table).Remove(qconf.Delete))
}

func (this *Mongodb) Drop(qconf MongodbQueryConf) error {
	return errors.Trace(this.conn.DB(qconf.Db).C(qconf.Table).DropCollection())
}

func (this *Mongodb) Update(qconf MongodbQueryConf) error {
	return errors.Trace(this.conn.DB(qconf.Db).C(qconf.Table).Update(qconf.Query, qconf.Update))
}

func (this *Mongodb) UpdateAll(qconf MongodbQueryConf) (*mgo.ChangeInfo, error) {
	var c, err = this.conn.DB(qconf.Db).C(qconf.Table).UpdateAll(qconf.Query, qconf.Update)
	return c, errors.Trace(err)
}

func (this *Mongodb) Upsert(qconf MongodbQueryConf) (*mgo.ChangeInfo, error) {
	var c, err = this.conn.DB(qconf.Db).C(qconf.Table).Upsert(qconf.Query, qconf.Update)
	return c, errors.Trace(err)
}

func (this *Mongodb) EnsureIndex(qconf MongodbQueryConf) error {
	return errors.Trace(this.conn.DB(qconf.Db).C(qconf.Table).EnsureIndexKey(qconf.Index...))
}

func (this *Mongodb) Close() {
	this.conn.Close()
}

func (this *Mongodb) getColl(qconf MongodbQueryConf) *mgo.Collection {
	return this.conn.DB(qconf.Db).C(qconf.Table)
}

type MongodbBufferWriter struct {
	mdb     *Mongodb
	mdbconf MongodbQueryConf
}

func NewMongodbBufferWriter(mdb *Mongodb, conf MongodbQueryConf) *MongodbBufferWriter {
	return &MongodbBufferWriter{mdb, conf}
}

// 批量添加，到阀值自动insert
func (this *MongodbBufferWriter) Write(value interface{}, size int) {
	this.mdbconf.Insert = append(this.mdbconf.Insert, value)
	if len(this.mdbconf.Insert) == size {
		this.mdb.Insert(this.mdbconf)
		this.mdbconf.Insert = make([]interface{}, 0)
	}
}

func (this *MongodbBufferWriter) Flush() {
	this.mdb.Insert(this.mdbconf)
}

type MongodbBulkWriter struct {
	bulk *BufferedBulkInserter
}

func NewMongodbBulkWriter(mdb *Mongodb, conf MongodbQueryConf, size int) *MongodbBulkWriter {
	return &MongodbBulkWriter{NewBufferedBulkInserter(mdb.getColl(conf), size, false)}
}

// 批量添加，到阀值自动insert
func (this *MongodbBulkWriter) Write(value interface{}) {
	this.bulk.Insert(value)
}

func (this *MongodbBulkWriter) Flush() {
	this.bulk.Flush()
}
