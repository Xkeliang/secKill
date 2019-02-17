package controller

import (
	"Seckill/SecProxy/service"
	"fmt"
	"github.com/beego"
	"github.com/beego/logs"
	"strings"
	"time"
)

type SkillController struct {
	beego.Controller
}

//抢购产品
func (p *SkillController) SecKill() {
	productId, err := p.GetInt("product_id")
	result := make(map[string]interface{})

	result["code"] = 0
	result["message"] = "success"

	defer func() {
		p.Data["json"] = result
		p.ServeJSON()
	}()

	if err != nil {
		result["code"] = 1001
		result["message"] = "invalid product_id"
		return
	}

	//获取请求信息
	/*

	*/
	source := p.GetString("src")
	authcode := p.GetString("autocode")
	secTime := p.GetString("time")
	nance := p.GetString("nance")

	secRequest := service.NewSecRequest()
	secRequest.AuthCode = authcode
	secRequest.Nance = nance
	secRequest.ProductId = productId
	secRequest.SecTime = secTime
	secRequest.Source = source
	secRequest.UserCookieSign = p.Ctx.GetCookie("userAuthSign")
	secRequest.UserId, err = p.GetInt("user_id")
	if err != nil {
		logs.Error("get user_id error:", err)
	}
	secRequest.AccessTime = time.Now()

	if len(p.Ctx.Request.RemoteAddr) > 0 {
		secRequest.ClientAddr = strings.Split(p.Ctx.Request.RemoteAddr, ":")[0]
	}

	logs.Debug("client Addr:[%v]", secRequest.ClientAddr)

	secRequest.ClientRefence = p.Ctx.Request.Referer()
	logs.Debug("secRequest.ClientReferce :%v",secRequest.ClientRefence)
	if err != nil {
		result["code"] = service.ErrInvalidRequest
		result["message"] = fmt.Sprintf("invalid cookie:userId")
		return
	}
	//抢购
	data, code, err := service.SecKill(secRequest)
	if err != nil {
		result["code"] = code
		result["message"] = err.Error()
		return
	}
	result["code"] = code
	result["data"] = data
	return
}

//查看抢购产品信息
func (p *SkillController) SecInfo() {
	productId, err := p.GetInt("product_id")
	result := make(map[string]interface{})

	//正常输出code和message
	result["code"] = 0
	result["message"] = "success"
	defer func() {
		p.Data["json"] = result
		p.ServeJSON()
	}()
	if err != nil {
		//获取产品id错误则输出所有产品
		data, code, err := service.SecInfoList()
		if err != nil {
			result["code"] = 1001
			result["message"] = "invalid product_id"
			logs.Error("invalid request,get product_id failed")
			return
		}
		result["code"] = code
		result["data"] = data
		return
	} else {
		data, code, err := service.SecInfo(productId)
		if err != nil {
			result["code"] = code
			result["message"] = err.Error()
			logs.Error("invalid request, get product_id failed, err:%v", err)
			return
		}
		result["code"] = code
		result["data"] = data
	}
}
