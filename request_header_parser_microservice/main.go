package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const welcomeHTML = `
<html>
    <head></head>
    <body>
        <h1>API Project: Request Header Parser Microservice</h1>
        <p>To use the service go to: <a href="/api/whoami">[base-url]/api/whoami</a></p>
        <br><br>
        <footer>
            <p>by: <a href="http://github.com/azehor">Azehor</a></p>
        </footer>
    </body>
</html>`

// routing to /api/whoami returns
// 1. ipaddress
// 2. language
// 3. software used to make the request

type Connection struct {
    Ipaddress string `json:"ipaddress"` 
    Language string `json:"language"` 
    Software string `json:"software"` 
}

func main() {
    welcome := func(w http.ResponseWriter, r *http.Request){
        fmt.Fprintf(w, "%s", welcomeHTML)
    }

    mux := http.NewServeMux()
    mux.HandleFunc("/", welcome)
    mux.HandleFunc("/api/whoami", request_header_parser)

    http.ListenAndServe(":8080", mux)
}

func request_header_parser(w http.ResponseWriter, r *http.Request){
        w.Header().Set("Content-Type", "application/json; charset=utf-8")
        var ip string
        ips := r.Header.Get("X-Forwarded-For")
        if ips != "" {
            fmt.Println(len(ips))
            ip = strings.Split(ips, ", ")[0]
        } else {
            ip = strings.Split(r.RemoteAddr, ":")[0]
        }
        conn := Connection{ip, r.Header.Get("Accept-Language"), r.Header.Get("User-Agent")}
        data, _ := json.Marshal(conn)
        fmt.Fprintf(w, "%s", data)
}
