package controllers

import (
	"regexp"
	"181112/models"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/utils"
	"github.com/astaxie/beego"
	"strconv"
	"encoding/base64"
	"github.com/gomodule/redigo/redis"
	"math"
	"strings"
)

type UserController struct {
	BaseController
}

func (this *UserController) ShowRegister()  {
	this.TplName="user/register.html"
}

func (this *UserController) HandleRegister()  {
	name:=this.GetString("user_name")
	pwd:=this.GetString("pwd")
	cpwd:=this.GetString("cpwd")
	email:=this.GetString("email")
	if name=="" || pwd=="" || cpwd=="" || email=="" {
		this.Data["json"]=&ResponseJSON{-1,nil,"imcomplete"}
		this.ServeJSON()
		return
	}
	allow:=this.GetString("allow")
	if allow=="" {
		this.Data["json"]=&ResponseJSON{-1,nil,"please agree"}
		this.ServeJSON()
		return
	}
	if pwd!=cpwd {
		this.Data["json"]=&ResponseJSON{-1,nil,"password not match"}
		this.ServeJSON()
		return
	}
	re:=regexp.MustCompile(`^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`)
	if re.FindString(email)=="" {
		this.Data["json"]=&ResponseJSON{-1,nil,"error email format"}
		this.ServeJSON()
		return
	}

	var user models.User
	user.Name=name
	user.PassWord=this.Sha256Str(pwd)
	user.Email=email
	o:=orm.NewOrm()
	if _,err:=o.Insert(&user);err!=nil {
		this.Data["json"]=&ResponseJSON{-1,nil,"failed to register"}
		this.ServeJSON()
		return
	}

	mail:=utils.NewEMail(`{"username":"563364657@qq.com","password":"cgapyzgkkczubdea","host":"smtp.qq.com","port":587}`)
	mail.From="563364657@qq.com"
	mail.To=[]string{email}
	mail.Subject="Register"
	mail.Text="Click http://"+beego.AppConfig.String("host")+":"+beego.AppConfig.String("httpport")+"/active?id="+strconv.Itoa(user.Id)
	mail.Send()

	this.Data["json"]=&ResponseJSON{0,nil,"OK"}
	this.ServeJSON()
}

func (this *UserController) Active()  {
	id,err:=this.GetInt("id")
	if err!=nil {
		this.Ctx.WriteString("not found")
		return
	}

	var user models.User
	user.Id=id
	o:=orm.NewOrm()
	if err:=o.Read(&user);err!=nil {
		this.Ctx.WriteString("not found")
		return
	}

	user.Active=true
	if _,err:=o.Update(&user);err!=nil {
		this.Ctx.WriteString("failed to active")
		return
	}

	this.Redirect("/login",302)
}

func (this *UserController) ShowLogin()  {
	username:=this.Ctx.GetCookie("username")
	if username!="" {
		t,_:=base64.StdEncoding.DecodeString(username)
		this.Data["username"]=string(t)
		this.Data["remember"]="checked"
	}
	this.TplName="user/login.html"
}

func (this *UserController) HandleLogin()  {
	name:=this.GetString("username")
	pwd:=this.GetString("pwd")
	if name=="" || pwd=="" {
		this.Data["json"]=&ResponseJSON{-1,nil,"imcomplete"}
		this.ServeJSON()
		return
	}

	var user models.User
	user.Name=name
	o:=orm.NewOrm()
	if err:=o.Read(&user,"Name");err!=nil {
		this.Data["json"]=&ResponseJSON{-1,nil,"user not exist"}
		this.ServeJSON()
		return
	}
	if user.PassWord!=this.Sha256Str(pwd) {
		this.Data["json"]=&ResponseJSON{-1,nil,"password error"}
		this.ServeJSON()
		return
	}
	if !user.Active {
		this.Data["json"]=&ResponseJSON{-1,nil,"user not active"}
		this.ServeJSON()
		return
	}

	remember:=this.GetString("remember")
	if remember!="" {
		this.Ctx.SetCookie("username",base64.StdEncoding.EncodeToString([]byte(name)),100)
	} else {
		this.Ctx.SetCookie("username","",-1)
	}

	this.SetSession("username",name)

	this.Data["json"]=&ResponseJSON{0,nil,"OK"}
	this.ServeJSON()
}

func (this *UserController) Logout()  {
	this.DelSession("username")
	this.Redirect("/login",302)
}

func (this *UserController) Info()  {
	username:=this.GetUser()
	this.Data["username"]=username

	var address models.Address
	o:=orm.NewOrm()
	if err:=o.QueryTable("Address").RelatedSel("User").Filter("User__Name",username).Filter("Isdefault",true).One(&address);err==nil {
		this.Data["address"]=address
	} else {
		this.Data["address"]=""
	}

	conn,_:=redis.Dial("tcp",beego.AppConfig.String("redis_host")+":"+beego.AppConfig.String("redis_port"))
	defer conn.Close()
	var user models.User
	user.Name=username
	o.Read(&user,"Name")
	visitedIds,_:=redis.Ints(conn.Do("lrange","history_"+strconv.Itoa(user.Id),0,4))
	var visitedGoodsList []models.GoodsSKU
	for _,visitedId:=range visitedIds{
		var visitedGoods models.GoodsSKU
		visitedGoods.Id=visitedId
		o.Read(&visitedGoods)
		visitedGoodsList=append(visitedGoodsList,visitedGoods)
	}
	this.Data["visitedGoodsList"]=visitedGoodsList

	this.ShowUserLayout("用户中心")
	this.TplName="user/info.html"
}

func (this *UserController) ShowAddress()  {
	username:=this.GetUser()
	var address models.Address
	o:=orm.NewOrm()
	if err:=o.QueryTable("Address").RelatedSel("User").Filter("User__Name",username).Filter("Isdefault",true).One(&address);err==nil {
		this.Data["address"]=address
	} else {
		this.Data["address"]=""
	}

	this.ShowUserLayout("用户中心")
	this.TplName="user/address.html"
}

func (this *UserController) HandleAddAddress()  {
	receiver:=this.GetString("receiver")
	addr:=this.GetString("addr")
	phone:=this.GetString("phone")
	if receiver=="" || addr=="" || phone=="" {
		this.Data["json"]=&ResponseJSON{-1,nil,"imcomplete"}
		this.ServeJSON()
		return
	}

	o:=orm.NewOrm()
	o.Begin()

	var oldaddress models.Address
	username:=this.GetUser()
	if err:=o.QueryTable("Address").RelatedSel("User").Filter("User__Name",username).Filter("Isdefault",true).One(&oldaddress);err==nil {
		oldaddress.Isdefault=false
		if _,err:=o.Update(&oldaddress);err!=nil {
			o.Rollback()
			this.Data["json"]=&ResponseJSON{-1,nil,"failed to add address"}
			this.ServeJSON()
			return
		}
	}

	var address models.Address
	address.Receiver=receiver
	address.Addr=addr
	address.Phone=phone
	address.Isdefault=true

	var user models.User
	user.Name=username
	o.Read(&user,"Name")
	address.User=&user

	if _,err:=o.Insert(&address);err!=nil {
		o.Rollback()
		this.Data["json"]=&ResponseJSON{-1,nil,"failed to add address"}
		this.ServeJSON()
		return
	}

	o.Commit()
	this.Data["json"]=&ResponseJSON{0,nil,"OK"}
	this.ServeJSON()
}

func (this *UserController) Cart()  {
	conn,_:=redis.Dial("tcp",beego.AppConfig.String("redis_host")+":"+beego.AppConfig.String("redis_port"))
	defer conn.Close()

	var user models.User
	user.Name=this.GetUser()
	o:=orm.NewOrm()
	o.Read(&user,"Name")

	cartdata,_:=redis.IntMap(conn.Do("hgetall","cart_"+strconv.Itoa(user.Id)))

	totalnum:=0
	totalprice:=0
	var goodsList []map[string]interface{}
	for k,v:=range cartdata{
		_v:=make(map[string]interface{})

		var goods models.GoodsSKU
		goodsId,_:=strconv.Atoi(k)
		goods.Id=goodsId
		o.Read(&goods)
		_v["goods"]=goods

		_v["num"]=v
		_v["totalPrice"]=goods.Price*v

		totalnum+=v
		totalprice+=goods.Price*v

		goodsList=append(goodsList,_v)
	}
	this.Data["goodsList"]=goodsList
	this.Data["totalnum"]=totalnum
	this.Data["totalprice"]=totalprice

	this.ShowUserLayout("购物车")
	this.TplName="user/cart.html"
}

func (this *UserController) PlaceOrder()  {
	goodsIds:=this.GetStrings("goods_id")
	if len(goodsIds)==0 {
		this.Redirect("/",302)
		return
	}

	username:=this.GetUser()
	o:=orm.NewOrm()
	var addressList []models.Address
	o.QueryTable("Address").RelatedSel("User").Filter("User__Name",username).OrderBy("-Isdefault").All(&addressList)
	this.Data["addressList"]=addressList

	conn,_:=redis.Dial("tcp",beego.AppConfig.String("redis_host")+":"+beego.AppConfig.String("redis_port"))
	defer conn.Close()

	var user models.User
	user.Name=username
	o.Read(&user,"Name")

	var goodsList []map[string]interface{}
	totalGoodsNum:=0
	totalGoodsPrice:=0
	for _,k:=range goodsIds{
		_v:=make(map[string]interface{})

		goodsId,_:=strconv.Atoi(k)
		var goods models.GoodsSKU
		goods.Id=goodsId
		o.Read(&goods)
		_v["goods"]=goods

		v,_:=redis.Int(conn.Do("hget","cart_"+strconv.Itoa(user.Id),goodsId))
		_v["num"]=v
		_v["totalPrice"]=goods.Price*v

		totalGoodsNum+=v
		totalGoodsPrice+=goods.Price*v

		goodsList=append(goodsList,_v)
	}
	this.Data["goodsList"]=goodsList
	this.Data["totalGoodsNum"]=totalGoodsNum
	this.Data["totalGoodsPrice"]=totalGoodsPrice

	transitPrice:=10
	this.Data["transitPrice"]=transitPrice

	totalPrice:=totalGoodsPrice+transitPrice
	this.Data["totalPrice"]=totalPrice

	this.Data["goodsIds"]=strings.Join(goodsIds,",")

	this.ShowUserLayout("提交订单")
	this.TplName="user/place_order.html"
}

func (this *UserController) Order()  {
	username:=this.GetUser()
	o:=orm.NewOrm()
	s:=o.QueryTable("OrderInfo").RelatedSel("User").Filter("User__Name",username)

	totalRows,_:=s.Count()
	pageSize:=2
	pageCount:=int(math.Ceil(float64(totalRows)/float64(pageSize)))
	this.Data["pageCount"]=pageCount

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
	this.Data["page"]=page

	pages:=this.PageTool(page,pageCount)
	this.Data["pages"]=pages

	var _orderList []models.OrderInfo
	s.Limit(pageSize,pageSize*(page-1)).OrderBy("-Time").All(&_orderList)

	var orderList []map[string]interface{}
	for _,v:=range _orderList{
		_v:=make(map[string]interface{})
		_v["order"]=v

		var goodsList []models.OrderGoods
		o.QueryTable("OrderGoods").RelatedSel("GoodsSKU","OrderInfo").Filter("OrderInfo__Id",v.Id).All(&goodsList)
		_v["goodsList"]=goodsList

		orderList=append(orderList,_v)
	}
	this.Data["orderList"]=orderList

	this.ShowUserLayout("用户中心")
	this.TplName="user/order.html"
}