package views

import (
	"risk-ext/models"
	"time"

	"github.com/kataras/iris"
	"gopkg.in/mgo.v2/bson"
)

type SharesView struct {
	Views
}

func (this *SharesView) Auth(ctx iris.Context) int {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT":    MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN}},
		"GET":    MA{"NOLOGIN": A{1}},
		"POST":   MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN, MEMBER_GENERAL}},
		"DELETE": MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN, MEMBER_GENERAL}}}
	return this.CheckPerms(perms[ctx.Method()])
}

func (this *SharesView) Post(ctx iris.Context) (statusCode int, data M) {
	data = make(M)
	statusCode = 400
	reportId := ctx.FormValue("reportId")
	amount := Session.User.Amount.QueryAiCar
	if amount <= 0 {
		data["error"] = "查询次数不足"
		data["code"] = 0
		return
	}

	flag := bson.IsObjectIdHex(reportId)
	if !flag {
		data["error"] = "参数有误"
		data["code"] = 0
		return
	}
	phone := ctx.FormValue("phone")
	if phone == "" {
		data["error"] = "请输入完整参数"
		data["code"] = 0
		return
	}
	rs, err := new(models.Reports).One(reportId)
	if err != nil {
		data["error"] = "参数有误，无数据"
		data["code"] = 0
		return
	}
	shareId := bson.NewObjectId()
	shareUser := models.Shares{}
	shareUser.ShareId = shareId.Hex()
	shareUser.ShareFname = Session.User.UserFname
	shareUser.ShareMobile = phone
	shareUser.ShareCreateAt = time.Now().Unix()
	rs.ReportShares[phone] = shareUser
	err = rs.Update()
	if err != nil {
		data["error"] = "添加分享人失败"
		data["code"] = 0
		return
	}
	comId := Session.User.UserCompany_id
	amount--
	am := models.Amounts{}
	am.CompanyId = comId
	am.QueryAiCar = amount
	new(models.Redis).Save("amounts", comId, am)
	statusCode = 200
	data["code"] = 1
	return
}

func (this *SharesView) Delete(ctx iris.Context) (statusCode int, data M) {
	data = make(M)
	statusCode = 400
	reportId := ctx.FormValue("reportId")
	phone := ctx.Params().Get("params")
	rep := new(models.Reports)
	port, _ := rep.One(reportId)
	err := port.RemoveShare(phone)
	if err != nil {
		data["error"] = "删除失败"
		data["code"] = 0
		return
	}
	statusCode = 200
	data["code"] = 1
	return
}

//校验验证码，获取报告信息
func (this *SharesView) Get(ctx iris.Context) (statusCode int, data M) {
	data = make(M)
	statusCode = 400
	mobile := ctx.FormValue("mobile")
	code := ctx.FormValue("code")
	reportId := ctx.Params().Get("params")
	bol := this.CheckCode(mobile, code)
	if bol {
		res, err := new(models.Reports).CheckPhone(mobile, reportId)
		if err != nil || res == nil {
			data["code"] = 0
			data["error"] = "无可看的报告"
			return
		}
		sharer := res.ReportShares[mobile]
		cou := sharer.ShareViews + 1
		sharer.ShareViews = cou
		res.ReportShares[mobile] = sharer
		res.Update()
		statusCode = 200
		data["list"] = res
		data["code"] = 1
		return
	} else {
		data["code"] = 0
		data["error"] = "验证码有误"
		return
	}
}

//更新操作待用
func (this *SharesView) Put(ctx iris.Context) (statuCode int, data M) {
	return
}
