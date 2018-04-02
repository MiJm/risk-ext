package views

import (
	"fmt"
	"risk-ext/models"
	"strconv"
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
	reportId := ctx.Params().Get("report_id")
	p, err := strconv.ParseInt(page, 10, 64)
	if err != nil {
		p = 1
	}
	s, err := strconv.ParseInt(size, 10, 64)
	if err != nil {
		s = 30
	}
	companyId := Session.User.UserCompany_id
	query := bson.M{}
	query["report_company_id"] = companyId
	if reportId != "" {
		query["_id"] = reportId
	}
	report := new(models.Reports)
	rs, num, err := report.Lists(query, int(p), int(s))
	if err != nil {
		data["error"] = err
		return
	} else {
		data = make(M)
		data["ai_amount"] = Session.User.Amount.QueryAiCar
		if len(rs) > 0 {
			data["list"] = rs
			data["num"] = num
		}
		data["code"] = 1
		statuCode = 200

	}
	return
}

//新增Report记录，发送获取Report请求
func (this *ReportView) Post(ctx iris.Context) (statuCode int, data interface{}) {
	statuCode = 400
	if Session.User.Amount.QueryAiCar <= 0 {
		data = "查询次数不足"
		return
	}
	//	token := ctx.PostValue("token")
	carNum := ctx.FormValueDefault("car_num", "")
	fmt.Println(Session)
	//	var rs = ""
	var reportFrom uint8
	if carNum == "" {
		reportFrom = 1
		//		title := []string{"device_latlng", "device_slatlng", "device_speed", "device_address", "device_loctype", "device_loctime"}
		//		f, head, err := ctx.FormFile("")
		//		b := make([]byte, head.Size)
		//		n, err := f.Read(b)
		//		if err != nil {
		//			data = "读取文件失败"
		//			return
		//		}
		//		da := string(b)
		//		str := make([]string, 0)
		//		result := strings.Split(da, "\n")
		//		for _, v := range result {
		//			v1 := strings.Split("v", ",")
		//			ma := make(map[string]string)
		//			for j, k := range v1 {
		//				ma[title[j]] = k
		//			}
		//			s, err := json.Marshal(ma)
		//			if err != nil {
		//				return
		//			}
		//			str = append(str, string(s))
		//		}
		//		openUrl := "devices/" + time.Now().Format("200601") + "/"
		//		saveUrl := beego.AppConfig.String("CarExport") + time.Now().Format("200601") + "/"
		//		err = utils.IsFile(saveUrl)
		//		if err != nil {
		//			return
		//		}
		//		saveUrl = fmt.Sprintf("%s%s轨迹.txt", saveUrl, carNum)
		//		openUrl = fmt.Sprintf("%s%s轨迹.txt", openUrl, carNum)
		//		err = ioutil.WriteFile(saveUrl, []byte(str), 0644)

	} else {
		//		url := "http://192.168.1.118:8080/v1/routes/analyse_track"
		//		req := httplib.Get(url)
		//		req.Header("Content-Type", "application/json;charset=UTF-8")
		//		req.Param("carNum", carNum)
		//		req.Param("token", token)
		//		rs1, err := req.String()
		//		if err != nil {
		//			data = "请求轨迹报表失败"
		//			return
		//		}
		//		rs = rs1
	}

	statuCode = 204
	//	data = rs
	report := new(models.Reports)
	report.ReportType = 0
	report.ReportPlate = carNum
	report.ReportDataFrom = reportFrom
	report.ReportOpenId = "123456"
	report.ReportCreateAt = time.Now().Unix()
	report.ReportCompanyId = Session.User.UserCompany_id
	report.Insert()
	return
}

//更新分享人信息
func (this *ReportView) Put(ctx iris.Context) (statusCode int, data interface{}) {
	statusCode = 400
	if Session.User.Amount.QueryAiCar <= 0 {
		data = "查询次数不足"
		return
	}
	reportId := ctx.FormValue("reportId")
	flag := bson.IsObjectIdHex(reportId)
	if !flag {
		data = "参数有误"
		return
	}
	phone := ctx.FormValue("phone")
	fname := ctx.FormValue("fname")
	if phone == "" || fname == "" {
		data = "请输入完整参数"
		return
	}
	rs, err := new(models.Reports).One(reportId)
	if err != nil {
		data = "参数有误，无数据"
		return
	}
	shareId := new(bson.ObjectId)
	shareUser := models.Shares{}
	shareUser.ShareId = shareId.Hex()
	shareUser.ShareFname = fname
	shareUser.ShareMobile = phone
	shareUser.ShareCreateAt = time.Now().Unix()
	rs.ReportShares[shareId.Hex()] = shareUser
	err = rs.Update()
	if err != nil {
		data = "添加分享人失败"
		return
	}
	statusCode = 204
	return

}
