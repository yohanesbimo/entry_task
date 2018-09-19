package main

import (
	u "entry_task/user"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	u.Init()

	r := mux.NewRouter()

	r.HandleFunc("/", u.Login)
	r.HandleFunc("/action-login", u.ActionLogin)

	r.HandleFunc("/profile", u.Profile)
	r.HandleFunc("/action-update-profile", u.ActionUpdateProfile)

	r.HandleFunc("/register", u.Register)

	r.HandleFunc("/logout", u.Logout)

	r.HandleFunc("/photo/{filename}", u.Photo)

	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}
}
