package model

type HTbroadband struct {

}

func (this *HTbroadband) Has(account string) (bool, error) {
	sql := myorm.BSQL().Select("id").From("221su_broadband").Where("account=?").GetSQL()
	list, err := myorm.Get(sql, account)
	if err != nil || len(list) == 0 {
		return false, err
	}
	return true, nil
}
