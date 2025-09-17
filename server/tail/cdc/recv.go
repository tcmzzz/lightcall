package cdc

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/pocketbase/pocketbase/core"
)

type RecvMsg struct {
	MsgOp   string          `json:"msg_op"`   // create/update/close
	MsgType string          `json:"msg_type"` // objective/task
	Event   json.RawMessage `json:"event"`
}

type RecvObjective struct {
	ExtID string `json:"ext_id"` // 三方系统对应标识
	Title string `json:"title"`  // 标题
	Info  struct {
		Company    string `json:"company"`    // 公司信息
		Background string `json:"background"` // 背景资料
	} `json:"info"`
}

type RecvTask struct {
	ObjectiveTitle string `json:"objective_title"` // 目标标题
	ExtID          string `json:"ext_id"`          // 三方系统对应标识
	Own            string `json:"own"`             // 所属人 传入Email, 处理时转换成内部用户ID
	Contact        string `json:"contact"`         // 联系人
	Callee         string `json:"callee"`          // 被叫人
	Desc           string `json:"desc"`            // 任务描述
}

func handleObjective(app core.App, msg *RecvMsg) error {
	var obj RecvObjective
	if err := json.Unmarshal(msg.Event, &obj); err != nil {
		return errors.Wrap(err, "解析目标数据失败")
	}

	collection, err := app.FindCollectionByNameOrId("objective")
	if err != nil {
		return errors.Wrap(err, "获取目标集合失败")
	}

	// 查找现有记录
	record, _ := app.FindFirstRecordByData(collection.Id, "ext_id", obj.ExtID)

	switch msg.MsgOp {
	case "create", "update":
		if record == nil {
			record = core.NewRecord(collection)
		}

		record.Set("ext_id", obj.ExtID)
		record.Set("title", obj.Title)
		record.Set("info", map[string]string{
			"company":    obj.Info.Company,
			"background": obj.Info.Background,
		})
		record.Set("open", true)

		if err := app.Save(record); err != nil {
			return errors.Wrap(err, "保存目标失败")
		}
		app.Logger().Info("目标已处理", "ext_id", obj.ExtID, "operation", msg.MsgOp)

	default:
		return errors.Errorf("未知的目标操作: %s", msg.MsgOp)
	}
	return nil
}

func handleTask(app core.App, msg *RecvMsg) error {
	var task RecvTask
	if err := json.Unmarshal(msg.Event, &task); err != nil {
		return errors.Wrap(err, "解析任务数据失败")
	}

	collection, err := app.FindCollectionByNameOrId("task")
	if err != nil {
		return errors.Wrap(err, "获取任务集合失败")
	}

	// 查找现有记录
	record, _ := app.FindFirstRecordByData(collection.Id, "ext_id", task.ExtID)

	switch msg.MsgOp {
	case "create":
		if record != nil {
			return errors.Errorf("任务已存在: %s", task.ExtID)
		}
		return createTask(app, collection, &task)
	default:
		return errors.Errorf("未知的任务操作: %s", msg.MsgOp)
	}

}

func createEmail(app core.App, email string) (*core.Record, error) {
	col, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		app.Logger().Info("找不到用户集合", "error", err)
		return nil, err
	}

	u, err := app.FindFirstRecordByData(col, "email", email)
	if err == nil {
		return u, nil
	}

	rec := core.NewRecord(col)
	rec.SetEmail(email)
	rec.SetEmailVisibility(true)
	rec.SetRandomPassword()
	rec.SetVerified(true)
	rec.Set("name", "New User")
	rec.Set("isAdmin", false)
	rec.Set("active", false)
	if err := app.Save(rec); err != nil {
		app.Logger().Info("创建用户失败", "email", email, "error", err)
		return nil, err
	}

	return app.FindFirstRecordByData(col, "email", email)
}
func createTask(app core.App, collection *core.Collection, task *RecvTask) error {

	user, err := createEmail(app, task.Own)
	if err != nil {
		return errors.Wrapf(err, "创建用户失败: %s", task.Own)
	}

	return app.RunInTransaction(func(txDao core.App) error {

		if user == nil {
			return errors.Errorf("用户不存在: %s", task.Own)
		}

		// 创建任务记录
		record := core.NewRecord(collection)
		record.Set("ext_id", task.ExtID)
		record.Set("own", user.Id)
		record.Set("contact", task.Contact)
		record.Set("callee", task.Callee)
		record.Set("desc", task.Desc)
		record.Set("open", true)

		// 保存任务
		if err := txDao.Save(record); err != nil {
			return errors.Wrap(err, "保存任务失败")
		}

		// 处理目标关联
		if task.ObjectiveTitle != "" { // nolint: nestif
			// 查找目标记录
			objCollection, err := txDao.FindCollectionByNameOrId("objective")
			if err != nil {
				return errors.Wrap(err, "查找目标集合失败")
			}

			objRecord, err := txDao.FindFirstRecordByData(objCollection.Id, "title", task.ObjectiveTitle)
			if err != nil {
				return errors.Wrapf(err, "查找目标失败: %s", task.ObjectiveTitle)
			}

			if objRecord != nil {
				// 更新目标的tasks字段
				tasks := objRecord.GetStringSlice("tasks")
				tasks = append(tasks, record.Id)
				objRecord.Set("tasks", tasks)

				// 保存目标更新
				if err := txDao.Save(objRecord); err != nil {
					return errors.Wrapf(err, "更新目标失败: %s", objRecord.Id)
				}
			}
		}

		app.Logger().Info("任务已创建", "ext_id", task.ExtID, "callee", task.Callee, "contact", task.Contact)
		return nil
	})
}
