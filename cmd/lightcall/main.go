package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/tcmzzz/lightcall/server"
	_ "github.com/tcmzzz/lightcall/sql/app"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/spf13/viper"
)

func setConfig() {

	exeFile, err := os.Executable()
	if err != nil {
		panic(err)
	}
	viper.AddConfigPath(filepath.Dir(exeFile))
	viper.AddConfigPath(".")

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	err = viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	viper.SetDefault("fsMasterFile", "/cdr-csv/Master.csv")
	viper.SetDefault("fsRecordDir", "/record")

	viper.SetDefault("cdcFile", "/cdc/cdc.log")
	viper.SetDefault("activityLogFile", "/cdc/activity.log")
	viper.SetDefault("changeLogFile", "/cdc/change.log")
}

func validateConfig() {
	// Validate required file paths
	requiredConfigs := map[string]string{
		"fsMasterFile":    "FS Master CSV file path",
		"fsRecordDir":     "FS Record directory path",
		"activityLogFile": "Activity log file path",
		"changeLogFile":   "Change log file path",
		"cdcFile":         "CDC file path",
	}

	for configKey, description := range requiredConfigs {
		value := viper.GetString(configKey)
		if value == "" {
			log.Fatalf("%s is required but not configured", description)
		}
	}
}

func main() {

	setConfig()
	validateConfig()

	pathConf := &server.FilePath{
		FsRecordDir:    viper.GetString("fsRecordDir"),
		TailFsCDR:      viper.GetString("fsMasterFile"),
		TailChange:     viper.GetString("cdcFile"),
		AppendActivity: viper.GetString("activityLogFile"),
		AppendChange:   viper.GetString("changeLogFile"),
	}

	app := pocketbase.New()

	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{Automigrate: isGoRun, Dir: "./sql/app"})

	// color.NoColor = true
	server.MustRegister(app, pathConf)

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
