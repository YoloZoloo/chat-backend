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
	if len(os.Args) > 1 {
		migrateArg := os.Args[1]
		if migrateArg == "migrate" {
			src.SetEnvVariables()
			model.Migrate()
			return
		}
	}
	src.SetEnvVariables()
	src.StartServer()
}
