/*
TODO:
	登陆验证
*/

package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"log"
	"net/http"
	"os"
	"session" // from https://github.com/mattn/go-session-manager
)

// 文件路径
const (
	Path_Favicon         string = "./page/ico/favicon.ico"
	Path_Login           string = "./page/login.html"
	Path_Student         string = "./page/student.html"
	Path_Modify_Password string = "./page/modify-password.html"
	Path_bootstrap_css   string = "./page/bootstrap/css/bootstrap.min.css"
	Path_bootstrap_js    string = "./page/bootstrap/js/bootstrap.min.js"
	Path_jquery          string = "./page/bootstrap/jquery.min.js"
)

// 页面动态数据
// 登陆页面动态数据
type tmp_Login struct {
	Id_not_exist   bool
	Password_error bool
}

// 用户主页动态数据
type tmp_Student struct {
	Name string
}

// 修改密码动态数据
type tmp_Modify_Password struct {
	Name           string
	Not_same       bool
	Not_form       bool
	Modify_success bool
}

type Session_struct struct {
	Name string
}

// log管理器和session管理器
var logger *log.Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
var session_manager *session.SessionManager = session.NewSessionManager(logger)
var db *sql.DB
var err error

// 静态文件请求的统一handle函数
func serveSingleFile(pattern string, filename string) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("\nURL: %s\nmethod: %s\nAddr: %s\n=========================", r.URL, r.Method, r.RemoteAddr)
		http.ServeFile(w, r, filename)
	})
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// 登陆操作的handle函数
func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// 登陆页面的POST方法
		r.ParseForm()
		tempName := r.Form["name"][0]
		tempPassword := r.Form["password"][0]
		logger.Printf("\nURL: %s\nmethod: %s\nAddr: %s\name: %s\npassword: %s\n=========================\n", r.URL, r.Method, r.RemoteAddr, r.Form["name"], r.Form["password"])

		//验证用户名密码
		row := db.QueryRow("select * from user where name='" + tempName + "'")
		var id int
		var username string
		var password string
		row.Scan(&id, &username, &password)

		if id != 0 { //搜索到用户名
			if tempPassword == password { //登陆成功
				session_manager.GetSession(w, r).Value = Session_struct{Name: tempName}
				http.Redirect(w, r, "/student", http.StatusFound)
				return //http.Redirect会执行后面的代码，return保证安全
			} else { //密码错误，登陆失败
				t, _ := template.ParseFiles(Path_Login)
				p := &tmp_Login{Id_not_exist: false, Password_error: true}
				t.Execute(w, p)
			}
		} else { // 无该用户名，登陆失败
			t, _ := template.ParseFiles(Path_Login)
			p := &tmp_Login{Id_not_exist: true, Password_error: false}
			t.Execute(w, p)
		}

		// 登陆页面的GET方法
	} else if r.Method == "GET" {
		logger.Printf("\nURL: %s\nmethod: %s\nAddr: %s\n=========================", r.URL, r.Method, r.RemoteAddr)
		t, _ := template.ParseFiles(Path_Login)
		p := &tmp_Login{Id_not_exist: false, Password_error: false}
		t.Execute(w, p)
	}
}

// 登出操作的handle函数
func logout(w http.ResponseWriter, r *http.Request) {
	logger.Printf("\nURL: %s\nmethod: %s\nAddr: %s\n=========================", r.URL, r.Method, r.RemoteAddr)
	session_manager.GetSession(w, r).Value = nil
	http.Redirect(w, r, "/", http.StatusFound)
	return
}

// 学生主页面操作的handle函数
func student(w http.ResponseWriter, r *http.Request) {
	logger.Printf("\nURL: %s\nmethod: %s\nAddr: %s\n=========================", r.URL, r.Method, r.RemoteAddr)
	// session没有记录，拒绝服务
	if session_manager.GetSession(w, r).Value == nil {
		logger.Println("No session, refuse")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	client_session := session_manager.GetSession(w, r).Value.(Session_struct)
	p := &tmp_Student{Name: client_session.Name}
	t, _ := template.ParseFiles(Path_Student)
	t.Execute(w, p)
}

// 修改密码handle函数
func modify_password(w http.ResponseWriter, r *http.Request) {
	logger.Printf("\nURL: %s\nmethod: %s\nAddr: %s\n=========================", r.URL, r.Method, r.RemoteAddr)
	// session没有记录，拒绝服务
	if session_manager.GetSession(w, r).Value == nil {
		logger.Println("No session, refuse")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// GET方法
	if r.Method == "GET" {
		client_session := session_manager.GetSession(w, r).Value.(Session_struct)
		p := &tmp_Modify_Password{Name: client_session.Name, Not_same: false, Not_form: false, Modify_success: false}
		t, _ := template.ParseFiles(Path_Modify_Password)
		t.Execute(w, p)

		// POST方法
	} else if r.Method == "POST" {
		r.ParseForm()
		client_session := session_manager.GetSession(w, r).Value.(Session_struct)
		tempPassword := r.Form["password"][0]
		tempRepeat := r.Form["repeat_password"][0]
		logger.Printf("\nURL: %s\nmethod: %s\nAddr: %s\nPassword: %s\nRepeat  : %s\n=========================\n", r.URL, r.Method, r.RemoteAddr, r.Form["password"], r.Form["repeat_password"])

		// 两次密码不一致
		if tempPassword != tempRepeat {
			p := &tmp_Modify_Password{Name: client_session.Name, Not_same: true, Not_form: false, Modify_success: false}
			t, _ := template.ParseFiles(Path_Modify_Password)
			t.Execute(w, p)

			// 密码不符合规范
		} else if len(tempPassword) < 4 || len(tempPassword) > 14 {
			p := &tmp_Modify_Password{Name: client_session.Name, Not_same: false, Not_form: true, Modify_success: false}
			t, _ := template.ParseFiles(Path_Modify_Password)
			t.Execute(w, p)

			// 可以修改密码
		} else {
			_, err = db.Exec("update user set password = '" + tempPassword + "' where name= '" + client_session.Name + "'")
			checkErr(err)
			p := &tmp_Modify_Password{Name: client_session.Name, Not_same: false, Not_form: false, Modify_success: true}
			t, _ := template.ParseFiles(Path_Modify_Password)
			t.Execute(w, p)
		}
	}
}

func admin(w http.ResponseWriter, r *http.Request) {
	logger.Printf("\nURL: %s\nmethod: %s\nAddr: %s\n=========================", r.URL, r.Method, r.RemoteAddr)
}

func main() {
	//连接数据库
	db, err = sql.Open("sqlite3", "./system.db")
	checkErr(err)

	//==================================

	/*
		for rows.Next() {
			var id int
			var username string
			var password string
			rows.Scan(&id, &username, &password)
			fmt.Println(id, username, password)
		}
	*/

	/*
		stmt, db_err := db.Prepare("INSERT INTO user(id, name, password) values(?,?,?)")
		checkErr(db_err)

		res, db_err := stmt.Exec(nil, "bbb", "ccc")
		checkErr(db_err)

		id, db_err := res.LastInsertId()
		checkErr(db_err)
		fmt.Println(id)
	*/
	//==================================

	fmt.Println("HTTP Server Start\n=========================")

	session_manager.OnStart(func(session *session.Session) {
		println("started new session")
	})
	// session_manager.OnEnd(func(session *session.Session) {
	// 	println("abandon")
	// })

	session_manager.SetTimeout(300)

	http.HandleFunc("/", login)
	http.HandleFunc("/student", student)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/admin", admin)
	http.HandleFunc("/modify-password", modify_password)

	serveSingleFile("/favicon.ico", Path_Favicon)
	serveSingleFile("/bootstrap/css/bootstrap.min.css", Path_bootstrap_css)
	serveSingleFile("/bootstrap/js/bootstrap.min.js", Path_bootstrap_js)
	serveSingleFile("/bootstrap/jquery.min.js", Path_jquery)

	err = http.ListenAndServe(":9090", nil)
	checkErr(err)
}
