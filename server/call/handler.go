package call

import (
	"encoding/json"
	"net/http"

	precall "lightcall/server/cloud/precall"
	"lightcall/server/config"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

func HandleCreateActivity(e *core.RequestEvent) error {
	taskID := e.Request.PathValue("id")
	user := e.Auth

	_, err := e.App.FindRecordById("task", taskID)
	if err != nil {
		return e.BadRequestError("invalid task", err)
	}

	result, err := makeCall(e.App, user, taskID)
	if err != nil {
		return e.InternalServerError("create activity fail", err)
	}

	// 创建activity记录
	c, err := e.App.FindCollectionByNameOrId("activity")
	if err != nil {
		return e.InternalServerError("create activity fail", err)
	}

	activity := core.NewRecord(c)
	rawlog := map[string]any{
		"call": result,
	}
	rawlogBytes, _ := json.Marshal(rawlog)
	activity.Load(map[string]any{
		"user":   user.Id,
		"isCall": true,
		"rawlog": string(rawlogBytes),
	})

	if err := e.App.Save(activity); err != nil {
		return e.InternalServerError("create activity fail", err)
	}

	return e.JSON(http.StatusOK, map[string]string{"id": activity.Id})
}

func HandlePreCall(conf config.Provider, handler *precall.Handler) func(*core.RequestEvent) error {

	return func(e *core.RequestEvent) error {
		activityID := e.Request.PathValue("activityId")
		activity, err := e.App.FindRecordById("activity", activityID)
		if err != nil {
			return e.NotFoundError("Activity not found", err)
		}

		// 解析rawlog.call
		var rawlog struct {
			Call struct{ Caller, Callee string }
		}
		if err := json.Unmarshal([]byte(activity.GetString("rawlog")), &rawlog); err != nil {
			return e.BadRequestError("Invalid rawlog format", err)
		}

		// 获取云配置
		cloudConf, err := conf.Cloud()
		if err != nil {
			return e.InternalServerError("Failed to get cloud config", err)
		}

		// 检查开关状态
		var shouldCall bool
		switch handler.Name {
		case "BlackList":
			shouldCall = cloudConf.Lifecycle.PreCall.Blacklist
		case "FlashCard":
			shouldCall = cloudConf.Lifecycle.PreCall.FlashCard
		}

		// 直接返回结果
		if !shouldCall {
			return e.JSON(http.StatusOK, precall.Result{Pass: true, Msg: "bypass"})
		}

		// 调用Hook
		cloudRespID, result, err := handler.Call(e.App, *cloudConf, &precall.Request{
			Caller: rawlog.Call.Caller,
			Callee: rawlog.Call.Callee,
		})

		if err != nil {
			return e.InternalServerError("Hook call failed", err)
		}

		// 关联cloudresp到activity
		activity.Set("hook+", cloudRespID)
		if err := e.App.Save(activity); err != nil {
			return e.InternalServerError("Failed to update activity", err)
		}

		return e.JSON(http.StatusOK, result)
	}
}

func HandleDirectCall(e *core.RequestEvent) error {
	// 解析请求体
	var req struct {
		Number string `json:"number"`
	}
	if err := json.NewDecoder(e.Request.Body).Decode(&req); err != nil {
		return e.BadRequestError("Invalid JSON", err)
	}

	// 验证号码参数
	if req.Number == "" {
		return e.BadRequestError("Number is required", nil)
	}

	// 获取当前用户
	user := e.Auth

	// 查找或创建临时目标
	tempObjective, err := e.App.FindFirstRecordByFilter("objective", "title='00000000#临时拨号记录'")
	if err != nil {
		// 不存在则创建
		objectiveCollection, err := e.App.FindCollectionByNameOrId("objective")
		if err != nil {
			return e.InternalServerError("Failed to find objective collection", err)
		}

		tempObjective = core.NewRecord(objectiveCollection)
		tempObjective.Load(map[string]any{
			"title": "00000000#临时拨号记录",
			"info": map[string]any{
				"company":    "临时拨号任务",
				"background": "用于直接拨号功能创建的临时任务",
			},
			"open": false,
		})

		if err := e.App.Save(tempObjective); err != nil {
			return e.InternalServerError("Failed to create temporary objective", err)
		}
	}

	// 检查是否已存在相同被叫号码且未关闭的临时任务
	existingTask, err := e.App.FindFirstRecordByFilter("task", "own={:userId} && callee={:callee} && open=true && contact='临时联系人'", dbx.Params{
		"userId": user.Id,
		"callee": req.Number,
	})
	if err == nil && existingTask != nil {
		return e.JSON(http.StatusOK, map[string]string{"taskId": existingTask.Id})
	}

	// 创建临时任务
	taskCollection, err := e.App.FindCollectionByNameOrId("task")
	if err != nil {
		return e.InternalServerError("Failed to find task collection", err)
	}

	task := core.NewRecord(taskCollection)
	task.Load(map[string]any{
		"own":                 user.Id,
		"contact":             "临时联系人",
		"callee":              req.Number,
		"desc":                "直接拨号任务",
		"open":                true,
		"objective_via_tasks": tempObjective.Id,
	})

	if err := e.App.Save(task); err != nil {
		return e.InternalServerError("Failed to create task", err)
	}

	// 将新任务ID添加到临时目标的tasks数组
	existingTasks := tempObjective.GetStringSlice("tasks")
	existingTasks = append(existingTasks, task.Id)
	tempObjective.Set("tasks", existingTasks)

	if err := e.App.Save(tempObjective); err != nil {
		return e.InternalServerError("Failed to update objective", err)
	}

	// 返回任务ID
	return e.JSON(http.StatusOK, map[string]string{"taskId": task.Id})
}
