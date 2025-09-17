package server

import (
	"lightcall/server/appender"
	"lightcall/server/appender/activity"
	"lightcall/server/appender/change"
	"lightcall/server/config"
	"lightcall/server/tail"
	"lightcall/server/tail/cdc"
	"lightcall/server/tail/fs"

	"github.com/pocketbase/pocketbase/core"
)

type FilePath struct {
	FsRecordDir    string
	TailFsCDR      string
	TailChange     string
	AppendActivity string
	AppendChange   string
}

func MustRegister(app core.App, path *FilePath) {

	configProvider := config.New(app)

	initData(app)
	initHook(app, configProvider)
	initRouter(app, configProvider)

	appender.MustRegister(app, &activity.Handler{LogFile: path.AppendActivity})
	appender.MustRegister(app, &change.Handler{LogFile: path.AppendChange})

	tail.MustRegister(app, &fs.Handler{MasterFile: path.TailFsCDR, RecordDir: path.FsRecordDir})
	tail.MustRegister(app, &cdc.Handler{CdcFile: path.TailChange})
}
