package controllers

import (
	"Project02/pyg/pyg/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/garyburd/redigo/redis"
	"github.com/smartwalle/alipay"
	"strconv"
	"strings"
	"time"
)

type OrderController struct {
	beego.Controller
}

func (this *OrderController) ShowOrder() {
	//获取数据
	goodsIds := this.GetStrings("checkGoods")

	//校验数据
	if len(goodsIds) == 0 {
		this.Redirect("/user/showCart", 302)
		return
	}
	//处理数据
	//获取当前用户的所有收货地址
	name := this.GetSession("name")
	o := orm.NewOrm()
	var addrs []models.Address
	o.QueryTable("Address").RelatedSel("User").Filter("User__Name", name.(string)).All(&addrs)
	this.Data["addrs"] = addrs

	conn, _ := redis.Dial("tcp", ":6379")

	//获取商品,获取总价和总件数
	var goods []map[string]interface{}
	var totalPrice, totalCount int

	for _, v := range goodsIds {
		temp := make(map[string]interface{})
		id, _ := strconv.Atoi(v)
		var goodsSku models.GoodsSKU
		goodsSku.Id = id
		o.Read(&goodsSku)

		//获取商品数量
		count, _ := redis.Int(conn.Do("hget", "cart_"+name.(string), id))

		//计算小计
		littlePrice := count * goodsSku.Price

		//把商品信息放到行容器
		temp["goodsSku"] = goodsSku
		temp["count"] = count
		temp["littlePrice"] = littlePrice

		totalPrice += littlePrice
		totalCount += 1

		goods = append(goods, temp)

	}
	//返回数据
	this.Data["totalPrice"] = totalPrice
	this.Data["totalCount"] = totalCount
	this.Data["truePrice"] = totalPrice + 10
	this.Data["goods"] = goods
	this.Data["goodsIds"] = goodsIds
	this.TplName = "place_order.html"

}

//提交订单
func (this *OrderController) HandlePushOrder() {
	//获取数据
	addrId, err1 := this.GetInt("addrId")
	payId, err2 := this.GetInt("payId")
	goodsIds := this.GetString("goodsIds")
	totalCount, err3 := this.GetInt("totalCount")
	totalPrice, err4 := this.GetInt("totalPrice")
	beego.Info(goodsIds)
	resp := make(map[string]interface{})
	defer RespFunc(&this.Controller, resp)

	name := this.GetSession("name")
	if name == nil {
		resp["errno"] = 2
		resp["errmsg"] = "当前用户未登录"
		return
	}

	//校验数据
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || goodsIds == "" {
		resp["errno"] = 1
		resp["errmsg"] = "传输数据不完整"
		return
	}
	/*if err1 != nil{
		resp["errno"] = 1
		resp["errmsg"] = "传输数据不完整1"
		return
	}
	if err2 != nil{
		resp["errno"] = 1
		resp["errmsg"] = "传输数据不完整2"
		return
	}
	if err3 != nil{
		resp["errno"] = 1
		resp["errmsg"] = "传输数据不完整3"
		return
	}
	if err4 != nil{
		resp["errno"] = 1
		resp["errmsg"] = "传输数据不完整4"
		return
	}
	if goodsIds == ""{
		resp["errno"] = 1
		resp["errmsg"] = "传输数据不完整5"
		return
	}*/
	//处理数据
	//把数据插入到mysql数据库中
	//获取用户对象和地址对象
	o := orm.NewOrm()
	var user models.User
	user.Name = name.(string)
	o.Read(&user, "Name")

	var address models.Address
	address.Id = addrId
	o.Read(&address)

	var orderInfo models.OrderInfo

	orderInfo.User = &user
	orderInfo.Address = &address
	orderInfo.PayMethod = payId
	orderInfo.TotalCount = totalCount
	orderInfo.TotalPrice = totalPrice
	orderInfo.TransitPrice = 10
	orderInfo.OrderId = time.Now().Format("20060102150405" + strconv.Itoa(user.Id))

	//开始事物
	o.Begin()
	o.Insert(&orderInfo)

	conn, _ := redis.Dial("tcp", "127.0.0.1:6379")

	defer conn.Close()
	//插入订单商品
	//goodsIds  //2  3  5
	goodsSlice := strings.Split(goodsIds[1:len(goodsIds)-1], " ")
	for _, v := range goodsSlice {
		//插入订单商品表

		//获取商品信息
		id, _ := strconv.Atoi(v)
		var goodsSku models.GoodsSKU
		goodsSku.Id = id
		o.Read(&goodsSku)

		oldStock := goodsSku.Stock
		beego.Info("原始库存等于", oldStock)
		//获取商品数量
		count, _ := redis.Int(conn.Do("hget", "cart_"+name.(string), id))

		//获取小计
		littlePrice := goodsSku.Price * count

		//插入
		var orderGoods models.OrderGoods
		orderGoods.OrderInfo = &orderInfo
		orderGoods.GoodsSKU = &goodsSku
		orderGoods.Count = count
		orderGoods.Price = littlePrice
		//插入之前需要更新商品库存和销量
		if goodsSku.Stock < count {
			resp["errno"] = 4
			resp["errmsg"] = "库存不足"
			return
		}
		//goodsSku.Stock -= count
		//goodsSku.Sales += count
		//o.Update(&goodsSku)

		//time.Sleep(time.Second*5) //手动加延迟

		o.Read(&goodsSku)

		qs := o.QueryTable("GoodsSKU").Filter("Id", id).Filter("Stock", oldStock)
		_, err := qs.Update(orm.Params{"Stock": goodsSku.Stock - count, "Sales": goodsSku.Sales + count})

		if err != nil {
			resp["errno"] = 7
			resp["errmsg"] = "购买失败，请重新排队！"
			o.Rollback()
			return
		}

		_, err = o.Insert(&orderGoods)
		if err != nil {
			resp["errno"] = 3
			resp["errmsg"] = "服务器异常"
			return
		}
		_, err = conn.Do("hdel", "cart_"+name.(string), id)
		if err != nil {
			resp["errno"] = 6
			resp["errmsg"] = "清空购物车失败"
			o.Rollback()
			return
		}

	}

	//
	o.Commit()

	resp["errno"] = 5
	resp["errmsg"] = "OK"
}

//支付
func (this *OrderController) Pay() {
	//获取数据
	orderId, err := this.GetInt("orderId")
	if err != nil {

		this.Redirect("/user/userOrder", 302)
		return
	}

	//处理数据
	o := orm.NewOrm()
	var orderInfo models.OrderInfo
	orderInfo.Id = orderId
	o.Read(&orderInfo)

	//支付
	//appId, aliPublicKey, privateKey string, isProduction bool
	privateKey := `MIIEpQIBAAKCAQEAvTkK5ifvEsjU6IU0wzo9uTsLUeZEcgbewRzV6mR1OH3dmOem
mo352YsXdLq2ZFWvEVUc5w1cKyNdWKzfLQWgYS0zU3V6RQBG0+4WFfOd+9ON6GKm
f2dVleOlFWQB//tfQRqRbN3g0dWUZ4s3XimxXBByZg3Cqiv5IbYpkdjvxLqOkDG/
LHXyq3M4WUu0Q0VAl4sTIHsGOVhhGOZej3iUNaCWoA9ruX9pEpxreK+z0RyMUcNW
GiaTcfBCyY7nJ2Izu1N1weWfpSBDCnPzDo6cAR4y9G1hbSePMTjKK8zcWWSjO0Gu
wuYAScqQv5vmSNLkZxHx6Cjvq7jhri0ebAI++wIDAQABAoIBAQCJfpqJ1PimWKJE
dw540agqIVo/X6fah11zS0WxNN/sdaEAy0rHQWUMi0I3ArknvQ8h9Au1ZILVobPh
jHP6nf0Ev7hs64819laBBp6rwsLIStfxxUUgjHCnIqxBF9NQM1Lq1qhXR/5l2uEk
QAeyd28164mE2Hjb+Gnl8hzQqqbG9loItCpyvFckYTbC8kYFgLvhEINPsiMDFyju
bsNCeP9DU0VgJaOAN3vl9LyGIqy/0BmxvaCPIAp2PygMil6FXRBmDaYmqq1VZzBs
d9+hqya7PHW/wDyNUdVNeFGxtTkcs4GutXjG4DsepvKRHG0kK8f8kq+BbcSjW+Ff
eZrXpsfxAoGBAPiUGqKgRoykD845ylqfVF14ccR7S8PxEwhs2H0/dscvd5nwXwSo
P5vSzoreuavfWgBIPNt5sSoH9GdkefxKDng5J5hlLCMT4Y1RUqQyKfWdPLXYylV+
w67rZEohNMc5zk8qpWgrSRdicCByM23m3Mk17sPhHDzxxrOU7yiApv5ZAoGBAMLf
R8hzDfTcBR9jW7FoFT3VXjC/2ffm3F+XyKcJrgEwW2s559mspZ+v6EUUrwSBsn9i
K5FNm/l2O11o4rDyhUhEa2UB7xanJKCkJZeN/nRJKuPB4KmDigDMy+Mky8sI6KrJ
Kl8B1Tn5+uCtZ0aE42GsAKwYJgMoeCE7geMj8EVzAoGBALYeVBFX2bhKruXJg063
sui0SK3KI21QH0Cp9kZ1C8HNLhQTfpn75nZ0kSw/F8srXVYdlrC5zKndoBtsCs9j
NoywWykU3qxocXTG4wQ3WHSBmawlQ8A1mop6HUUOZQudd2Ca/wp9xBQk479xy+o2
HQYxxFewgq7H+Gszr7B96VspAoGBAL7X1/7w3nwsdT/WGFhXbGYP7ZykZpynFK7x
gOpFSomTiBQss2iz8ce/iCMPLI+nTN3/kFdOwC/AoEbMjyVnfSvXBa34BOQUcIR5
/O69erL7bOt8Vb7tOVurNQmQYZzHbsTDGaHNs7qBnDYo2/lt7xkaT9Y6GBADtBIn
qv59lbMNAoGAWtHNn/hchX+80xVDBgjzmJioJjhw7k7w06V0aNKKZITPynxg2nH1
A+itrxlfm5+3livWMDCFXWzMxI6wBTqHfxTm+TIe1PuCmvq3ogYt2Q05o0bDXk8S
YJvICnk8C8HF0/Sz8SJfH7gjhrcI45z2d6WPakdVmjaETdgZNws3hsI=`


	publicKey := `MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvTkK5ifvEsjU6IU0wzo9
uTsLUeZEcgbewRzV6mR1OH3dmOemmo352YsXdLq2ZFWvEVUc5w1cKyNdWKzfLQWg
YS0zU3V6RQBG0+4WFfOd+9ON6GKmf2dVleOlFWQB//tfQRqRbN3g0dWUZ4s3Ximx
XBByZg3Cqiv5IbYpkdjvxLqOkDG/LHXyq3M4WUu0Q0VAl4sTIHsGOVhhGOZej3iU
NaCWoA9ruX9pEpxreK+z0RyMUcNWGiaTcfBCyY7nJ2Izu1N1weWfpSBDCnPzDo6c
AR4y9G1hbSePMTjKK8zcWWSjO0GuwuYAScqQv5vmSNLkZxHx6Cjvq7jhri0ebAI+
+wIDAQAB`

	client := alipay.New("2016093000634080",publicKey,privateKey,false)
	var p = alipay.TradePagePay{}
	p.NotifyURL = "http://127.0.0.1:8080/payOK"
	p.ReturnURL = "http://127.0.0.1:8080/payOK"
	p.Subject = "品优购"
	p.OutTradeNo = orderInfo.OrderId
	p.TotalAmount = strconv.Itoa(orderInfo.TotalPrice)
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"
	url, err := client.TradePagePay(p)
	if err != nil {
		beego.Error("支付失败")
	}
	payUrl := url.String()

	orderInfo.Orderstatus = 1
	o.Update(&orderInfo,"Orderstatus")

	this.Redirect(payUrl,302)

}
