package services

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/MixinNetwork/supergroup/handlers/common"
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
		if client.Client.SpeakStatus == common.ClientSpeckStatusOpen && i%100 == 0 {
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
		log.Println("block.sql open fail...")
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
	session.Redis(ctx).QDel(ctx, fmt.Sprintf("client_user:%s:%s", clientID, userID))
	if err := session.DB(ctx).Save(&models.ClientUser{
		ClientID:   clientID,
		UserID:     userID,
		Priority:   models.ClientUserPriorityLow,
		Status:     models.ClientUserStatusAudience,
		IsReceived: false,
	}).Error; err != nil {
		tools.Println(err)
	}

}

func handleUserLine(ctx context.Context, client *clientInfo, line string) {
	u := getUserInfoFromLine(ctx, client, line)
	if u == nil || u.UserID == "" {
		w.Done()
		return
	}

	if client.Client.ClientID == "47cdbc9e-e2b9-4d1f-b13e-42fec1d8853d" && time.Since(u.CreatedAt).Hours() > 14*24 {
		u.Priority = models.ClientUserStatusAudience
		u.IsReceived = false
	}
	session.Redis(ctx).QDel(ctx, fmt.Sprintf("client_user:%s:%s", u.ClientID, u.UserID))
	if err := session.DB(ctx).Save(&u).Error; err != nil {
		tools.Println(err)
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
		go common.SearchUser(ctx, u.ClientID, u.UserID)
		if tools.Includes(client.ManagerList, u.UserID) {
			u.Status = models.ClientUserStatusAdmin
			u.Priority = models.ClientUserPriorityHigh
		} else {
			u.Status = models.ClientUserStatusLarge
			u.Priority = models.ClientUserPriorityHigh
		}
		return &u
	} else {
		log.Println(users)
		return nil
	}
}
