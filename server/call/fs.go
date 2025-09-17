package call

import (
	"bytes"
	"fmt"
	"log/slog"
	"text/template"

	"github.com/pocketbase/pocketbase/core"
)

type fsCallForm struct {
	TaskID     string `json:"variable_sip_i_ring_taskid"`
	ActivityID string `json:"variable_sip_i_ring_activityid"`
	UserID     string `json:"variable_sip_i_ring_userid"`
	AuthToken  string `json:"variable_sip_i_ring_auth"`
}

func HandleFsCall(se *core.RequestEvent) error {

	app, form := se.App, &fsCallForm{}

	if err := se.BindBody(form); err != nil {
		app.Logger().Error("bind req fail", "err", err)
		return se.String(200, fsFmtFailTpl(400, "Invalid Request", app.Logger()))
	}

	// verify auth and number
	user, err := se.App.FindAuthRecordByToken(form.AuthToken, core.TokenTypeAuth)

	if err != nil {
		app.Logger().Error("find auth fail", "err", err)
		return se.String(200, fsFmtFailTpl(401, "Unauthorized", app.Logger()))
	}

	if user.Id != form.UserID {
		app.Logger().Error("user id not same", "form", form.UserID, "record", user.Id)
		return se.String(200, fsFmtFailTpl(401, "Unauthorized", app.Logger()))
	}

	if _, err := app.FindRecordById("activity", form.ActivityID); err != nil {
		app.Logger().Error("can not find activity", "form", form.ActivityID)
		return se.String(200, fsFmtFailTpl(400, "Invalid Request", app.Logger()))
	}

	result, err := makeCall(app, user, form.TaskID)
	if err != nil {
		app.Logger().Error("make call fail", "err", err)
		return se.String(200, fsFmtFailTpl(400, "Invalid Request", app.Logger()))
	}

	// Format template
	param := fsTplBridgeParam{
		UserID:     form.UserID,
		TaskID:     form.TaskID,
		ActivityID: form.ActivityID,
		OriCaller:  result.OriCaller,
		OriCallee:  result.OriCallee,
		Caller:     result.Caller,
		Callee:     result.Callee,
		DialStr:    fmt.Sprintf("%s@%s", result.Callee, result.Addr),
	}

	return se.String(200, param.Fmt(app.Logger()))
}

type fsTplBridgeParam struct {
	UserID     string
	TaskID     string
	ActivityID string
	OriCaller  string
	OriCallee  string
	Caller     string
	Callee     string
	DialStr    string
}

func (p fsTplBridgeParam) Fmt(logger *slog.Logger) string {

	buf := bytes.NewBufferString("")
	t := template.Must(template.New("bridge").Parse(fsTplBridge))
	if err := t.Execute(buf, p); err != nil {
		logger.Warn("format tpl fail", "err", err)
	}
	return buf.String()
}

// <action application="answer"/>
const fsTplBridge = `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
 <document type="freeswitch/xml">
  <section name="dialplan" description=""">
   <context name="public">
    <extension name="hold_music" continue="true">
      <condition>
        <action application="set" data="userId={{.UserID}}" />
        <action application="set" data="taskId={{.TaskID}}"/>
        <action application="set" data="activityId={{.ActivityID}}"/>
        <action application="set" data="oriCaller={{.OriCaller}}"/>
        <action application="set" data="oriCallee={{.OriCallee}}"/>
        <action application="set" data="realCaller={{.Caller}}"/>
        <action application="set" data="realCallee={{.Callee}}"/>
        <action application="set" data="effective_caller_id_name={{.Caller}}"/>
        <action application="set" data="effective_caller_id_number={{.Caller}}"/>
        <action application="set" data="RECORD_STEREO=true"/>
        <action application="set" data="RECORD_DATE=${strftime(%Y-%m-%d %H:%M)}"/>
        <action application="set" data="record_file=${strftime(%Y-%m-%d-%H-%M-%S)}_${destination_number}_${effective_caller_id_number}.mp3"/>
        <action application="record_session" data="$${recordings_dir}/${record_file}"/>
        <action application="set" data="hangup_after_bridge=true"/>
        <action application="bridge" data="sofia/internal/{{.DialStr}}"/>
      </condition>
    </extension>
   </context>
  </section>
</document>`

func fsFmtFailTpl(code int, msg string, logger *slog.Logger) string {

	buf := bytes.NewBufferString("")
	t := template.Must(template.New("fail").Parse(fsTplFail))
	err := t.Execute(buf, map[string]interface{}{
		"Code": code,
		"Msg":  msg,
	})

	if err != nil {
		logger.Warn("format tpl fail", "error", err)
	}
	return buf.String()

}

const fsTplFail = `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
 <document type="freeswitch/xml">
  <section name="dialplan" description=""">
   <context name="public">
    <extension name="dialFail">
      <condition>
        <action application="respond" data="{{.Code}} {{.Msg}}"/>
      </condition>
    </extension>
   </context>
  </section>
</document>`
