package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ResponseFormat struct {
	Code int         `json:"code"` //0:成功 1：失败
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func FailResponse(context *gin.Context, msg string) {
	format := ResponseFormat{
		Code: 1,
		Msg:  msg,
	}

	context.JSON(http.StatusOK, format)
}

func SuccessResponse(context *gin.Context, data interface{}) {
	resp := ResponseFormat{
		Code: 0,
		Msg:  "ok",
		Data: data,
	}
	context.JSON(http.StatusOK, resp)
}
