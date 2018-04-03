package views

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"risk-ext/models"
	"risk-ext/utils"
	"strconv"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/astaxie/beego"
	"github.com/kataras/iris"
)

type ReportView struct {
	Views
}

func (this *ReportView) Auth(ctx iris.Context) bool {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT":  MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN}},
		"GET":  MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN, MEMBER_GENERAL}},
		"POST": MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN, MEMBER_GENERAL}}}
	return this.CheckPerms(perms[ctx.Method()])
}

func (this *ReportView) Get(ctx iris.Context) (statuCode int, data M) {
	statuCode = 400
	data = make(M)
	token := ctx.FormValueDefault("token", "123")
	fmt.Println(token)
	page := ctx.FormValue("page")
	size := ctx.FormValue("size")
	reportId := ctx.FormValue("report_id")
	p, err := strconv.ParseInt(page, 10, 64)
	if err != nil {
		p = 1
	}
	s, err := strconv.ParseInt(size, 10, 64)
	if err != nil {
		s = 30
	}
	report := new(models.Reports)
	companyId := Session.User.UserCompany_id
	query := bson.M{}
	query["report_company_id"] = companyId
	query["report_deleteat"] = 0
	data = make(M)
	data["ai_amount"] = Session.User.Amount.QueryAiCar
	if reportId != "" {
		rs, err := report.One(reportId)
		if err != nil {
			data["error"] = err
			return
		} else {
			data["list"] = rs.ReportShares
			data["code"] = 1
			statuCode = 200
		}
	} else {
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
	}

	return
}

type Result struct {
	Status int8        `json:"status"`
	Data   interface{} `json:"data"`
	Msg    string      `json:"msg"`
}

//新增Report记录，发送获取Report请求
func (this *ReportView) Post(ctx iris.Context) (statuCode int, data M) {
	data = make(M)
	open := "http://192.168.1.118:8080/"
	statuCode = 400
	if Session.User.Amount.QueryAiCar <= 0 {
		data["error"] = "查询次数不足"
		data["code"] = 0
		return
	}
	token := ctx.PostValue("token")
	carNum := ctx.FormValueDefault("car_num", "")
	var reportFrom uint8
	if carNum == "" {
		reportFrom = 1
		f, head, err := ctx.FormFile("file")
		b := make([]byte, head.Size)
		_, err = f.Read(b)
		if err != nil {
			data["code"] = 0
			data["error"] = "读取文件失败"
			return
		}

		da := string(b)
		result := strings.Split(da, "\n")
		rout_arr := make([]models.Routes, 0)
		for k, v := range result {
			if k == 0 {
				//第一行标题去掉
				continue
			}
			v1 := strings.Split(v, ",")
			routes := models.Routes{}
			routes.Device_address = v1[5]
			loctime := utils.Str2Time(v1[7])
			routes.Device_loctime = loctime
			typ, err := strconv.Atoi(v1[6])
			speed, err4 := strconv.Atoi(v1[4])
			if err != nil || err4 != nil {
				data["code"] = 0
				data["error"] = "文件格式有误"
				return
			}
			routes.Device_loctype = uint8(typ)
			routes.Device_speed = uint8(speed)
			var latlng = make([]float64, 0)
			var slatlng = make([]float64, 0)
			lat, err := strconv.ParseFloat(v1[0], 64)
			lng, err1 := strconv.ParseFloat(v1[1], 64)
			slat, err2 := strconv.ParseFloat(v1[2], 64)
			slng, err3 := strconv.ParseFloat(v1[3], 64)
			if err != nil || err1 != nil || err2 != nil || err3 != nil {
				data["code"] = 0
				data["error"] = "文件格式有误"
				return
			}
			latlng = append(latlng, lng, lat)
			slatlng = append(slatlng, slng, slat)
			routes.Device_latlng.Coordinates = latlng
			routes.Device_slatlng.Coordinates = slatlng
			rout_arr = append(rout_arr, routes)
		}
		re, err := json.Marshal(rout_arr)
		openUrl := "devices/" + time.Now().Format("200601") + "/"
		saveUrl := beego.AppConfig.String("CarExport") + time.Now().Format("200601") + "/"
		err = utils.IsFile(saveUrl)
		if err != nil {
			return
		}
		saveUrl = fmt.Sprintf("%s%s轨迹.txt", saveUrl, carNum)
		openUrl = fmt.Sprintf("%s%s轨迹.txt", openUrl, carNum)
		err = ioutil.WriteFile(saveUrl, re, 0644)
		open = open + openUrl
	} else {
		//请求内部数据接口
		parame := "carNum=" + carNum + "&" + "token=" + token
		rs1 := new(Views).GetMainData("routes/analyse_track", parame)
		result := rs1.(map[string]interface{})
		data1 := result["data"]
		open = open + data1.(string)
	}
	//	post_param := "file_url=" + open
	//	post_res := new(Views).GetAnalysisData("task/file/submit", post_param)
	//	res := post_res.(M)
	//	if res["code"].(int) != 0 {
	//		data["error"] = "上传数据失败"
	//		data["code"] = 0
	//		return
	//	}
	report := new(models.Reports)
	report.ReportType = 0
	report.ReportPlate = carNum
	report.ReportDataFrom = reportFrom
	//	report.ReportOpenId = res["task_id"].(string)
	report.ReportCompanyId = Session.User.UserCompany_id
	fmt.Println(report)
	err := report.Insert()
	if err != nil {
		data["error"] = "上传数据失败"
		data["code"] = 0
		return
	}
	data["code"] = 1
	statuCode = 200
	return
}

//更新分享人信息
func (this *ReportView) Put(ctx iris.Context) (statusCode int, data M) {
	data = make(M)
	statusCode = 400
	typ := ctx.FormValueDefault("type", "0")
	reportId := ctx.FormValue("reportId")
	if typ == "1" { //删除分享人
		shareId := ctx.FormValue("shareId")
		flag := bson.IsObjectIdHex(shareId)
		if !flag {
			data["error"] = "参数有误"
			data["code"] = 0
			return
		}
		fmt.Println(shareId)
		rep := new(models.Reports)
		port, _ := rep.One(reportId)
		err := port.RemoveShare(shareId)
		if err != nil {
			data["error"] = "删除失败"
			data["code"] = 0
			return
		}
	} else if typ == "0" { //新增分享人

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
		rs.ReportShares[shareId.Hex()] = shareUser
		err = rs.Update()
		if err != nil {
			data["error"] = "添加分享人失败"
			data["code"] = 0
			return
		}

	}
	statusCode = 200
	data["code"] = 1
	return
}
