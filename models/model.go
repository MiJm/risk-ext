package models

import (
	"encoding/json"
	"math/rand"
	"reflect"
	"regexp"
	"risk-ext/config"
	"strconv"
	"strings"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	colls   = make(map[string]*mgo.Collection)
	EmptyId bson.ObjectId //空objectId
)

const (
	MANAGER_ADMIN     = 0
	MANAGER_SERVICE   = 1
	MANAGER_STORE     = 2
	MANAGER_ASSISTANT = 3

	MEMBER_SUPER   = 2
	MEMBER_ADMIN   = 1
	MEMBER_GENERAL = 0
	MEMBER_STORE   = 3

	HTTP_OK_200                  = 200
	HTTP_100_Continue            = 100
	HTTP_101_Switching_Protocols = 101
	HTTP_102_Processing          = 102
)

type Model struct {
}

type Redis struct {
}

func initColl(coll interface{}) string {
	key := _disName(coll)
	if colls[key] == nil {
		colls[key] = config.Mongo.C(key)
	}
	return key
}

func (this *Redis) Map(key, field string, result interface{}) (err error) {

	data, err := config.Redis.HGet(key, field).Bytes()

	if err != nil {
		return
	}
	err = json.Unmarshal(data, result)
	return
}

func (this *Redis) Get(key string) (rs string, err error) {
	rs, err = config.Redis.Get(key).Result()
	return
}

func (this *Redis) Delete(key string) (err error) {
	err = config.Redis.Del(key).Err()
	return
}

func (this *Redis) Save(key, field string, result interface{}) (err error) {
	data, err := json.Marshal(result)
	if err != nil {
		return
	}
	strData := string(data)
	err = config.Redis.HSet(key, field, []byte(strData)).Err()
	return
}

func (this *Redis) ListPush(key string, result interface{}) (err error) {
	data, err := json.Marshal(result)
	if err != nil {
		return
	}
	strData := string(data)
	err = config.Redis.LPush(key, []byte(strData)).Err()
	return
}

func (this *Model) Collection(coll interface{}) (c *mgo.Collection) {
	key := initColl(coll)
	c = colls[key]
	return
}

func (this *Model) RouteCollection(key string) (c *mgo.Collection) {
	c = config.RouteMongo.C(key)
	return
}

func (this *Model) One(coll interface{}) {
	key := initColl(coll)
	c := colls[key]
	v := reflect.ValueOf(coll).Elem()

	if &v != nil {
		fieldName := strings.ToUpper(key[0:1]) + key[1:len(key)-1] + "Id"
		idobj := v.FieldByName(fieldName)
		if idobj.IsValid() {
			c.FindId(idobj.Interface()).One(coll)
		}

	}
}

func _disName(coll interface{}) string {
	collName := reflect.TypeOf(coll).String()
	collName = strings.ToLower(collName)
	strs := strings.Split(collName, ".")
	key := strs[len(strs)-1]
	return key
}

//检测手机号码是否合法
func CheckPhone(phone string) bool {
	reg := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return reg.MatchString(phone)
}

func GetRandCode() string {
	rand.Seed(time.Now().Unix())
	var code string = strconv.Itoa(rand.Intn(10)) +
		strconv.Itoa(rand.Intn(10)) +
		strconv.Itoa(rand.Intn(10)) +
		strconv.Itoa(rand.Intn(10))

	return code
}
