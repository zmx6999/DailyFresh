package controllers

import (
	"181112/models"
	"github.com/astaxie/beego/orm"
	"time"
	"strings"
	"strconv"
	"github.com/gomodule/redigo/redis"
	"github.com/astaxie/beego"
)

type OrderController struct {
	BaseController
}

func (this *OrderController) AddOrder()  {
	goodsIds:=strings.Split(this.GetString("goods_id"),",")
	if len(goodsIds)==0 || strings.Trim(goodsIds[0]," ")=="" {
		this.Data["json"]=&ResponseJSON{-1,nil,"no goods"}
		this.ServeJSON()
		return
	}

	username:=this.GetUser()
	if username=="" {
		this.Data["json"]=&ResponseJSON{-1,nil,"please login"}
		this.ServeJSON()
		return
	}
	var user models.User
	user.Name=username
	o:=orm.NewOrm()
	o.Read(&user,"Name")

	addressId,_:=this.GetInt("address_id")
	var address models.Address
	if err:=o.QueryTable("Address").RelatedSel("User").Filter("User__Name",username).Filter("Id",addressId).One(&address);err!=nil {
		this.Data["json"]=&ResponseJSON{-1,nil,"invalid address"}
		this.ServeJSON()
		return
	}

	paystyle,_:=this.GetInt("paystyle")

	conn,err:=redis.Dial("tcp",beego.AppConfig.String("redis_host")+":"+beego.AppConfig.String("redis_port"))
	if err!=nil {
		this.Data["json"]=&ResponseJSON{-1,nil,"failed to connect to redis"}
		this.ServeJSON()
		return
	}
	defer conn.Close()

	o.Begin()

	var order models.OrderInfo
	loc,_:=time.LoadLocation("Asia/Saigon")
	order.OrderId=""
	order.User=&user
	order.Address=&address
	order.PayMethod=paystyle
	order.TotalCount=0
	order.TotalPrice=0
	order.TransitPrice=10
	order.Orderstatus=1
	order.Time=time.Now().In(loc)
	if _,err:=o.Insert(&order);err!=nil {
		o.Rollback()
		this.Data["json"]=&ResponseJSON{-1,nil,"failed to add order"}
		this.ServeJSON()
		return
	}

	totalnum:=0
	totalprice:=0
	for _,v:=range goodsIds{
		var ordergoods models.OrderGoods
		ordergoods.OrderInfo=&order

		goodsId,_:=strconv.Atoi(v)
		var goods models.GoodsSKU
		goods.Id=goodsId
		o.Read(&goods)
		ordergoods.GoodsSKU=&goods

		num,_:=redis.Int(conn.Do("hget","cart_"+strconv.Itoa(user.Id),goodsId))
		if num>goods.Stock {
			o.Rollback()
			this.Data["json"]=&ResponseJSON{-1,nil,goods.Name+" shortage"}
			this.ServeJSON()
			return
		}
		ordergoods.Count=num
		ordergoods.Price=goods.Price*num

		if _,err:=o.Insert(&ordergoods);err!=nil {
			o.Rollback()
			this.Data["json"]=&ResponseJSON{-1,nil,"failed to add order"}
			this.ServeJSON()
			return
		}

		goods.Stock-=num
		goods.Sales+=num
		if _,err:=o.Update(&goods);err!=nil {
			o.Rollback()
			this.Data["json"]=&ResponseJSON{-1,nil,"failed to add order"}
			this.ServeJSON()
			return
		}

		conn.Do("hdel","cart_"+strconv.Itoa(user.Id),goodsId)

		totalnum+=num
		totalprice+=goods.Price*num
	}

	order.TotalPrice=totalprice+order.TransitPrice
	order.TotalCount=totalnum
	order.OrderId=this.GetId(order.Id,"OrderInfo",10)
	if _,err=o.Update(&order);err!=nil {
		o.Rollback()
		this.Data["json"]=&ResponseJSON{-1,nil,"failed to add order"}
		this.ServeJSON()
		return
	}

	o.Commit()
	this.Data["json"]=&ResponseJSON{0,nil,"OK"}
	this.ServeJSON()
}