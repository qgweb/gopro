package orm

import (
	"database/sql"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

type QGORM struct {
	mutex sync.RWMutex
	db    *sql.DB
	debug bool
	bs    *BuildSQL
}

func NewORM() *QGORM {
	return &QGORM{bs: &BuildSQL{}}
}

func (this *QGORM) Open(driveSql string) error {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	var err error
	this.db, err = sql.Open("mysql", driveSql)
	if err != nil {
		return err
	}
	return nil
}

func (this *QGORM) SetMaxIdleConns(c int) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.db.SetMaxIdleConns(c)
}

func (this *QGORM) SetMaxOpenConns(c int) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.db.SetMaxOpenConns(c)
}

func (this *QGORM) Debug(d bool) {
	this.debug = d
}

func (this *QGORM) Query(data ...interface{}) ([]map[string]string, error) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	var list = make([]map[string]string, 0)

	rows, err := this.db.Query(this.bs.GetSQL(), data...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	col, _ := rows.Columns()

	var trows = make([]interface{}, len(col))
	for i := 0; i < len(col); i++ {
		var c sql.RawBytes
		trows[i] = &c
	}

	for rows.Next() {
		err := rows.Scan(trows...)
		if err != nil {
			return nil, err
		}

		var d = make(map[string]string)
		for i := 0; i < len(col); i++ {
			d[col[i]] = string(*trows[i].(*sql.RawBytes))
		}
		list = append(list, d)
	}

	return list, nil
}

func (this *QGORM) Update(data ...interface{}) (int64, error) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	res, err := this.db.Exec(this.bs.GetSQL(), data...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (this *QGORM) Insert(data ...interface{}) (int64, error) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	res, err := this.db.Exec(this.bs.GetSQL(), data...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (this *QGORM) Delete(data ...interface{}) (int64, error) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	res, err := this.db.Exec(this.bs.GetSQL(), data...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (this *QGORM) LastSql() string {
	return this.bs.GetSQL()
}

func (this *QGORM) BSQL() *BuildSQL {
	this.bs.Reset()
	return this.bs
}
