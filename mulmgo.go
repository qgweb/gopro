package main
import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"time"
	"fmt"
	"github.com/qgweb/gopro/lib/convert"
)

func main() {
	db,_:= leveldb.OpenFile("/tmp", &opt.Options{BlockCacheCapacity: 64,
		WriteBuffer: 32 * 1024 * 1024})
	bt:=time.Now()
	//sync := &opt.WriteOptions{Sync: true}
	for i:=0;i<100000;i++ {
		key := []byte(convert.ToString(i))
		db.Has(key,nil)
		db.
		//db.Put(key,key,nil)
	}
	fmt.Println(time.Now().Sub(bt).Seconds())
}
