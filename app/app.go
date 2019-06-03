package app

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"risk-ext/config"
	"strings"

	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
)

//view 接口
type V interface {
	Auth(iris.Context) int
	Get(iris.Context) (int, interface{})
	Post(iris.Context) (int, interface{})
	Put(iris.Context) (int, interface{})
	Delete(iris.Context) (int, interface{})
}

type M map[string]interface{}

var (
	app    = iris.New()
	method = []string{"Get", "Post", "Put", "Delete"}
	paths  = make(map[string]V)
)

var conf = iris.Configuration{ // default configuration:
	DisableStartupLog:                 false,
	DisableInterruptHandler:           false,
	DisablePathCorrection:             false,
	EnablePathEscape:                  false,
	FireMethodNotAllowed:              false,
	DisableBodyConsumptionOnUnmarshal: false,
	EnableOptimizations:               true,
	DisableAutoFireStatusCode:         false,
	TimeFormat:                        "2006-1-2 15:04:05",
	Charset:                           "UTF-8",
}

func init() {
	if config.GetBool("debug") {
		app.Logger().SetLevel("debug")
	}

	if config.GetBool("logs") {
		app.Use(recover.New())
		app.Use(logger.New())
	}
}
func AddPath(path string, view V) {
	paths[path] = view
}

func callback(ctx iris.Context, v V) {
	authResult := v.Auth(ctx)
	if authResult == 403 {
		ctx.StatusCode(403)
		ctx.JSON("没有权限")
		return
	} else if authResult == 401 {
		ctx.StatusCode(401)
		ctx.JSON("登录失效")
		return
	}
	var code, data = 404, M{}
	var m = ctx.Method()
	switch m {
	case "GET":
		code, data = v.Get(ctx)
	case "POST":
		code, data = v.Post(ctx)
	case "PUT":
		code, data = v.Put(ctx)
	case "DELETE":
		code, data = v.Delete(ctx)
		break
	default:
		code, data = 404, M{}
	}
	ctx.StatusCode(code)
	ctx.JSON(data)
}

func App() *iris.Application {
	v1 := app.Party("v2")
	for k, m := range paths {
		v1.Get(k, func(ctx iris.Context) {
			callback(ctx, m)
		})
	}
	return app
}

func Run() {
	app := App()
	host := config.GetString("host")
	port := config.GetString("port")
	if port == "" {
		port = "80"
	}
	static := config.GetString("staticPath")
	if strings.TrimSpace(static) != "" {
		staticArr := strings.Split(static, " ")
		for _, item := range staticArr {
			if strings.TrimSpace(item) != "" {
				static_items := strings.Split(item, ":")
				if len(static_items) == 2 {
					app.StaticWeb(static_items[0], static_items[1])
				}
			}

		}
	}
	app.Run(iris.Addr(host+":"+port), iris.WithConfiguration(conf))
}
func HttpClient(url string, args interface{}, method string, result interface{}, contentType string, token ...string) (err error, jsonStr string) {
	client := &http.Client{}
	var params = ""
	//	var contentType = "application/x-www-form-urlencoded"

	if reflect.TypeOf(args).String() != "string" {
		jsonData, err := json.Marshal(args)
		if err != nil {
			return err, ""
		}
		params = string(jsonData)
	} else {
		params = args.(string)
	}

	var req *http.Request
	if method == "GET" {
		req, err = http.NewRequest(method, url, nil)
	} else {
		req, err = http.NewRequest(method, url, strings.NewReader(params))
	}
	if err != nil {
		return err, ""
	}

	req.Header.Add("Content-Type", contentType)

	if len(token) != 0 {
		req.Header.Add("Authorization", token[0])
	}
	resp, err := client.Do(req)
	if err != nil {
		return err, ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err, ""
	}
	jsonStr = string(body)
	//	err = json.Unmarshal([]byte(jsonStr), result)
	return err, jsonStr
}
