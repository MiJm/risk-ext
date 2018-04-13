package views

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"risk-ext/config"
	"risk-ext/models"
	"risk-ext/utils"
	"strconv"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/kataras/iris"
)

type ReportView struct {
	Views
}

func (this *ReportView) Auth(ctx iris.Context) bool {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT":    MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN}},
		"GET":    MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN, MEMBER_GENERAL}},
		"POST":   MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN, MEMBER_GENERAL}},
		"DELETE": MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN}}}
	return this.CheckPerms(perms[ctx.Method()])
}

func (this *ReportView) Detail(ctx iris.Context) (statuCode int, data interface{}) {
	statuCode = 400
	reportId := ctx.Params().Get("report_id")
	if !bson.IsObjectIdHex(reportId) {
		data = "报表ID不正确"
		return
	}
	report := new(models.Reports)
	report.ReportId = bson.ObjectIdHex(reportId)
	report.Model.One(report)
	if report.ReportCreateAt == 0 {
		statuCode = 404
		data = "报表不存在"
		return
	}

	if report.ReportDeleteAt > 0 {
		statuCode = 410
		data = "报表已被删除"
		return
	}

	if report.ReportStatus != 1 {
		statuCode = 400
		data = "报表当前状态不可用"
		return
	}
	type Poi struct {
		Device_address string  `json:"device_address"`
		Device_lat     float64 `json:"device_lat"`
		Device_lng     float64 `json:"device_lng"`
		Device_loctime uint64  `json:"device_loctime"`
	}
	statuCode = 200
	data = report
	return
}

func (this *ReportView) Get(ctx iris.Context) (statuCode int, result interface{}) {
	statuCode = 400

	page := ctx.FormValue("page")
	size := ctx.FormValue("size")
	reportId := ctx.Params().Get("report_id")
	if reportId != "" {
		statuCode, result = this.Detail(ctx)
		return
	}

	data := make(M)

	p, err := strconv.ParseInt(page, 10, 16)
	if err != nil {
		p = 1
	}
	s, err := strconv.ParseInt(size, 10, 16)
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

type Result struct {
	Status int8        `json:"status"`
	Data   interface{} `json:"data"`
	Msg    string      `json:"msg"`
}

//新增Report记录，发送获取Report请求
func (this *ReportView) Post(ctx iris.Context) (statuCode int, data M) {
	data = make(M)
	open := ""
	statuCode = 400
	if Session.User.Amount.QueryAiCar <= 0 {
		data["error"] = "查询次数不足"
		data["code"] = 0
		return
	}
	token := ctx.PostValue("token")
	from, err := ctx.PostValueInt("data_from") //from为0是内部 1是外部
	if err != nil {
		data["code"] = 0
		data["error"] = "请选择数据来源"
		return
	}
	carNum := ctx.FormValueDefault("car_num", "")
	if carNum == "" {
		data["code"] = 0
		data["error"] = "请输入车牌号或者任务名称"
		return
	}
	var reportFrom uint8
	if from == 1 {
		reportFrom = 1
		f, head, err := ctx.FormFile("file")
		if err != nil {
			data["code"] = 0
			data["error"] = "上传文件失败"
			return
		}
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
			if len(v1) < 8 {
				continue
			}
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
		if err != nil {
			data["code"] = 0
			data["error"] = "文件转换失败"
			return
		}
		openUrl := config.GetString("CarExport") + time.Now().Format("200601") + "/"
		saveUrl := config.GetString("CarExport") + time.Now().Format("200601") + "/"
		err = utils.IsFile(saveUrl)
		if err != nil {
			data["code"] = 0
			data["error"] = "保存文件失败"
			return
		}
		saveUrl = fmt.Sprintf("%s%s轨迹.json", saveUrl, strconv.Itoa(int(time.Now().Unix())))
		openUrl = fmt.Sprintf("%s%s轨迹.json", openUrl, strconv.Itoa(int(time.Now().Unix())))
		err = ioutil.WriteFile(saveUrl, re, 0644)
		if err != nil {
			data["code"] = 0
			data["error"] = "文件写入失败"
			return
		}
		open = openUrl
	} else {
		//请求内部数据接口
		parame := "carNum=" + carNum + "&" + "token=" + token
		result := struct {
			Status int
			Data   string
			Msg    string
		}{}
		err := this.GetMainData("routes/analyse_track", parame, &result)
		if err != nil {
			data["code"] = 0
			data["error"] = "查询轨迹失败"
			return
		}
		if result.Status == 1 {
			data1 := result.Data
			open = data1
		} else {
			data["error"] = result.Data
			data["code"] = 0
			return
		}

	}

	report := new(models.Reports)
	report.ReportType = 0
	report.ReportPlate = carNum
	report.ReportCompanyId = Session.User.UserCompany_id
	report.ReportDataFrom = reportFrom
	err = report.Insert()
	if err != nil {
		data["error"] = "上传数据失败"
		data["code"] = 0
		return
	}
	Task := struct {
		ReportId string //报表ID
		Path     string //分析数据文件路径
	}{}
	reportId := report.ReportId.Hex()
	Task.Path = open
	Task.ReportId = reportId
	fmt.Println("Task", Task)

	err = new(models.Redis).ListPush("analysis_tasks", Task)
	if err != nil {
		data["error"] = "建立任务失败"
		data["code"] = 0
		return
	}
	data["code"] = 1
	statuCode = 200
	return
}

func (this *ReportView) Delete(ctx iris.Context) (statusCode int, data M) {
	data = make(M)
	statusCode = 400
	reportId := ctx.Params().Get("report_id")
	flag := bson.IsObjectIdHex(reportId)
	if !flag {
		data["error"] = "参数有误"
		data["code"] = 0
		return
	}
	rep := new(models.Reports)
	port, _ := rep.One(reportId)
	port.ReportDeleteAt = time.Now().Unix()
	err := port.Update()
	if err != nil {
		data["error"] = "删除失败"
		data["code"] = 0
		return
	}
	statusCode = 200
	data["code"] = 1
	return
}
