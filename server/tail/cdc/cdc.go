package cdc

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cast"
)

type Handler struct {
	CdcFile string
}

func (c *Handler) File() string { return c.CdcFile }
func (c *Handler) Deal(app core.App, line string) error {
	msg := &RecvMsg{}
	if err := json.Unmarshal([]byte(line), msg); err != nil {
		return err
	}

	if msg.MsgOp == "open" {
		return handleRecordStatus(app, msg, true)
	}
	if msg.MsgOp == "close" {
		return handleRecordStatus(app, msg, false)
	}

	if msg.MsgType == "objective" {
		return handleObjective(app, msg)
	}
	if msg.MsgType == "task" {
		return handleTask(app, msg)
	}

	return errors.Errorf("未知的消息类型: %s", msg.MsgType)
}

func handleRecordStatus(app core.App, msg *RecvMsg, open bool) error {

	m := map[string]string{}
	if err := json.Unmarshal(msg.Event, &m); err != nil {
		return errors.Wrap(err, "unmashal fail")
	}

	collectionName := msg.MsgType
	collection, err := app.FindCollectionByNameOrId(collectionName)
	if err != nil {
		return errors.Wrap(err, "获取目标集合失败")
	}

	// 查找现有记录
	extID := m["ext_id"]
	record, err := app.FindFirstRecordByData(collection.Id, "ext_id", extID)
	if err != nil {
		return errors.Wrapf(err, "%s不存在: %s", collectionName, extID)
	}

	if record == nil {
		return errors.Errorf("%s不存在: %s", collectionName, extID)
	}

	record.Set("open", open)
	if err := app.Save(record); err != nil {
		return errors.Wrapf(err, "设置 %s(%s) open %s 失 败", collectionName, extID, cast.ToString(open))
	}

	app.Logger().Info("设置 open 状态成功", "collection", collectionName, "extId", extID, "open", cast.ToString(open))
	return nil
}
