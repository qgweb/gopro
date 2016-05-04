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
}

func NewORM() *QGORM {
	return &QGORM{}
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

func (this *QGORM) Query(osql string, data ...interface{}) ([]map[string]string, error) {
	var list = make([]map[string]string, 0)

	rows, err := this.db.Query(osql, data...)
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

func (this *QGORM) Update(sql string, data ...interface{}) (int64, error) {
	res, err := this.db.Exec(sql, data...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (this *QGORM) Insert(sql string, data ...interface{}) (int64, error) {
	res, err := this.db.Exec(sql, data...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (this *QGORM) Delete(sql string, data ...interface{}) (int64, error) {
	res, err := this.db.Exec(sql, data...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}


func (this *QGORM) BSQL() *BuildSQL {
	return &BuildSQL{}
}

func (this *QGORM) Begin() (*sql.Tx, error){
	return this.db.Begin()
}

