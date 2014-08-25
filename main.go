/*
TODO:
	登陆验证
*/

package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"session" // from https://github.com/mattn/go-session-manager
)

// 文件路径
const (
	Path_Favicon string = "./page/ico/favicon.ico"
	Path_Login   string = "./page/login.html"
	Path_Student string = "./page/student.html"
)

type Page struct {
	Name string
}

// log管理器和session管理器
var logger *log.Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
var session_manager *session.SessionManager = session.NewSessionManager(logger)

// 静态文件请求的统一handle函数
func serveSingle(pattern string, filename string) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("\nURL: %s\nmethod: %s\nAddr: %s\n=========================", r.URL, r.Method, r.RemoteAddr)
		http.ServeFile(w, r, filename)
	})
}

// 登陆操作的handle函数
func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// 登陆页面的GET方法
		logger.Printf("\nURL: %s\nmethod: %s\nAddr: %s\n=========================", r.URL, r.Method, r.RemoteAddr)
		t, _ := template.ParseFiles(Path_Login)
		t.Execute(w, nil)
	} else if r.Method == "POST" {
		// 登陆页面的POST方法
		r.ParseForm()
		logger.Printf("\nURL: %s\nmethod: %s\nAddr: %s\nusername: %s\npassword: %s\n=========================\n", r.URL, r.Method, r.RemoteAddr, r.Form["username"], r.Form["password"])
		//验证用户名密码
		if r.Form["username"][0] == "admin" && r.Form["password"][0] == "admin" {
			// set session
			session_manager.GetSession(w, r).Value = "bbb"
			http.Redirect(w, r, "/student", http.StatusFound)
			return //http.Redirect会执行后面的代码，return保证安全
		} else { // 登陆失败
			t, _ := template.ParseFiles(Path_Login)
			t.Execute(w, nil)
		}
	}
}

// 学生主页面操作的handle函数
func student(w http.ResponseWriter, r *http.Request) {
	logger.Printf("\nURL: %s\nmethod: %s\nAddr: %s\n=========================", r.URL, r.Method, r.RemoteAddr)
	if session_manager.GetSession(w, r).Value == nil {
		// TODO 拒绝服务
		fmt.Println("拒绝服务")
		return
	}
	p := &Page{Name: session_manager.GetSession(w, r).Value.(string)}
	t, _ := template.ParseFiles(Path_Student)
	t.Execute(w, p)
}

func admin(w http.ResponseWriter, r *http.Request) {
	logger.Printf("\nURL: %s\nmethod: %s\nAddr: %s\n=========================", r.URL, r.Method, r.RemoteAddr)
}

func main() {
	fmt.Println("HTTP Server Start\n=========================")
	// session_manager.OnStart(func(session *session.Session) {
	// 	println("started new session")
	// })
	// session_manager.OnEnd(func(session *session.Session) {
	// 	println("abandon")
	// })

	session_manager.SetTimeout(300)

	http.HandleFunc("/", login)
	http.HandleFunc("/student", student)
	http.HandleFunc("/admin", admin)
	serveSingle("/favicon.ico", Path_Favicon)

	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
