package views

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"new/system/utils"
	"risk-ext/models"
	"strconv"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
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
func (this *ReportView) Post(ctx iris.Context) (statuCode int, data interface{}) {
	open := "http://192.168.1.118:8080/"
	statuCode = 400
	if Session.User.Amount.QueryAiCar <= 0 {
		data = "查询次数不足"
		return
	}
	token := ctx.PostValue("token")
	carNum := ctx.FormValueDefault("car_num", "")
	var rs = ""
	var reportFrom uint8
	if carNum == "" {
		reportFrom = 1
		title := []string{"device_latlng", "device_slatlng", "device_speed", "device_address", "device_loctype", "device_loctime"}
		f, head, err := ctx.FormFile("")
		b := make([]byte, head.Size)
		_, err = f.Read(b)
		if err != nil {
			data = "读取文件失败"
			return
		}
		da := string(b)
		str := "["
		result := strings.Split(da, "\n")
		for k, v := range result {
			v1 := strings.Split(v, ",")
			ma := make(map[string]string)
			for j, k := range v1 {
				ma[title[j]] = k
			}
			s, err := json.Marshal(ma)
			if err != nil {
				return
			}

			if k == len(result)-1 {
				str = fmt.Sprint("s%,s%]", str, s)
			} else {
				str = fmt.Sprint("s%,s%", str, s)
			}
		}
		openUrl := "devices/" + time.Now().Format("200601") + "/"
		saveUrl := beego.AppConfig.String("CarExport") + time.Now().Format("200601") + "/"
		err = utils.IsFile(saveUrl)
		if err != nil {
			return
		}
		saveUrl = fmt.Sprintf("%s%s轨迹.txt", saveUrl, carNum)
		openUrl = fmt.Sprintf("%s%s轨迹.txt", openUrl, carNum)
		err = ioutil.WriteFile(saveUrl, []byte(str), 0644)
		open = open + openUrl
		fmt.Println(open)

	} else {
		url := "http://192.168.1.118:8080/v1/routes/analyse_track"
		req := httplib.Get(url)
		req.Header("Content-Type", "application/json;charset=UTF-8")
		req.Param("carNum", carNum)
		req.Param("token", token)
		rs1, err := req.String()
		if err != nil {
			data = "请求轨迹报表失败"
			return
		}
		result := Result{}
		err = json.Unmarshal([]byte(rs1), &result)
		data := result.Data
		open = open + data.(string)
		fmt.Println(open)
	}

	statuCode = 204
	report := new(models.Reports)
	report.ReportType = 0
	report.ReportPlate = carNum
	report.ReportDataFrom = reportFrom
	report.ReportOpenId = "123456"
	report.ReportCreateAt = time.Now().Unix()
	report.ReportCompanyId = Session.User.UserCompany_id
	report.Insert()
	data = rs
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
