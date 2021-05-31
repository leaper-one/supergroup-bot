package tools

import (
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
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
