package router

import (
	"github.com/astaxie/beego"
	"Seckill/SecAdmin/controller/product"
	"Seckill/SecAdmin/controller/activity"
)

func init(){
	//注册路由
	beego.Router("/",&product.ProductController{},"*:ListProduct")
	beego.Router("/product/list",&product.ProductController{},"*:ListProduct")
	beego.Router("/product/create",&product.ProductController{},"*:CreateProduct")
	beego.Router("/product/submit",&product.ProductController{},"*:SubmitProduct")

	beego.Router("/activity/create",&activity.ActivityController{},"*:CreateActivity")
	beego.Router("/activity/list",&activity.ActivityController{},"*:ListActivity")
	beego.Router("/activity/submit",&activity.ActivityController{},"*:SubmitActivity")

}
