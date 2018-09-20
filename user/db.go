package user

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
)

const (
	HOST      = "127.0.0.1"
	PORT      = 3306
	USER      = "entry_task"
	PASSWORD  = "user12345"
	DATABASE  = "user"
	REDISHOST = "localhost:6379"
)

var (
	redisDB redis.Conn

	db  *sql.DB
	err error

	stmtRegister   *sql.Stmt
	stmtUpdateUser *sql.Stmt

	stmtGetUserByUsername *sql.Stmt
	stmtGetUserByID       *sql.Stmt
)

func Init() {
	dialRedis()
	connect()
	prepareStmt()
}

func dialRedis() {
	redisDB, err = redis.Dial("tcp", REDISHOST)
	if err != nil {
		log.Fatal("Cannot connect redis:", err)
	}
}

func connect() {
	mysql := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8",
		USER, PASSWORD, HOST, PORT, DATABASE)

	db, err = sql.Open("mysql", mysql)
	if err != nil {
		log.Fatal("Failed to connect DB:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to connect DB:", err)
	}

	//db.SetMaxIdleConns(100)
	db.SetMaxOpenConns(1024)
}

func prepareStmt() {
	var err error = nil

	stmtRegister, err = db.Prepare("INSERT INTO profile (username, password, name) VALUES( ?, ?, ?) --")
	if err != nil {
		log.Fatal("Cannot prepare query:", err)
	}

	stmtGetUserByUsername, err = db.Prepare("SELECT id, username, password FROM profile WHERE username=? --")
	if err != nil {
		log.Fatal("Cannot prepare query:", err)
	}

	stmtGetUserByID, err = db.Prepare("SELECT id, username, name, nickname, photo FROM profile WHERE id=? --")
	if err != nil {
		log.Fatal("Cannot prepare query:", err)
	}

	stmtUpdateUser, err = db.Prepare("UPDATE profile SET nickname=?, photo=? WHERE id=? --")
	if err != nil {
		log.Fatal("Cannot prepare query:", err)
	}
}

func getUserByUsername(username string) (user UserProfile, err error) {
	res, err := stmtGetUserByUsername.Query(username)
	if err != nil {
		log.Println("Failed to check username:", err)
		return user, err
	}

	defer res.Close()

	if res.Next() {
		if err := res.Scan(&user.ID, &user.Username, &user.Password); err != nil {
			log.Println("Failed read user data from DB:", err)
			return user, err
		}
	}

	return user, err
}

func getUserByID(userID int64) (user UserProfile) {
	res, err := stmtGetUserByID.Query(userID)
	if err != nil {
		log.Println("Failed to check user ID:", err)
		return user
	}

	defer res.Close()

	for res.Next() {
		if err := res.Scan(&user.ID, &user.Username, &user.Name, &user.Nickname, &user.Photo); err != nil {
			log.Fatal("Failed read user data from DB:", err)
		}
	}

	return user
}

func updateUser(user UserProfile) (err error) {
	_, err = stmtUpdateUser.Exec(user.Nickname, user.Photo, user.ID)
	if err != nil {
		log.Fatal("Failed update user on DB:", err)
	}

	return err
}

func registerUser(username string, password string, name string) error {
	_, err := stmtRegister.Exec(username, password, name)
	if err != nil {
		log.Fatal("Failed register user on DB:", err)
	}

	return err
}

func setSessionRedis(key string, value string, expires int) bool {
	_, err := redisDB.Do("SET", key, value)
	if err != nil {
		log.Fatal("Cannot set redis value:", err)
		return false
	}

	_, err = redisDB.Do("EXPIRE", key, expires)
	if err != nil {
		log.Fatal("Cannot set redis expires:", err)
		return false
	}

	return true
}

func getSessionFromRedis(userID string) (string, error) {
	result, err := redis.Bytes(redisDB.Do("GET", userID))
	if err != nil {
		log.Println("Cannot read from redis:", err)
		return "", nil
	}

	return string(result), nil
}

func removeSessionFromRedis(userID string) (int64, error) {
	result, err := redis.Int64(redisDB.Do("DEL", userID))
	if err != nil {
		log.Println("Cannot remove session from redis:", err)
		return 0, err
	}
	return result, err
}
