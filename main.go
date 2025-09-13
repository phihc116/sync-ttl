package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/phihc116/sync-ttl/internals/infrastructure"
	"github.com/phihc116/sync-ttl/internals/server"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	ctx := context.Background()

	shardIDStr := os.Getenv("SHARD_ID")
	shardID, err := strconv.Atoi(shardIDStr)
	if err != nil {
		log.Fatalf("invalid SHARD_ID: %v", err)
	}

	shardCountStr := os.Getenv("SHARD_COUNT")
	shardCount, err := strconv.Atoi(shardCountStr)
	if err != nil {
		log.Fatalf("invalid SHARD_COUNT: %v", err)
	}

	log.SetOutput(&lumberjack.Logger{
		Filename:   "/app/logs/update_" + shardIDStr + ".log",
		MaxSize:    50,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	})
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	if err := infrastructure.InitDataStore(ctx); err != nil {
		log.Fatal(err)
	}

	// if err := migrations.RunMigrations(); err != nil {
	// 	log.Fatal(err)
	// }

	sqlDb := infrastructure.GetSQLClient()
	dynamo := infrastructure.GetDynamoDbClient()

	log.Println("SQL ready:", sqlDb != nil)
	log.Println("Dynamo ready:", dynamo != nil)

	// filePath := filepath.Join("user_data.txt")
	// // server.LoadUserFromFile(filePath)

	// filePath := filepath.Join("us.csv")
	// server.LoadUserFromFileCSV(filePath)

	log.Printf("Starting UpdateAllUsers for shard %d/%d", shardID, shardCount)
	if err := server.UpdateAllUsers(ctx, 100, shardID, shardCount); err != nil {
		log.Fatal(err)
	}
}
