package main

import (
	_ "Project02/pyg/pyg/routers"
	"github.com/astaxie/beego"
	_ "Project02/pyg/pyg/models"
)

func main() {
	beego.Run()
}

