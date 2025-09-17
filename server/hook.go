package server

import (
	"strings"

	"github.com/tcmzzz/lightcall/server/config"

	"github.com/pocketbase/pocketbase/core"
)

func initHook(app core.App, config config.Provider) {

	// block user not active
	app.OnRecordAuthRequest("users").BindFunc(func(e *core.RecordAuthRequestEvent) error {
		active := e.Record.GetBool("active")
		if !active {
			return e.BadRequestError("user is not active", "")
		}
		return e.Next()
	})

	// id != ext_id 视为来自外部系统的同步, 系统维护的添加时自动设置 ext_id = id
	app.OnRecordCreate("task", "objective").BindFunc(func(e *core.RecordEvent) error {
		id := e.Record.GetString("id")
		extID := e.Record.GetString("ext_id")
		if extID == "" {
			e.Record.Set("ext_id", id)
		}
		return e.Next()
	})

	// clear config cache when config changed
	app.OnRecordAfterUpdateSuccess("config").BindFunc(func(e *core.RecordEvent) error {
		config.ClearCache()
		return e.Next()
	})

	// hide callee number when privacy.HideNumber is open
	app.OnRecordEnrich("task").BindFunc(func(e *core.RecordEnrichEvent) error {

		privacy, err := config.Privacy()
		if err != nil {
			return err
		}

		if privacy.HideNumber && !e.RequestInfo.Auth.GetBool("isAdmin") {

			callee := e.Record.GetString("callee")
			masked := ""
			if len(callee) > 6 { // 保留前3位和后3位, 中间部分用*填充
				prefix := callee[:3]
				suffix := callee[len(callee)-3:]
				masked = prefix + strings.Repeat("*", len(callee)-6) + suffix
			} else { // 如果号码长度不足6位，全部显示为*
				masked = strings.Repeat("*", len(callee))
			}
			e.Record.Set("callee", masked)
		}

		return e.Next()
	})
}
