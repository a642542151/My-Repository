package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"newProject/controllers"
)

func init() {
	//路由过滤器，参数1 过滤匹配（支持正则） 参数2 过滤位置  参数3 过滤操作   参数是 context
	beego.InsertFilter("/article/*",beego.BeforeExec,filterFunc)


	beego.Router("/", &controllers.MainController{})
	//注册业务
	beego.Router("/register", &controllers.UserController{}, "get:ShowRegister;post:HandleRegister")
	//登录业务
	beego.Router("/login", &controllers.UserController{}, "get:ShowLogin;post:HandleLogin")
	//展示首页
	beego.Router("/article/index", &controllers.ArticleController{}, "get,post:ShowIndex")
	//添加文章业务
	beego.Router("/article/addArticle", &controllers.ArticleController{}, "get:ShowAddArticle;post:HandleAddArticle")
	//展示文章详情
	beego.Router("/article/content", &controllers.ArticleController{}, "get:ShowContent")
	//编辑文章内容
	beego.Router("/article/update",&controllers.ArticleController{},"get:ShowUpdate;post:HandleUpdate")
	//删除文章条目
	beego.Router("/article/delete",&controllers.ArticleController{},"get:HandleDelete")
	//展示添加文章类别页面
	beego.Router("/article/addType",&controllers.ArticleController{},"get:ShowAddType;post:HandleAddType")
	//删除文章类别条目
	beego.Router("/article/deleteType",&controllers.ArticleController{},"get:HandleTypeDelete")
	//退出登录
	beego.Router("/article/logout",&controllers.UserController{},"get:Logout")

}


func filterFunc (ctx *context.Context){
	//登录校验
	userName := ctx.Input.Session("userName")
	if userName == nil{
		//context 包中的跳转
		ctx.Redirect(302,"/login")
		return
	}
}
