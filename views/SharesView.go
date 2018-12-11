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

func (this *SharesView) Auth(ctx iris.Context) bool {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT":    MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN}},
		"GET":    MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN, MEMBER_GENERAL}},
		"POST":   MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN, MEMBER_GENERAL}},
		"DELETE": MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN, MEMBER_GENERAL}}}
	return this.CheckPerms(perms[ctx.Method()])
}

func (this *SharesView) Post(ctx iris.Context) (statusCode int, data M) {
	data = make(M)
	statusCode = 400
	reportId := ctx.FormValue("reportId")
	if Session.User.Amount.QueryAiCar <= 0 {
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
	fname := ctx.FormValue("fname")
	if phone == "" || fname == "" {
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
	shareUser.ShareFname = fname
	shareUser.ShareMobile = phone
	shareUser.ShareCreateAt = time.Now().Unix()
	rs.ReportShares[phone] = shareUser
	err = rs.Update()
	if err != nil {
		data["error"] = "添加分享人失败"
		data["code"] = 0
		return
	}
	statusCode = 200
	data["code"] = 1
	return
}

func (this *SharesView) Delete(ctx iris.Context) (statusCode int, data M) {
	data = make(M)
	statusCode = 400
	reportId := ctx.FormValue("reportId")
	shareId := ctx.Params().Get("share_id")
	flag := bson.IsObjectIdHex(shareId)
	if !flag {
		data["error"] = "参数有误"
		data["code"] = 0
		return
	}
	rep := new(models.Reports)
	port, _ := rep.One(reportId)
	err := port.RemoveShare(shareId)
	if err != nil {
		data["error"] = "删除失败"
		data["code"] = 0
		return
	}
	statusCode = 200
	data["code"] = 1
	return
}
