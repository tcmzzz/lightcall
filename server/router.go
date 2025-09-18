package server

import (
	"os"

	"github.com/tcmzzz/lightcall/server/call"
	"github.com/tcmzzz/lightcall/server/cloud/mock"
	"github.com/tcmzzz/lightcall/server/cloud/precall"
	"github.com/tcmzzz/lightcall/server/config"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func initRouter(app core.App, config config.Provider) {

	if app.IsDev() {

		app.OnServe().BindFunc(func(se *core.ServeEvent) error {
			gPrecall := se.Router.Group("/api/mockcloud").Group("/precall")
			gPrecall.POST(precall.BlackList.Path, mock.HandleMockBlacklist)
			gPrecall.POST(precall.FlashCard.Path, mock.HandleMockFlashcard)
			return se.Next()
		})

	}

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {

		if _, err := os.Stat("./public"); err == nil {
			se.Router.GET("/{path...}", apis.Static(os.DirFS("./public"), false))
		}

		g := se.Router.Group("/api/custom/call")

		g.GET("/new/{id}", call.HandleCreateActivity).Bind(apis.RequireAuth())
		g.GET("/precall/blacklist/{activityId}", call.HandlePreCall(config, precall.BlackList))
		g.GET("/precall/flashcard/{activityId}", call.HandlePreCall(config, precall.FlashCard))
		g.POST("/direct", call.HandleDirectCall).Bind(apis.RequireAuth())
		g.POST("/sip/fs", call.HandleFsCall) // TODO: check fs ip

		return se.Next()
	})

}
