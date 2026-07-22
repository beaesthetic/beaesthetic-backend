package template

import (
	"fmt"
	"strings"
	"time"

	"github.com/goodsign/monday"
)

const (
	defaultTimezone   = "Europe/Rome"
	defaultLocale     = monday.LocaleItIT
	defaultDateLayout = "Monday 2 January"
)

func formatDateWithLayoutDefault(layout string, value any) (string, error) {
	return formatDateWithLayoutIn(string(defaultLocale), defaultTimezone, layout, value)
}

func formatDateWithLayoutIn(locale string, timezone string, layout string, value any) (string, error) {
	parsed, loc, err := valueInTimezone(timezone, value)
	if err != nil {
		return "", err
	}
	resolvedLocale := monday.Locale(strings.TrimSpace(locale))
	if resolvedLocale == "" {
		resolvedLocale = defaultLocale
	}
	resolvedLayout := strings.TrimSpace(layout)
	if resolvedLayout == "" {
		resolvedLayout = defaultDateLayout
	}
	return monday.Format(parsed.In(loc), resolvedLayout, resolvedLocale), nil
}

func isChristmasHoliday(value any) (bool, error) {
	inDecember, err := dateInMonthDayRange(time.December, 8, time.December, 31, value)
	if err != nil {
		return false, err
	}
	inJanuary, err := dateInMonthDayRange(time.January, 1, time.January, 6, value)
	if err != nil {
		return false, err
	}
	return inDecember || inJanuary, nil
}

func dateInMonthDayRange(startMonth time.Month, startDay int, endMonth time.Month, endDay int, value any) (bool, error) {
	parsed, loc, err := valueInTimezone(defaultTimezone, value)
	if err != nil {
		return false, err
	}
	current := monthDay{month: time.Month(parsed.In(loc).Month()), day: parsed.In(loc).Day()}
	start := monthDay{month: startMonth, day: startDay}
	end := monthDay{month: endMonth, day: endDay}
	if monthDayBefore(start, end) {
		return !monthDayBefore(current, start) && !monthDayBefore(end, current), nil
	}
	return !monthDayBefore(current, start) || !monthDayBefore(end, current), nil
}

type monthDay struct {
	month time.Month
	day   int
}

func monthDayBefore(a monthDay, b monthDay) bool {
	if a.month != b.month {
		return a.month < b.month
	}
	return a.day < b.day
}

func valueInTimezone(timezone string, value any) (time.Time, *time.Location, error) {
	loc := loadTimezone(timezone)
	switch typed := value.(type) {
	case time.Time:
		return typed, loc, nil
	case string:
		parsed, err := parseTime(typed, loc)
		if err != nil {
			return time.Time{}, loc, err
		}
		return parsed, loc, nil
	default:
		return time.Time{}, loc, fmt.Errorf("unsupported time value %T", value)
	}
}

func parseTime(value string, loc *time.Location) (time.Time, error) {
	layouts := []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02 15:04:05", "2006-01-02T15:04", "2006-01-02 15:04", "2006-01-02"}
	var lastErr error
	for _, layout := range layouts {
		if layout == time.RFC3339 {
			parsed, err := time.Parse(layout, value)
			if err == nil {
				return parsed, nil
			}
			lastErr = err
			continue
		}
		parsed, err := time.ParseInLocation(layout, value, loc)
		if err == nil {
			return parsed, nil
		}
		lastErr = err
	}
	return time.Time{}, fmt.Errorf("parse time %q: %w", value, lastErr)
}

func loadTimezone(timezone string) *time.Location {
	if strings.TrimSpace(timezone) == "" {
		timezone = defaultTimezone
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.UTC
	}
	return loc
}
