package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// TODO: Implement an actual SQL Database instead of mock data

// API endpoints:
// /api/users :
//     - GET: returns a JSON array containing all users and their respective id
//     - POST: adds a new username and generates their id, returns a JSON containing username and generated id
// /api/users/:_id/exercises :
//     - POST: adds a new exercise to the user with id :_id, must contain description and duration, and an optional date
//             if no date is provided, uses the current date
// /api/users/:_id/logs :
//     - GET: returns a full exercise log of the user, accepts from, to and limit parameters
//            from and to are dates in yyyy-mm-dd format
//            limit is an integer of how many logs to return (if there's more than ?limit amount of logs, return the latest ones)

// JSON formats:
// Exercise:
// {
//     username: string,
//     description: string,
//     duration: int,
//     date: string, (format: "Mon Jan 01 1990")
//     _id: string (format: 24 digit hex number)
// }
//
// User:
// {
//     username: string,
//     _id: string
// }
//
// Log:
// {
//     username: string,
//     count: int,
//     _id: string,
//     from: string, (format: date) 'json:omitempty'
//     to: string, (format: date) 'json:omitempty'
//     log: [{
//         description: string,
//         duration: int,
//         date: string (aforementioned format)
//     }]  (Ordered by date descending)
// }

const welcomePage = `
<!DOCTYPE html>
<html>
    <head></head>
    <body>
        <h1>API Project: Exercise Tracker REST API</h1>
        <form method="POST" action="/api/users">
            <h3>POST /api/users</h3>
            <input id="username" type="text" name="username" placeholder="username"> 
            <input type="submit">
        </form>
        <br><br>
        <form method="POST" action="/">
            <h3>POST /api/users/:_id/exercises</h3>
            <input id="uid" type="text" name="userID" placeholder="userID"> 
            <input id="desc" type="text" name="description" placeholder="description"> 
            <input id="dur" type="text" name="duration" placeholder="duration* (mins.)"> 
            <input id="date" type="text" name="date" placeholder="Date (yyyy-mm-dd)"> 
            <input type="submit">
        </form>
        <br><br>
        <footer>
        <p>by: <a href="http://github.com/azehor">Azehor</a></p>
        </footer>
    </body>
</html>`

const DateFormat = "Mon Jan 02 2006"
const ymdDateFormat = "2006-01-02"

type UserType struct{
    UserID string `json:"_id"`
    Username string `json:"username"`
}

type Exercise struct{
    UserID string `json:"_id,omitempty"`
    Username string `json:"username,omitempty"`
    Description string `json:"description"`
    Duration int `json:"duration"`
    Date string `json:"date"`
}

type ExerciseLog struct {
    UserID string `json:"_id"`
    Username string `json:"username"`
    From string `json:"from,omitempty"`
    To string `json:"to,omitempty"`
    Count int `json:"count"`
    ExerciseLog []Exercise `json:"log"`
}

var users = []*UserType{
    {Username: "Juan", UserID: "0f0f"},
    {Username: "Julia", UserID: "ea63"},
}

var exercises = []*Exercise{
    {users[0].UserID, users[0].Username, "Jogging", 30, "Mon Feb 02 2023"},
    {users[1].UserID, users[1].Username, "Bench Press", 15, "Thu Jan 24 2020"},
    {users[1].UserID, users[1].Username, "Running", 20, "Thu Jan 24 2022"},
}

func main(){
    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Get("/", func(w http.ResponseWriter, r *http.Request){
        w.Write([]byte(welcomePage))
    })
    r.Post("/", validateExerciseInput)
    r.Route("/api/users", func(r chi.Router) {
        r.Get("/", listUsers)
        r.Post("/", createUser)

        r.Post("/{userID}/exercises", createExercise)

        r.Get("/{userID}/logs", listExercises)
    })


    http.ListenAndServe(":3000", r)
}

func listUsers(w http.ResponseWriter, r *http.Request) {
    w = setJsonHeader(w)
    res, _ := json.Marshal(users)
    w.Write(res)
}

func createUser(w http.ResponseWriter, r *http.Request) {
    u := &UserType{}
    name := r.FormValue("username")
    if name == "" {
        errInternal(w)
        return
    } else {
        // This block of assignations should be moved to 
        // func (u *UserType) Bind() error
        u.Username = name
        u.UserID = generateNewId()
        dbNewUser(u)
        res, _ := json.Marshal(u)
        w = setJsonHeader(w)
        w.Write(res)
    }
}

func createExercise(w http.ResponseWriter, r *http.Request) {
    e := &Exercise{}
    id := r.FormValue("userID")
    u, _ := dbGetUserById(id)
    name := u.Username
    description := r.FormValue("description")
    dur := r.FormValue("duration")
    duration, _ := strconv.Atoi(dur)
    date := formatDate(r.FormValue("date"))
    if date == "" {
        date  = time.Now().Format(DateFormat)
    }
    // This block of assignations should be moved to 
    // func (e *Exercise) Bind() error
    e.UserID = id
    e.Username = name
    e.Description = description
    e.Duration = duration
    e.Date = date
    dbNewExercise(e)
    w = setJsonHeader(w)
    res, _ := json.Marshal(e)
    w.Write(res)
}

func listExercises(w http.ResponseWriter, r *http.Request) {
    var l ExerciseLog
    from := r.URL.Query().Get("from")
    to := r.URL.Query().Get("to")
    limit := r.URL.Query().Get("limit")
    ex, err := dbGetExercises(chi.URLParam(r, "userID"), from, to, limit)
    if err != nil {
        errInternal(w) //only on user not found
        return
    }
    w = setJsonHeader(w)

    if len(ex) > 0 {
        l = ExerciseLog{
            UserID: ex[0].UserID,
            Username : ex[0].Username,
            From: formatDate(from),
            To: formatDate(to),
        }
        for _, e := range(ex){
            l.Count++
            e.UserID = ""
            e.Username = ""
            l.ExerciseLog = append(l.ExerciseLog, e)
        }
    } else {
        u, _ := dbGetUserById(chi.URLParam(r, "userID"))
        l = ExerciseLog{
            UserID: u.UserID,
            Username: u.Username,
            From: formatDate(from),
            To: formatDate(to),
            Count: 0,
        }
    }
    res, _ := json.Marshal(l)
    w.Write(res)
}

func validateExerciseInput(w http.ResponseWriter, r *http.Request) {
    var err error
    id := r.FormValue("userID")
    description := r.FormValue("description")
    duration := r.FormValue("duration")
    if description == "" || duration == "" {
        //Validation Error
        errInternal(w)
        return
    }
    if _, err = dbGetUserById(id); err != nil {
        //Validation error
        errNotFound(w)
        return
    }
    if _, err = strconv.Atoi(duration); err != nil {
        errInternal(w)
        return
    }
    http.Redirect(w, r, fmt.Sprintf("/api/users/%s/exercises", id), http.StatusPermanentRedirect)
}

func setJsonHeader(w http.ResponseWriter) http.ResponseWriter {
    w.Header().Set("Content-Type", "application/json, charset=UTF-8")
    return w
}

func formatDate(date string) string {
    if date == "" {
        return ""
    }
    t, err := time.Parse(ymdDateFormat, date)
    if err != nil {
        return ""
    }
    res := t.Format(DateFormat)
    return res
}

func generateNewId() string {
    src := rand.New(rand.NewSource(time.Now().UnixNano()))
    b := make([]byte, 12)
    if _, err := src.Read(b); err != nil {
        panic(err)
    }
    return hex.EncodeToString(b)[:24]
}

func errInternal(w http.ResponseWriter) {
    http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func errNotFound(w http.ResponseWriter) {
    http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func dbGetUserById(id string) (*UserType, error) {
    for _, u := range(users) {
        if u.UserID == id {
            return u, nil
        }
    }
    return nil, errors.New("Not Found")
}

func dbNewUser(newUser *UserType) {
    _, err := dbGetUserById(newUser.UserID)
    for err == nil {
        newUser.UserID = generateNewId()
        _, err = dbGetUserById(newUser.UserID)
    }
    users = append(users, newUser)
}

func dbNewExercise(newExercise *Exercise) {
    exercises = append(exercises, newExercise)
}

func dbGetExercises(uid string, from string, to string, limit string) ([]Exercise, error){
    if _, err := dbGetUserById(uid); err != nil {
        return nil, err 
    }
    var err error
    var res []Exercise
    var formattedFrom time.Time
    var formattedTo time.Time
    var lim int
    if lim, err = strconv.Atoi(limit); err != nil {
        lim = math.MaxInt
    }
    formattedFrom, err = time.Parse(ymdDateFormat, from)
    formattedTo, err = time.Parse(ymdDateFormat, to)
    if formattedTo.IsZero(){
        formattedTo = time.Unix(1<<63-62135596801, 999999999) //Max time.Time useful for comparisons
    }
    for _, e := range(exercises) {
        if e.UserID == uid {
            if len(res) < lim {
                d, _ := time.Parse(DateFormat, e.Date)
                if d.After(formattedFrom) && d.Before(formattedTo) {
                    res = append(res, *e)
                }
            }
        }
    }
    return res, nil
}

