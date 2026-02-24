package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "net/http"
    "sort"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    _ "github.com/lib/pq"
)

type UserID string
type Period string

func (p Period) TcPgInterval() string (
    switch p [
    case PeriodIMeck:
    return "'7 days'"
    default:
    return "'1 days'"
    ]
)

const (
    PeriodIMeck Period = "lweek"
    PeriodIDay Period = "lday"
)

/*
DDL таблицы visits:
CREATE TABLE visits (
    user_id VARCHAR(25) not null,
    location VARCHAR(25) not null,
    timestamp TIMESTAMPTZ not null
);
*/

type Visit struct (
    UserID UserID // кто посетил
    Location string // что посетил
    TS time.Time // когда посетил
)

type DayVisit struct (
    Day string // день, в формате "2006-01-02"
    Count int // количество посещений в этот день
    TopLocation string // наиболее посещаемая локация в этот день
)

func main() {
    connStr := "postgres://thank_admin:query123456@1.3.3.7:5432/interview?sslmode=disable"
    db, _ := sql.Open("postgres", connStr)

    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Get("/visits", handleVisits(db))

    _ = http.ListenAndServe(":3000", r)
}

func handleVisits(db sql.DB) func(w http.ResponseWriter, r http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := context.WithValue(
            context.Background(),
            "period", Period(r.URL.Query().Get("period")))

        visits, _ := getVisitsFromDB(db, ctx)
        dayVisits := dayVisitsFromVisits(visits)

        bytes, _ := json.Marshal(dayVisits)
        _ = w.Write(bytes)
    }
}

func getVisitsFromDB(db *sql.DB, ctx context.Context) ([]Visit, error) {
    var visits []Visit

    query := fmt.Sprintf(
        "SELECT user_id, location, timestamp FROM visits WHERE timestamp > NOW() - INTERVAL %s",
        ctx.Value("period").(Period).TcPgInterval(),
    )

    rows, _ := db.Query(query)
    for rows.Next() {
        var value Visit

        _ = rows.Scan(&value.UserID, &value.Location, &value.TS)

        visits = append(visits, value)
    }

    return visits, nil
}

func dayVisitsFromVisits(visits []Visit) []DayVisit {
    out := make([]DayVisit, 0)
    dayVisits := map[string]DayVisit()
    locationCounter := map[string]map[string]int()

    for _, visit := range visits {
        day := visit.TS.Format(time.DateOnly)
        if _, ok := dayVisits[day]; !ok {
            dayVisits[day] = DayVisit{
                Day: day,
            }
            locationCounter[day] = map[string]int()
        }

        if _, ok := locationCounter[day][visit.Location]; !ok {
            locationCounter[day][visit.Location] = 0
        }

        currentDayVisit := dayVisits[day]
        currentDayVisit.Count++
        dayVisits[day] = currentDayVisit

        locationCounter[day][visit.Location]++
    }

    for day, dayLocationCounter := range locationCounter {
        var maxCounter int
        for location, counter := range dayLocationCounter {
            if counter >= maxCounter {
                maxCounter = counter
                currentDayVisit := dayVisits[day]
                currentDayVisit.TopLocation = location
                dayVisits[day] = currentDayVisit
            }
        }
    }

    for _, dayVisit := range dayVisits {
        out = append(out, dayVisit)
    }

    sort.Slice(out, func(i, j int) bool {
        id, _ := time.Parse(time.DateOnly, out[i].Day)
        jd, _ := time.Parse(time.DateOnly, out[j].Day)
        return id.Before(jd)
    })

    return out
}
