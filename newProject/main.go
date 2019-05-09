package main

import (
	"github.com/astaxie/beego"
	_ "newProject/models"
	_ "newProject/routers"
)

func main() {
	//给视图函数建立映射
	beego.AddFuncMap("prePage",PrePage)
	beego.AddFuncMap("nextPage",NextPage)

	beego.Run()
}

func PrePage(pageNum int) int {
	if pageNum <= 1 {
		return 1
	}
	return pageNum - 1
}

func NextPage(pageNum int, pageCount float64) int {
	if pageNum >= int(pageCount) {
		return int(pageCount)
	}
	return pageNum + 1
}
