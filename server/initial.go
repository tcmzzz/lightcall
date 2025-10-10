package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/k0kubun/pp/v3"
	"github.com/pkg/errors"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/security"
)

// InitVars contains variables for template substitution during initialization
type InitVars struct {
	AdminEmail    string
	AdminID       string
	HasIceServers bool
	IceURLs       string
	IceUsername   string
	IceCredential string
}

func makeInitVars() InitVars {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		panic("INIT_ADMIN_EMAIL not set")
	}

	iceURLs := os.Getenv("ICE_URLS")
	hasIceServers := iceURLs != ""

	return InitVars{
		AdminEmail:    adminEmail,
		HasIceServers: hasIceServers,
		IceURLs:       iceURLs,
		IceUsername:   os.Getenv("ICE_USERNAME"),
		IceCredential: os.Getenv("ICE_CREDENTIAL"),
	}
}

func loadDefaultAdmin(app core.App, vars InitVars) string {
	adminEmail := vars.AdminEmail

	if adminEmail == "" {
		app.Logger().Info("AdminEmail not set, skipping admin creation")
		panic("AdminEmail not set, skipping admin creation")
	}

	adminPassword := security.RandomString(12)

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

func loadDefaultData(app core.App, collectionName string, rawData string, vars InitVars) {
	tmpl, err := template.New(collectionName).Parse(rawData)
	if err != nil {
		panic("Failed to parse template for collection " + collectionName + ": " + err.Error())
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		panic("Failed to execute template for collection " + collectionName + ": " + err.Error())
	}

	processedRawData := buf.String()

	objs := []map[string]any{}
	if err := json.Unmarshal([]byte(processedRawData), &objs); err != nil {
		panic("Failed to unmarshal raw data for collection " + collectionName + ": " + err.Error())
	}

	col, err := app.FindCollectionByNameOrId(collectionName)
	if err != nil {
		panic("Failed to find collection " + collectionName + ": " + err.Error())
	}

	for _, obj := range objs {
		record := core.NewRecord(col)
		record.Load(obj)
		if err := app.Save(record); err != nil {
			panic("Failed to save record in collection " + collectionName + ": " + err.Error())
		}
		app.Logger().Info("Record created successfully", "collection", collectionName, "data", obj)
	}
}

func checkInitFlag(app core.App) (bool, error) {
	record, err := app.FindFirstRecordByData("config", "name", "initial")
	if err != nil {
		return false, err
	}

	return record.GetString("value") == "true", nil
}

func saveInitFlag(app core.App) error {
	col, err := app.FindCollectionByNameOrId("config")
	if err != nil {
		return errors.Wrap(err, "failed to find config collection")
	}

	initialRecord, err := app.FindFirstRecordByData("config", "name", "initial")
	if err != nil || initialRecord == nil {
		initialRecord = core.NewRecord(col)
		initialRecord.Set("name", "initial")
	}

	initialRecord.Set("value", "true")
	if err := app.Save(initialRecord); err != nil {
		return errors.Wrap(err, "failed to mark initialization as complete")
	}

	app.Logger().Info("Initialization completed and marked as complete")
	return nil
}

func initData(app core.App) {

	if app.IsDev() {
		return
	}

	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		initialized, err := checkInitFlag(e.App)
		if err != nil {
			e.App.Logger().Debug("Initialization flag not found, proceeding with initialization")
		} else if initialized {
			e.App.Logger().Info("skip initial data, already loaded")
			return e.Next()
		}

		vars := makeInitVars()
		adminID := loadDefaultAdmin(e.App, vars)

		vars.AdminID = adminID
		pp.Println(vars)

		loadDefaultData(e.App, "config", rawInitConfig, vars)
		loadDefaultData(e.App, "outgw", rawOutgwData, vars)
		loadDefaultData(e.App, "number", rawNumberData, vars)
		loadDefaultData(e.App, "task", rawTaskData, vars)
		loadDefaultData(e.App, "objective", rawObjectiveData, vars)

		if vars.HasIceServers {
			loadDefaultData(e.App, "config", rawIceEnvConfig, vars)
		} else {
			loadDefaultData(e.App, "config", rawIceEmptyConfig, vars)
		}

		if err := saveInitFlag(e.App); err != nil {
			panic("Failed to save initialization flag: " + err.Error())
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
	  "own": "{{.AdminID}}",
    "id": "exampletask0001",
    "ext_id": "exampletask0001"
  }
]
`

var rawIceEmptyConfig = `
[
  {
    "name": "ice_servers",
    "value": []
  }
]
`

var rawIceEnvConfig = `
[
  {
    "name": "ice_servers",
    "value": [
      {
        "urls": "{{.IceURLs}}",
        "username": "{{.IceUsername}}",
        "credential": "{{.IceCredential}}"
      }
    ]
  }
]
`
