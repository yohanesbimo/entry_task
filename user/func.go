package user

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

const WORKER_NUM = 3

var (
	jobs   chan UserProfile
	result chan ResponseMessage
	user   UserProfile
)

func InitWorker() {
	jobs = make(chan UserProfile, 100)
	result = make(chan ResponseMessage, 100)

	for i := 0; i < WORKER_NUM; i++ {
		go tryingToLogin()
	}
}

func renderHTML(w http.ResponseWriter, tmpl *template.Template, data interface{}) {
	err := tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func hashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed Hashing Password")
	}

	return string(hash[:])
}

func Register(w http.ResponseWriter, r *http.Request) {
	/*form := r.URL.Query()

	uname := form.Get("username")
	pwd := form.Get("password")
	name := form.Get("name")*/
	for i := 1000; i < 10000; i++ {
		uname := fmt.Sprintf("yohanesbimo%d", i)
		pwd := "bimo123456"
		name := fmt.Sprintf("Yohanes Bimo%d", i)
		_, err := getUserByUsername(uname)
		if err != nil {
			log.Println("Username already taken")
			return
		} else {
			hash := hashPassword(pwd)

			err := registerUser(uname, hash, name)
			if err != nil {
				log.Println("Failed register user:", err)
			}

			//log.Println("Registration success")
		}
	}
	log.Println("done")
}

func Login(w http.ResponseWriter, r *http.Request) {
	user := isLogin(w, r)
	if user.ID != 0 {
		http.Redirect(w, r, "/profile", http.StatusFound)
	} else {
		tmpl, err := templateView("login")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		var data interface{}

		renderHTML(w, tmpl, data)
	}
}

func tryingToLogin() {
	for j := range jobs {
		log.Println("Jobs Started", j.Username)

		user, err := getUserByUsername(j.Username)
		if err != nil {
			log.Println("Failed retreive user data:", err)
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(j.Password))
		resp := ResponseMessage{}

		if err != nil {
			resp.Status = "NOK"
			resp.Message = "Wrong password, please try again"
		} else {
			now := time.Now().Format(time.RFC850)
			sess, err := bcrypt.GenerateFromPassword([]byte(strconv.FormatInt(user.ID, 10)+now), bcrypt.DefaultCost)
			if err != nil {
				log.Println("Cannot create session:", err)
				resp.Status = "NOK"
				resp.Message = "System error. Please try again"
			} else {
				//expires := 1 * time.Hour
				session := setSessionRedis(strconv.FormatInt(user.ID, 10), string(sess[:]), 1*60*60)
				if session {
					/*cookie := http.Cookie{
						Name:    "session",
						Value:   string(sess[:]),
						Domain:  "localhost",
						Expires: time.Now().Add(expires),
					}
					http.SetCookie(w, &cookie)

					cookie = http.Cookie{
						Name:    "ID",
						Value:   strconv.FormatInt(user.ID, 10),
						Domain:  "localhost",
						Expires: time.Now().Add(expires),
					}
					http.SetCookie(w, &cookie)

					cookie = http.Cookie{
						Name:    "last-login",
						Value:   now,
						Domain:  "localhost",
						Expires: time.Now().Add(expires),
					}
					http.SetCookie(w, &cookie)*/

					resp.Status = "OK"
					resp.Message = "Login Successful"
				} else {
					resp.Status = "NOK"
					resp.Message = "System error. Please try again"
				}
			}
		}

		result <- resp

		log.Println("Jobs Finished")
	}
}

func ActionLogin(w http.ResponseWriter, r *http.Request) {
	user := isLogin(w, r)
	if user.ID != 0 {
		http.Redirect(w, r, "/profile", http.StatusFound)
	} else {
		err := r.ParseForm()
		if err != nil {
			return
		}

		uname := r.Form.Get("username")
		pwd := r.Form.Get("password")

		jobs <- UserProfile{
			Username: uname,
			Password: pwd,
		}

		/*for r := range result {
			r, _ := json.Marshal(r.Status)
			w.Write(r)
		}*/

		r := <-result
		log.Println("status:", r.Status)
		res, err := json.Marshal(r)
		w.Write(res)
	}
}

func isLogin(w http.ResponseWriter, r *http.Request) UserProfile {
	session, sessionErr := r.Cookie("session")
	userID, userIDErr := r.Cookie("ID")
	lastLogin, lastLoginErr := r.Cookie("last-login")

	if sessionErr != nil || userIDErr != nil && lastLoginErr != nil {
		return UserProfile{}
	}

	userSession, err := getSessionFromRedis(userID.Value)
	if err != nil {
		log.Println("Cannot authenticate session from Redis:", err)
		return UserProfile{}
	}

	err = bcrypt.CompareHashAndPassword([]byte(session.Value), []byte(userID.Value+lastLogin.Value))
	if err != nil {
		log.Println("Cannot authenticate", err)
		return UserProfile{}
	}

	if session.Value != userSession {
		log.Println("Session not valid")
		return UserProfile{}
	}

	expires := 1 * time.Hour
	session.Expires = time.Now().Add(expires)
	userID.Expires = time.Now().Add(expires)
	lastLogin.Expires = time.Now().Add(expires)

	http.SetCookie(w, session)
	http.SetCookie(w, userID)
	http.SetCookie(w, lastLogin)

	setSessionRedis(userID.Value, session.Value, 1*60*60)

	user.ID, _ = strconv.ParseInt(userID.Value, 10, 64)
	return user
}

func Profile(w http.ResponseWriter, r *http.Request) {
	user := isLogin(w, r)

	if user.ID == 0 {
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		tmpl, err := templateView("profile")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		user = getUserByID(user.ID)

		renderHTML(w, tmpl, user)
	}
}

func ActionUpdateProfile(w http.ResponseWriter, r *http.Request) {
	res := isLogin(w, r)
	if res.ID == 0 {
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		response := ResponseMessage{}

		err := r.ParseMultipartForm(0)
		if err != nil {
			log.Println(err)
			return
		}

		user.Username = r.FormValue("username")
		user.Name = r.FormValue("name")
		user.Nickname = r.FormValue("nickname")
		file, handler, err := r.FormFile("photo")
		if err != nil {
			log.Println("Unable to get photo")
		}

		if file != nil {
			defer file.Close()
			pwd, _ := os.Getwd()
			filename := user.Username + "-" + handler.Filename
			f, err := os.OpenFile(pwd+"/user/images/"+filename, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				fmt.Println("File not found:", err)

			}
			defer f.Close()
			io.Copy(f, file)

			user.Photo = filename
		}

		err = updateUser(user)
		if err != nil {
			log.Println("Cannot update user data:", err)
			response.Status = "NOK"
			response.Message = "Failed update user data"
		} else {
			response.Status = "OK"
			response.Message = "User data updated"
		}

		json, _ := json.Marshal(response)
		w.Write(json)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	user := isLogin(w, r)

	if user.ID != 0 {
		_, err := removeSessionFromRedis(strconv.FormatInt(user.ID, 10))
		if err != nil {
			log.Println("Failed to logout:", err)
			http.Redirect(w, r, "/profile", http.StatusFound)
		} else {
			http.Redirect(w, r, "/", http.StatusFound)
		}
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func Photo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	pwd, _ := os.Getwd()
	img, err := os.Open(pwd + "/user/images/" + vars["filename"])
	if err != nil {
		log.Println(err)
	}
	defer img.Close()
	w.Header().Set("Content-Type", "image/jpeg")
	io.Copy(w, img)
}
