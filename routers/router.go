package routers

import (
	"181112/controllers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

func init() {
	beego.InsertFilter("/user/*",beego.BeforeExec, func(ctx *context.Context) {
		username:=ctx.Input.Session("username")
		if username==nil {
			ctx.Redirect(302,"/login")
		}
	})

	beego.Router("/",&controllers.GoodsController{},"get:Index")
	beego.Router("/goods/detail",&controllers.GoodsController{},"get:Detail")
	beego.Router("/goods/list",&controllers.GoodsController{},"get:List")
	beego.Router("/goods/search",&controllers.GoodsController{},"post:Search")

    beego.Router("/register", &controllers.UserController{},"get:ShowRegister;post:HandleRegister")
	beego.Router("/login", &controllers.UserController{},"get:ShowLogin;post:HandleLogin")
	beego.Router("/logout", &controllers.UserController{},"get:Logout")
	beego.Router("/active", &controllers.UserController{},"get:Active")

	beego.Router("/user/info",&controllers.UserController{},"get:Info")
	beego.Router("/user/address",&controllers.UserController{},"get:ShowAddress;post:HandleAddAddress")
	beego.Router("/user/cart",&controllers.UserController{},"get:Cart")
	beego.Router("/user/place_order",&controllers.UserController{},"post:PlaceOrder")
	beego.Router("/user/order",&controllers.UserController{},"get:Order")

	beego.Router("/cart/add",&controllers.CartController{},"post:AddCart")
	beego.Router("/cart/update",&controllers.CartController{},"post:UpdateCart")
	beego.Router("/cart/delete",&controllers.CartController{},"post:DeleteCart")

	beego.Router("/order/add",&controllers.OrderController{},"post:AddOrder")
}