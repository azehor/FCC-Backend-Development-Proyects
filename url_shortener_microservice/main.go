package main

import (
    "encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const index = `
<!DOCTYPE html>
<html>
    <head></head>
    <body>
        <h1>API Project: URL Shortener Microservice</h1>
        <form method="POST" action="/api/shorturl/">
            <label for="urlValue">URL:<label>
            <input id="urlValue" name="urlValue" type="text" placeholder="https://www.google.com/">
            <input type="submit">
        </form>
        <h2>Example usage:</h2>
        <p><a href="/api/shorturl/1">[Base URL]/api/shorturl/1</a></p>
        <p>Will redirect to:</p>
        <p>https://www.linkedin.com/in/juan-ignacio-piazza</p>
        <br><br>
        <footer>
            <p>by: <a href="http://github.com/azehor">Azehor</a></p>
        </footer>
    </body>
</html>`

// Requests to /api/shorturl/ can have 3 responses:
// 1. if method == POST and the post form value is a valid URL sends a JSON
//    response with the following format:
//        original_url: "the url sent via POST"
//        short_url: id of the shortened URL (we treat the id as the shortened url)
// 2. if method == POST and the post form value is not a valid URL sends a
//    JSON response: error: "invalid URL"
// 3. if method == GET and path contains a valid ID redirects to the original url
// 4. if method == GET and path contains a valid ID that's not on the DB responds with a JSON containing
//    error: "No short URL found for the given input"
// 5. if method == GET and path contains an invalid ID responds with a JSON containing:
//    error: "Wrong format"

type JsonResponse struct {
    OriginalUrl string `json:"original_url,omitempty"`
    ShortUrl int `json:"short_url,omitempty"`
    Error string `json:"error,omitempty"`
}


type apiHandler struct{
    urlStore map[int]string
}

func (a apiHandler) searchURlbyName(name string) (int, error){
    for k, v := range(a.urlStore) {
        if v == name {
            return k, nil
        }
    }
    return len(a.urlStore)+1, errors.New("url not found")
}

func (a apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request){
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    if r.Method == http.MethodGet {
        id, err := strconv.Atoi(r.URL.EscapedPath()[14:])
        if err != nil {
            res, _ := json.Marshal(JsonResponse{Error: "Wrong Format"})
            fmt.Fprintf(w, "%s", res)
        } else {
            if a.urlStore[id] != "" {
                // Replace with redirect
                http.Redirect(w, r, a.urlStore[id], http.StatusFound)
            } else {
                res, _ := json.Marshal(JsonResponse{Error: "No short URL found for the given input"})
                fmt.Fprintf(w, "%s", res)
            }
        }
    } else if r.Method == http.MethodPost {
        err := r.ParseForm()
        if err != nil {
            res, _ := json.Marshal(JsonResponse{Error: "Error parsing form values"})
            fmt.Fprintf(w, "%s", res)
        }
        originalUrl := r.FormValue("urlValue")
        if originalUrl == "" {
            res, _ := json.Marshal(JsonResponse{Error: "URL to shorten Not Found"})
            fmt.Fprintf(w, "%s", res)
        } else {
            if !IsUrl(originalUrl){
                res, _ := json.Marshal(JsonResponse{Error: "invalid URL"})
                fmt.Fprintf(w, "%s", res)
            } else {
            i, err := a.searchURlbyName(originalUrl)
            if err != nil {
                a.urlStore[i] = originalUrl
            }
            res, _ := json.Marshal(JsonResponse{OriginalUrl: originalUrl, ShortUrl: i})
            fmt.Fprintf(w, "%s", res)
            }
        }
    }   
}

func main(){
    mux := http.NewServeMux()
    mux.Handle("/api/shorturl/", apiHandler{urlStore: map[int]string{1: "https://www.linkedin.com/in/juan-ignacio-piazza/"}})

    welcomePage := func(w http.ResponseWriter, r *http.Request){
        fmt.Fprintf(w, "%s", index)
    }
    mux.HandleFunc("/", welcomePage)

    http.ListenAndServe(":8080", mux)
}

func IsUrl(s string) bool {
    u, err := url.Parse(s)
    return err == nil && u.Scheme != "" && u.Host != ""
}
