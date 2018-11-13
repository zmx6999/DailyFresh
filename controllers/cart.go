package controllers

import (
	"github.com/gomodule/redigo/redis"
	"github.com/astaxie/beego"
	"181112/models"
	"github.com/astaxie/beego/orm"
	"strconv"
)

type CartController struct {
	BaseController
}

func (this *CartController) AddCart()  {
	username:=this.GetUser()
	if username=="" {
		this.Data["json"]=&ResponseJSON{-1,nil,"please login"}
		this.ServeJSON()
		return
	}

	goodsId,err1:=this.GetInt("goods_id")
	num,err2:=this.GetInt("num")
	if err1!=nil || err2!=nil {
		this.Data["json"]=&ResponseJSON{-1,nil,"parameter error"}
		this.ServeJSON()
		return
	}

	conn,err:=redis.Dial("tcp",beego.AppConfig.String("redis_host")+":"+beego.AppConfig.String("redis_port"))
	if err!=nil {
		this.Data["json"]=&ResponseJSON{-1,nil,"failed to connect redis server"}
		this.ServeJSON()
		return
	}
	defer conn.Close()

	var user models.User
	user.Name=username
	o:=orm.NewOrm()
	o.Read(&user,"Name")

	currentNum,_:=redis.Int(conn.Do("hget","cart_"+strconv.Itoa(user.Id),goodsId))
	conn.Do("hset","cart_"+strconv.Itoa(user.Id),goodsId,currentNum+num)

	cartnum:=0
	cartdata,_:=redis.IntMap(conn.Do("hgetall","cart_"+strconv.Itoa(user.Id)))
	for _,v:=range cartdata{
		cartnum+=v
	}
	this.Data["json"]=&ResponseJSON{0,cartnum,"OK"}
	this.ServeJSON()
}

func (this *CartController) UpdateCart()  {
	username:=this.GetUser()
	if username=="" {
		this.Data["json"]=&ResponseJSON{-1,nil,"please login"}
		this.ServeJSON()
		return
	}

	goodsId,err1:=this.GetInt("goods_id")
	num,err2:=this.GetInt("num")
	if err1!=nil || err2!=nil {
		this.Data["json"]=&ResponseJSON{-1,nil,"parameter error"}
		this.ServeJSON()
		return
	}

	conn,err:=redis.Dial("tcp",beego.AppConfig.String("redis_host")+":"+beego.AppConfig.String("redis_port"))
	if err!=nil {
		this.Data["json"]=&ResponseJSON{-1,nil,"failed to connect redis server"}
		this.ServeJSON()
		return
	}
	defer conn.Close()

	var user models.User
	user.Name=username
	o:=orm.NewOrm()
	o.Read(&user,"Name")

	conn.Do("hset","cart_"+strconv.Itoa(user.Id),goodsId,num)

	this.Data["json"]=&ResponseJSON{0,nil,"OK"}
	this.ServeJSON()
}

func (this *CartController) DeleteCart()  {
	username:=this.GetUser()
	if username=="" {
		this.Data["json"]=&ResponseJSON{-1,nil,"please login"}
		this.ServeJSON()
		return
	}

	goodsId,err:=this.GetInt("goods_id")
	if err!=nil {
		this.Data["json"]=&ResponseJSON{-1,nil,"parameter error"}
		this.ServeJSON()
		return
	}

	conn,err:=redis.Dial("tcp",beego.AppConfig.String("redis_host")+":"+beego.AppConfig.String("redis_port"))
	if err!=nil {
		this.Data["json"]=&ResponseJSON{-1,nil,"failed to connect redis server"}
		this.ServeJSON()
		return
	}
	defer conn.Close()

	var user models.User
	user.Name=username
	o:=orm.NewOrm()
	o.Read(&user,"Name")

	conn.Do("hdel","cart_"+strconv.Itoa(user.Id),goodsId)

	this.Data["json"]=&ResponseJSON{0,nil,"OK"}
	this.ServeJSON()
}