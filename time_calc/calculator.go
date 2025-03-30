package time_calc

import (
	"fmt"
	"time"
)

// TimeCalculator は基本的な時間計算を行うための構造体です
type TimeCalculator struct {
	// 必要に応じてフィールドを追加できます
}

// stringToTime は、指定された日付文字列をtime.Time型に変換する関数です。
func (c *TimeCalculator) StringToTime(dateStr string) (time.Time, error) {
	// Goのレイアウトは "2006-01-02" を用います
	layout := "2006-01-02"
	t, err := time.Parse(layout, dateStr)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

// AddDuration は、string型の日付に time.Duration を加算し、フォーマット済みの文字列を返すメソッドです
func (c *TimeCalculator) AddDuration(t string, d time.Duration) string {
	var t1 time.Time
	var err error
	if t1, err = c.StringToTime(t); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
	newTime := t1.Add(d)
	// フォーマット例: "2006-01-02 15:04:05"
	return newTime.Format("2006-01-02 15:04:05")
}

// SubtractDuration は、string型の日付から time.Duration を減算し、フォーマット済みの文字列を返すメソッドです
func (c *TimeCalculator) SubtractDuration(t string, d time.Duration) string {
	var t1 time.Time
	var err error
	if t1, err = c.StringToTime(t); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
	newTime := t1.Add(-d)
	return newTime.Format("2006-01-02 15:04:05")
}

// DiffTime は、二つの time.Time の差を計算し、time.Duration を時間単位でフォーマットした文字列を返すメソッドです
func (c *TimeCalculator) DiffTime(t1, t2 time.Time) string {
	diff := t1.Sub(t2)
	// time.Duration の String() メソッドは "72h3m0.5s" のように返す
	return diff.String()
}
