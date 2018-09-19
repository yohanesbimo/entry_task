package user

import (
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/gomodule/redigo/redis"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

var (
	mock sqlmock.Sqlmock
	s    *miniredis.Miniredis
)

func mockDBConn(t *testing.T) {
	db, mock, err = sqlmock.New()
	if err != nil {
		t.Errorf("Cannot mock db")
	}

	mock.ExpectPrepare("INSERT INTO profile")
	mock.ExpectPrepare("SELECT id, username, password FROM profile")
	mock.ExpectPrepare("SELECT id, username, name, nickname, photo FROM profile")
	mock.ExpectPrepare("UPDATE profile")
	prepareStmt()
}

func mockRedisConn() {
	s, err = miniredis.Run()
	if err != nil {
		panic(err)
	}
	//defer s.Close()

	redisDB, err = redis.Dial("tcp", s.Addr())
}

func TestRegisterUser(t *testing.T) {
	mockDBConn(t)
	defer db.Close()

	mock.ExpectExec("INSERT INTO profile").WithArgs("shopee", "shopee123", "shopee").WillReturnResult(sqlmock.NewResult(1, 1))

	if err := registerUser("shopee", "shopee123", "shopee"); err != nil {
		t.Errorf("Register user was not expected: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %s", err)
	}
}

func TestGetUserByUsername(t *testing.T) {
	mockDBConn(t)
	defer db.Close()

	columns := []string{"id", "username", "password"}
	rs := sqlmock.NewRows(columns)
	rs.AddRow("1", "shopee", "shopee123")

	mock.ExpectQuery("SELECT id, username, password FROM profile").WithArgs("shopee").WillReturnRows(rs)

	if _, err := getUserByUsername("shopee"); err != nil {
		t.Errorf("Incorrect user data")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %s", err)
	}
}

func TestGetUserByID(t *testing.T) {
	mockDBConn(t)
	defer db.Close()

	columns := []string{"id", "username", "name", "nickname", "photo"}
	rs := sqlmock.NewRows(columns)
	rs.FromCSVString("1, shopee, shopee, shopee, shopee.jpg")

	mock.ExpectQuery("SELECT id, username, name, nickname, photo FROM profile").WithArgs(1).WillReturnRows(rs)

	if user := getUserByID(1); user.ID == 0 {
		t.Errorf("Incorrect user data")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %s", err)
	}
}

func TestUpdateUser(t *testing.T) {
	mockDBConn(t)
	defer db.Close()

	columns := []string{"id", "username", "name", "nickname", "photo"}
	rs := sqlmock.NewRows(columns)
	rs.AddRow("1", "shopee", "shopee", "shopee", "shopee.jpg")

	user := UserProfile{
		ID:       1,
		Username: "shopee",
		Name:     "shopee",
		Nickname: "Shopee",
		Photo:    "Shopee.png",
	}

	mock.ExpectExec("UPDATE profile").WithArgs(user.Nickname, user.Photo, user.ID).WillReturnResult(sqlmock.NewResult(1, 1))

	if err := updateUser(user); err != nil {
		t.Errorf("Incorrect user data")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %s", err)
	}
}

func TestSetSessionRedis(t *testing.T) {
	mockRedisConn()
	defer redisDB.Close()

	if res := setSessionRedis("shopee", "shopee123", 1*60); res == false {
		t.Errorf("Incorrect result, got %t want true", res)
	}
}

func TestGetSessionFromRedis(t *testing.T) {
	mockRedisConn()
	defer redisDB.Close()

	key := "shopee"
	expected := "shopee123"

	s.Set(key, expected)
	s.SetTTL(key, 1*10)

	if res, _ := getSessionFromRedis("shopee"); res != expected {
		t.Errorf("Incorrect session, got %s, want %s", res, expected)
	}
}

func TestRemoveSessionFromRedis(t *testing.T) {
	mockRedisConn()
	defer redisDB.Close()

	key := "shopee"
	value := "shopee123"

	s.Set(key, value)

	if _, err := removeSessionFromRedis(key); err != nil {
		t.Errorf("Unable remove session, got %v, want %v", err, nil)
	}
}
