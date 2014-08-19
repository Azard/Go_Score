package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

const (
	URL_Favicon string = "./static/ico/favicon.ico"
	URL_Login   string = "./static/login.gtpl"
	URL_Client  string = "./static/client.gtpl"
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

func home(w http.ResponseWriter, r *http.Request) {
	console_show_Req(r)
	if r.Method == "GET" {
		t, _ := template.ParseFiles(URL_Login)
		t.Execute(w, nil)
	} else if r.Method == "POST" {
		fmt.Println("username:", r.Form["username"])
		fmt.Println("password:", r.Form["password"])
		t, _ := template.ParseFiles(URL_Client)
		t.Execute(w, nil)
	}
}

func main() {
	http.HandleFunc("/", home)
	serveSingle("/favicon.ico", URL_Favicon)

	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
