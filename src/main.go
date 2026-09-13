package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
)

var (
	sourceURL  string
	listenAddr string
)

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing required env var: %s", key)
	}
	return v
}

func handler(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(sourceURL)
	if err != nil {
		http.Error(w, "failed to fetch source", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "failed to read source", http.StatusBadGateway)
		return
	}

	cal, err := ics.ParseCalendar(strings.NewReader(string(body)))
	if err != nil {
		http.Error(w, "failed to parse ics", http.StatusInternalServerError)
		return
	}

	for _, event := range cal.Events() {
		startProp := event.GetProperty(ics.ComponentPropertyDtStart)
		if startProp == nil {
			continue
		}
		start, err := event.GetStartAt()
		if err != nil {
			continue
		}
		date := start.Truncate(24 * time.Hour)

		event.SetProperty(ics.ComponentPropertyDtStart,
			date.Format("20060102"),
			ics.WithValue(string(ics.ValueDataTypeDate)))

		if _, err := event.GetEndAt(); err == nil {
			end := date.AddDate(0, 0, 1)
			event.SetProperty(ics.ComponentPropertyDtEnd,
				end.Format("20060102"),
				ics.WithValue(string(ics.ValueDataTypeDate)))
		}

		alarm := event.AddAlarm()
		alarm.SetAction(ics.ActionDisplay)
		alarm.SetTrigger("-PT9H")
		alarm.SetDescription("Reminder")
	}

	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	fmt.Fprint(w, cal.Serialize())
}

func main() {
	sourceURL = mustEnv("ICAL_PROXY_SOURCE_URL")
	listenAddr = os.Getenv("ICAL_PROXY_LISTEN")
	if listenAddr == "" {
		listenAddr = "127.0.0.1:8080"
	}

	http.HandleFunc("/calendar.ics", handler)
	log.Printf("listening on %s", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
