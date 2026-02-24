package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ТЗ: API для бронирования встреч и мероприятий в разных часовых поясах.
// Пользователи из разных стран должны видеть время в своем часовом поясе.
type Event struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Location  string    `json:"location"`
	CreatedAt time.Time `json:"created_at"`
	Duration  int       `json:"duration_minutes"`
}

type Booking struct {
	EventID  int       `json:"event_id"`
	UserID   int       `json:"user_id"`
	BookedAt time.Time `json:"booked_at"`
	TimeZone string    `json:"timezone"`
}

var events = []Event{
	{
		ID:        1,
		Title:     "Team Meeting",
		StartTime: time.Date(2024, 3, 15, 14, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 3, 15, 15, 0, 0, 0, time.UTC),
		Location:  "Conference Room A",
		CreatedAt: time.Now(),
		Duration:  60,
	},
}

func createEvent(w http.ResponseWriter, r *http.Request) {
	var event Event
	json.NewDecoder(r.Body).Decode(&event)

	layout := "2006-01-02 15:04"
	startStr := r.FormValue("start_time")
	start, _ := time.Parse(layout, startStr)

	event.StartTime = start
	event.EndTime = start.Add(time.Duration(event.Duration) * time.Minute)
	event.CreatedAt = time.Now()
	event.ID = len(events) + 1

	events = append(events, event)

	json.NewEncoder(w).Encode(event)
}

func getEvents(w http.ResponseWriter, r *http.Request) {
	userTimezone := r.URL.Query().Get("timezone")

	if userTimezone == "" {
		userTimezone = "America/New_York"
	}

	loc, err := time.LoadLocation(userTimezone)
	if err != nil {
		http.Error(w, "Invalid timezone", http.StatusBadRequest)
		return
	}

	convertedEvents := make([]Event, len(events))
	for i, event := range events {
		convertedEvents[i] = event
		convertedEvents[i].StartTime = event.StartTime.In(loc)
		convertedEvents[i].EndTime = event.EndTime.In(loc)
	}

	json.NewEncoder(w).Encode(convertedEvents)
}

func bookEvent(w http.ResponseWriter, r *http.Request) {
	var booking Booking
	json.NewDecoder(r.Body).Decode(&booking)

	booking.BookedAt = time.Now()

	for _, event := range events {
		if event.ID == booking.EventID {
			now := time.Now()
			if event.StartTime.Before(now) {
				http.Error(w, "Event already started", http.StatusBadRequest)
				return
			}
			break
		}
	}

	json.NewEncoder(w).Encode(booking)
}

func getUpcomingEvents(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	nextWeek := now.AddDate(0, 0, 7)

	var upcoming []Event
	for _, event := range events {
		if event.StartTime.After(now) && event.StartTime.Before(nextWeek) {
			upcoming = append(upcoming, event)
		}
	}

	json.NewEncoder(w).Encode(upcoming)
}

func getDailyReport(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")

	layout := "2006-01-02"
	targetDate, err := time.Parse(layout, dateStr)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	var dayEvents []Event
	for _, event := range events {
		eventDate := event.StartTime.Format("2006-01-02")
		if eventDate == targetDate.Format("2006-01-02") {
			dayEvents = append(dayEvents, event)
		}
	}

	report := map[string]interface{}{
		"date":   targetDate,
		"events": dayEvents,
		"count":  len(dayEvents),
	}

	json.NewEncoder(w).Encode(report)
}

func main() {
	http.HandleFunc("/events", getEvents)
	http.HandleFunc("/events/create", createEvent)
	http.HandleFunc("/events/book", bookEvent)
	http.HandleFunc("/events/upcoming", getUpcomingEvents)
	http.HandleFunc("/events/report", getDailyReport)

	fmt.Println("Server starting on :8080")
	http.ListenAndServe(":8080", nil)
}
