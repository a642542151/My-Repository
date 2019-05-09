package models

import (
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

type UserData struct {
	Id   int
	Name string
	Pwd  string
	//多对多   多用户对多文章
	Articles []*Article `orm:"rel(m2m)"`
}

//文章表	单表
type Article struct {
	Id    int    `orm:"pk;auto"`
	Title string `orm:"unique;size(40)"`
	//文章内容
	Content string `orm:"size(500)"`
	//图片在数据库中保存的是路径
	Img string `orm:"null"`
	//阅读时间
	Time time.Time `orm:"type(datetime);auto_now_add"`
	//阅读数
	ReadCount int `orm:"default(0)"`

	//多文章对多用户
	Users []*UserData `orm:"reverse(many)"`
	//单文章对多类型
	ArticleType *ArticleType `orm:"rel(fk);null;on_delete(set_null)"`
}

//文章类型表	一对多
type ArticleType struct {
	Id       int
	TypeName string     `orm:"unique"`
	Article  []*Article `orm:"reverse(many)"`
}

//用户阅读文章表	多对多
//在多对多的关系建立中，orm 会自动建立另一张新表来表示该关系

func init() {

	orm.RegisterDataBase("default", "mysql", "root:123456@tcp(192.168.11.60:3306)/userData?charset=utf8")

	orm.RegisterModel(new(UserData), new(Article), new(ArticleType))

	orm.RunSyncdb("default", false, true)
}
