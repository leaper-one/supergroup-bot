package tools

import (
	"crypto/md5"
	"github.com/MixinNetwork/supergroup/config"
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"io"
	"math/big"
	"strings"
)

func GetUUID() string {
	id, _ := uuid.NewV4()
	return id.String()
}

func CompareTwoString(s1, s2 string) (decimal.Decimal, decimal.Decimal, error) {
	d1, err := decimal.NewFromString(s1)
	if err != nil {
		return decimal.Decimal{}, decimal.Decimal{}, err
	}
	d2, err := decimal.NewFromString(s2)
	if err != nil {
		return decimal.Decimal{}, decimal.Decimal{}, err
	}
	return d1, d2, nil
}

func NumberFixed(s string, i int32) string {
	d, _ := decimal.NewFromString(s)
	return d.StringFixed(i)
}
func ShardId(cid, uid string) (string) {
	minId, maxId := cid, uid
	if strings.Compare(cid, uid) > 0 {
		maxId, minId = cid, uid
	}
	h := md5.New()
	io.WriteString(h, minId)
	io.WriteString(h, maxId)

	b := new(big.Int).SetInt64(config.MessageShardSize)
	c := new(big.Int).SetBytes(h.Sum(nil))
	m := new(big.Int).Mod(c, b)
	h = md5.New()
	h.Write([]byte(config.MessageShardModifier))
	h.Write(m.Bytes())
	s := h.Sum(nil)
	s[6] = (s[6] & 0x0f) | 0x30
	s[8] = (s[8] & 0x3f) | 0x80
	sid, _ := uuid.FromBytes(s)
	return sid.String()
}
