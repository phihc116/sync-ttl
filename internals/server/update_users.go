package server

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/phihc116/sync-ttl/internals/infrastructure"
	"github.com/phihc116/sync-ttl/internals/models"
	"gorm.io/gorm"
)

const (
	queryLimit     = 500
	maxConcurrency = 50
)

func UpdateAllUsers(ctx context.Context, batchSize int, shardID, shardCount int) error {
	sqlDB := infrastructure.GetSQLClient()
	dynamo := infrastructure.GetDynamoDbClient()

	nowMillis := time.Now().UnixMilli()
	threshold := nowMillis - int64(90*24*time.Hour.Milliseconds())
	ninetyDays := int64(90 * 24 * time.Hour.Seconds())

	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	return sqlDB.Model(&models.User{}).Where("is_updated = ?", false).
		Where("is_updated = ? AND client_id % ? = ?", false, shardCount, shardID).
		FindInBatches(&[]models.User{}, batchSize, func(tx *gorm.DB, batch int) error {
			usersPtr, ok := tx.Statement.Dest.(*[]models.User)
			if !ok {
				return fmt.Errorf("unexpected type for batch destination")
			}
			users := *usersPtr

			for _, u := range users {
				wg.Add(1)
				sem <- struct{}{}
				go func(u models.User) {
					defer wg.Done()
					defer func() { <-sem }()

					total := 0
					var lastKey map[string]types.AttributeValue

					for {
						out, err := dynamo.Query(ctx, &dynamodb.QueryInput{
							TableName:              aws.String(infrastructure.Table),
							KeyConditionExpression: aws.String("ClientID = :cid"),
							FilterExpression:       aws.String("Ctime >= :threshold AND DataType IN (:t1, :t2)"),
							ExpressionAttributeValues: map[string]types.AttributeValue{
								":cid":       &types.AttributeValueMemberS{Value: fmt.Sprintf("%d", u.ClientID)},
								":threshold": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", threshold)},
								":t1":        &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", models.HistoryTypeID)},
								":t2":        &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", models.HistoryDeleteDirectiveTypeID)},
							},
							ExclusiveStartKey: lastKey,
							Limit:             aws.Int32(queryLimit),
						})
						if err != nil {
							log.Printf("query failed for ClientID %d: %v", u.ClientID, err)
							break
						}

						var items []models.SyncEntity
						if err := attributevalue.UnmarshalListOfMaps(out.Items, &items); err != nil {
							log.Printf("unmarshal failed for ClientID %d: %v", u.ClientID, err)
							break
						}

						for _, it := range items {
							ctimeSec := *it.Ctime / 1000
							expiration := ctimeSec + ninetyDays

							_, err := dynamo.UpdateItem(ctx, &dynamodb.UpdateItemInput{
								TableName: aws.String(infrastructure.Table),
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
								log.Printf("Update failed for %s/%s (DataType=%v, Ctime=%v): %v",
									it.ClientID, it.ID, *it.DataType, *it.Ctime, err)
							} else {
								total++
								log.Printf("Update success: ClientID=%s, ID=%s, DataType=%v, Ctime=%v",
									it.ClientID, it.ID, *it.DataType, *it.Ctime)
							}
						}

						if out.LastEvaluatedKey == nil {
							break
						}
						lastKey = out.LastEvaluatedKey
					}

					if total > 0 {
						if err := sqlDB.Model(&u).
							Updates(map[string]any{
								"is_updated":    true,
								"updated_count": total,
							}).Error; err != nil {
							log.Printf("failed to mark updated for ClientID %d: %v", u.ClientID, err)
						} else {
							log.Printf("ClientID %d updated (%d items)", u.ClientID, total)
						}
					}
				}(u)
			}

			wg.Wait()
			return nil
		}).Error
}
