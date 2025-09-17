package mock

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/tcmzzz/lightcall/server/cloud/precall"

	"github.com/pocketbase/pocketbase/core"
)

// 黑名单模拟处理函数
func HandleMockBlacklist(e *core.RequestEvent) error {
	// 解析请求体到precall.Request结构
	var req precall.Request
	if err := e.BindBody(&req); err != nil {
		return e.BadRequestError("无效的请求格式", err)
	}

	rd := rand.New(rand.NewSource(time.Now().UnixNano()))

	result := &precall.Result{
		Pass: true,
		Msg:  "ok",
	}
	// 模拟黑名单判断
	if rd.Intn(100) < 50 { // 10%的概率命中黑名单
		result.Pass = false
		result.Msg = "命中黑名单"
	}

	return e.JSON(http.StatusOK, precall.Response{Code: 0, Msg: "success", Data: result})
}

// 闪卡模拟处理函数
func HandleMockFlashcard(e *core.RequestEvent) error {
	// 解析请求体到precall.Request结构
	var req precall.Request
	if err := e.BindBody(&req); err != nil {
		return e.BadRequestError("无效的请求格式", err)
	}

	rd := rand.New(rand.NewSource(time.Now().UnixNano()))

	result := &precall.Result{
		Pass: true,
		Msg:  "发送成功",
	}

	if rd.Intn(100) < 10 {
		result.Pass = false
		result.Msg = "发送失败"
	}

	return e.JSON(http.StatusOK, precall.Response{Code: 0, Msg: "success", Data: result})
}
