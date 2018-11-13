package controllers

import (
	"181112/models"
	"github.com/astaxie/beego/orm"
	"math"
	"github.com/gomodule/redigo/redis"
	"github.com/astaxie/beego"
	"strconv"
)

type GoodsController struct {
	BaseController
}

func (this *GoodsController) Index()  {
	types:=this.GetTypes()
	this.Data["types"]=types

	var bannerList []models.IndexGoodsBanner
	o:=orm.NewOrm()
	o.QueryTable("IndexGoodsBanner").OrderBy("Index").All(&bannerList)
	this.Data["bannerList"]=bannerList

	var promotionList []models.IndexPromotionBanner
	o.QueryTable("IndexPromotionBanner").OrderBy("Index").All(&promotionList)
	this.Data["promotionList"]=promotionList

	var goodsList []map[string]interface{}
	for _,v:=range types{
		_v:=make(map[string]interface{})
		_v["type"]=v

		var textGoodsList []models.IndexTypeGoodsBanner
		o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType","GoodsSKU").Filter("GoodsType__Name",v.Name).Filter("DisplayType",0).OrderBy("Index").All(&textGoodsList)
		_v["textGoodsList"]=textGoodsList

		var imageGoodsList []models.IndexTypeGoodsBanner
		o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType","GoodsSKU").Filter("GoodsType__Name",v.Name).Filter("DisplayType",1).OrderBy("Index").All(&imageGoodsList)
		_v["imageGoodsList"]=imageGoodsList

		goodsList=append(goodsList,_v)
	}
	this.Data["goodsList"]=goodsList

	this.ShowGoodsLayout("首页",true)
	this.TplName="goods/index.html"
}

func (this *GoodsController) Detail()  {
	id,err:=this.GetInt("id")
	if err!=nil {
		this.Redirect("/",302)
		return
	}

	var goods models.GoodsSKU
	goods.Id=id
	o:=orm.NewOrm()
	if err:=o.QueryTable("GoodsSKU").RelatedSel("Goods","GoodsType").Filter("Id",id).One(&goods);err!=nil {
		this.Redirect("/",302)
		return
	}
	this.Data["goods"]=goods

	var newgoods []models.GoodsSKU
	o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",goods.GoodsType.Id).OrderBy("-Time").Limit(2,0).All(&newgoods)
	this.Data["newgoods"]=newgoods

	conn,_:=redis.Dial("tcp",beego.AppConfig.String("redis_host")+":"+beego.AppConfig.String("redis_port"))
	defer conn.Close()
	username:=this.GetUser()
	if username!="" {
		var user models.User
		user.Name=username
		o.Read(&user,"Name")
		conn.Do("lrem","history_"+strconv.Itoa(user.Id),0,id)
		conn.Do("lpush","history_"+strconv.Itoa(user.Id),id)
	}

	this.ShowGoodsLayout("商品详情-"+goods.Name,false)
	this.TplName="goods/detail.html"
}

func (this *GoodsController) List()  {
	id,err:=this.GetInt("id")
	if err!=nil {
		this.Redirect("/",302)
		return
	}

	var goodsType models.GoodsType
	goodsType.Id=id
	o:=orm.NewOrm()
	if err:=o.Read(&goodsType);err!=nil {
		this.Redirect("/",302)
		return
	}
	this.Data["goodsType"]=goodsType

	var goodsList []models.GoodsSKU
	s:=o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",goodsType.Id)
	sort:=this.GetString("sort")
	if sort=="price" {
		s=s.OrderBy("Price")
	} else if sort=="sales" {
		s=s.OrderBy("Sales")
	}
	totalRows,_:=s.Count()
	pageSize:=3
	pageCount:=int(math.Ceil(float64(totalRows)/float64(pageSize)))
	page,err:=this.GetInt("p")
	if err!=nil {
		page=1
	}
	if page<1 {
		page=1
	}
	if page>pageCount {
		page=pageCount
	}
	s.Limit(pageSize,pageSize*(page-1)).All(&goodsList)
	this.Data["goodsList"]=goodsList
	this.Data["page"]=page
	this.Data["pageCount"]=pageCount
	this.Data["pages"]=this.PageTool(page,pageCount)

	var newgoods []models.GoodsSKU
	o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id",goodsType.Id).OrderBy("-Time").Limit(2,0).All(&newgoods)
	this.Data["newgoods"]=newgoods

	this.Data["sort"]=sort

	this.ShowGoodsLayout("商品列表-"+goodsType.Name,false)
	this.TplName="goods/list.html"
}

func (this *GoodsController) Search()  {
	name:=this.GetString("goodsname")
	var goodsList []models.GoodsSKU
	o:=orm.NewOrm()
	o.QueryTable("GoodsSKU").Filter("Name__icontains",name).All(&goodsList)
	this.Data["goodsList"]=goodsList
	this.ShowGoodsLayout("商品搜索",false)
	this.TplName="goods/search.html"
}