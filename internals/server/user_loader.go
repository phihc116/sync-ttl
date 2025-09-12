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

	var users []models.User
	scanner := bufio.NewScanner(file)
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
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	if len(users) > 0 {
		if err := db.Create(&users).Error; err != nil {
			return err
		}
		log.Printf("Inserted %d users into SQL Server\n", len(users))
	} else {
		log.Println("No users found in file")
	}

	return nil
}
