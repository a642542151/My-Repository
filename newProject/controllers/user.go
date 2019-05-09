package controllers

import (
	"encoding/base64"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"newProject/models"
)

type UserController struct {
	beego.Controller
}

func (r *UserController) ShowRegister() {
	r.TplName = "register.html"

}

func (r *UserController) HandleRegister() {
	//获取数据
	userName := r.GetString("userName")
	pwd := r.GetString("password")
	//校验数据
	if userName == "" || pwd == "" {
		beego.Error("输入数据不完整")
		r.TplName = "register.html"
		return
	}
	//处理数据
	o := orm.NewOrm()
	var user models.UserData
	user.Name = userName
	user.Pwd = pwd
	id, err := o.Insert(&user)
	if err != nil {
		beego.Error("用户注册失败")
		r.TplName = "register.html"
		return
	}
	beego.Info("插入数据的 id 为", id)
	//返回数据
	r.Redirect("/article/login",302)
}

func (this *UserController) ShowLogin() {
	//获取 cookie 数据，如果获取查到了，说明上一次记住用户名，反之，不记住用户名
	userName := this.Ctx.GetCookie("userName")
	//解密
	dec,_ := base64.StdEncoding.DecodeString(userName)
	if userName != ""{
		this.Data["userName"] = string(dec)
		this.Data["checked"] = "checked"
	}else {
		this.Data["userName"] = ""
		this.Data["checked"] = ""
	}
	this.TplName = "login.html"
}

func (this *UserController) HandleLogin() {
	//获取数据
	userName := this.GetString("userName")
	pwd := this.GetString("password")
	//校验数据
	if userName == "" || pwd == "" {
		beego.Error("数据输入不完整")
		this.TplName = "login.html"
		return
	}
	//处理数据
	o := orm.NewOrm()
	var user models.UserData
	user.Name = userName
	err := o.Read(&user, "Name")
	if err != nil {
		beego.Error("用户不存在")
		this.TplName = "login.html"
		return
	}
	if user.Pwd != pwd {
		beego.Error("密码错误")
		this.TplName = "login.html"
		return
	}
	//实现记住用户名功能。
	//设置 cookie
	remember:= this.GetString("remember")
	//给 userName 加密
	enc := base64.StdEncoding.EncodeToString([]byte(userName))
	if remember =="on"{
		this.Ctx.SetCookie("userName",enc,60)
	}else {
		this.Ctx.SetCookie("userName",userName,-1)
	}

	//session 存储
	this.SetSession("userName",userName)
	//返回数据
	this.Redirect("/article/index",302)
}

//退出登录
func(this *UserController)Logout(){
	//删除 session 然后跳转登录页面
	this.DelSession("userName")

	this.Redirect("/login",302)
}
