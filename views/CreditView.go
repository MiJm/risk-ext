package views

import (
	"io"
	"os"
	"risk-ext/app"
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

func (this *CreditView) Post(ctx iris.Context) (statuCode int, data app.M) {
	data = make(M)
	statuCode = 400
	amount := Session.User.Amount.QueryCredit
	comId := Session.User.UserCompany_id
	if amount <= 0 {
		data["error"] = "查询次数不足"
		data["code"] = 0
		return
	}
	name := ctx.FormValue("name")
	idcard := ctx.FormValueDefault("idcard", "")
	tel := ctx.FormValueDefault("tel", "")
	if name == "" || idcard == "" || tel == "" {
		data["error"] = "参数不足"
		data["code"] = 0
		return
	}

	ctx.SetMaxRequestBodySize(2 << 31)
	file, _, err := ctx.FormFile("authorize")
	if err != nil {
		data["code"] = 0
		data["error"] = "上传文件失败"
		return
	}
	defer file.Close()
	//	title := head.Filename
	title := "authImg.png"
	AuthSaveUrl := config.GetString("CreditFiles") + time.Now().Format("200601") + "/" + idcard + "/"
	AuthOpenUrl := "imgs/" + time.Now().Format("200601") + "/" + idcard + "/"
	err = utils.IsFile(AuthSaveUrl)
	if err != nil {
		data["code"] = 0
		data["error"] = "创建文件失败"
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
	AuthThumbUrl := config.GetString("CreditFiles") + time.Now().Format("200601") + "/" + idcard + "/thumb"
	AuthThumbOpenUrl := "imgs/" + time.Now().Format("200601") + "/" + idcard + "/thumb"
	err = utils.IsFile(AuthThumbUrl)
	if err != nil {
		data["code"] = 0
		data["error"] = "创建文件失败"
		return
	}
	err = utils.ImageResize(AuthSaveUrl+title, AuthThumbUrl+title)
	if err != nil {
		data["code"] = 0
		data["error"] = "生成文件失败"
		return
	}

	ImgSaveUrl := config.GetString("CreditFiles") + "Image_url/"
	ImgThumbUrl := config.GetString("CreditFiles") + "Image_thumb/"
	ImgOpenUrl := "imgs/Image_url/"
	ImgOpenThumbUrl := "imgs/Image_thumb/"
	frontFile, _, err := ctx.FormFile("front")
	if err != nil {
		data["code"] = 0
		data["error"] = "上传文件失败"
		return
	}
	//	frontTitle := frontHead.Filename
	//	frontTile := frontTitle
	frontTile := "idCardFront.png"
	frontImgSaveUrl := ImgSaveUrl + idcard + "/front"
	frontImgThumbUrl := ImgThumbUrl + idcard + "/front/"
	frontImgOpenUrl := ImgOpenUrl + idcard + "/front"
	frontImgThumbOpenUrl := ImgOpenThumbUrl + idcard + "/front/"

	err = utils.IsFile(frontImgSaveUrl)
	if err != nil {
		data["code"] = 0
		data["error"] = "创建文件失败"
		return
	}
	err = utils.IsFile(frontImgThumbUrl)
	if err != nil {
		data["code"] = 0
		data["error"] = "创建文件失败"
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

	backFile, _, err := ctx.FormFile("back")
	if err != nil {
		data["code"] = 0
		data["error"] = "上传文件失败"
		return
	}
	//	backTitle := backHead.Filename
	//	backTile := backTitle
	backTile := "idCardBack.png"
	backImgSaveUrl := ImgSaveUrl + idcard + "/back"
	backImgThumbUrl := ImgThumbUrl + idcard + "/back/"
	backImgOpenUrl := ImgOpenUrl + idcard + "/back"
	backImgThumbOpenUrl := ImgOpenThumbUrl + idcard + "/back/"
	err = utils.IsFile(backImgSaveUrl)
	if err != nil {
		data["code"] = 0
		data["error"] = "创建文件失败"
		return
	}
	err = utils.IsFile(backImgThumbUrl)
	if err != nil {
		data["code"] = 0
		data["error"] = "创建文件失败"
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
	im.AuthImage = AuthOpenUrl + title
	im.AuthImageThumb = AuthThumbOpenUrl + title
	im.FrontImageUrl = frontImgOpenUrl + "/" + frontTile
	im.FrontImageThumb = frontImgThumbOpenUrl + frontTile
	im.BackImageUrl = backImgOpenUrl + "/" + backTile
	im.BackImageThumb = backImgThumbOpenUrl + backTile
	report.ReportImage = im
	report.ReportIdCard = idcard
	report.ReportCheckName = Session.User.UserFname
	report.ReportName = name
	report.ReportCompanyName = Session.User.UserCompanyName
	str := bson.NewObjectId().Hex()
	numStr := utils.SubString(str, 12)
	report.ReportNumber = numStr
	report.ReportMobile = tel
	err = report.Insert()
	if err != nil {
		data["error"] = "上传数据失败"
		data["code"] = 0
		return
	}
	amount--
	am := Session.User.Amount
	am.QueryCredit = amount
	new(models.Redis).Save("amounts", comId, am)
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
func (this *CreditView) Put(ctx iris.Context) (statusCode int, data app.M) {
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
	report.ReportAuditName = Session.Manager.Manager_fname
	report.ReportAuditTime = int64(time.Now().Unix())
	report.ReportStatus = int8(status)
	err = report.Update()
	if err == nil {
		if status == 0 {
			Task := struct {
				ReportId  string //报表ID
				CompanyId string //企业ID
				Type      int8   //类型
				Name      string //姓名
				IdCard    string //身份证号
			}{}
			Task.ReportId = reportId
			Task.CompanyId = report.ReportCompanyId
			Task.Type = int8(report.ReportType)
			Task.Name = report.ReportName
			Task.IdCard = report.ReportIdCard
			err = new(models.Redis).ListPush("analysis_tasks", Task)
			if err != nil {
				data["error"] = "建立任务失败"
				data["code"] = 0
				return
			}
		} else {
			Session.ChangeAmount(report.ReportCompanyId, 1, 3)
		}

		statusCode = 200
		data["code"] = 1
	} else {
		data["error"] = err.Error()
		data["code"] = 0
	}
	return
}

//删除操作待用
func (this *CreditView) Delete(ctx iris.Context) (statuCode int, data app.M) {
	return
}
