package server

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
)

func initData(app core.App) {

	if app.IsDev() { // skip when in dev mode
		return
	}

	app.OnServe().BindFunc(func(e *core.ServeEvent) error {

		objs := []map[string]any{}
		if err := json.Unmarshal([]byte(rawInitConfig), &objs); err != nil {
			panic(err)
		}
		col, err := e.App.FindCollectionByNameOrId("config")
		if err != nil {
			panic(err)
		}

		for _, obj := range objs {
			exist, err := e.App.FindFirstRecordByData(col, "name", obj["name"])
			if err == nil && exist != nil {
				e.App.Logger().Debug("skip init config", "name", obj["name"])
				continue
			}

			record := core.NewRecord(col)
			record.Load(obj)
			if err := e.App.Save(record); err != nil {
				panic(err)
			}
		}
		return e.Next()
	})

}

var rawInitConfig = `
[
  {
    "name": "dial",
    "value": {
      "caller": {
        "affinity": true
      }
    }
  },
  {
    "name": "privacy",
    "value": {
      "hideNumber": true
    }
  },
  {
    "name": "cloud",
    "value": {
      "addr": "",
      "appid": "",
      "secret": "",
      "lifecycle": {
        "precall": {
          "blacklist": false,
          "flashCard": false
        }
      }
    }
  },
  {
    "name": "ice_servers",
    "value": []
  }
]
`
