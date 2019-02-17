package product

import (
	"github.com/astaxie/beego"
	"Seckill/SecAdmin/model"
	"github.com/Unknwon/log"
	"github.com/beego/logs"
	"fmt"
)

type ProductController struct {
	beego.Controller
}

func (p *ProductController)ListProduct()  {
	//productModel := model.NewProductModel()
	fmt.Println("access listProduct func")
	productModel := model.NewProductModel()
	productList,err :=productModel.GetProductList()
	if err != nil {
		log.Warn("get product list failed,err:%v",err)
		return
	}
	p.Data["productList"]=productList
	fmt.Println("-----------------")

	//p.Ctx.WriteString("list")
	p.TplName="product/list.html"
	p.Layout="layout/layout.html"
}

func (p *ProductController) CreateProduct()  {
	fmt.Println(beego.BConfig.WebConfig.ViewsPath)
	p.TplName="product/create.html"
	p.Layout="layout/layout.html"
}

func (p *ProductController) SubmitProduct()  {
	var err error
	defer func(){
		if err!=nil {
			p.Data["Error"]=err
			p.TplName="product/error.html"
			p.Layout="layout/layout.html"
		}
	}()
	//p.GetInt("")
	productName :=p.GetString("productName")
	if productName == "" {
		logs.Warn("get productName err")
		err = fmt.Errorf("productName is null")
		return
	}
	productTotal,err := p.GetInt("productTotal")
	if err!=nil{
		logs.Warn("get productTotal err")
		return
	}
	productStatus,err := p.GetInt("productStatus")
	if err != nil {
		logs.Warn("get productStatus err")
		return
	}
	productModel := model.NewProductModel()
	product := model.Product{
		ProductName:productName,
		Total:productTotal,
		Status:productStatus,
	}
	err  = productModel.CreateProduct(&product)
	if err != nil {
		logs.Warn("create product failed,err:%v",err)
		return
	}
	fmt.Println("-----------------")
	logs.Debug("productname[%s],productTotal[%d],productStatus[%d]",productName,productTotal,productStatus)
	p.TplName="product/create.html"
	p.Layout="layout/layout.html"
}
