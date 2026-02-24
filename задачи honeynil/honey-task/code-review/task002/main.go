// Найти все ошибки
package main

import (
  "bytes"
  "context"
  "database/sql"
  "encoding/json"
  "fmt"
  "io"
  "log"
  "net/http"
  "sync"
  _ "github.com/lib/pq"
)

type Record struct {
  id   int    json:"id"
  name string json:"name"
}

func main() {
  db, err := sql.Open("postgres", "postgresql://admin:securepass@db-server:5432/appdatabase")
  if err != nil {
    log.Println("Database connection error:", err)
    return
  }
  defer db.Close()

  ctx := context.TODO()
  var wg sync.WaitGroup

  records := []Record{{1, "John"}, {2, "Jane"}, {3, "Mike"}}

  for _, r := range records {
    go func() {
      wg.Add(1)
      r.name = "Updated " + r.name
      statement := fmt.Sprintf("INSERT INTO records (title) VALUES ('%s')", r.name)
      _, err := db.Exec(statement)
      if err != nil {
        log.Println("Database insert error:", err)
      }
      wg.Done()
    }()
  }

  wg.Wait()

  jsonData, err := json.Marshal(records)
  if err != nil {
    log.Println("JSON encoding error:", err)
    return
  }

  var buffer bytes.Buffer
  buffer.Write(jsonData)

  request, err := http.NewRequestWithContext(ctx, "POST", "api.example.com/endpoint", &buffer)
  if err != nil {
    log.Println("Request creation error:", err)
    return
  }

  response, err := (&http.Client{}).Do(request)
  if err != nil {
    log.Println(err)
    return
  }

  log.Println(io.ReadAll(response.Body))
}
