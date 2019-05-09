package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"math"
	"newProject/models"
	"path"
	"strconv"
	"time"
)

type ArticleController struct {
	beego.Controller
}

//展示首页
func (this *ArticleController) ShowIndex() {
	//校验登录状态
	userName := this.GetSession("userName")
	if userName == nil {
		this.Redirect("/article/login", 302)
		return
	}

	//断言
	this.Data["userName"] = userName.(string)

	//获取所有文章数据，展示到页面

	o := orm.NewOrm()
	qs := o.QueryTable("Article")
	var articles []models.Article
	//qs.All(&articles)

	//获取选中的类型
	typeName := this.GetString("select")
	var count int64
	if typeName == "" {
		//获取总记录数
		count, _ = qs.RelatedSel("ArticleType").Count()
	} else {
		count, _ = qs.RelatedSel("ArticleType").Filter("ArticleType__TypeName", typeName).Count()
	}
	//获取总页数
	pageIndex := 2
	pageCount := math.Ceil(float64(count) / float64(pageIndex))
	//获取首页跟末页数据
	//获取页码
	pageNum, err := this.GetInt("pageNum")
	if err != nil {
		pageNum = 1
	}
	beego.Info("数据总页数末:", pageNum)

	//获取对应页的数据   获取几条数据   起始位置
	//ORM多表查询的时候默认是惰性查询，关联查询之后，如果关联字段为空，字段查询不到

	//where ArticleType.typeName = typename  filter 相当于 Where
	if typeName == "" {
		qs.Limit(pageIndex, pageIndex*(pageNum-1)).RelatedSel("ArticleType").All(&articles)
	} else {
		qs.Limit(pageIndex, pageIndex*(pageNum-1)).RelatedSel("ArticleType").Filter("ArticleType__TypeName", typeName).All(&articles)
	}
	//查询所有文章类型，并展示
	var articleTypes []models.ArticleType
	o.QueryTable("ArticleType").All(&articleTypes)

	this.Data["articleTypes"] = articleTypes

	//传入数据
	this.Data["articles"] = articles
	this.Data["count"] = count
	this.Data["pageCount"] = pageCount
	this.Data["pageNum"] = pageNum
	//传输给后台的数据，$.TypeName  加$来区分是后台传来的数据，在视图条件判断中进行比较
	this.Data["TypeName"] = typeName
	this.Layout = "layout.html"
	this.LayoutSections = make(map[string]string)
	this.LayoutSections["indexJs"] = "indexJs.html"
	this.TplName = "index.html"

}

//展示添加页面
func (this *ArticleController) ShowAddArticle() {
	//获取所有类型并绑定下拉框
	o := orm.NewOrm()
	var articleTypes []models.ArticleType
	o.QueryTable("ArticleType").All(&articleTypes)

	this.Data["articleTypes"] = articleTypes

	this.Layout = "layout.html"
	this.TplName = "add.html"
}

//处理添加文章业务
func (this *ArticleController) HandleAddArticle() {
	//获取数据
	articleTitle := this.GetString("articleName")
	content := this.GetString("content")
	typeName := this.GetString("select")

	//校验数据
	if articleTitle == "" || content == "" {
		beego.Error("获取文章数据错误")
		this.Data["errmsg"] = "获取数据错误"
		this.Layout = "layout.html"
		this.TplName = "add.html"
		return
	}
	//获取图片
	file, head, err := this.GetFile("uploadname")
	if err != nil {
		beego.Error("获取图片数据错误")
		this.Data["errmsg"] = "图片上传失败"
		this.Layout = "layout.html"
		this.TplName = "add.html"
		return
	}
	defer file.Close()
	//校验图片大小
	if head.Size > 5000000 {
		beego.Error("获取图片数据过大")
		this.Data["errmsg"] = "图片数据过大"
		this.Layout = "layout.html"
		this.TplName = "add.html"
		return
	}
	//校验图片格式 获取图片文件后缀
	ext := path.Ext(head.Filename)
	if ext != ".jpg" && ext != ".png" && ext != ".jpeg" {
		beego.Error("获取图片格式错误")
		this.Data["errmsg"] = "上传图片格式错误"
		this.Layout = "layout.html"
		this.TplName = "add.html"
		return
	}
	//防止重名
	fileName := time.Now().Format("2006010215040522")
	//根据公司业务 校验其他 操作

	//把上传的文件存储到项目文件夹	//html 中上传文件按钮上传的 name ，第二参数存储的文件路径+防止重名的文件名+文件后缀
	this.SaveToFile("uploadname", "./static/img/"+fileName+ext)
	//处理数据
	//把数据存到数据库
	//创建对象
	o := orm.NewOrm()
	//新建结构体变量
	var article models.Article
	//给结构体变量赋值
	article.Title = articleTitle
	article.Content = content
	article.Img = "/static/img/" + fileName + ext
	//获取文章类型对象，并插入到文章中
	var articleType models.ArticleType
	articleType.TypeName = typeName
	o.Read(&articleType, "TypeName")

	article.ArticleType = &articleType
	//插入数据
	_, err = o.Insert(&article)
	if err != nil {
		beego.Error("获取数据错误", err)
		this.Data["errmsg"] = "数据插入失败"
		this.Layout = "layout.html"
		this.TplName = "add.html"
		return
	}

	//返回数据  跳转页面
	this.Redirect("/article/index", 302)

}

//展示文章内容页面
func (this *ArticleController) ShowContent() {
	//获取数据
	id, err := this.GetInt("id")
	//校验数据
	if err != nil {
		beego.Error("获取 id 数据错误")
		this.Redirect("/article/index", 302)
		return
	}
	//处理数据
	//查找文章数据
	o := orm.NewOrm()
	//获取查询对象
	var article models.Article
	//给查询条件赋值
	article.Id = id
	//查询
	o.Read(&article)

	//
	var users []models.UserData
	o.QueryTable("UserData").Filter("Articles__Article__Id", id).Distinct().All(&users)
	this.Data["users"] = users

	//给更新条件赋值
	article.ReadCount += 1
	o.Update(&article)

	//返回数据
	this.Data["article"] = article

	//插入多对多关系   根据用户名获取用户对象
	userName := this.GetSession("userName")
	var user models.UserData
	user.Name = userName.(string)
	o.Read(&user, "Name")
	//多对多的插入操作
	//获取 orm 对象
	//前面几十行  已经获取过 orm 对象 （o := orm.NewOrm()
	//获取被插入数据的对象(文章)
	//前面几十行  已经获取过 被插入数据的对象（文章）（var article models.Article
	//获取多对多操作对象
	m2m := o.QueryM2M(&article, "Users")
	//用多对多操作对象插入
	m2m.Add(user)

	this.Layout = "layout.html"
	this.TplName = "content.html"

}

//编辑文章页面
func (this *ArticleController) ShowUpdate() {
	//获取数据
	id, err := this.GetInt("id")
	//校验数据
	if err != nil {
		beego.Error("获取 id 数据错误")
		this.Redirect("/article/index", 302)
		return
	}
	//处理数据

	//查找文章数据
	o := orm.NewOrm()
	//获取查询对象
	var article models.Article
	article.Id = id
	o.Read(&article)

	//返回数据
	this.Data["article"] = article
	this.Layout = "layout.html"
	this.TplName = "update.html"
}

//封装文件上传处理函数
func UploadFile(this *ArticleController, filePath string, errHtml string) string {
	//获取图片
	//返回值 文件二进制流  文件头    错误信息
	file, head, err := this.GetFile(filePath)
	if err != nil {
		beego.Error("获取数据错误")
		this.Data["errmsg"] = "图片上传失败"
		this.TplName = errHtml
		return ""
	}
	defer file.Close()
	//校验文件大小
	if head.Size > 5000000 {
		beego.Error("获取数据错误")
		this.Data["errmsg"] = "图片数据过大"
		this.TplName = errHtml
		return ""
	}

	//校验格式 获取文件后缀
	ext := path.Ext(head.Filename)
	if ext != ".jpg" && ext != ".png" && ext != ".jpeg" {
		beego.Error("获取数据错误")
		this.Data["errmsg"] = "上传文件格式错误"
		this.TplName = errHtml
		return ""
	}

	//防止重名
	fileName := time.Now().Format("200601021504052222")

	//其他校验操作

	//把上传的文件存储到项目文件夹
	this.SaveToFile(filePath, "./static/img/"+fileName+ext)
	return "/static/img/" + fileName + ext
}

//处理文章编辑业务
func (this *ArticleController) HandleUpdate() {
	//获取数据
	articleName := this.GetString("articleName")
	content := this.GetString("content")
	savePath := UploadFile(this, "uploadname", "update.html")
	id, _ := this.GetInt("id") //隐藏域传值
	//校验数据
	if articleName == "" || content == "" || savePath == "" {
		beego.Error("获取数据失败")
		this.Redirect("/article/update?id="+strconv.Itoa(id), 302)
		return
	}
	//处理数据
	//更新操作
	o := orm.NewOrm()
	var article models.Article
	//先查询更新的文章是否存在  如果不查询是否存在，更新操作提交时会报错
	article.Id = id
	//必须查询
	o.Read(&article)
	//更新  需要先赋新值  beego 中的 orm 如果需要更新，更新的对象 id 必须有值
	article.Title = articleName
	article.Content = content
	article.Img = savePath
	o.Update(&article)

	//返回数据
	this.Redirect("/article/index", 302)
}

//删除文章条目
func (this *ArticleController) HandleDelete() {
	//获取数据
	id, err := this.GetInt("id")
	//校验数据
	if err != nil {
		beego.Error("获取 id 错误")
		this.Redirect("/article/index", 302)
		return
	}
	//处理数据
	o := orm.NewOrm()
	var article models.Article
	article.Id = id
	o.Delete(&article)
	//返回数据
	this.Redirect("/article/index", 302)
}

//展示文章类别页面
func (this *ArticleController) ShowAddType() {
	//获取所有类型，并展示在页面上
	o := orm.NewOrm()
	var articleTypes []models.ArticleType
	o.QueryTable("ArticleType").OrderBy("Id").All(&articleTypes)

	this.Data["articleTypes"] = articleTypes
	this.Layout = "layout.html"
	this.TplName = "addType.html"
}

//处理添加类型操作
func (this *ArticleController) HandleAddType() {
	//获取数据
	typeName := this.GetString("typeName")
	//校验数据
	if typeName == "" {
		beego.Error("类型名称传输失败")
		this.Redirect("/article/addType", 302)
		return
	}

	//处理数据
	//插入操作
	o := orm.NewOrm()
	var articleType models.ArticleType
	articleType.TypeName = typeName
	o.Insert(&articleType)

	//返回数据
	this.Redirect("/article/addType", 302)
}

//删除文章类型条目
func (this *ArticleController) HandleTypeDelete() {
	//获取数据
	id, err := this.GetInt("id")
	//校验数据
	if err != nil {
		beego.Error("获取 id 错误")
		this.Redirect("/article/addType", 302)
		return
	}

	//处理数据
	o := orm.NewOrm()
	var articleType models.ArticleType
	articleType.Id = id
	o.Delete(&articleType)

	//返回数据
	this.LayoutSections = make(map[string]string)
	this.LayoutSections["addTypeJs"] = "addTypeJs.html"

	this.Redirect("/article/addType", 302)

}
