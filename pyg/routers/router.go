package routers

import (
	"Project02/pyg/pyg/controllers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

func init() {
	//路由过滤器
	beego.InsertFilter("/user/*",beego.BeforeExec,filterFunc)

    beego.Router("/", &controllers.MainController{})
    //展示注册页面
    beego.Router("/register",&controllers.UserController{},"get:ShowRegister;post:HandleRegister")
	//发送短信
	beego.Router("/sendMsg",&controllers.UserController{},"post:HandleSendMsg")
	//展示邮箱激活页面
	beego.Router("/register-email",&controllers.UserController{},"get:ShowEmail;post:HandleEmail")
	//激活用户	active
	beego.Router("/active",&controllers.UserController{},"get:Active")
	//展示登录页面
	beego.Router("/login",&controllers.UserController{},"get:ShowLogin;post:HandleLogin")
	//退出登录
	beego.Router("/user/logout",&controllers.UserController{},"get:Logout")
	//展示用户中心页面
	beego.Router("/user/userCenterInfo",&controllers.UserController{},"get:ShowUserCenterInfo")
	//展示收货地址页面
	beego.Router("/user/site",&controllers.UserController{},"get:ShowSite;post:HandleSite")
	//展示用户中心订单页面
	beego.Router("/user/userOrder",&controllers.UserController{},"get:ShowUserOrder")


	//展示主页
	beego.Router("/index",&controllers.GoodsController{},"get:ShowIndex")
	//展示生鲜首页
	beego.Router("/index_sx",&controllers.GoodsController{},"get:ShowIndexSx")
	//商品详情
	beego.Router("/goodsDetail",&controllers.GoodsController{},"get:ShowDetail")
	//同一类型所有商品
	beego.Router("/goodsType",&controllers.GoodsController{},"get:ShowList")
	//商品搜索
	beego.Router("/search",&controllers.GoodsController{},"post:HandleSearch")


	//添加购物车
	beego.Router("/addCart",&controllers.CartController{},"post:HandleAddCart")
	//展示购物车
	beego.Router("/user/showCart",&controllers.CartController{},"get:ShowCart")
	//更改购物车数量-添加
	beego.Router("/upCart",&controllers.CartController{},"post:HandleUpCart")
	//删除购物车商品
	beego.Router("/deleteCart",&controllers.CartController{},"post:HandleDeleteCart")


	//添加商品到订单
	beego.Router("/user/addOrder",&controllers.OrderController{},"post:ShowOrder")
	//提交订单
	beego.Router("/pushOrder",&controllers.OrderController{},"post:HandlePushOrder")

	//支付
	beego.Router("/pay",&controllers.OrderController{},"get:Pay")


}

func filterFunc(ctx *context.Context){
	//过滤校验
	name := ctx.Input.Session("name")
	if name == nil {
		ctx.Redirect(302,"/login")
		return
	}
}
