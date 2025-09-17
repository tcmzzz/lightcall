package dev

import (
	"embed"
	"encoding/json"
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"github.com/spf13/cast"
)

//go:embed *
var fs embed.FS

func init() {
	m.Register(func(app core.App) error {

		collections := []string{"cloudresp", "config", "outgw", "number", "users", "activity", "task", "objective"}

		for _, name := range collections {
			if err := loadRecord(app, name); err != nil {
				app.Logger().Debug("[load dev data] fail", "name", name, "err", err)
			} else {
				app.Logger().Debug("[load dev data] success", "name", name)
			}
		}

		return loadDevAdmin(app)

	}, func(app core.App) error {
		return nil
	})
}

func loadDevAdmin(app core.App) error {
	col, err := app.FindCollectionByNameOrId("_superusers")
	if err != nil {
		panic(err)
	}
	col.AuthToken.Secret = "14QmX12jPKmlgPdqKrY6jAWNaNXUJa0o1X9ZuobFGmHBwnm8w6"
	if err := app.Save(col); err != nil {
		panic(err)
	}

	superusers, err := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	if err != nil {
		return err
	}

	record := core.NewRecord(superusers)

	record.Set(core.FieldNameId, "rw775b2siw236d7")
	record.Set("email", "test@test.com")
	record.Set("password", "123123123")
	record.SetTokenKey("U8OD8W3EvS20t025czUglw7MOq06TbXm6nrj6B4KZ2mPcFa389")

	return app.Save(record)

}

func loadRecord(app core.App, name string) error {

	bts, err := fs.ReadFile(fmt.Sprintf("%s.%s", name, "json"))
	if err != nil {
		return err
	}

	col, err := app.FindCollectionByNameOrId(name)
	if err != nil {
		app.Logger().Error("[load dev data] failed to find collection", "name", name, "error", err)
		return err
	}

	objs := []map[string]any{}
	if err := json.Unmarshal(bts, &objs); err != nil {
		app.Logger().Error("[load dev data] failed to unmarshal json", "name", name, "error", err)
		return err
	}

	for _, obj := range objs {
		record := core.NewRecord(col)
		record.Load(obj)

		if name == "activity" {

			rec := cast.ToString(obj["record"])

			if rec != "" {
				bts, err := fs.ReadFile(rec)
				if err != nil {
					app.Logger().Error("[load dev data] failed to read record file", "rec", rec, "error", err)
					continue
				}
				f1, err := filesystem.NewFileFromBytes(bts, rec)
				if err != nil {
					app.Logger().Error(
						"[load dev data] failed to create file from bytes",
						"rec", rec, "error", err,
					)
					continue
				}
				record.Set("record", f1)
			}
		}

		if name == "objective" {
			docs := cast.ToStringSlice(obj["docs"])
			files := []*filesystem.File{}
			for _, d := range docs {
				if d == "" {
					continue
				}
				bts, err := fs.ReadFile(d)
				if err != nil {
					app.Logger().Error(
						"[load dev data] failed to read file",
						"doc", d, "error", err,
					)
					continue
				}
				f1, err := filesystem.NewFileFromBytes(bts, d)
				if err != nil {
					app.Logger().Error(
						"[load dev data] failed to create file",
						"doc", d, "error", err,
					)
					continue
				}
				files = append(files, f1)
			}
			record.Set("docs", files)
		}

		if name == "users" {
			if value, ok := obj["password"].(string); ok {
				record.SetPassword(value)
			}
		}

		if err := app.Save(record); err != nil {
			app.Logger().Error(
				"[load dev data] failed to save record",
				"name", name,
				"error", err,
			)
		}
	}
	return nil
}
