package views

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"risk-ext/app"
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

func (this *ReportView) Auth(ctx iris.Context) int {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT":    MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN}},
		"GET":    MA{"USER": A{MEMBER_SUPER, MEMBER_ADMIN, MEMBER_GENERAL}, "ADMIN": A{MANAGER_ADMIN, MANAGER_SERVICE, MANAGER_ASSISTANT}},
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
	stype := Session.Type
	if stype == 1 {
		if report.ReportStatus != 1 {
			statuCode = 400
			data = "报表当前状态不可用"
			return
		}
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
	data := make(app.M)
	defer func() {
		result = data
	}()
	statuCode = 400
	page := ctx.FormValue("page")
	size := ctx.FormValue("size")
	reportId := ctx.Params().Get("report_id")
	reportType := ctx.FormValueDefault("type", "0")
	if reportId != "" {
		statuCode, result = this.Detail(ctx)
		if statuCode != 200 && statuCode != 204 {
			data["code"] = 0
			data["error"] = result
		} else {
			data["list"] = result
			data["code"] = 1
		}
		return
	}

	p, err := strconv.ParseInt(page, 10, 16)
	if err != nil {
		p = 1
	}
	s, err := strconv.ParseInt(size, 10, 16)
	if err != nil {
		s = 30
	}
	rType, err := strconv.ParseInt(reportType, 10, 64)
	if err != nil {
		data["code"] = 0
		data["error"] = "参数有误"
		return
	}
	report := new(models.Reports)
	companyId := Session.User.UserCompany_id
	query := bson.M{}
	query["report_company_id"] = companyId
	query["report_deleteat"] = 0
	query["report_type"] = uint8(rType)
	data = make(app.M)
	data["ai_amount"] = Session.User.Amount.QueryAiCar
	data["dianhua_amount"] = Session.User.Amount.QueryDianHua
	data["credit_amount"] = Session.User.Amount.QueryCredit
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
	return
}

type Result struct {
	Status int8        `json:"status"`
	Data   interface{} `json:"data"`
	Msg    string      `json:"msg"`
}

//新增Report记录，发送获取Report请求
func (this *ReportView) Post(ctx iris.Context) (statuCode int, result interface{}) {
	data := make(app.M)
	defer func() {
		result = data
	}()
	open := ""
	statuCode = 400
	amount := Session.User.Amount.QueryAiCar
	comId := Session.User.UserCompany_id
	groId := Session.User.UserGroupId
	if amount <= 0 {
		data["error"] = "查询次数不足"
		data["code"] = 0
		return
	}
	//	token := ctx.PostValue("token")
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
		ctx.SetMaxRequestBodySize(2 << 31)
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
		result1 := strings.Split(da, "\n")
		rout_arr := make([]models.Routes, 0)
		for k, v := range result1 {
			if k == 0 {
				//第一行标题去掉
				continue
			}
			v1 := strings.Split(v, ",")
			if len(v1) < 3 {
				continue
			}
			routes := models.Routes{}
			loctime := utils.Str2Time(v1[2])
			if loctime == 0 {
				continue
			}
			routes.Device_loctime = loctime
			var latlng = make([]float64, 2)
			lat, err := strconv.ParseFloat(v1[0], 64)
			lng, err1 := strconv.ParseFloat(v1[1], 64)
			if err != nil || err1 != nil {
				continue
			}
			latlng[0] = lng
			latlng[1] = lat
			routes.Device_latlng.Coordinates = latlng
			rout_arr = append(rout_arr, routes)
		}
		if len(rout_arr) < 1000 {
			data["code"] = 0
			data["error"] = "可用数据不足1000条，无法生成产生报告"
			return
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
		cou, err, car := new(models.Cars).OneCar(carNum)
		if err != nil || cou == 0 {
			data["error"] = "不存在该车辆"
			data["code"] = 0
			return
		}

		level := Session.User.UserLevel
		if comId != "5a618eea8a5da54404b68e41" && comId != "5a4e41f78a5da57a9f95df11" {
			if level == 2 {
				if comId != car.Car_company_id {
					data["error"] = "您没有权限查看该车辆"
					data["code"] = 0
					return
				}
			} else {
				gro, err := new(models.Groups).One(groId)
				if err != nil {
					data["error"] = "不存在该组织"
					data["code"] = 0
					return
				}
				flg := false
				for _, v := range gro.Group_sub {
					if v.Group_id.Hex() == car.Car_group_id {
						flg = true
					}
				}
				if !flg {
					data["error"] = "您没有权限查看该车辆"
					data["code"] = 0
					return
				}
			}
		}

		if len(car.Car_devices) == 0 {
			data["error"] = "该车辆未绑定设备，无法分析产生报表"
			data["code"] = 0
			return
		}
	}

	report := new(models.Reports)
	report.ReportType = 0
	report.ReportPlate = carNum
	report.ReportCompanyId = comId
	report.ReportDataFrom = reportFrom
	str := bson.NewObjectId().Hex()
	numStr := utils.SubString(str, 12)
	report.ReportNumber = numStr
	err = report.Insert()
	if err != nil {
		data["error"] = "上传数据失败"
		data["code"] = 0
		return
	}
	task := models.Task{}
	reportId := report.ReportId.Hex()
	task.Path = open
	task.ReportId = reportId
	task.CompanyId = comId
	task.CarNum = carNum
	task.Type = 0

	err = new(models.Redis).ListPush("analysis_tasks", task)
	if err != nil {
		data["error"] = "建立任务失败"
		data["code"] = 0
		return
	}
	amount--
	am := Session.User.Amount
	am.QueryAiCar = amount
	new(models.Redis).Save("amounts", comId, am)
	data["code"] = 1
	statuCode = 200
	return
}

func (this *ReportView) Delete(ctx iris.Context) (statusCode int, result interface{}) {
	data := make(app.M)
	defer func() {
		result = data
	}()
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

//更新操作待用
func (this *ReportView) Put(ctx iris.Context) (statuCode int, data interface{}) {
	return
}
