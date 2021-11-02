package services

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
)

type MigrationService struct{}

var w sync.WaitGroup

func (service *MigrationService) Run(ctx context.Context) error {
	client, err := addClient(ctx)
	if err != nil {
		return err
	}
	f, err := os.Open("users.sql")
	if err != nil {
		log.Println("users.sql open fail...")
		return err
	}
	buf := bufio.NewReader(f)
	i := 0
	for {
		w.Add(1)
		i++
		if client.Client.SpeakStatus == models.ClientSpeckStatusOpen && i%100 == 0 {
			time.Sleep(10 * time.Second)
		}
		line, err := buf.ReadString('\n')
		go handleUserLine(ctx, client, strings.TrimSpace(line))
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}
	w.Wait()
	f, err = os.Open("block.sql")
	if err != nil {
		log.Println("users.sql open fail...")
		return err
	}
	buf = bufio.NewReader(f)
	i = 0
	for {
		i++
		log.Println(i)
		line, err := buf.ReadString('\n')
		handleBlockLine(ctx, client.Client.ClientID, strings.TrimSpace(line))
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}
	return nil
}

func handleBlockLine(ctx context.Context, clientID, userID string) {
	if userID == "" {
		return
	}
	query := durable.InsertQueryOrUpdate("client_users", "client_id,user_id", "priority,status")
	_, err := session.Database(ctx).Exec(ctx, query, clientID, userID, models.ClientUserPriorityStop, models.ClientUserStatusAudience)
	if err != nil {
		session.Logger(ctx).Println(err)
	}
}

func handleUserLine(ctx context.Context, client *clientInfo, line string) {
	u := getUserInfoFromLine(ctx, client, line)
	if u == nil || u.UserID == "" {
		w.Done()
		return
	}

	if client.Client.ClientID == "47cdbc9e-e2b9-4d1f-b13e-42fec1d8853d" && time.Since(u.CreatedAt).Hours() > 14*24 {
		u.Priority = models.ClientUserPriorityStop
	}
	if err := models.CreateOrUpdateClientUser(ctx, u); err != nil {
		session.Logger(ctx).Println(err)
	}
	w.Done()
}

func getUserInfoFromLine(ctx context.Context, client *clientInfo, line string) *models.ClientUser {
	users := strings.Split(line, "|")
	if len(users) == 4 {
		var u models.ClientUser
		u.ClientID = client.Client.ClientID
		for i, d := range users {
			d = strings.TrimSpace(d)
			switch i {
			case 0:
				u.UserID = d
			case 1:
				u.AccessToken = d
			case 2:
				log.Println(d)
				t, _ := time.Parse("2006-01-02 15:04:05.000000+00", d)
				log.Println(t)
				u.CreatedAt = t
			case 3:
				t, _ := time.Parse("2006-01-02 15:04:05.000000+00", d)
				u.DeliverAt = t
				u.ReadAt = t
			}
		}
		if u.UserID == "" {
			return &u
		}
		go models.SearchUser(ctx, u.UserID)
		if tools.Includes(client.ManagerList, u.UserID) {
			u.Status = models.ClientUserStatusAdmin
			u.Priority = models.ClientUserPriorityHigh
		} else {
			curStatus, err := models.GetClientUserStatusByClientUser(ctx, &u)
			if err != nil {
				session.Logger(ctx).Println(err)
			}
			var priority int
			if curStatus != models.ClientUserStatusAudience {
				priority = models.ClientUserPriorityHigh
			} else {
				priority = models.ClientUserPriorityLow
			}
			u.Status = curStatus
			u.Priority = priority
		}
		return &u
	} else {
		log.Println(users)
		return nil
	}
}
