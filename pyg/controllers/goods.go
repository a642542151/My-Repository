package controllers

import (
	"Project02/pyg/pyg/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/garyburd/redigo/redis"
	"math"
)

type GoodsController struct {
	beego.Controller
}

func (this *GoodsController) ShowIndex() {
	name := this.GetSession("name")
	if name != nil {
		this.Data["name"] = name.(string)
	} else {
		this.Data["name"] = ""
	}

	//三级联动
	//获取类型信息并传递给前端
	//获取一级菜单
	o := orm.NewOrm()
	//新建接收对象
	var oneclass []models.TpshopCategory
	//将一级菜单的数据存放至 oneClass
	o.QueryTable("TpshopCategory").Filter("Pid", 0).All(&oneclass)

	//定义一个总容器
	var types []map[string]interface{}
	//遍历取出所有一级菜单的条目
	for _, v := range oneclass {
		//定义一个行容器
		//行容器中存放一个一级菜单的名和多个二级菜单的名称
		t := make(map[string]interface{})
		//定义二级菜单
		var secondClass []models.TpshopCategory
		//将二级菜单的数据存放至 secondClass
		o.QueryTable("TpshopCategory").Filter("Pid", v.Id).All(&secondClass)
		//行容器中t1存放一级菜单的所有数据
		t["t1"] = v
		//行容器中t2存放二级菜单的所有数据
		t["t2"] = secondClass
		//将行容器添加到总容器中
		types = append(types, t)
	}

	//获取第三级菜单
	//遍历总容器,取出总容器types 中的t2
	for _, v1 := range types {
		//定义一个二级菜单的容器
		var erji []map[string]interface{}
		//再次遍历types中的t2取出所有的二级菜单
		for _, v2 := range v1["t2"].([]models.TpshopCategory) {
			//定义一个二级菜单的行容器
			//二级菜单的行容器中存放一个二级菜单的名称和多个三级菜单的名称
			t := make(map[string]interface{})
			//定义三级菜单
			var thirdClass []models.TpshopCategory
			//获取三级菜单
			o.QueryTable("TpshopCategory").Filter("Pid", v2.Id).All(&thirdClass)
			//将再次获取到的二级菜单保存到 t22中
			t["t22"] = v2
			//将刚刚获取到的三级菜单保存到 t23中
			t["t23"] = thirdClass
			//将二级菜单的行容器添加到二级菜单容器中
			erji = append(erji, t)
		}
		//将二级菜单容器添加到总容器中
		v1["t3"] = erji
	}

	this.Data["types"] = types
	this.TplName = "index.html"

}

//展示生鲜首页
func (this *GoodsController) ShowIndexSx() {
	//获取生鲜首页内容
	//获取商品类型
	o:= orm.NewOrm()
	//获取所有商品类型
	var goodsTypes []models.GoodsType
	o.QueryTable("GoodsType").All(&goodsTypes)
	this.Data["goodsTypes"] = goodsTypes

	//获取轮播图片
	var goodsBanners []models.IndexGoodsBanner
	o.QueryTable("IndexGoodsBanner").All(&goodsBanners)
	this.Data["goodsBanners"] = goodsBanners

	//获取促销商品
	var promotionBanners []models.IndexPromotionBanner
	o.QueryTable("IndexPromotionBanner").All(&promotionBanners)
	this.Data["promotions"] = promotionBanners

	//获取首页商品展示
	//创建总容器
	var goods []map[string]interface{}
	//循环获取商品分类
	for _,v := range goodsTypes{
		//创建接收首页文字商品的对象
		var textGoods []models.IndexTypeGoodsBanner
		//创建接收首页图片商品的对象
		var imageGoods []models.IndexTypeGoodsBanner
		//获取首页商品展示的所有相关数据
		qs:= o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsSKU","GoodsType").Filter("GoodsType__Id",v.Id).OrderBy("Index")
		//获取文字商品
		qs.Filter("DisplayType",0).All(&textGoods)
		//获取图片商品
		qs.Filter("DisplayType",1).All(&imageGoods)

		//定义行容器
		temp := make(map[string]interface{})
		//将首页商品分类添加到行容器中
		temp["goodsType"] = v
		//将首页展示文字商品添加到行容器中
		temp["textGoods"] = textGoods
		//将首页展示图片商品添加到行容器中
		temp["imageGoods"] = imageGoods

		//把行容器添加到总容器中
		goods = append(goods,temp)
	}

	this.Data["goods"] = goods

	this.TplName = "index_sx.html"

}

//独立于beego框架的
func PageEdit(pageCount int,pageIndex int)[]int{
	//不足五页
	var pages []int
	if pageCount < 5{
		for i:=1;i<=pageCount;i++{
			pages = append(pages,i)
		}
	}else if pageIndex <= 3{
		for i:=1;i<=5;i++{
			pages = append(pages,i)
		}
	}else if pageIndex >= pageCount -2{
		for i:=pageCount - 4;i<=pageCount;i++{
			pages = append(pages,i)
		}
	}else {
		for i:=pageIndex - 2;i<=pageIndex + 2;i++{
			pages = append(pages,i)
		}
	}

	return pages
}

//展示商品详情页
func(this *GoodsController)ShowDetail(){
	id,err := this.GetInt("Id")
	//校验数据
	if err!= nil {
		beego.Error("商品链接错误")
		this.Redirect("/index_sx",302)
		return
	}
	//处理数据
	//根据 id 获取商品有关数据
	o:= orm.NewOrm()
	var goodsSku models.GoodsSKU
	/*goodsSku.Id = id
	o.Read(&goodsSku)*/
	//获取商品详情
	o.QueryTable("GoodsSKU").RelatedSel("Goods","GoodsType").Filter("Id",id).One(&goodsSku)

	//获取同一类型的新品推荐
	var newGoods []models.GoodsSKU
	qs := o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Name",goodsSku.GoodsType.Name)
	qs.OrderBy("-Time").Limit(2,0).All(&newGoods)
	this.Data["newGoods"] = newGoods

	//获取所有商品类型
	var goodsTypes []models.GoodsType
	o.QueryTable("GoodsType").All(&goodsTypes)
	this.Data["goodsTypes"] = goodsTypes

	//存储浏览记录
	name := this.GetSession("name")
	if name != nil {
		//把历史浏览记录存储在redis中
		conn,err := redis.Dial("tcp","127.0.0.1:6379")
		if err == nil {
			defer conn.Close()
			conn.Do("lrem","history_"+name.(string),0,id)
			_,err = conn.Do("lpush","history_"+name.(string),id)
			beego.Info(err)
		}
	}

	//传递数据
	this.Layout = "sx_layout.html"
	this.LayoutSections = make(map[string]string)
	this.LayoutSections["detailJS"] = "detailJS.html"
	this.Data["goodsSku"] = goodsSku
	this.TplName="detail.html"
}

//展示商品列表页
func(this*GoodsController)ShowList(){
	//获取数据
	id,err := this.GetInt("id")
	//校验数据
	if err != nil {
		beego.Error("类型不存在")
		this.Redirect("/index_sx",302)
		return
	}
	//处理数据
	o := orm.NewOrm()
	var goods []models.GoodsSKU
	//获取排序方式
	sort := this.GetString("sort")

	//实现分页

	qs := o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",id)
	//获取总页码
	count,_ := qs.Count()
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
	if sort == ""{
		qs.All(&goods)
	}else if sort == "price"{
		qs.OrderBy("Price").All(&goods)
	}else {
		qs.OrderBy("-Sales").All(&goods)
	}

	this.Data["sort"] = sort

	//var goodsSku models.GoodsSKU
	/*goodsSku.Id = id
	o.Read(&goodsSku)*/
	//获取商品详情
	//o.QueryTable("GoodsSKU").RelatedSel("Goods","GoodsType").Filter("Id",id).One(&goodsSku)
	//获取所有商品类型
	var goodsTypes []models.GoodsType
	o.QueryTable("GoodsType").All(&goodsTypes)
	this.Data["goodsTypes"] = goodsTypes

	//获取同一类型的新品推荐
	var newGoods []models.GoodsSKU
	qs = o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",id)
	qs.OrderBy("-Time").Limit(2,0).All(&newGoods)
	this.Data["newGoods"] = newGoods


	//返回数据
	this.Layout = "sx_layout.html"
	this.Data["pageIndex"] = pageIndex
	this.Data["id"] = id
	this.Data["goods"] = goods
	this.TplName = "list.html"
}

//搜索页面
func(this*GoodsController)HandleSearch(){
	//获取数据
	goodsName := this.GetString("goodsName")
	//校验数据
	if goodsName == ""{
		this.Redirect("/index_sx",302)
		return
	}
	//处理数据
	o := orm.NewOrm()
	var goods []models.GoodsSKU
	//模糊查询
	o.QueryTable("GoodsSKU").Filter("Name__icontains",goodsName).All(&goods)

	//获取所有商品类型
	var goodsTypes []models.GoodsType
	o.QueryTable("GoodsType").All(&goodsTypes)
	this.Data["goodsTypes"] = goodsTypes

	//返回数据
	this.Layout = "sx_layout.html"
	this.Data["goods"] = goods
	this.TplName = "search.html"
}
