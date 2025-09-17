package precall

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/tcmzzz/lightcall/server/config"

	"github.com/pkg/errors"
	"github.com/pocketbase/pocketbase/core"
)

var BlackList = &Handler{Name: "BlackList", Path: "/blacklist", ParseFunc: commonParse}
var FlashCard = &Handler{Name: "FlashCard", Path: "/flashcard", ParseFunc: commonParse}

func commonParse(bts []byte) (*Result, error) {
	obj := &Response{}
	if err := json.Unmarshal(bts, obj); err != nil {
		return nil, errors.Wrap(err, "json Unmarshal fail")
	}

	if obj.Data == nil {
		return nil, errors.New("empty data")
	}
	return obj.Data, nil
}

type Request struct {
	Caller string `json:"caller"`
	Callee string `json:"callee"`
}

type Response struct {
	Code int     `json:"code"`
	Msg  string  `json:"msg"`
	Data *Result `json:"data"`
}

type Result struct {
	Pass bool   `json:"pass"`
	Msg  string `json:"msg"`
}

type parseFunc func([]byte) (*Result, error)

type Handler struct {
	Name      string
	Path      string
	ParseFunc parseFunc
}

func (h *Handler) Call(app core.App, conf config.Cloud, req *Request) (string, *Result, error) {
	url := fmt.Sprintf("%s/precall%s", conf.Addr, h.Path)

	// 序列化请求
	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", nil, errors.Wrap(err, "序列化请求失败")
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return "", nil, errors.Wrap(err, "创建HTTP请求失败")
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("AppID", conf.AppID)
	httpReq.Header.Set("Secret", conf.Secret)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", nil, errors.Wrap(err, "发送请求失败")
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, errors.Wrap(err, "读取响应失败")
	}

	if resp.StatusCode != 200 {
		return "", nil, errors.Errorf("请求失败，状态码：%d，响应内容：%s", resp.StatusCode, respBody)
	}

	// 解析响应
	result, err := h.ParseFunc(respBody)
	if err != nil {
		return "", nil, errors.Wrap(err, "解析响应失败")
	}

	// 创建cloudresp记录
	collection, err := app.FindCollectionByNameOrId("cloudresp")
	if err != nil {
		return "", nil, errors.Wrap(err, "获取cloudresp集合失败")
	}

	record := core.NewRecord(collection)
	record.Set("type", "pre-call")
	record.Set("name", h.Name)
	record.Set("result", result)
	record.Set("rawresp", string(respBody))

	if err := app.Save(record); err != nil {
		return "", nil, errors.Wrap(err, "保存cloudresp记录失败")
	}

	return record.Id, result, nil
}
