package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
)

const index = `
<!DOCTYPE html>
<html>
    <head></head>
    <body>
        <h1>API Project: File Metadata Microservice</h1>
        <form method="POST" enctype="multipart/form-data" action="/api/fileanalyse">
            <label for="uploadFile">Select a file:</label>
            <input type="file" id="uploadFile" name="uploadFile"><br><br>
            <input type="submit">
        </form>
        <p>by <a href="github.com/azehor">Azehor</a></p>
    </body>
</html>`

func main(){
    welcomePage := func(w http.ResponseWriter, r *http.Request){
        fmt.Fprintf(w, "%s", index)
    }
    mux := http.NewServeMux()
    mux.HandleFunc("/", welcomePage)
    mux.HandleFunc("/api/fileanalyse", file_metadata_analyse)

    http.ListenAndServe(":8080", mux)
}

type FileMetadata struct{
    Name string `json:"name"`
    Type string `json:"type"`
    Size int64 `json:"size"`
}

func file_metadata_analyse(w http.ResponseWriter, r *http.Request){
    err := r.ParseMultipartForm(math.MaxInt32)
    if err != nil {
        panic(err)
    }
    fileHeader := r.MultipartForm.File["uploadFile"]
    if len(fileHeader) == 0 {
        fmt.Fprintln(w, "An error ocurred with the file upload, please try again")
    } else {
        header := fileHeader[0]
        w.Header().Set("Content-Type", "application/json; charset=utf-8")
        metadata := FileMetadata{header.Filename, header.Header.Get("Content-Type"), header.Size}
        data, err := json.Marshal(metadata)
        if err != nil {
            panic(err)
        }
        fmt.Fprintf(w, "%s", data)
    }
}
