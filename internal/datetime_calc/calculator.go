package datetime_calc

import (
	"time"
)

// DatetimeCalculator は基本的な時間計算を行うための構造体です
type DatetimeCalculator struct {
	// 必要に応じてフィールドを追加できます
}

// NewTime は指定された年月日時分秒をもとに time.Time を返す関数です。
func (c *DatetimeCalculator) NewTime(year, month, day, hour, minute, second int) time.Time {
	// time.Date の第2引数は time.Month 型であるため、int から変換します。
	return time.Date(year, time.Month(month), day, hour, minute, second, 0, time.Local)
}

// stringToTime は、指定された日付文字列をtime.Time型に変換する関数です。
func (c *DatetimeCalculator) StringToTime(dateStr string) (time.Time, error) {
	// Goのレイアウトは "2006-01-02" を用います
	layout := "2006-01-02 15:04:05"
	t, err := time.Parse(layout, dateStr)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

// AddDuration は、string型の日付に time.Duration を加算し、フォーマット済みの文字列を返すメソッドです
func (c *DatetimeCalculator) AddDuration(year, month, day, hour, minute, second int, d time.Duration) string {
	t1 := c.NewTime(year, month, day, hour, minute, second)
	newTime := t1.Add(d)
	// フォーマット例: "2006-01-02 15:04:05"
	return newTime.Format("2006-01-02 15:04:05")
}

// AddDatetime は、指定された日付に年月日時分秒を加算し、フォーマット済みの文字列を返すメソッドです
func (c *DatetimeCalculator) AddDatetime(year, month, day, hour, minute, second int, addYears, addMonths, addDays, addHours, addMinutes, addSeconds int) string {
	t1 := c.NewTime(year, month, day, hour, minute, second)

	// 年月日を加算
	t2 := t1.AddDate(addYears, addMonths, addDays)

	// 時分秒を加算
	duration := time.Duration(addHours)*time.Hour + time.Duration(addMinutes)*time.Minute + time.Duration(addSeconds)*time.Second
	newTime := t2.Add(duration)

	return newTime.Format("2006-01-02 15:04:05")
}

// AddDatetimeFloat は、AddDatetime関数のラッパー関数で、すべての引数をfloat64型で受け取ります
func (c *DatetimeCalculator) AddDatetimeFloat(year, month, day, hour, minute, second float64, addYears, addMonths, addDays, addHours, addMinutes, addSeconds float64) string {
	// float64からintに変換
	yearInt := int(year)
	monthInt := int(month)
	dayInt := int(day)
	hourInt := int(hour)
	minuteInt := int(minute)
	secondInt := int(second)
	addYearsInt := int(addYears)
	addMonthsInt := int(addMonths)
	addDaysInt := int(addDays)
	addHoursInt := int(addHours)
	addMinutesInt := int(addMinutes)
	addSecondsInt := int(addSeconds)

	// 元のAddDatetime関数を呼び出す
	return c.AddDatetime(yearInt, monthInt, dayInt, hourInt, minuteInt, secondInt,
		addYearsInt, addMonthsInt, addDaysInt, addHoursInt, addMinutesInt, addSecondsInt)
}

// SubtractDuration は、string型の日付から time.Duration を減算し、フォーマット済みの文字列を返すメソッドです
func (c *DatetimeCalculator) SubtractDuration(year, month, day, hour, minute, second int, d time.Duration) string {
	t1 := c.NewTime(year, month, day, hour, minute, second)
	newTime := t1.Add(-d) // 減算する時間を正しく計算
	return newTime.Format("2006-01-02 15:04:05")
}

// SubtractDatetime は、指定された日付から年月日時分秒を減算し、フォーマット済みの文字列を返すメソッドです
func (c *DatetimeCalculator) SubtractDatetime(year, month, day, hour, minute, second int, subYears, subMonths, subDays, subHours, subMinutes, subSeconds int) string {
	t1 := c.NewTime(year, month, day, hour, minute, second)

	// 年月日を減算（負の値を渡して減算）
	t2 := t1.AddDate(-subYears, -subMonths, -subDays)

	// 時分秒を減算（負の値を渡して減算）
	duration := time.Duration(-subHours)*time.Hour + time.Duration(-subMinutes)*time.Minute + time.Duration(-subSeconds)*time.Second
	newTime := t2.Add(duration)

	return newTime.Format("2006-01-02 15:04:05")
}

// SubtractDatetimeFloat は、SubtractDatetime関数のラッパー関数で、すべての引数をfloat64型で受け取ります
func (c *DatetimeCalculator) SubtractDatetimeFloat(year, month, day, hour, minute, second float64, subYears, subMonths, subDays, subHours, subMinutes, subSeconds float64) string {
	// float64からintに変換
	yearInt := int(year)
	monthInt := int(month)
	dayInt := int(day)
	hourInt := int(hour)
	minuteInt := int(minute)
	secondInt := int(second)
	subYearsInt := int(subYears)
	subMonthsInt := int(subMonths)
	subDaysInt := int(subDays)
	subHoursInt := int(subHours)
	subMinutesInt := int(subMinutes)
	subSecondsInt := int(subSeconds)

	// 元のSubtractDatetime関数を呼び出す
	return c.SubtractDatetime(yearInt, monthInt, dayInt, hourInt, minuteInt, secondInt,
		subYearsInt, subMonthsInt, subDaysInt, subHoursInt, subMinutesInt, subSecondsInt)
}

// DiffTime は、二つの time.Time の差を計算し、time.Duration を時間単位でフォーマットした文字列を返すメソッドです
func (c *DatetimeCalculator) DiffTime(t1, t2 time.Time) string {
	diff := t1.Sub(t2)
	// time.Duration の String() メソッドは "72h3m0.5s" のように返す
	return diff.String()
}
