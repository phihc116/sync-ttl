package main

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/phihc116/sync-ttl/server"
)

// 1000
const scanLimit = 3

func main() {
	totalSegments := int32(runtime.GOMAXPROCS(0))
	fmt.Printf("Using TotalSegments = %d\n", totalSegments)

	ctx := context.TODO()
	db, err := server.NewDynamoDb(ctx)
	if err != nil {
		log.Fatal(err)
	}

	nowMillis := time.Now().UnixMilli()
	threshold := nowMillis - int64(90*24*time.Hour.Milliseconds())
	ninetyDays := int64(90 * 24 * time.Hour.Seconds())

	var wg sync.WaitGroup
	var total int64
	mu := sync.Mutex{}

	for seg := range totalSegments {
		wg.Add(1)
		go func(segment int32) {
			defer wg.Done()
			var lastKey map[string]types.AttributeValue

			for {
				out, err := db.Scan(ctx, &dynamodb.ScanInput{
					TableName:         aws.String(server.Table),
					Segment:           aws.Int32(segment),
					TotalSegments:     aws.Int32(totalSegments),
					ExclusiveStartKey: lastKey,
					FilterExpression:  aws.String("Ctime >= :threshold AND DataType IN (:t1, :t2)"),
					ExpressionAttributeValues: map[string]types.AttributeValue{
						":threshold": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", threshold)},
						":t1":        &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", server.HistoryTypeID)},
						":t2":        &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", server.HistoryDeleteDirectiveTypeID)},
					},
					Limit: aws.Int32(scanLimit),
				})
				if err != nil {
					log.Printf("[seg %d] scan failed: %v", segment, err)
					return
				}

				var items []server.SyncEntity
				if err := attributevalue.UnmarshalListOfMaps(out.Items, &items); err != nil {
					log.Printf("[seg %d] unmarshal failed: %v", segment, err)
					return
				}

				for _, it := range items {
					ctimeSec := *it.Ctime / 1000
					expiration := ctimeSec + ninetyDays

					_, err := db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
						TableName: aws.String(server.Table),
						Key: map[string]types.AttributeValue{
							"ClientID": &types.AttributeValueMemberS{Value: it.ClientID},
							"ID":       &types.AttributeValueMemberS{Value: it.ID},
						},
						UpdateExpression: aws.String("SET ExpirationTime = :exp"),
						ExpressionAttributeValues: map[string]types.AttributeValue{
							":exp": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", expiration)},
						},
					})

					if err != nil {
						log.Printf("[seg %d] update failed for %s/%s (DataType=%v, Ctime=%v): %v",
							segment, it.ClientID, it.ID, *it.DataType, *it.Ctime, err)
					} else {
						log.Printf("[seg %d] updated %s/%s (DataType=%v, Ctime=%v)",
							segment, it.ClientID, it.ID, *it.DataType, *it.Ctime)
					}

					mu.Lock()
					total++
					mu.Unlock()
				}

				if out.LastEvaluatedKey == nil {
					break
				}
				lastKey = out.LastEvaluatedKey
			}
		}(int32(seg))
	}

	wg.Wait()
	fmt.Println("Total items updated:", total)
}
