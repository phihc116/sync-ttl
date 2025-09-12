package server

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/phihc116/sync-ttl/internals/infrastructure"
	"github.com/phihc116/sync-ttl/internals/models"
)

func LoadUserFromFile(filePath string) error {
	db := infrastructure.GetSQLClient()

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	const maxCapacity = 1024 * 1024
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, maxCapacity)

	var users []models.User
	const batchSize = 100
	totalInserted := 0

	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "\t")
		if len(parts) < 3 {
			continue
		}

		clientID, _ := strconv.ParseInt(parts[1], 10, 64)

		s := strings.TrimSpace(parts[0])
		s = strings.TrimPrefix(s, "\ufeff")
		externalID, _ := strconv.ParseInt(s, 10, 64)

		user := models.User{
			ID:           uuid.NewString(),
			ClientID:     clientID,
			ExternalID:   externalID,
			Email:        parts[2],
			IsUpdated:    false,
			UpdatedCount: 0,
		}
		users = append(users, user)

		if len(users) >= batchSize {
			if err := db.CreateInBatches(users, batchSize).Error; err != nil {
				return err
			}
			totalInserted += len(users)
			log.Printf("Inserted %d users (total %d)", len(users), totalInserted)
			users = users[:0]
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	if len(users) > 0 {
		if err := db.CreateInBatches(users, batchSize).Error; err != nil {
			return err
		}
		totalInserted += len(users)
		log.Printf("Inserted %d users (total %d)", len(users), totalInserted)
	}

	if totalInserted == 0 {
		log.Println("No users found in file")
	} else {
		log.Printf("Finished inserting %d users into SQL Server", totalInserted)
	}

	return nil
}
