package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis"
)

func main() {

	db := ConnectDB()
	cache := ConnectCache()
	producer := ConnectProducer()

	go WatchGameUpdates()

	log.Printf("Connected to database")

	port := os.Getenv("PORT")

	if port == "" {
		panic("PORT must be specified")
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", rootRoute(cache))
	http.HandleFunc("/login", loginRoute(db, cache))
	http.HandleFunc("/events", SocketHandler(db, cache, producer))

	log.Printf("Server started on port %s", port)
	err := http.ListenAndServe(":"+port, nil)

	if err != nil {
		panic(err)
	}
}

func rootRoute(cache *redis.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		userID := getUserID(cache, r)

		if userID == -1 {
			http.Redirect(w, r, "/login", 302)
		} else {
			http.ServeFile(w, r, "templates/home.html")
		}
	}
}

func loginRoute(db *sql.DB, cache *redis.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			// Am I logged in
			http.ServeFile(w, r, "templates/login.html")
		} else if r.Method == "POST" {
			email := r.FormValue("email")
			password := r.FormValue("password")

			id, err := ValidateLogin(db, email, password)

			if err != nil {
				w.Write([]byte("Error " + err.Error()))
			} else {
				sessionID := GenerateSession(cache, id)
				http.SetCookie(w, &http.Cookie{
					Name:    "GameSession",
					Value:   sessionID,
					Expires: time.Now().Add(SessionTime),
				})
				//w.Write([]byte("Authenticated " + sessionID))
				http.Redirect(w, r, "/", 302)
			}
		}
	}
}

func getUserID(cache *redis.Client, r *http.Request) int {
	cookie, err := r.Cookie("GameSession")
	if err != nil || cookie.Value == "" {
		return -1
	}

	userID, err := CheckSession(cache, cookie.Value)
	if err != nil {
		return -1
	}

	return userID
}
