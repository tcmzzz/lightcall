package server

import (
	"github.com/tcmzzz/lightcall/server/appender"
	"github.com/tcmzzz/lightcall/server/appender/activity"
	"github.com/tcmzzz/lightcall/server/appender/change"
	"github.com/tcmzzz/lightcall/server/config"
	"github.com/tcmzzz/lightcall/server/tail"
	"github.com/tcmzzz/lightcall/server/tail/cdc"
	"github.com/tcmzzz/lightcall/server/tail/fs"

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
