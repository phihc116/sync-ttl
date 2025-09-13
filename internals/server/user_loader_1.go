package server

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/phihc116/sync-ttl/internals/infrastructure"
	"github.com/phihc116/sync-ttl/internals/models"
)

func LoadUserFromFileCSV(filePath string) error {
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
	const batchSize = 300
	totalInserted := 0

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		if len(parts) < 6 {
			log.Printf("skip invalid line: %s", line)
			continue
		}

		clientID, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil {
			log.Printf("invalid clientID: %s", parts[1])
			continue
		}

		externalID, err := strconv.ParseInt(strings.TrimSpace(parts[2]), 10, 64)
		if err != nil {
			log.Printf("invalid externalID: %s", parts[2])
			continue
		}

		isUpdated, err := strconv.ParseBool(strings.TrimSpace(parts[4]))
		if err != nil {
			isUpdated = parts[4] == "1"
		}

		updatedCount, err := strconv.ParseInt(strings.TrimSpace(parts[5]), 10, 64)
		if err != nil {
			log.Printf("invalid updated_count: %s", parts[5])
			continue
		}

		user := models.User{
			ID:           strings.TrimSpace(parts[0]),
			ClientID:     clientID,
			ExternalID:   externalID,
			Email:        strings.TrimSpace(parts[3]),
			IsUpdated:    isUpdated,
			UpdatedCount: updatedCount,
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

	log.Printf("Finished inserting %d users into SQL Server", totalInserted)
	return nil
}
