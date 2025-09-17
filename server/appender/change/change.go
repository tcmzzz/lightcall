package change

import (
	"encoding/json"
	"time"

	"lightcall/server/appender"

	"github.com/pkg/errors"
	"github.com/pocketbase/pocketbase/core"
)

type RequestEvent struct {
	EventType   string    `json:"eventType"`   // "close_request", "open_request"
	ID          string    `json:"id"`          // 内部ID
	ExtID       string    `json:"ext_id"`      // 外部系统ID
	Timestamp   time.Time `json:"timestamp"`   // 请求时间
	RequestedBy string    `json:"requestedBy"` // 请求用户ID
	Reason      string    `json:"reason"`      // 关闭原因
	EntityType  string    `json:"entityType"`  // "task" 或 "objective"
}

type Handler struct {
	LogFile string
}

func (h *Handler) File() string {
	return h.LogFile
}

func (h *Handler) MustRegister(app core.App, appender *appender.Appender) error {

	app.OnRecordUpdateRequest("task", "objective").BindFunc(func(e *core.RecordRequestEvent) error {

		record := e.Record

		extID := record.GetString("ext_id")
		if extID == "" || extID == record.Id {
			return e.Next()
		}

		originalOpen := record.Original().GetBool("open")
		currentOpen := record.GetBool("open")

		if originalOpen == currentOpen { // skip if open not changed
			return e.Next()
		}

		col := e.Collection.Name
		user := e.Auth.GetString("email")

		event := RequestEvent{
			ID:          record.Id,
			ExtID:       extID,
			Timestamp:   time.Now(),
			RequestedBy: user,
			Reason:      "用户操作",
			EntityType:  col,
		}

		if currentOpen {
			event.EventType = "open_request"
		} else {
			event.EventType = "close_request"
		}
		// 写入文件
		data, err := json.Marshal(event)
		if err != nil {
			return errors.Wrap(err, "failed to marshal close request event")
		}

		// 写入文件
		appender.Append(string(data))

		return e.JSON(200, map[string]any{
			"status":  200,
			"message": "操作进入队列等待处理(来自三方系统)",
		})
		// return e.String(401, "send to request queue")
	})

	return nil
}
