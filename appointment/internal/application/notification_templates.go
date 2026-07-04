package application

import (
	"fmt"
	"time"

	"github.com/petretiandrea/beaesthetic-backend/appointment/internal/domain"
)

func ReminderNotificationPayload(event *domain.AgendaEvent) (string, string) {
	day := formatDateToRome(event.Start)
	hour := formatTimeToRome(event.Start)
	if timeIncludeDayRange(event.Start, monthDay{month: time.December, day: 10}, monthDay{month: time.January, day: 6}) {
		return "", fmt.Sprintf("Il centro Be Aesthetic ti ricorda il tuo appuntamento di %s alle ore %s.\nBuona giornata e buone feste!", day, hour)
	}
	return "", fmt.Sprintf("Il centro Be Aesthetic ti ricorda il tuo appuntamento di %s alle ore %s.\nBuona giornata!", day, hour)
}

func confirmationNotificationPayload(event *domain.AgendaEvent, isRescheduled bool) (string, string) {
	day := formatDateToRome(event.Start)
	hour := formatTimeToRome(event.Start)
	if isRescheduled {
		return "", fmt.Sprintf("Il centro Be Aesthetic ti informa che il tuo appuntamento è stato spostato. La nuova data è %s alle ore %s.\nBuona giornata!", day, hour)
	}
	return "", fmt.Sprintf("Il centro Be Aesthetic ti conferma la prenotazione del tuo appuntamento per il giorno %s alle ore %s.\nBuona giornata!", day, hour)
}

type monthDay struct {
	month time.Month
	day   int
}

func timeIncludeDayRange(value time.Time, start monthDay, end monthDay) bool {
	current := monthDay{month: value.UTC().Month(), day: value.UTC().Day()}
	if monthDayBefore(start, end) {
		return !monthDayBefore(current, start) && !monthDayBefore(end, current)
	}
	return !monthDayBefore(current, start) || !monthDayBefore(end, current)
}

func monthDayBefore(a monthDay, b monthDay) bool {
	if a.month != b.month {
		return a.month < b.month
	}
	return a.day < b.day
}

func formatDateToRome(value time.Time) string {
	local := value.In(romeLocation())
	return fmt.Sprintf("%s %d %s, %d", italianWeekday(local.Weekday()), local.Day(), italianMonth(local.Month()), local.Year())
}

func formatTimeToRome(value time.Time) string {
	return value.In(romeLocation()).Format("15:04")
}

func romeLocation() *time.Location {
	loc, err := time.LoadLocation("Europe/Rome")
	if err != nil {
		return time.UTC
	}
	return loc
}

func italianWeekday(day time.Weekday) string {
	weekdays := map[time.Weekday]string{
		time.Sunday:    "domenica",
		time.Monday:    "lunedì",
		time.Tuesday:   "martedì",
		time.Wednesday: "mercoledì",
		time.Thursday:  "giovedì",
		time.Friday:    "venerdì",
		time.Saturday:  "sabato",
	}
	return weekdays[day]
}

func italianMonth(month time.Month) string {
	months := map[time.Month]string{
		time.January:   "gennaio",
		time.February:  "febbraio",
		time.March:     "marzo",
		time.April:     "aprile",
		time.May:       "maggio",
		time.June:      "giugno",
		time.July:      "luglio",
		time.August:    "agosto",
		time.September: "settembre",
		time.October:   "ottobre",
		time.November:  "novembre",
		time.December:  "dicembre",
	}
	return months[month]
}
