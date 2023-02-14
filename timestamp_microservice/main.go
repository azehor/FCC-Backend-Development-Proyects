package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// Requesting /api/____ returns a JSON object with either:
// 1. "Invalid Date" if ____ is not a valid utc date (yyyy-MM-dd) or unix timestamp
// 2. unix: unix_timestamp and utc: utc_date if ____ is a valid utc date or unix timestamp
// 2. unix: unix_timestamp and utc: utc_date for time.Now() if ____ is empty

const index = `
<!DOCTYPE html>
<html>
    <head></head>
    <body>
        <h1>API Project: Timestamp Microservice</h1>
        <h2>Example usage:</h2>
        <p><a href="/api/2020-06-16">[base-url]/api/2020-06-16</a></p>
        <p><a href="/api/1451001600000">[base-url]/api/1451001600000</a></p>
        <br><br>
        <footer>
            <p>by: <a href="http://github.com/azehor">Azehor</a></p>
        </footer>
    </body>
</html>`

type apiHandler struct{}

type JsonResponse struct{
    UnixTimestamp int64 `json:"unix"`
    UTCDate string `json:"utc"`
}

func (apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request){
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    path := r.URL.EscapedPath()[5:]
    var res []byte
    if path == "" {
        res = jsonMarshal(time.Now())
    } else if date, err := time.Parse("2006-01-02", path); err == nil {
        res = jsonMarshal(date)
    } else if unixTime, err := strconv.ParseInt(path, 10, 64); err == nil{
        date := time.Unix(unixTime, 0)
        res = jsonMarshal(date)
    } else if err != nil {
        res, _ = json.Marshal("Invalid Date")
    }
    fmt.Fprintf(w, "%s", res)
}

func jsonMarshal(date time.Time) []byte {
    response := JsonResponse{date.Unix(), date.UTC().Format(http.TimeFormat)}
    res, _ := json.Marshal(response)
    return res
}

func main(){
    mux := http.NewServeMux()
    mux.Handle("/api/", apiHandler{})

    welcomePage := func(w http.ResponseWriter, r *http.Request){
        fmt.Fprintf(w, "%s", index)
    }
    mux.HandleFunc("/", welcomePage)

    http.ListenAndServe(":8080", mux)
}
