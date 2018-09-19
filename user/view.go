package user

import (
	"html/template"
	"os"
)

func viewPath(view string) string {
	pwd, _ := os.Getwd()
	return pwd + "/user/html/" + view + ".html"
}

func templateView(page string)(*template.Template, error) {
	tmpl, err := template.ParseFiles(viewPath(page))
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}