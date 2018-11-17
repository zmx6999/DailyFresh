package otp

import (
	"hash"
	"crypto/hmac"
	"math"
	"strconv"
	"time"
		)

type OTPConfig struct {
	Key string
	HashFcn func() hash.Hash
}

func (o *OTPConfig) GenerateTOTP(interval int,digits int) string {
	x:=int(time.Now().Unix())/interval
	return o.GenerateOTP(x,digits)
}

func (o *OTPConfig) GenerateOTP(x int,digits int) string {
	y:=o.getHmac(x)
	offset:=y[len(y)-1]&0xf
	z:=(int(y[offset]&0x7f)<<24)|(int(y[offset+1]&0xff)<<16)|(int(y[offset+2]&0xff)<<8)|int(y[offset+3]&0xff)
	r:=z%int(math.Pow10(digits))
	rs:=strconv.Itoa(r)
	for len(rs)<digits {
		rs="0"+rs
	}
	return rs
}

func (o *OTPConfig) getHmac(x int) []byte {
	mhash:=hmac.New(o.HashFcn,[]byte(o.Key))
	mhash.Write([]byte{byte(x)})
	return mhash.Sum(nil)
}