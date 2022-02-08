package tools

import (
	"unicode"

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

func LanguageCount(str1 string, lang *unicode.RangeTable) decimal.Decimal {
	var count, totalCount int64
	for _, char := range str1 {
		if unicode.IsLetter(char) {
			totalCount++
		}
		if lang == nil {
			if (char >= 'a' && char <= 'z') ||
				(char >= 'A' && char <= 'Z') {
				count++
			}
		} else if unicode.Is(lang, char) {
			count++
		}
	}
	if totalCount == 0 {
		totalCount = 1
	}
	return decimal.NewFromInt(count).Div(decimal.NewFromInt(totalCount))
}
