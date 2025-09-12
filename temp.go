// package main

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"time"

// 	"github.com/aws/aws-sdk-go-v2/aws"
// 	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
// 	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
// 	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
// 	"github.com/phihc116/sync-ttl/server"
// )

// const queryLimit = 50

// func main() {
// 	ctx := context.TODO()
// 	db, err := server.NewDynamoDb(ctx)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	clientID := "1721045845"

// 	nowMillis := time.Now().UnixMilli()
// 	threshold := nowMillis - int64(90*24*time.Hour.Milliseconds())
// 	ninetyDays := int64(90 * 24 * time.Hour.Seconds())

// 	var lastKey map[string]types.AttributeValue
// 	var total int

// 	for {
// 		// Query theo partition key = ClientID
// 		out, err := db.Query(ctx, &dynamodb.QueryInput{
// 			TableName:              aws.String(server.Table),
// 			KeyConditionExpression: aws.String("ClientID = :cid"),
// 			FilterExpression:       aws.String("Ctime >= :threshold AND DataType IN (:t1, :t2)"),
// 			ExpressionAttributeValues: map[string]types.AttributeValue{
// 				":cid":       &types.AttributeValueMemberS{Value: clientID},
// 				":threshold": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", threshold)},
// 				":t1":        &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", server.HistoryTypeID)},
// 				":t2":        &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", server.HistoryDeleteDirectiveTypeID)},
// 			},
// 			ExclusiveStartKey: lastKey,
// 			Limit:             aws.Int32(queryLimit),
// 		})
// 		if err != nil {
// 			log.Fatalf("query failed: %v", err)
// 		}

// 		var items []server.SyncEntity
// 		if err := attributevalue.UnmarshalListOfMaps(out.Items, &items); err != nil {
// 			log.Fatalf("unmarshal failed: %v", err)
// 		}

// 		for _, it := range items {
// 			ctimeSec := *it.Ctime / 1000
// 			expiration := ctimeSec + ninetyDays

// 			_, err := db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
// 				TableName: aws.String(server.Table),
// 				Key: map[string]types.AttributeValue{
// 					"ClientID": &types.AttributeValueMemberS{Value: it.ClientID},
// 					"ID":       &types.AttributeValueMemberS{Value: it.ID},
// 				},
// 				UpdateExpression: aws.String("SET ExpirationTime = :exp"),
// 				ExpressionAttributeValues: map[string]types.AttributeValue{
// 					":exp": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", expiration)},
// 				},
// 			})
// 			if err != nil {
// 				log.Printf("update failed for %s/%s: %v", it.ClientID, it.ID, err)
// 			} else {
// 				total++
// 			}
// 		}

// 		if out.LastEvaluatedKey == nil {
// 			break
// 		}
// 		lastKey = out.LastEvaluatedKey
// 	}

// 	fmt.Println("Total items updated for client", clientID, ":", total)
// }
