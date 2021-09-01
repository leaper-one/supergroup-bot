package durable

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/shopspring/decimal"
)

var client http.Client

type Api struct {
}

type respData struct {
	Data interface{} `json:"data"`
}

var retry = 0

func (c *Api) httpGetRetry(url string, data interface{}, err error) error {
	if retry >= 10 {
		return err
	}
	time.Sleep(time.Second * 5)
	retry++
	return c.Get(url, data)
}

func (c *Api) RawGet(url string) []byte {
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		return c.RawGet(url)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.RawGet(url)
	}
	return body
}

func (c *Api) Get(url string, data interface{}) error {
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		return c.httpGetRetry(url, data, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.httpGetRetry(url, data, err)
	}
	if strings.HasPrefix(url, "https://serpapi.com/") {
		err = json.Unmarshal(body, data)
		if err != nil {
			return c.httpGetRetry(url, data, err)
		}
	} else {
		var res respData
		res.Data = data
		err = json.Unmarshal(body, &res)
		if err != nil {
			return c.httpGetRetry(url, data, err)
		}
	}
	retry = 0
	return nil
}

type AssetMap map[string]decimal.Decimal
type UserSharesMap map[string]AssetMap

func (c *Api) FoxSharesCheck(userIDs []string, respData *UserSharesMap) error {
	var postData io.Reader
	if userIDs != nil {
		dataByte, err := json.Marshal(userIDs)
		if err != nil {
			return err
		}
		postData = bytes.NewReader(dataByte)
	}
	req, _ := http.NewRequest("POST", "https://f1-defi-api.firesbox.com/app/positions", postData)
	req.Header.Add("Authorization", config.Config.FoxToken)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, respData)
	if err != nil {
		if retry >= 10 {
			return err
		}
		time.Sleep(time.Second * 5)
		return c.FoxSharesCheck(userIDs, respData)
	}
	retry = 0
	return nil
}

type exinAssetItem map[string]decimal.Decimal

type exinShareUserItem struct {
	UserID     string        `json:"user_id"`
	Interest   exinAssetItem `json:"interest,omitempty"`
	FundPool   exinAssetItem `json:"fund-pool,omitempty"`
	Investment exinAssetItem `json:"investment,omitempty"`
	Invest     exinAssetItem `json:"invest,omitempty"`
}

type exinSharesResp struct {
	Code    int                 `json:"code"`
	Success bool                `json:"success"`
	Data    []exinShareUserItem `json:"data"`
}

type exinSharesOneResp struct {
	Code    int               `json:"code"`
	Success bool              `json:"success"`
	Data    exinShareUserItem `json:"data"`
}

func (c *Api) ExinSharesCheck(userIDs []string, assetIDs []string, respData *UserSharesMap) error {
	req, _ := http.NewRequest("POST",
		fmt.Sprintf("https://eiduwejdk.com/mixin-task?api_key=%s&user_id=%s&asset_uuid=%s&api_subject=mixin-social",
			config.Config.ExinToken, strings.Join(userIDs, ","), strings.Join(assetIDs, ",")), nil)
	resp, err := client.Do(req)
	if err != nil {
		if retry > 10 {
			return err
		}
		time.Sleep(time.Second * 5)
		return c.ExinSharesCheck(userIDs, assetIDs, respData)
	}
	retry = 0
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}
	var sharesResp exinSharesResp

	if len(userIDs) == 1 {
		var t exinSharesOneResp
		err = json.Unmarshal(body, &t)
		if err != nil {
			log.Println("ExinSharesCheck error", err, string(body))
			return err
		}
		sharesResp.Data = []exinShareUserItem{t.Data}
	} else {
		err = json.Unmarshal(body, &sharesResp)
		if err != nil {
			log.Println("ExinSharesCheck error", err, string(body))
			return err
		}
	}

	for _, shares := range sharesResp.Data {
		(*respData)[shares.UserID] = make(AssetMap)
		tmp := (*respData)[shares.UserID]
		for _, assetID := range assetIDs {
			assetAmount := decimal.Zero

			if shares.FundPool[assetID].GreaterThan(decimal.Zero) {
				assetAmount = assetAmount.Add(shares.FundPool[assetID])
			}
			if shares.Investment[assetID].GreaterThan(decimal.Zero) {
				assetAmount = assetAmount.Add(shares.Investment[assetID])
			}
			if shares.Invest[assetID].GreaterThan(decimal.Zero) {
				assetAmount = assetAmount.Add(shares.Invest[assetID])
			}
			if assetAmount.GreaterThan(decimal.Zero) {
				tmp[assetID] = assetAmount
			}
		}
	}
	return nil
}
