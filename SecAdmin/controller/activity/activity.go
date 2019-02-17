package activity

import (
	"github.com/astaxie/beego"
	"Seckill/SecAdmin/model"
	"github.com/beego/logs"
	"fmt"
)

type ActivityController struct {
	beego.Controller
}

func (p *ActivityController)CreateActivity()  {
	p.TplName="activity/create.html"
	p.Layout="layout/layout.html"
	return
}

//获取活动产品mysql
func (p *ActivityController)ListActivity()  {
	activityModel := model.NewActivityModel()
	activityList,err := activityModel.GetActivityList()
	if err != nil {
		logs.Warn("get activityList failed,err:%v",err)
		return
	}
	p.Data["activityList"]=activityList
	p.TplName = "activity/list.html"
	p.Layout = "layout/layout.html"
}

//提交互动清单到mysql
func (p *ActivityController)SubmitActivity()  {
	activityModel := model.NewActivityModel()
	var activity model.Activity

	var err error
	var Error string
	p.TplName = "activity/list.html"
	p.Layout = "layout/layout.html"
	defer func() {
		if err != nil {
			p.Data["Error"]=Error
			p.TplName="activity/error.html"
		}
	}()
	name := p.GetString("activityName")
	if name == "" {
		Error  ="活动不能为空"
		err = fmt.Errorf("activity name can not null")
		return
	}

	productId,err := p.GetInt("productId")
	if err != nil {
		err = fmt.Errorf("invalud procuctId err:%v",err)
		Error = err.Error()
	}
	startTime,err := p.GetInt64("startTime")
	if err != nil {
		err = fmt.Errorf("invalud start_time err:%v",err)
		Error = err.Error()
	}
	endTime,err := p.GetInt64("endTime")
	if err != nil {
		err = fmt.Errorf("invalud end_time err:%v",err)
		Error = err.Error()
	}
	total,err := p.GetInt("total")
	if err != nil {
		err = fmt.Errorf("invalud total err:%v",err)
		Error = err.Error()
	}
	speed,err := p.GetInt("speed")
	if err != nil {
		err = fmt.Errorf("invalud speed err:%v",err)
		Error = err.Error()
	}
	limit,err := p.GetInt("buyLimit")
	if err != nil {
		err = fmt.Errorf("invalud buy_limit err:%v",err)
		Error = err.Error()
	}
	rate,err := p.GetFloat("buyRate")
	if err != nil {
		err = fmt.Errorf("invalud buy_rate err:%v",err)
		Error = err.Error()
	}

	activity.ActivityName=name
	activity.ProductId =productId
	activity.StartTime =startTime
	activity.EndTime=endTime
	activity.Total=total
	activity.Speed=speed
	activity.BuyLimit=limit
	activity.BuyRate=rate

	err = activityModel.CreateActivity(&activity)
	if err != nil {
		err = fmt.Errorf("创建活动失败,err:%v",err)
		Error = err.Error()
		return
	}

	p.Redirect("/activity/list",301)
	return
}