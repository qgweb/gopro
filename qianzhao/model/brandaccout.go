package model

//宽带账户|属地|
//校园名称|校园组别|上行带宽|下
//行带宽
type BrandAccount struct {
	Id            string `json:"id"`              //编号
	Province      string `json:"province"`        //属地（带争议）
	Account       string `json:"account"`         //宽带账户
	SchoolName    string `json:"school_name"`     //校园名称
	SchoolGroup   string `json:"school_group"`    //校园组别
	UpBroadband   string `json:"broadband_up"`    //上行带宽
	DownBroadband string `json:"broadboand_down"` //下行带宽
}
