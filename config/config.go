package config

import (
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/go-redis/redis"

	"gopkg.in/mgo.v2"
	"gopkg.in/yaml.v2"
)

type M map[interface{}]interface{}

var (
	_config = M{}
	Mongo   *mgo.Database
	Redis   *redis.Client
)

func Type(obj interface{}) string {
	return reflect.TypeOf(obj).String()
}

func init() {
	paths := os.Args
	path := "" //配置文件路径
	if len(paths) > 1 {
		path = paths[1]
	}

	var yamlFile []byte
	var err error

	if path != "" {
		yamlFile, err = ioutil.ReadFile(path)
	} else {
		yamlFile, err = ioutil.ReadFile("app.conf")
	}

	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(yamlFile, &_config)
	if err != nil {
		log.Fatal(err)
	}
	initConfig()
	initMongo()
	initRedis()
}

func initConfig() {
	model := _config["model"]

	if Type(model) == "string" {
		modelMap := _config[model.(string)]
		if Type(modelMap) == "config.M" {
			mm := modelMap.(M)
			for k, v := range mm {
				_config[k.(string)] = v
			}
			delete(_config, model.(string))
		}
	}
}

func Get(key string) interface{} {
	return _config[key]
}

func GetString(key string) string {
	val := _config[key]
	if Type(val) == "string" {
		return _config[key].(string)
	}
	return ""
}

func GetBool(key string) bool {
	val := _config[key]
	if Type(val) == "bool" {
		return _config[key].(bool)
	}
	return false
}

func GetInt(key string) int {
	val := _config[key]
	if Type(val) == "int" {
		return _config[key].(int)
	}
	return 0
}

func initMongo() {
	db := Get("db")

	if Type(db) != "config.M" {
		return
	}

	dbMap := db.(M)

	if dbMap["type"] != "mongodb" {
		return
	}
	hosts := dbMap["host"].([]interface{})
	dbName := dbMap["name"].(string)
	user := dbMap["user"].(string)
	pwd := dbMap["pwd"].(string)

	var hs = make([]string, 0)

	for _, h := range hosts {
		if Type(h) == "string" && h != "" {
			hs = append(hs, h.(string))
		}
	}

	info := &mgo.DialInfo{
		Addrs:    hs,
		Database: dbName,
		Timeout:  60 * time.Second,
		Username: user,
		Password: pwd,
	}
	session, err := mgo.DialWithInfo(info)
	if err != nil {
		log.Fatal("mongodb:", err)
	}
	session.SetMode(mgo.Eventual, true)
	Mongo = session.DB(dbName)
}

func initRedis() {
	redis_ := Get("redis")

	if Type(redis_) != "config.M" {
		return
	}
	redisMap := redis_.(M)
	Redis = redis.NewClient(&redis.Options{
		Addr:     redisMap["host"].(string),
		Password: redisMap["pwd"].(string), // no password set
		DB:       redisMap["name"].(int),   // use default DB
	})
}
