package views

import (
	"github.com/astaxie/beego/httplib"
	"github.com/kataras/iris"
)

type TrackView struct {
	Views
}

func (this *TrackView) Auth(ctx iris.Context) bool {
	this.Views.Auth(ctx)
	var perms = PMS{
		"PUT": MA{"ADMIN": A{MANAGER_ADMIN, MANAGER_ASSISTANT, MANAGER_SERVICE}, "USER": A{}},
		"GET": MA{"ADMIN": A{MANAGER_ADMIN, MANAGER_ASSISTANT, MANAGER_SERVICE}, "USER": A{MEMBER_SUPER, MEMBER_ADMIN}, "NOLOGIN": A{1}}}
	return this.CheckPerms(perms[ctx.Method()])
}

func (this *TrackView) Get(ctx iris.Context) (statuCode int, data interface{}) {
	statuCode = 400
	token := ctx.FormValueDefault("token", "")
	carNum := ctx.FormValueDefault("car_num", "")
	var rs = ""
	if carNum == "" {
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
		rs = rs1
	}

	statuCode = 200
	data = rs
	return
}
