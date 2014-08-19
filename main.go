package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

const (
	Path_Favicon string = "./static/ico/favicon.ico"
	Path_Login   string = "./static/login.gtpl"
	Path_Student string = "./static/student.gtpl"
)

func console_show_Req(r *http.Request) {
	r.ParseForm()
	fmt.Println("=========================")
	fmt.Println("URL:", r.URL)
	fmt.Println("method:", r.Method)
}

func serveSingle(pattern string, filename string) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		console_show_Req(r)
		http.ServeFile(w, r, filename)
	})
}

func login(w http.ResponseWriter, r *http.Request) {
	console_show_Req(r)
	if r.Method == "GET" {
		t, _ := template.ParseFiles(Path_Login)
		t.Execute(w, nil)
	} else if r.Method == "POST" {
		fmt.Println("username:", r.Form["username"])
		fmt.Println("password:", r.Form["password"])
		//验证用户名密码
		if r.Form["username"][0] == "admin" && r.Form["password"][0] == "admin" {
			http.Redirect(w, r, "/student", http.StatusFound)
			return //http.Redirect会执行后面的代码，return保证安全
		} else { // 登陆失败
			t, _ := template.ParseFiles(Path_Login)
			t.Execute(w, nil)
		}
	}
}

func student(w http.ResponseWriter, r *http.Request) {
	console_show_Req(r)
	t, _ := template.ParseFiles(Path_Student)
	t.Execute(w, nil)
}

func admin(w http.ResponseWriter, r *http.Request) {
	console_show_Req(r)
}

func main() {
	http.HandleFunc("/", login)
	http.HandleFunc("/student", student)
	http.HandleFunc("/admin", admin)
	serveSingle("/favicon.ico", Path_Favicon)

	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
