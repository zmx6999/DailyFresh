package main

import (
	_ "181112/routers"
	"github.com/astaxie/beego"
	"bytes"
	"time"
)

func main() {
	beego.AddFuncMap("getconfig",GetConfig)
	beego.AddFuncMap("prepage",ShowPrePage)
	beego.AddFuncMap("nextpage",ShowNextPage)
	beego.AddFuncMap("phone",ShowPhone)
	beego.AddFuncMap("showtime",ShowTime)
	beego.Run()
}

func GetConfig(key string) string {
	return beego.AppConfig.String(key)
}

func ShowPrePage(x int) int {
	if x==1 {
		return x
	} else {
		return x-1
	}
}

func ShowNextPage(x int,pageCount int) int {
	if x==pageCount {
		return x
	} else {
		return x+1
	}
}

func ShowPhone(phone string) string {
	y:=bytes.Runes([]byte(phone))
	return string(y[:3])+"****"+string(y[7:])
}

func ShowTime(x time.Time,timezone string) string {
	loc,_:=time.LoadLocation(timezone)
	return x.In(loc).Format("2006-01-02 15:04:05")
}