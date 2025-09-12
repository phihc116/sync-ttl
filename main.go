package main

import (
	"context"
	"log"
	"path/filepath"

	"github.com/phihc116/sync-ttl/internals/infrastructure"
	"github.com/phihc116/sync-ttl/internals/migrations"
	"github.com/phihc116/sync-ttl/internals/server"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	ctx := context.Background()

	log.SetOutput(&lumberjack.Logger{
		Filename:   "update.log",
		MaxSize:    100,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	})
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	if err := infrastructure.InitDataStore(ctx); err != nil {
		log.Fatal(err)
	}

	if err := migrations.RunMigrations(); err != nil {
		log.Fatal(err)
	}

	sqlDb := infrastructure.GetSQLClient()
	dynamo := infrastructure.GetDynamoDbClient()

	log.Println("SQL ready:", sqlDb != nil)
	log.Println("Dynamo ready:", dynamo != nil)

	filePath := filepath.Join("user_data.txt")
	server.LoadUserFromFile(filePath)

	//server.UpdateAllUsers(ctx, 200)
}
