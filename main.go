package main

import (
	model "chat-backend/model"
	"chat-backend/src"
	"os"

	_ "github.com/astaxie/session"
	_ "github.com/astaxie/session/providers/memory"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	migrateArg := os.Args[1]
	if migrateArg == "migrate" {
		model.Migrate()
		return
	}
	src.SetEnvVariables()
	src.StartServer()
}
