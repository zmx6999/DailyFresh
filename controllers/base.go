package controllers

import (
	"github.com/astaxie/beego"
	"crypto/sha256"
	"encoding/hex"
	"181112/models"
	"github.com/gomodule/redigo/redis"
	"encoding/gob"
	"bytes"
	"github.com/astaxie/beego/orm"
	"strconv"
)

type BaseController struct {
	beego.Controller
}

type ResponseJSON struct {
	Code int `json:"code"`
	Data interface{} `json:"data"`
	Msg string `json:"msg"`
}

func (this *BaseController) Sha256Str(x string) string {
	y:=sha256.Sum256([]byte(x))
	return hex.EncodeToString(y[:])
}

func (this *BaseController) GetUser() string {
	username:=this.GetSession("username")
	if username==nil {
		return ""
	} else {
		return username.(string)
	}
}

func (this *BaseController) GetTypes() []models.GoodsType {
	var types []models.GoodsType
	conn,_:=redis.Dial("tcp",beego.AppConfig.String("redis_host")+":"+beego.AppConfig.String("redis_port"))
	defer conn.Close()
	tb,_:=redis.Bytes(conn.Do("get","types"))
	decoder:=gob.NewDecoder(bytes.NewReader(tb))
	decoder.Decode(&types)
	if len(types)==0 {
		o:=orm.NewOrm()
		o.QueryTable("GoodsType").All(&types)
		var buffer bytes.Buffer
		encoder:=gob.NewEncoder(&buffer)
		encoder.Encode(&types)
		conn.Do("set","types",buffer.Bytes())
	}
	return types
}

func (this *BaseController) ShowGoodsLayout(title string,isHome bool)  {
	this.Data["username"]=this.GetUser()
	if !isHome {
		this.Data["types"]=this.GetTypes()
	}
	this.Data["isHome"]=isHome
	this.Data["title"]=title
	this.Data["goodsname"]=this.GetString("goodsname")
	this.Data["cartnum"]=this.GetCartNum()
	this.Layout="goods/layout.html"
}

func (this *BaseController) PageTool(page int,pageCount int) []int {
	if pageCount<5 {
		pages:=make([]int,pageCount)
		for i:=0; i<pageCount; i++ {
			pages[i]=i+1
		}
		return pages
	} else if page<3 {
		return []int{1,2,3,4,5}
	} else if page>pageCount-2 {
		return []int{pageCount-4,pageCount-3,pageCount-2,pageCount-1,pageCount}
	} else {
		return []int{page-2,page-1,page,page+1,page+2}
	}
}

func (this *BaseController) ShowUserLayout(title string)  {
	this.Data["username"]=this.GetUser()
	this.Data["title"]=title
	this.Layout="user/layout.html"
}

func (this *BaseController) GetCartNum() int {
	username:=this.GetUser()
	if username=="" {
		return 0
	}

	conn,err:=redis.Dial("tcp",beego.AppConfig.String("redis_host")+":"+beego.AppConfig.String("redis_port"))
	if err!=nil {
		return 0
	}
	defer conn.Close()

	var user models.User
	user.Name=username
	o:=orm.NewOrm()
	o.Read(&user,"Name")

	cartdata,_:=redis.IntMap(conn.Do("hgetall","cart_"+strconv.Itoa(user.Id)))

	cartnum:=0
	for _,v:=range cartdata{
		cartnum+=v
	}
	return cartnum
}