package server

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/security"
)

func loadDefaultAdmin(app core.App) string {
	adminEmail := os.Getenv("INIT_ADMIN_EMAIL")

	if adminEmail == "" {
		app.Logger().Info("INIT_ADMIN_EMAIL not set, skipping admin creation")
		panic("INIT_ADMIN_EMAIL not set, skipping admin creation")
	}

	adminPassword := security.RandomString(12)

	// Create PocketBase superuser
	superusers, err := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	if err != nil {
		app.Logger().Error("Failed to find superusers collection", "error", err)
		panic("Failed to find superusers collection: " + err.Error())
	}

	existingSuperuser, err := app.FindFirstRecordByData(core.CollectionNameSuperusers, "email", adminEmail)
	if err != nil || existingSuperuser == nil {
		superuser := core.NewRecord(superusers)
		superuser.Set("email", adminEmail)
		superuser.Set("password", adminPassword)
		superuser.Set("passwordConfirm", adminPassword)

		if err := app.Save(superuser); err != nil {
			app.Logger().Error("Failed to create superuser", "error", err)
			panic("Failed to create superuser: " + err.Error())
		}

		app.Logger().Info("Superuser created successfully", "email", adminEmail)
		fmt.Printf("Superuser created successfully, username: %s, password: %s\n", adminEmail, adminPassword)

	} else {
		app.Logger().Info("Superuser already exists", "email", adminEmail)
	}

	// Create regular user with random password
	userEmail := adminEmail
	user, err := app.FindFirstRecordByData("users", "email", userEmail)
	if err != nil || user == nil {
		password := security.RandomString(10)

		userCollection, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			app.Logger().Error("Failed to find users collection", "error", err)
			panic("Failed to find users collection: " + err.Error())
		}

		user = core.NewRecord(userCollection)
		user.Set("email", userEmail)
		user.Set("password", password)
		user.Set("passwordConfirm", password)

		// Extract name from email (part before @)
		name := userEmail
		if idx := strings.Index(userEmail, "@"); idx > 0 {
			name = userEmail[:idx]
		}

		user.Set("name", name)
		user.Set("isAdmin", true)
		user.Set("active", true)
		user.Set("verified", true)

		if err := app.Save(user); err != nil {
			app.Logger().Error("Failed to create regular user", "error", err)
			panic("Failed to create regular user: " + err.Error())
		}
		app.Logger().Info("Regular user created successfully", "email", userEmail)
		fmt.Printf("Regular user created successfully: %s (password: %s)\n", userEmail, password)
	} else {
		app.Logger().Info("Regular user already exists", "email", userEmail)
	}

	return user.Id
}

func loadDefaultData(app core.App, collectionName string, rawData string, adminID string) {
	// Replace ADMIN_ID placeholder with actual admin ID
	processedRawData := strings.ReplaceAll(rawData, "ADMIN_ID", adminID)

	objs := []map[string]any{}
	if err := json.Unmarshal([]byte(processedRawData), &objs); err != nil {
		panic("Failed to unmarshal raw data for collection " + collectionName + ": " + err.Error())
	}

	col, err := app.FindCollectionByNameOrId(collectionName)
	if err != nil {
		panic("Failed to find collection " + collectionName + ": " + err.Error())
	}

	for _, obj := range objs {
		// Directly create new record without checking if it exists
		record := core.NewRecord(col)
		record.Load(obj)
		if err := app.Save(record); err != nil {
			panic("Failed to save record in collection " + collectionName + ": " + err.Error())
		}
		app.Logger().Info("Record created successfully", "collection", collectionName, "data", obj)
	}
}

func initData(app core.App) {

	if app.IsDev() {
		return
	}

	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		record, err := e.App.FindFirstRecordByData("config", "name", "initial")
		if err == nil && record.GetString("value") == "true" {
			e.App.Logger().Info("skip initial data, already loaded")
			return e.Next()
		}

		adminID := loadDefaultAdmin(e.App)
		loadDefaultData(app, "config", rawInitConfig, adminID)
		loadDefaultData(e.App, "outgw", rawOutgwData, adminID)
		loadDefaultData(e.App, "number", rawNumberData, adminID)
		loadDefaultData(e.App, "task", rawTaskData, adminID)
		loadDefaultData(e.App, "objective", rawObjectiveData, adminID)

		col, err := e.App.FindCollectionByNameOrId("config")
		if err != nil {
			panic("Failed to find config collection: " + err.Error())
		}

		initialRecord, err := e.App.FindFirstRecordByData("config", "name", "initial")
		if err != nil || initialRecord == nil {
			initialRecord = core.NewRecord(col)
			initialRecord.Set("name", "initial")
		}
		initialRecord.Set("value", "true")
		if err := e.App.Save(initialRecord); err != nil {
			panic("Failed to mark initialization as complete: " + err.Error())
		} else {
			e.App.Logger().Info("Initialization completed and marked as complete")
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

var rawOutgwData = `
[
  {
    "id": "t88gsc1c7123c92",
    "name": "演示SIP网关",
    "protocol": "SIP",
    "addr": "192.168.66.30:5080",
    "enable": true
  }
]
`

var rawNumberData = `
[
  {
    "outgw": "t88gsc1c7123c92",
    "number": "1234567",
    "enable": true,
    "tag": {
      "city": "北京市",
      "province": "北京市"
    }
  }
]
`

var rawObjectiveData = `
[
  {
    "id": "exaobjective001",
    "ext_id": "exaobjective001",
    "title": "00000001#示例工作目标",
    "info": {
      "company": "示例公司",
      "background": "这是一个示例目标，用于演示系统功能"
    },
    "tasks": [
      "exampletask0001"
    ],
    "open": true
  }
]
`

var rawTaskData = `
[
  {
    "contact": "示例联系人",
    "callee": "1234567",
    "desc": "这是一个示例任务，用于演示系统功能",
    "open": true,
	  "own": "ADMIN_ID",
    "id": "exampletask0001",
    "ext_id": "exampletask0001"
  }
]
`
