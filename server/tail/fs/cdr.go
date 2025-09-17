package fs

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"github.com/pocketbase/pocketbase/tools/types"
)

type CdrLine struct {
	UUID                 string `json:"uuid"`
	Originator           string `json:"originator"`
	StartEpoch           int64  `json:"start_epoch"`
	AnswerEpoch          int64  `json:"answer_epoch"`
	ProgressMediaEpoch   int64  `json:"progress_media_epoch"`
	EndEpoch             int64  `json:"end_epoch"`
	Duration             int64  `json:"duration"`
	Billmsec             int64  `json:"billmsec"`
	HangupCause          string `json:"hangup_cause"`
	SipHangupDisposition string `json:"sip_hangup_disposition"`
	SipTermStatus        string `json:"sip_term_status"`
	UserID               string `json:"userId"`
	TaskID               string `json:"taskId"`
	ActivityID           string `json:"activityId"`
	OriCaller            string `json:"oriCaller"`
	OriCallee            string `json:"oriCallee"`
	RealCaller           string `json:"realCaller"`
	RealCallee           string `json:"realCallee"`
	Record               string `json:"record"`
}

type State struct {
	Caller      string `json:"caller"`
	Callee      string `json:"callee"`
	ProviderOK  bool   `json:"provider_ok"`
	ConnectOK   bool   `json:"connect_ok"`
	StartEpoch  int64  `json:"start_epoch"`
	Billmsec    int64  `json:"billmsec"`
	Duration    int64  `json:"duration"`
	ALegCause   string `json:"a_leg_cause"`
	ALegSipTerm string `json:"a_leg_sip_term"`
	BLegCause   string `json:"b_leg_cause"`
	BLegSipTerm string `json:"b_leg_sip_term"`
}

func (s *State) Comment() string { // 电话开始于 2024-09-13 13:22, 总用时 5 分钟

	start := time.Unix(s.StartEpoch, 0).Format("2006-01-02 15:04")
	sb := strings.Builder{}

	sb.WriteString(fmt.Sprintf("电话开始于 %s", start))
	if s.ConnectOK {
		billsec := s.Billmsec / 1000
		minutes := billsec / 60
		seconds := billsec % 60
		sb.WriteString(fmt.Sprintf(", 通话时长 %d 分钟 %d 秒", minutes, seconds))
		return sb.String()
	}

	// if not connected, show the cause
	sb.WriteString(", 未接通(")

	if s.BLegSipTerm != "" && s.BLegSipTerm != "200" {
		sb.WriteString(fmt.Sprintf("线路商响应错误码 %s ", s.BLegSipTerm))
	} else if s.ProviderOK {
		sb.WriteString("客户未接听")
	} else {
		sb.WriteString(fmt.Sprintf("用时%d秒线路未接通", s.Duration))
	}

	sb.WriteString(")")

	return sb.String()
}

func GenerateCallState(aleg, bleg *CdrLine) (*State, error) {
	if aleg == nil || bleg == nil {
		return nil, errors.New("无效的通话记录")
	}

	return &State{
		Caller:      aleg.OriCaller,
		Callee:      aleg.OriCallee,
		ProviderOK:  aleg.ProgressMediaEpoch != 0,
		ConnectOK:   aleg.Billmsec != 0,
		StartEpoch:  aleg.StartEpoch,
		Billmsec:    aleg.Billmsec,
		Duration:    aleg.Duration,
		ALegCause:   aleg.HangupCause,
		ALegSipTerm: aleg.SipTermStatus,
		BLegCause:   bleg.HangupCause,
		BLegSipTerm: bleg.SipTermStatus,
	}, nil
}

func (l *CdrLine) LoadWithBleg(app core.App, recordDir string, bleg *CdrLine) error {

	record, err := app.FindRecordById("activity", l.ActivityID)
	if err != nil {
		return errors.Wrapf(err, "find activity fail(activity_id: %s).", l.ActivityID)
	}

	task, err := app.FindRecordById("task", l.TaskID)
	if err != nil {
		return errors.Wrapf(err, "find task fail(task_id: %s).", l.TaskID)
	}

	task.Set("activity+", record.Id)

	state, err := GenerateCallState(l, bleg)
	if err != nil {
		return errors.Wrapf(err, "generate call state fail")
	}

	str := record.GetString("rawlog")
	if str == "" {
		str = "{}"
	}

	rawlog := map[string]interface{}{}
	if err := json.Unmarshal([]byte(str), &rawlog); err != nil {
		return errors.Wrapf(err, "activity rawlog invalid(activity_id: %s).", l.ActivityID)
	}
	rawlog["fslega"] = l
	rawlog["fslegb"] = bleg
	rawlog["state"] = state
	rl, err := json.Marshal(rawlog)
	if err != nil {
		return errors.Wrapf(err, "activity rawlog invalid(activity_id: %s).", l.ActivityID)
	}
	record.Set("rawlog", string(rl))

	created, _ := types.ParseDateTime(l.AnswerEpoch)

	record.Load(map[string]any{
		"user":    l.UserID,
		"comment": state.Comment(),
		"created": created,
		"updated": created,
		"isCall":  true,
	})

	if state.ConnectOK {

		rFile, err := filesystem.NewFileFromPath(path.Join(recordDir, l.Record))
		if err != nil {
			app.Logger().Warn("record file fail", "error", err)
		} else {
			record.Set("record", rFile)
		}
	}

	return app.RunInTransaction(func(txApp core.App) error {

		if err := txApp.Save(record); err != nil {
			return errors.Wrap(err, "save record fail")
		}

		if err := txApp.Save(task); err != nil {
			return errors.Wrap(err, "save task fail")
		}
		return nil
	})
}
