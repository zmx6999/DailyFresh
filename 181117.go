package main

import (
	"encoding/base64"
	"fmt"
	"strconv"
)

func main()  {
	x:=1
	y:=base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(x)))
	fmt.Print(y)
	t,_:=base64.StdEncoding.DecodeString(y)
	fmt.Print(string(t))
}