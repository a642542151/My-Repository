package controllers

import (
	"Project02/pyg/pyg/models"
	"encoding/json"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/utils"
	"github.com/garyburd/redigo/redis"
	"math"
	"math/rand"
	"regexp"
	"time"
)

type UserController struct {
	beego.Controller
}

//展示注册页面
func (this *UserController) ShowRegister() {
	this.TplName = "register.html"
}

func RespFunc(this *beego.Controller, resp map[string]interface{}) {
	//3.把容器传递给前段
	this.Data["json"] = resp
	//4.指定传递方式
	this.ServeJSON()
}

type Message struct {
	Message   string
	RequestId string
	BizId     string
	Code      string
}

//发送短信
func (this *UserController) HandleSendMsg() {
	//接受数据
	phone := this.GetString("phone")
	resp := make(map[string]interface{})

	defer RespFunc(&this.Controller, resp)
	//返回json格式数据
	//校验数据
	if phone == "" {
		beego.Error("获取电话号码失败")
		//2.给容器赋值
		resp["errno"] = 1
		resp["errmsg"] = "获取电话号码错误"
		return
	}
	//检查电话号码格式是否正确
	reg, _ := regexp.Compile(`^1[3-9][0-9]{9}$`)
	result := reg.FindString(phone)
	if result == "" {
		beego.Error("电话号码格式错误")
		//2.给容器赋值
		resp["errno"] = 2
		resp["errmsg"] = "电话号码格式错误"
		return
	}
	//发送短信   SDK调用
	client, err := sdk.NewClientWithAccessKey("cn-hangzhou", "LTAIu4sh9mfgqjjr", "sTPSi0Ybj0oFyqDTjQyQNqdq9I9akE")
	if err != nil {
		beego.Error("电话号码格式错误")
		//2.给容器赋值
		resp["errno"] = 3
		resp["errmsg"] = "初始化短信错误"
		return
	}
	//生成6位数随机数
	//rand.Seed(time.Now().UnixNano())

	//6位随机数
	//方法一：

	//var a []string
	//for i:=0;i<6;i++ {
	//	a = append(a,strconv.Itoa(rand.Intn(9)))
	//}
	//code := strings.Join(a,"")

	//方法二：
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	vcode := fmt.Sprintf("%06d", rnd.Int31n(1000000))

	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Scheme = "https" // https | http
	request.Domain = "dysmsapi.aliyuncs.com"
	request.Version = "2017-05-25"
	request.ApiName = "SendSms"
	request.QueryParams["RegionId"] = "cn-hangzhou"
	request.QueryParams["PhoneNumbers"] = phone
	request.QueryParams["SignName"] = "品优购"
	request.QueryParams["TemplateCode"] = "SMS_164275022"
	request.QueryParams["TemplateParam"] = `{"code":` + vcode + `}`

	response, err := client.ProcessCommonRequest(request)
	if err != nil {
		beego.Error("电话号码格式错误")
		//2.给容器赋值
		resp["errno"] = 4
		resp["errmsg"] = "短信发送失败"
		return
	}
	//json数据解析
	var message Message
	json.Unmarshal(response.GetHttpContentBytes(), &message)
	if message.Message != "OK" {
		beego.Error("电话号码格式错误")
		//2.给容器赋值
		resp["errno"] = 6
		resp["errmsg"] = message.Message
		return
	}

	resp["errno"] = 5
	resp["errmsg"] = "发送成功"
	resp["code"] = vcode

}

//处理注册业务
func (this *UserController) HandleRegister() {
	//获取数据
	phone := this.GetString("phone")
	pwd := this.GetString("password")
	rpwd := this.GetString("repassword")
	//校验数据
	if phone == "" || pwd == "" || rpwd == "" {
		beego.Error("获取数据错误1")
		this.Data["errmsg"] = "获取数据错误1"
		this.TplName = "register.html"
		return
	}
	if pwd != rpwd {
		beego.Error("两次密码输入不一致1")
		this.Data["errmsg"] = "两次密码输入不一致1"
		this.TplName = "register.html"
		return
	}
	//处理数据
	//向数据库中插入数据
	o := orm.NewOrm()
	var user models.User
	user.Name = phone
	user.Pwd = pwd
	user.Phone = phone
	o.Insert(&user)

	//跳转邮箱激活页面
	this.Ctx.SetCookie("userName", user.Name, 60*10)
	this.Redirect("/register-email", 302)

	//返回数据

}

//展示邮箱激活页面
func (this *UserController) ShowEmail() {
	this.TplName = "register-email.html"

}

//处理邮箱激活业务
func (this *UserController) HandleEmail() {
	//获取数据
	email := this.GetString("email")
	pwd := this.GetString("password")
	rpwd := this.GetString("repassword")
	//校验数据
	if email == "" || pwd == "" || rpwd == "" {
		beego.Error("获取数据错误2")
		this.Data["errmsg"] = "获取数据错误2"
		this.TplName = "register.html"
		return
	}
	if pwd != rpwd {
		beego.Error("两次密码输入不一致2")
		this.Data["errmsg"] = "两次密码输入不一致2"
		this.TplName = "register-email.html"
		return
	}
	//校验邮箱格式
	//吧字符串全部大写
	reg, _ := regexp.Compile(`^\w[\w\.-]*@[0-9a-z][0-9a-z-]*(\.[a-z]+)*\.[a-z]{2,6}$`)
	result := reg.FindString(email)
	if result == "" {
		beego.Error("邮箱格式错误")
		this.Data["errmsg"] = "邮箱格式错误"
		this.TplName = "register-email.html"
		return
	}
	//处理数据
	//发送邮件
	//utils   全局通用接口 工具类 邮箱配置
	config := `{"username":"czbkttsx@163.com","password":"czbkpygbj3q","host":"smtp.163.com","port":25}`
	emailReg := utils.NewEMail(config)
	//内容配置
	emailReg.Subject = "品优购用户激活"
	emailReg.From = "czbkttsx@163.com"
	emailReg.To = []string{email}
	userName := this.Ctx.GetCookie("userName")
	emailReg.HTML = `<a href="http://127.0.0.1:8081/active?userName=` + userName + `"> 点击激活该用户</a>`

	//发送邮件
	emailReg.Send()

	//插入邮箱  更新数据库邮箱界面
	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	err := o.Read(&user, "Name")
	if err != nil {
		beego.Error("邮箱插入错误", err)
		return
	}
	user.Email = email
	o.Update(&user)

	//返回数据
	this.Ctx.WriteString("邮件已发送，请去目标邮箱激活用户！")
}

//处理激活业务
func (this *UserController) Active() {
	//获取数据
	userName := this.GetString("userName")
	//校验数据
	if userName == "" {
		beego.Error("用户名错误")
		this.Redirect("/register-email", 302)
		return
	}
	//处理数据		本质上是更新 active
	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	err := o.Read(&user, "Name")
	if err != nil {
		beego.Error("用户名不存在")
		this.Redirect("/register-email", 302)
		return
	}
	user.Active = true
	o.Update(&user)
	//返回数据
	this.Redirect("/login", 302)
}

//展示登录页面
func (this *UserController) ShowLogin() {
	name := this.Ctx.GetCookie("LoginName")
	if name != "" {
		this.Data["checked"] = "checked"
	} else {
		this.Data["checked"] = ""
	}
	this.Data["name"] = name
	this.TplName = "login.html"
}

//处理登录业务
func (this *UserController) HandleLogin() {
	//获取数据
	userName := this.GetString("name")
	pwd := this.GetString("pwd")
	//校验数据
	if userName == "" || pwd == "" {
		this.Data["errmsg"] = "获取数据错误"
		this.TplName = "login.html"
		return
	}
	//处理数据
	o := orm.NewOrm()
	var user models.User
	reg, _ := regexp.Compile(`^\w[\w\.-]*@[0-9a-z][0-9a-z-]*(\.[a-z]+)*\.[a-z]{2,6}$`)
	result := reg.FindString(userName)

	if result != "" {
		user.Email = userName
		err := o.Read(&user, "Email")
		if err != nil {
			this.Data["errmsg"] = "邮箱未注册"
			this.TplName = "login.html"
			return
		}
		if user.Pwd != pwd {
			this.Data["errmsg"] = "邮箱密码错误"
			this.TplName = "login.html"
			return
		}
		//记住邮箱
		m1 := this.GetString("m1")
		if m1 == "2" {
			this.Ctx.SetCookie("LoginName", user.Email, 60*60)
		} else {
			this.Ctx.SetCookie("LoginName", user.Email, -1)
		}
	} else {
		user.Name = userName
		err := o.Read(&user, "Name")
		if err != nil {
			this.Data["errmsg"] = "用户名不存在"
			this.TplName = "login.html"
			return
		}
		if user.Pwd != pwd {
			this.Data["errmsg"] = "用户密码错误"
			this.TplName = "login.html"
			return
		}
		//记住用户名
		m1 := this.GetString("m1")
		if m1 == "2" {
			this.Ctx.SetCookie("LoginName", user.Name, 60*60)
		} else {
			this.Ctx.SetCookie("LoginName", user.Name, -1)
		}
	}

	//校验用户是否激活
	if user.Active == false {
		this.Data["errmsg"] = "当前用户未激活，请去邮箱激活！"
		this.TplName = "login.html"
		return
	}

	//返回数据
	this.SetSession("name", user.Name)
	this.Redirect("/index", 302)

}

//退出登录
func (this *UserController) Logout() {
	//删除 session 然后跳转登录页面
	this.DelSession("name")

	this.Redirect("/login", 302)
}

//展示用户中心页面
func (this *UserController) ShowUserCenterInfo() {
	//展示默认地址
	o:=orm.NewOrm()
	var user models.User
	//给查询对象赋值
	name := this.GetSession("name")
	user.Name = name.(string)
	o.Read(&user,"Name")
	this.Data["user"] = user


	//传地址
	var address models.Address
	qs:= o.QueryTable("Address").RelatedSel("User").Filter("User__Name",user.Name)
	qs.Filter("IsDefault",true).One(&address)
	this.Data["address"] = address
	/*//手机号加密
	qian := address.User.Phone[:3]
	hou := address.User.Phone[7:]
	address.User.Phone = qian + "****" + hou
	*/

	//从 redis中取出数据
	conn,err:= redis.Dial("tcp","127.0.0.1:6379")
	if err!= nil {
		beego.Error("redis 连接错误")
		return
	}
	defer conn.Close()
	goodsId,_ := redis.Ints(conn.Do("lrange","history_"+name.(string),0,4))
	var goods []models.GoodsSKU
	for _,v := range goodsId{
		var goodsSku models.GoodsSKU
		goodsSku.Id = v
		o.Read(&goodsSku)

		goods = append(goods,goodsSku)
	}

	this.Data["goods"] = goods

	this.Data["tplName"] = "个人信息"
	this.Layout = "user_center_layout.html"
	this.TplName = "user_center_info.html"

}

//展示收货地址页面
func (this *UserController) ShowSite() {
	//展示默认地址
	o := orm.NewOrm()
	var address models.Address
	//获取当前用户的默认地址
	name := this.GetSession("name")
	qs := o.QueryTable("Address").RelatedSel("User").Filter("User__Name", name.(string))
	qs.Filter("IsDefault", true).One(&address)

	/*//手机号加密
	qian := address.Phone[:3]
	hou := address.Phone[7:]
	address.Phone = qian + "****" + hou
	*/
	this.Data["tplName"] = "收货地址"
	this.Data["address"] = address
	this.Layout = "user_center_layout.html"
	this.TplName = "user_center_site.html"
}

//添加收货地址业务
func (this *UserController) HandleSite() {
	//获取数据
	receiver := this.GetString("receiver")
	addr := this.GetString("addr")
	postcode := this.GetString("postcode")
	phone := this.GetString("phone")
	//校验数据
	if receiver == "" || addr == "" || postcode == "" || phone == "" {
		this.Data["errmsg"] = "获取数据错误"
		this.TplName = "user_center_site.html"
		return
	}
	//操作数据
	o := orm.NewOrm()

	var userAddr models.Address
	userAddr.Receiver = receiver
	userAddr.Addr = addr
	userAddr.PostCode = postcode
	userAddr.Phone = phone

	//是哪个用户添加的地址？
	//获取当前登录用户的 name
	name := this.GetSession("name")
	var user models.User
	user.Name = name.(string)
	o.Read(&user, "Name")
	userAddr.User = &user

	//查询看有没有默认地址，如果有，把默认地址修改为非默认，如果没有，将插入的地址设置为默认地址
	//查询当前用户是否有默认地址
	var oldAddress models.Address
	qs := o.QueryTable("Address").RelatedSel("User").Filter("User__Name", name.(string))
	err := qs.Filter("IsDefault", true).One(&oldAddress)
	if err == nil {
		oldAddress.IsDefault = false
		o.Update(&oldAddress, "IsDefault")
	}
	userAddr.IsDefault = true

	_, err = o.Insert(&userAddr)
	if err != nil {
		beego.Info("插入数据错误", err)
	}
	//返回数据
	this.Redirect("/user/site", 302)

}

//展示用户中心订单页面
func (this *UserController)ShowUserOrder (){
//从数据库中获取当前用户所有订单信息
	name := this.GetSession("name")
	//获取订单信息
	o :=orm.NewOrm()
	var orderinfos []models.OrderInfo
	o.QueryTable("OrderInfo").RelatedSel("User").Filter("User__Name",name.(string)).OrderBy("-Time").All(&orderinfos)

	var orders []map[string]interface{}
	for _,v := range orderinfos{
		temp := make(map[string]interface{})
		//获取当前订单所有的订单商品
		var orderGoods []models.OrderGoods
		o.QueryTable("OrderGoods").RelatedSel("OrderInfo","GoodsSKU").Filter("OrderInfo__Id",v.Id).All(&orderGoods)

		temp["orderGoods"] = orderGoods
		temp["orderInfo"] = v

		orders = append(orders,temp)
	}

	//实现分页
	qs:= o.QueryTable("OrderInfo")
	count,_ := qs.Count()
	beego.Info(count)
	pageSize := 1
	pageCount := int(math.Ceil(float64(count) / float64(pageSize)))
	//获取当前页码
	pageIndex,err := this.GetInt("pageIndex")
	if err != nil {
		pageIndex = 1
	}
	pages := PageEdit(pageCount,pageIndex)
	this.Data["pages"] = pages

	//获取上一页，下一页的值
	var prePage,nextPage int
	//设置个范围
	if pageIndex -1 <= 0{
		prePage = 1
	}else {
		prePage = pageIndex - 1
	}


	if pageIndex +1 >= pageCount{
		nextPage = pageCount
	}else {
		nextPage = pageIndex + 1
	}


	this.Data["prePage"] = prePage
	this.Data["nextPage"] = nextPage

	qs = qs.Limit(pageSize,pageSize*(pageIndex - 1))

	//获取排序
	//if sort == ""{
	//	qs.All(&goods)
	//}else if sort == "price"{
	//	qs.OrderBy("Price").All(&goods)
	//}else {
	//	qs.OrderBy("-Sales").All(&goods)
	//}
	//
	//this.Data["sort"] = sort

	this.Data["pageIndex"] = pageIndex
	this.Data["orders"] = orders
	this.Data["tplName"] = "全部订单"
	this.Layout = "user_center_layout.html"
	this.TplName = "user_center_order.html"

}
