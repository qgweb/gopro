package models

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
)

var (
	order_table  = "SHRTB_ORDER"
	order_prefix = "SORDER_"
)

type Order struct {
	Name  string                 `json:"name" form:"name" comm:"订单名称"`
	Price string                 `json:"price" form:"price" comm:"出价"`
	Size  string                 `json:"size" form:"size" comm:"投放尺寸"`
	Btime string                 `json:"btime" form:"btime" comm:"投放开始时间"`
	Etime string                 `json:"etime" form:"etime" comm:"投放结束时间"`
	Url   string                 `json:"url" form:"url" comm:"投放地址"`
	Purl  map[string]interface{} `json:"purls" form:"purls" comm:"投放地址"`
}

// 序列化
func (this *Order) marshal(or Order) ([]byte, error) {
	or.Purl = make(map[string]interface{})
	return json.Marshal(or)
}

// 活取列表
func (this *Order) List() (list []Order, err error) {
	client, err := dataDb.NewClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	vals, err := client.MultiHgetAll(order_table)
	if err != nil {
		return nil, err
	}
	list = make([]Order, 0, len(vals))
	for _, v := range vals {
		o := Order{}
		err := json.Unmarshal([]byte(v.String()), &o)
		if err != nil {
			continue
		}
		list = append(list, o)
	}
	return list, nil
}

// 添加订单
func (this *Order) Add(order Order) error {
	client, err := dataDb.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	v, err := this.marshal(order)
	if err != nil {
		return err
	}
	beego.Error(client.Hset(order_table, order.Name, string(v)))
	beego.Error(client.MultiHset(order_prefix+order.Name, order.Purl))
	// 推送到投放系统
	puturl := fmt.Sprintf("%s\t%s\t%s", order.Price, order.Size, order.Url)
	for v := range order.Purl {
		putDb.Sadd(v, puturl)
	}
	return nil
}

// 获取订单详情
func (this *Order) Get(name string) (o Order, err error) {
	client, err := dataDb.NewClient()
	if err != nil {
		return o, err
	}
	defer client.Close()
	v, err := client.Hget(order_table, name)
	if err != nil {
		return o, err
	}
	err = json.Unmarshal([]byte(v), &o)
	if err != nil {
		return o, err
	}

	v1, err1 := client.MultiHgetAll(order_prefix + o.Name)
	if err1 != nil {
		return o, err1
	}
	o.Purl = make(map[string]interface{}, len(v))
	for url := range v1 {
		o.Purl[url] = 1
	}
	return o, nil
}

// 删除订单
func (this *Order) Del(name string) error {
	client, err := dataDb.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()

	// 推送到投放系统
	norder, err := this.Get(name)
	if err != nil {
		return err
	}

	beego.Info(norder)
	puturl := fmt.Sprintf("%s\t%s\t%s", norder.Price, norder.Size, norder.Url)
	for v := range norder.Purl {
		putDb.Srem(v, puturl)
	}

	// 删除订单信息
	beego.Error(client.Hdel(order_table, name))
	beego.Error(client.Hclear(order_prefix + name))
	return nil
}
