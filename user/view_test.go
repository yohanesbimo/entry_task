package user

import (
	"os"
	"testing"
)

type Test struct {
	Input    interface{}
	Expected interface{}
}

func TestViewPath(t *testing.T) {
	pwd, _ := os.Getwd()
	tests := []Test{
		{
			Input:    "login",
			Expected: pwd + "/user/html/login.html",
		}, {
			Input:    "profile",
			Expected: pwd + "/user/html/profile.html",
		},
	}

	for _, test := range tests {
		result := viewPath(test.Input.(string))
		if result != test.Expected.(string) {
			t.Errorf("Incorrect view path, got %s, want %s", result, test.Expected.(string))
		}
	}
}
