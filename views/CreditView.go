package views

import (
	"io"
	"os"
	"risk-ext/config"
	"risk-ext/models"
	"risk-ext/utils"
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/kataras/iris"
)

type CreditView struct {
	Views
}

func (this *CreditView) Auth(ctx iris.Context) int {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT":    MA{"ADMIN": A{MANAGER_ADMIN, MANAGER_SERVICE, MANAGER_ASSISTANT}},
		"GET":    MA{"ADMIN": A{MANAGER_ADMIN, MANAGER_SERVICE, MANAGER_ASSISTANT}},
		"POST":   MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN}},
		"DELETE": MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN}}}
	return this.CheckPerms(perms[ctx.Method()])
}

func (this *CreditView) Post(ctx iris.Context) (statuCode int, data M) {
	data = make(M)
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
	defer file.Close()
	title := head.Filename
	AuthSaveUrl := config.GetString("CreditExport") + time.Now().Format("200601") + "/" + idcard
	err = utils.IsFile(AuthSaveUrl)
	if err != nil {
		data["code"] = 0
		data["error"] = "创建文件失败1"
		return
	}
	f, err := os.OpenFile(AuthSaveUrl+title, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		data["code"] = 0
		data["error"] = err.Error()
		return
	}
	defer f.Close()
	_, err = io.Copy(f, file)
	if err != nil {
		data["code"] = 0
		data["error"] = "保存文件失败"
		return
	}
	ImgSaveUrl := config.GetString("CreditExport")
	//+ "Image_url/" + time.Now().Format("200601") + "/"
	ImgThumbUrl := config.GetString("CreditExport") + "Image_thumb/"
	frontFile, frontHead, err := ctx.FormFile("front")
	if err != nil {
		data["code"] = 0
		data["error"] = "上传文件失败"
		return
	}
	frontTitle := frontHead.Filename
	frontTile := frontTitle
	frontImgSaveUrl := ImgSaveUrl + idcard + "/front"
	frontImgThumbUrl := ImgThumbUrl + idcard + "/front/"

	err = utils.IsFile(frontImgSaveUrl)
	if err != nil {
		data["code"] = 0
		data["error"] = "创建文件失败2"
		return
	}
	err = utils.IsFile(frontImgThumbUrl)
	if err != nil {
		data["code"] = 0
		data["error"] = "创建文件失败2"
		return
	}

	front, err := os.OpenFile(frontImgSaveUrl+"/"+frontTile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		data["code"] = 0
		data["error"] = err.Error()
		return
	}
	defer front.Close()
	_, err = io.Copy(front, frontFile)
	if err != nil {
		data["code"] = 0
		data["error"] = "保存文件失败"
		return
	}
	err = utils.ImageResize(frontImgSaveUrl+"/"+frontTile, frontImgThumbUrl+frontTile)
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
	backTile := backTitle
	backImgSaveUrl := ImgSaveUrl + idcard + "/back"
	backImgThumbUrl := ImgThumbUrl + idcard + "/back/"
	err = utils.IsFile(backImgSaveUrl)
	if err != nil {
		data["code"] = 0
		data["error"] = "创建文件失败3"
		return
	}
	err = utils.IsFile(backImgThumbUrl)
	if err != nil {
		data["code"] = 0
		data["error"] = "创建文件失败3"
		return
	}
	back, err := os.OpenFile(backImgSaveUrl+"/"+backTile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		data["code"] = 0
		data["error"] = err.Error()
		return
	}
	defer back.Close()
	_, err = io.Copy(back, backFile)
	if err != nil {
		data["code"] = 0
		data["error"] = "保存文件失败"
		return
	}
	err = utils.ImageResize(backImgSaveUrl+"/"+backTile, backImgThumbUrl+backTile)
	if err != nil {
		data["code"] = 0
		data["error"] = "生成文件失败"
		return
	}
	//建立报表记录

	report := new(models.Reports)
	report.ReportType = 3
	report.ReportCompanyId = comId
	report.ReportStatus = 2
	im := new(models.Image)
	im.AuthImage = AuthSaveUrl + title
	im.FrontImageUrl = frontImgSaveUrl + "/" + frontTile
	im.FrontImageThumb = frontImgThumbUrl + frontTile
	im.BackImageUrl = backImgSaveUrl + "/" + backTile
	im.BackImageThumb = backImgThumbUrl + backTile
	report.ReportImage = im
	report.ReportIdCard = idcard
	err = report.Insert()
	if err != nil {
		data["error"] = "上传数据失败"
		data["code"] = 0
		return
	}
	data["code"] = 1
	statuCode = 200
	return
}

func (this *CreditView) Get(ctx iris.Context) (statuCode int, result interface{}) {
	statuCode = 400
	data := make(M)
	page := ctx.FormValue("page")
	size := ctx.FormValue("size")
	status := ctx.FormValueDefault("status", "-4")
	reportType := 3
	p, err := strconv.ParseInt(page, 10, 16)
	if err != nil {
		p = 1
	}
	s, err := strconv.ParseInt(size, 10, 16)
	if err != nil {
		s = 30
	}
	Tstatus, err := strconv.Atoi(status)
	if err != nil {
		data["code"] = 0
		data["error"] = "参数有误"
		result = data
		return
	}
	report := new(models.Reports)
	query := bson.M{}
	query["report_deleteat"] = 0
	if Tstatus != -4 {
		if Tstatus == 1 {
			var types = []int8{0, 1, -1}
			query["report_status"] = bson.M{"$in": types}
		} else {
			query["report_status"] = int8(Tstatus)
		}
	}
	query["report_type"] = uint8(reportType)
	rs, num, err := report.Lists(query, int(p), int(s))
	if err != nil {
		data["error"] = err
		return
	} else {
		data["list"] = rs
		data["num"] = num
		data["code"] = 1
		statuCode = 200
	}
	result = data
	return
}

//征信查询审核
func (this *CreditView) Put(ctx iris.Context) (statusCode int, data M) {
	data = make(M)
	statusCode = 400
	reportId := ctx.FormValue("report_id")
	report, err := new(models.Reports).One(reportId)
	if err != nil {
		data["error"] = err.Error()
		data["code"] = 0
		return
	}
	access := ctx.FormValue("access")
	status, err := strconv.Atoi(access)
	if err != nil {
		data["error"] = "参数有误"
		data["code"] = 0
		return
	}
	if status == 1 {
		status = 0
	} else if status == 0 {
		status = 3
	}
	report.ReportStatus = int8(status)
	err = report.Update()
	if err == nil {
		if status == 0 {
			amount := Session.User.Amount.QueryAiCar
			if amount <= 0 {
				data["error"] = "查询次数不足"
				data["code"] = 0
				return
			}
			task := new(models.Task)
			task.ReportId = reportId
			task.CompanyId = report.ReportCompanyId
			task.Type = int8(report.ReportType)
			task.Name = report.ReportName
			task.IdCard = report.ReportIdCard
			err = new(models.Redis).ListPush("analysis_tasks", task)
			if err != nil {
				data["error"] = "建立任务失败"
				data["code"] = 0
				return
			}
			amount--
			am := models.Amounts{}
			am.CompanyId = report.ReportCompanyId
			am.QueryAiCar = amount
			new(models.Redis).Save("amounts", report.ReportCompanyId, am)

		}
		statusCode = 200
		data["code"] = 1
	} else {
		data["error"] = err.Error()
		data["code"] = 0
	}
	return
}
