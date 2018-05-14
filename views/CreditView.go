package views

import (
	"io"
	"os"
	"risk-ext/config"
	"risk-ext/models"
	"risk-ext/utils"
	"time"

	"github.com/kataras/iris"
)

type CreditView struct {
	Views
}

func (this *CreditView) Auth(ctx iris.Context) int {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT":    MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN}},
		"GET":    MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN, MEMBER_GENERAL}},
		"POST":   MA{"NOLOGIN": A{1}},
		"DELETE": MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN}}}
	return this.CheckPerms(perms[ctx.Method()])
}

func (this *CreditView) Post(ctx iris.Context) (data M, statuCode int) {
	data = make(M)
	open := ""
	statuCode = 400
	amount := Session.User.Amount.QueryCredit
	comId := Session.User.UserCompany_id
	//	groId := Session.User.UserGroupId
	if amount <= 0 {
		data["error"] = "查询次数不足"
		data["code"] = 0
		return
	}
	name := ctx.FormValue("name")
	idcard := ctx.FormValueDefault("idcard", "")
	if name == "" || idcard == "" {
		data["error"] = "参数不足"
		data["code"] = 0
		return
	}
	ctx.SetMaxRequestBodySize(2 << 31)
	file, head, err := ctx.FormFile("authorize")
	if err != nil {
		data["code"] = 0
		data["error"] = "上传文件失败"
		return
	}
	title := head.Filename
	AuthSaveUrl := config.GetString("CreditExport") + time.Now().Format("200601") + "/" + idcard + title
	f, err := os.Create(AuthSaveUrl)
	if err != nil {
		data["code"] = 0
		data["error"] = "创建文件失败"
		return
	}
	defer f.Close()
	_, err = io.Copy(f, file)
	if err != nil {
		data["code"] = 0
		data["error"] = "保存文件失败"
		return
	}
	ImgSaveUrl := config.GetString("CreditExport") + "Image_url/" + time.Now().Format("200601") + "/"
	ImgThumbUrl := config.GetString("CreditExport") + "Image_thumb/" + time.Now().Format("200601") + "/"
	frontFile, frontHead, err := ctx.FormFile("front")
	if err != nil {
		data["code"] = 0
		data["error"] = "上传文件失败"
		return
	}
	frontTitle := frontHead.Filename
	frontTile := "front" + frontTitle
	frontImgSaveUrl := ImgSaveUrl + idcard + frontTile
	frontImgThumbUrl := ImgThumbUrl + idcard + frontTile
	front, err := os.Create(frontImgSaveUrl)
	if err != nil {
		data["code"] = 0
		data["error"] = "创建文件失败"
		return
	}
	defer f.Close()
	_, err = io.Copy(front, frontFile)
	if err != nil {
		data["code"] = 0
		data["error"] = "保存文件失败"
		return
	}
	err = utils.ImageResize(frontImgSaveUrl, frontImgThumbUrl)
	if err != nil {
		data["code"] = 0
		data["error"] = "生成文件失败"
		return
	}

	backFile, backHead, err := ctx.FormFile("back")
	if err != nil {
		data["code"] = 0
		data["error"] = "上传文件失败"
		return
	}
	backTitle := backHead.Filename
	backTile := "back" + backTitle
	backImgSaveUrl := ImgSaveUrl + idcard + backTile
	backImgThumbUrl := ImgThumbUrl + idcard + backTile
	back, err := os.Create(backImgSaveUrl)
	if err != nil {
		data["code"] = 0
		data["error"] = "创建文件失败"
		return
	}
	defer f.Close()
	_, err = io.Copy(back, backFile)
	if err != nil {
		data["code"] = 0
		data["error"] = "保存文件失败"
		return
	}
	err = utils.ImageResize(backImgSaveUrl, backImgThumbUrl)
	if err != nil {
		data["code"] = 0
		data["error"] = "生成文件失败"
		return
	}
	//建立报表记录
	report := new(models.Reports)
	report.ReportType = 3
	report.ReportCompanyId = comId
	err = report.Insert()
	if err != nil {
		data["error"] = "上传数据失败"
		data["code"] = 0
		return
	}
	Task := struct {
		ReportId  string //报表ID
		CompanyId string //企业ID
		CarNum    string //车牌号
		Path      string //分析数据文件路径
	}{}
	reportId := report.ReportId.Hex()
	Task.Path = open
	Task.ReportId = reportId
	Task.CompanyId = comId

	err = new(models.Redis).ListPush("analysis_tasks", Task)
	if err != nil {
		data["error"] = "建立任务失败"
		data["code"] = 0
		return
	}
	amount--
	am := models.Amounts{}
	am.CompanyId = comId
	am.QueryAiCar = amount
	new(models.Redis).Save("amounts", comId, am)
	data["code"] = 1
	statuCode = 200
	return
}
