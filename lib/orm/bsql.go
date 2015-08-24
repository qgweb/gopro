//构建SQL
package orm

import (
	"strconv"
	"strings"
)

type BuildSQL struct {
	sql string
}

func (this *BuildSQL) Select(fields ...string) *BuildSQL {
	this.sql += "SELECT " + strings.Join(fields, ",")
	return this
}

func (this *BuildSQL) From(fields string) *BuildSQL {
	this.sql += " FROM " + fields
	return this
}

func (this *BuildSQL) Having(fields string) *BuildSQL {
	this.sql += " HAVING " + fields
	return this
}

func (this *BuildSQL) Order(fields string) *BuildSQL {
	this.sql += " ORDER BY " + fields
	return this
}

func (this *BuildSQL) Group(fields string) *BuildSQL {
	this.sql += " GROUP BY " + fields
	return this
}

func (this *BuildSQL) Limit(b ...int) *BuildSQL {
	if len(b) == 2 {
		this.sql += " LIMIT " + strconv.Itoa(b[0]) + "," + strconv.Itoa(b[1])
	} else if len(b) == 1 {
		this.sql += " LIMIT " + strconv.Itoa(b[0])
	}

	return this
}

func (this *BuildSQL) Where(fields string) *BuildSQL {
	this.sql += " WHERE " + fields
	return this
}

func (this *BuildSQL) And(fields string) *BuildSQL {
	this.sql += " AND " + fields
	return this
}

func (this *BuildSQL) Or(fields string) *BuildSQL {
	this.sql += " OR " + fields
	return this
}

func (this *BuildSQL) Like(fields string) *BuildSQL {
	this.sql += " " + fields + " LIKE ?"
	return this
}

func (this *BuildSQL) Set(fields ...string) *BuildSQL {
	for k, v := range fields {
		fields[k] = v + "=?"
	}

	this.sql += " SET " + strings.Join(fields, ",")
	return this
}

func (this *BuildSQL) Update(fields string) *BuildSQL {
	this.sql += " UPDATE " + fields
	return this
}

func (this *BuildSQL) Insert(fields string) *BuildSQL {
	this.sql += " INSERT INTO " + fields
	return this
}

func (this *BuildSQL) Values(fields ...string) *BuildSQL {
	var vlen = []rune(strings.Repeat("?,", len(fields)))
	this.sql += "(" + strings.Join(fields, ",") + ") VALUES(" + string(vlen[0:len(vlen)-1]) + ")"
	return this
}

func (this *BuildSQL) Fields(fields ...string) *BuildSQL {
	this.sql += " (" + strings.Join(fields, ",") + ")"
	return this
}

func (this *BuildSQL) Reset() *BuildSQL {
	this.sql = ""
	return this
}

func (this *BuildSQL) SetSQL(sql string) {
	this.sql = sql
}

func (this *BuildSQL) GetSQL() string {
	return this.sql
}
