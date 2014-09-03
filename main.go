package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
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
	Path_Admin           string = "./page/admin.html"
	Path_Admin_Student   string = "./page/admin-student.html"
	Path_Admin_Grade     string = "./page/admin-grade.html"

	Path_bootstrap_css string = "./page/bootstrap/css/bootstrap.min.css"
	Path_bootstrap_js  string = "./page/bootstrap/js/bootstrap.min.js"
	Path_jquery        string = "./page/bootstrap/jquery.min.js"
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

type temp_Admin struct {
	Name         string
	Find_Project []Project
}

type temp_Admin_Student struct {
	Name         string
	Find_Student []Student
}

type temp_Admin_Grade struct {
	Name           string
	Find_Project   []Project
	Select_Project string
	Find_Grade     []Admin_Grade
	Not_Have_Grade []string
}

type temp_Student_Grade struct {
	Name       string
	Find_Grade []Student_Grade
}

type Project struct {
	Pid        int
	Pname      string
	Full_grade int
}

type Student struct {
	Uid      int
	Name     string
	Password string
}

type Admin_Grade struct {
	Name       string
	Score      int
	Full_Grade int
	Remark     string
}

type Student_Grade struct {
	Name       string
	Pname      string
	Score      int
	Full_Grade int
	Remark     string
}

type Session_struct struct {
	Name       string
	Admin_flag bool
}

var db *sql.DB
var f, err = os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

// log管理器和session管理器
var logger *log.Logger = log.New(f, "", log.Ldate|log.Ltime)
var session_manager *session.SessionManager = session.NewSessionManager(logger)

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

func getMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// 登陆操作的handle函数
func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// 登陆页面的POST方法
		r.ParseForm()
		tempName := r.Form["name"][0]
		tempPassword := getMd5String(r.Form["password"][0])
		logger.Printf("\nURL: %s\nmethod: %s\nAddr: %s\nname: %s\npassword: %s\n=========================\n", r.URL, r.Method, r.RemoteAddr, r.Form["name"], getMd5String(r.Form["password"][0]))

		// 验证是否是管理员
		row := db.QueryRow("select * from admin where aname = '" + tempName + "'")
		var aname string = ""
		var apassword string
		row.Scan(&aname, &apassword)
		if aname != "" && tempPassword == apassword {
			session_manager.GetSession(w, r).Value = Session_struct{Name: tempName, Admin_flag: true}
			http.Redirect(w, r, "/admin", http.StatusFound)
			logger.Printf("Admin Login, ID:%s\n", tempName)
			return //http.Redirect会执行后面的代码，return保证安全
		}

		// 普通用户
		// 验证用户名密码
		row = db.QueryRow("select * from user where name='" + tempName + "'")
		var id int
		var username string
		var password string
		row.Scan(&id, &username, &password)

		if id != 0 { //搜索到用户名
			if tempPassword == password { //登陆成功
				session_manager.GetSession(w, r).Value = Session_struct{Name: tempName, Admin_flag: false}
				http.Redirect(w, r, "/student", http.StatusFound)
				logger.Printf("User Login, ID:%s\n", tempName)
				return //http.Redirect会执行后面的代码，return保证安全
			} else { //密码错误，登陆失败
				t, _ := template.ParseFiles(Path_Login)
				p := &tmp_Login{Id_not_exist: false, Password_error: true}
				t.Execute(w, p)
				logger.Printf("User Login fail, password wrong, ID:%s\n", tempName)
			}
		} else { // 无该用户名，登陆失败
			t, _ := template.ParseFiles(Path_Login)
			p := &tmp_Login{Id_not_exist: true, Password_error: false}
			t.Execute(w, p)
			logger.Printf("User Login fail, ID not exist\n")
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
	logger.Printf("Logout, ID:%s\n", session_manager.GetSession(w, r).Value.(Session_struct).Name)
	session_manager.GetSession(w, r).Value = nil
	http.Redirect(w, r, "/", http.StatusFound)
	return
}

// 学生主页面操作的handle函数
func student(w http.ResponseWriter, r *http.Request) {
	logger.Printf("\nURL: %s\nmethod: %s\nAddr: %s\n=========================", r.URL, r.Method, r.RemoteAddr)
	// session不是student，拒绝服务
	if session_manager.GetSession(w, r).Value == nil || session_manager.GetSession(w, r).Value.(Session_struct).Admin_flag != false {
		logger.Println("session wrong, refuse")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	client_session := session_manager.GetSession(w, r).Value.(Session_struct)
	logger.Printf("Student search grade, ID: %s\n", client_session.Name)

	rows, err := db.Query("select name,pname,score,full_grade,remark from user natural join grade natural join project where name = ?", client_session.Name)
	checkErr(err)
	var rows_grade []Student_Grade
	for i := 0; rows.Next(); i++ {
		rows_grade = append(rows_grade, Student_Grade{"", "", 0, 0, ""})
		rows.Scan(&rows_grade[i].Name, &rows_grade[i].Pname, &rows_grade[i].Score, &rows_grade[i].Full_Grade, &rows_grade[i].Remark)
	}
	p := &temp_Student_Grade{Name: client_session.Name, Find_Grade: rows_grade}
	t, _ := template.ParseFiles(Path_Student)
	t.Execute(w, p)
}

// 修改密码handle函数
func modify_password(w http.ResponseWriter, r *http.Request) {
	logger.Printf("\nURL: %s\nmethod: %s\nAddr: %s\n=========================", r.URL, r.Method, r.RemoteAddr)
	// session不是student，拒绝服务
	if session_manager.GetSession(w, r).Value == nil || session_manager.GetSession(w, r).Value.(Session_struct).Admin_flag != false {
		logger.Println("session wrong, refuse")
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
		logger.Printf("\nURL: %s\nmethod: %s\nAddr: %s\nPassword: %s\nRepeat  : %s\n=========================\n", r.URL, r.Method, r.RemoteAddr, getMd5String(r.Form["password"][0]), getMd5String(r.Form["repeat_password"][0]))

		// 两次密码不一致
		if tempPassword != tempRepeat {
			p := &tmp_Modify_Password{Name: client_session.Name, Not_same: true, Not_form: false, Modify_success: false}
			t, _ := template.ParseFiles(Path_Modify_Password)
			t.Execute(w, p)
			logger.Printf("Student modify password fail, 2 times not same, ID: %s\n", client_session.Name)

			// 密码不符合规范
		} else if len(tempPassword) < 4 || len(tempPassword) > 14 {
			p := &tmp_Modify_Password{Name: client_session.Name, Not_same: false, Not_form: true, Modify_success: false}
			t, _ := template.ParseFiles(Path_Modify_Password)
			t.Execute(w, p)
			logger.Printf("Student modify password fail, not format, ID: %s\n", client_session.Name)

			// 可以修改密码
		} else {
			tempPassword = (getMd5String(tempPassword))
			fmt.Println(tempPassword)
			_, err = db.Exec("update user set password = '" + tempPassword + "' where name= '" + client_session.Name + "'")
			checkErr(err)
			p := &tmp_Modify_Password{Name: client_session.Name, Not_same: false, Not_form: false, Modify_success: true}
			t, _ := template.ParseFiles(Path_Modify_Password)
			t.Execute(w, p)
			logger.Printf("Student modify password success, ID: %s\n", client_session.Name)
		}
	}
}

// admin操作projec页面
func admin(w http.ResponseWriter, r *http.Request) {
	logger.Printf("\nURL: %s\nmethod: %s\nAddr: %s\n=========================", r.URL, r.Method, r.RemoteAddr)
	// session没有记录，拒绝服务
	if session_manager.GetSession(w, r).Value == nil || session_manager.GetSession(w, r).Value.(Session_struct).Admin_flag != true {
		logger.Println("session wrong, refuse")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// GET方法
	if r.Method == "GET" {

		// POST方法
	} else if r.Method == "POST" {
		r.ParseForm()
		delete_pid := r.Form.Get("delete_pid")
		add_full_grade := r.Form.Get("add_full_grade")
		add_pname := r.Form.Get("add_pname")

		// Add操作
		if delete_pid == "" {
			db.Exec("insert into project values(?,?,?)", nil, add_pname, add_full_grade)
			// Delete操作
		} else {
			db.Exec("delete from project where pid = ?", delete_pid)
		}

	}
	rows, err := db.Query("select * from project")
	checkErr(err)
	var rows_project []Project
	for i := 0; rows.Next(); i++ {
		rows_project = append(rows_project, Project{0, "", 0})
		rows.Scan(&rows_project[i].Pid, &rows_project[i].Pname, &rows_project[i].Full_grade)
	}
	p := &temp_Admin{Name: session_manager.GetSession(w, r).Value.(Session_struct).Name, Find_Project: rows_project}
	t, _ := template.ParseFiles(Path_Admin)
	t.Execute(w, p)

}

// admin操作student页面
func admin_student(w http.ResponseWriter, r *http.Request) {
	logger.Printf("\nURL: %s\nmethod: %s\nAddr: %s\n=========================", r.URL, r.Method, r.RemoteAddr)
	// session没有记录，拒绝服务
	if session_manager.GetSession(w, r).Value == nil || session_manager.GetSession(w, r).Value.(Session_struct).Admin_flag != true {
		logger.Println("session wrong, refuse")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// GET方法
	if r.Method == "GET" {

		// POST方法
	} else if r.Method == "POST" {
		r.ParseForm()
		add_name := r.Form.Get("add_name")
		add_password := r.Form.Get("add_password")
		delete_uid := r.Form.Get("delete_uid")
		modify_uid := r.Form.Get("modify_uid")
		modify_password := r.Form.Get("modify_password")

		if add_name != "" {
			db.Exec("insert into user values(?,?,?)", nil, add_name, getMd5String(add_password))
		} else if delete_uid != "" {
			db.Exec("delete from user where uid = ?", delete_uid)
		} else if modify_uid != "" {
			db.Exec("update user set password = ? where uid = ?", getMd5String(modify_password), modify_uid)
		}
	}

	rows, err := db.Query("select * from user")
	checkErr(err)
	var rows_student []Student
	for i := 0; rows.Next(); i++ {
		rows_student = append(rows_student, Student{0, "", ""})
		rows.Scan(&rows_student[i].Uid, &rows_student[i].Name, &rows_student[i].Password)
	}
	p := &temp_Admin_Student{Name: session_manager.GetSession(w, r).Value.(Session_struct).Name, Find_Student: rows_student}
	t, _ := template.ParseFiles(Path_Admin_Student)
	t.Execute(w, p)
}

// Admin打分界面
func admin_grade(w http.ResponseWriter, r *http.Request) {
	logger.Printf("\nURL: %s\nmethod: %s\nAddr: %s\n=========================", r.URL, r.Method, r.RemoteAddr)
	// session没有记录，拒绝服务
	if session_manager.GetSession(w, r).Value == nil || session_manager.GetSession(w, r).Value.(Session_struct).Admin_flag != true {
		logger.Println("session wrong, refuse")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	client_session := session_manager.GetSession(w, r).Value.(Session_struct)
	rows, err := db.Query("select * from project")
	checkErr(err)
	var rows_project []Project
	for i := 0; rows.Next(); i++ {
		rows_project = append(rows_project, Project{0, "", 0})
		rows.Scan(&rows_project[i].Pid, &rows_project[i].Pname, &rows_project[i].Full_grade)
	}

	if r.Method == "GET" {
		p := &temp_Admin_Grade{Name: client_session.Name, Find_Project: rows_project, Select_Project: "select a project"}
		t, _ := template.ParseFiles(Path_Admin_Grade)
		t.Execute(w, p)

	} else if r.Method == "POST" {
		r.ParseForm()
		change_project := r.Form.Get("change_project")
		update_project := r.Form.Get("update_project")
		update_student_number := r.Form.Get("update_student_number")
		update_grade := r.Form.Get("update_grade")
		update_remark := r.Form.Get("update_remark")

		// 修改了select project
		if change_project != "" {
			rows, err = db.Query("select name, score, full_grade, remark from user natural join project natural join grade where pname = ?", change_project)
			checkErr(err)
			var rows_grade []Admin_Grade
			for i := 0; rows.Next(); i++ {
				rows_grade = append(rows_grade, Admin_Grade{"", 0, 0, ""})
				rows.Scan(&rows_grade[i].Name, &rows_grade[i].Score, &rows_grade[i].Full_Grade, &rows_grade[i].Remark)
			}

			rows, err = db.Query("select user.name from user except select user.name from user natural join grade natural join project where pname = ?", change_project)
			checkErr(err)
			var rows_not_have_grade []string
			for i := 0; rows.Next(); i++ {
				rows_not_have_grade = append(rows_not_have_grade, "")
				rows.Scan(&rows_not_have_grade[i])
			}

			p := &temp_Admin_Grade{Name: client_session.Name, Find_Project: rows_project, Select_Project: change_project, Find_Grade: rows_grade, Not_Have_Grade: rows_not_have_grade}
			t, _ := template.ParseFiles(Path_Admin_Grade)
			t.Execute(w, p)

			// 修改分数
		} else {
			row := db.QueryRow("select uid from user where name = ?", update_student_number)
			var ttemp_uid int
			row.Scan(&ttemp_uid)

			row = db.QueryRow("select pid from project where pname = ?", update_project)
			var ttemp_pid int
			row.Scan(&ttemp_pid)

			db.Exec("delete from grade where uid = ? and pid = ?", ttemp_uid, ttemp_pid)
			db.Exec("insert into grade values(?,?,?,?)", ttemp_uid, ttemp_pid, update_grade, update_remark)

			rows, err = db.Query("select name, score, full_grade, remark from user natural join project natural join grade where pname = ?", update_project)
			checkErr(err)
			var rows_grade []Admin_Grade
			for i := 0; rows.Next(); i++ {
				rows_grade = append(rows_grade, Admin_Grade{"", 0, 0, ""})
				rows.Scan(&rows_grade[i].Name, &rows_grade[i].Score, &rows_grade[i].Full_Grade, &rows_grade[i].Remark)
			}

			rows, err = db.Query("select user.name from user except select user.name from user natural join grade natural join project where pname = ?", update_project)
			checkErr(err)
			var rows_not_have_grade []string
			for i := 0; rows.Next(); i++ {
				rows_not_have_grade = append(rows_not_have_grade, "")
				rows.Scan(&rows_not_have_grade[i])
			}

			p := &temp_Admin_Grade{Name: client_session.Name, Find_Project: rows_project, Select_Project: update_project, Find_Grade: rows_grade, Not_Have_Grade: rows_not_have_grade}
			t, _ := template.ParseFiles(Path_Admin_Grade)
			t.Execute(w, p)
		}
	}

}

func main() {
	//连接数据库
	db, err = sql.Open("sqlite3", "./system.db")
	checkErr(err)

	fmt.Println("HTTP Server Start\n=========================")

	session_manager.OnStart(func(session *session.Session) {
		println("started new session")
	})
	session_manager.OnEnd(func(session *session.Session) {
		println("session abandon")
	})

	session_manager.SetTimeout(1200)

	http.HandleFunc("/", login)
	http.HandleFunc("/student", student)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/admin", admin)
	http.HandleFunc("/admin-student", admin_student)
	http.HandleFunc("/admin-grade", admin_grade)
	http.HandleFunc("/modify-password", modify_password)

	serveSingleFile("/favicon.ico", Path_Favicon)
	serveSingleFile("/bootstrap/css/bootstrap.min.css", Path_bootstrap_css)
	serveSingleFile("/bootstrap/js/bootstrap.min.js", Path_bootstrap_js)
	serveSingleFile("/bootstrap/jquery.min.js", Path_jquery)

	err = http.ListenAndServe(":9090", nil)
	checkErr(err)
}
