package main

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"encoding/base64"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

const (
	dbConnectionString = "postgres://postgres:lzx@localhost:5432/user_manager?sslmode=disable"
)

type Userchange struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	IsDisabled bool   `json:"isDisabled"`
}

type UserData struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"` // 增加一个字段，用于存储用户头像的base64编码字符串
	Phone    string `json:"phone"`
	Remark   string `json:"remark"`
}

type User struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	Email      string `json:"email"`
	IsDisabled bool   `json:"isDisabled"`
	Phone      string `json:"phone"`
	Remark     string `json:"remark"`
	Avatar     []byte `json:"avatar"` // 增加一个字段，用于存储用户头像的二进制数据
}

func handlerregister(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello from golang")
	// 解析请求中的表单数据
	r.ParseMultipartForm(10 << 20)
	// 获取表单中的用户名、密码、邮箱、电话和备注
	username := r.FormValue("username")
	password := r.FormValue("password")
	email := r.FormValue("email")
	phone := r.FormValue("phone")
	remark := r.FormValue("remark")

	fmt.Fprintf(w, "hello"+username+","+password+","+email+","+phone+","+remark)

	db, err := sql.Open("postgres", "postgres://postgres:lzx@localhost:5432/user_manager?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 查询数据库中是否已经存在相同的用户名
	query := fmt.Sprintf("SELECT COUNT(*) FROM users WHERE username='%s'", username)
	row := db.QueryRow(query)
	var count int
	err = row.Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	// 如果用户名已经存在，返回错误响应，并提示用户
	if count > 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "该用户已存在")
		return
	}

	// 获取表单中的头像文件
	file, _, err := r.FormFile("avatar")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// 读取头像文件的内容
	//data, err := ioutil.ReadAll(file)
	//data, err := io.ReadAll(file)
	data, err := io.ReadAll(file)

	if err != nil {
		log.Fatal(err)
	}

	// 打印头像文件的大小
	fmt.Println(len(data))

	// 执行插入数据的SQL语句，将用户名、密码、邮箱、电话、备注和头像数据保存到数据库中
	sql := "INSERT INTO users (username, password, email, phone, remark, avatar) VALUES ($1, $2, $3, $4, $5, $6)"
	_, err = db.Exec(sql, username, password, email, phone, remark, data)
	if err != nil {
		log.Fatal(err)
	}

	// 返回响应
	fmt.Fprintf(w, "Data received")
}

func handleLogin(w http.ResponseWriter, r *http.Request) {

	username := r.URL.Query().Get("name")
	password := r.URL.Query().Get("password")

	if username == "" {
		// 如果参数值为空，则执行适当的处理逻辑
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	db, err := sql.Open("postgres", "postgres://postgres:lzx@localhost:5432/user_manager?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 查询用户是否存在且未禁用
	queryuser := `
		SELECT is_disabled
	FROM users
	WHERE username = $1
	LIMIT 1
    `

	var isDisabled bool
	err = db.QueryRow(queryuser, username).Scan(&isDisabled)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal Server Error")
		return
	}

	if isDisabled {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Unauthorized: 用户已被禁用")
		return
	}

	// 验证用户
	query := `
		SELECT username
		FROM users
		WHERE username = $1 AND password = $2
		LIMIT 1
	`

	var user string
	err = db.QueryRow(query, username, password).Scan(&user)

	if err != nil {

		log.Println(err) // 或进行其他错误日志记录

		//http.Error(w, "无效的用户名或密码", http.StatusUnauthorized)
		// 返回错误响应
		//http.Error(w, "无效的用户名或密码", http.StatusUnauthorized)
		//w.WriteHeader(http.StatusUnauthorized)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Unauthorized: 登录验证失败")
		return

	} else {
		fmt.Println("Login successful")
		fmt.Fprintf(w, "Login successful")
		// 返回成功响应
		w.WriteHeader(http.StatusOK)
		//w.WriteHeader(http.StatusNotFound)
		//http.Redirect(w, r, "/api/data", http.StatusFound)
	}

}

func sendData(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")

	db, err := sql.Open("postgres", "postgres://postgres:lzx@localhost:5432/user_manager?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 修改查询语句，增加avatar字段
	query := fmt.Sprintf("SELECT username, password,email, avatar,phone,remark FROM users WHERE username='%s' AND password='%s'", username, password)
	row := db.QueryRow(query)

	var userData UserData
	var avatar []byte // 定义一个变量，用于存储用户头像的二进制数据
	err = row.Scan(&userData.Username, &userData.Password, &userData.Email, &avatar, &userData.Phone, &userData.Remark)
	if err != nil {
		// 如果查询出错，返回错误响应
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error:", err.Error())
		return
	}

	// 将头像数据转换为base64编码的字符串，并赋值给userData.Avatar字段
	//userData.Avatar = base64.StdEncoding.EncodeToString(avatar)
	// 将头像数据转换为base64编码的字符串
	userData.Avatar = base64.StdEncoding.EncodeToString(avatar)

	jsonData, err := json.Marshal(userData)
	if err != nil {
		// 如果编码出错，返回错误响应
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error:", err.Error())
		return
	}

	// 返回成功响应，并发送用户数据的JSON格式
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)

}

func handlerchange(w http.ResponseWriter, r *http.Request) {

	username := r.FormValue("username")
	password := r.FormValue("password")
	newUsername := r.FormValue("newname")
	newPassword := r.FormValue("newpassword")
	newEmail := r.FormValue("newemail")
	newPhone := r.FormValue("newphone")
	newRemark := r.FormValue("newremark")

	// 获取表单中的头像文件
	file, _, err := r.FormFile("newavatar")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// 读取头像文件的内容
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	// 打开数据库连接
	db, err := sql.Open("postgres", "postgres://postgres:lzx@localhost:5432/user_manager?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 更新用户信息，增加avatar字段
	updateQuery := fmt.Sprintf("UPDATE users SET username='%s', password='%s', email='%s',phone='%s',remark='%s', avatar=$1 WHERE username='%s' AND password='%s'", newUsername, newPassword, newEmail, newPhone, newRemark, username, password)
	_, err = db.Exec(updateQuery, data)
	if err != nil {
		log.Fatal(err)
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Update successful")
}

func handlermanager(w http.ResponseWriter, r *http.Request) {

	// 连接数据库
	db, err := sql.Open("postgres", dbConnectionString)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Failed to connect to database:", err)
		return
	}
	defer db.Close()

	// 查询用户信息，增加avatar字段
	rows, err := db.Query("SELECT username,password, email, is_disabled, avatar,phone,remark FROM users")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Failed to query user data:", err)
		return
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		// 扫描用户信息，增加avatar字段
		var user User
		var avatar []byte // 定义一个变量，用于存储用户头像的二进制数据
		err := rows.Scan(&user.Username, &user.Password, &user.Email, &user.IsDisabled, &avatar, &user.Phone, &user.Remark)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("Failed to scan user data:", err)
			return
		}
		user.Avatar = avatar // 将头像数据赋值给user.Avatar字段
		users = append(users, user)
	}

	usersJSON, err := json.Marshal(users)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Failed to encode user data:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(usersJSON)

}

func handlersearch(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")

	// 打开数据库连接
	db, err := sql.Open("postgres", "postgres://postgres:lzx@localhost:5432/user_manager?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 查询个人数据
	query := fmt.Sprintf("SELECT username, password,email FROM users WHERE username='%s'", username)
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// 读取查询结果
	var userData UserData
	for rows.Next() {
		err = rows.Scan(&userData.Username, &userData.Password, &userData.Email)
		if err != nil {
			log.Fatal(err)
		}
	}

	// 返回个人数据
	if userData.Username != "" {
		jsonData, err := json.Marshal(userData)
		if err != nil {
			log.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	} else {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "No user data found")
	}

}

func handlerban(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// 解析请求体中的数据
	var user Userchange
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 连接数据库
	db, err := sql.Open("postgres", dbConnectionString)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Failed to connect to database:", err)
		return
	}
	defer db.Close()

	// 更新用户状态
	query := "UPDATE users SET is_disabled = $1 WHERE username = $2"
	_, err = db.Exec(query, !user.IsDisabled, user.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Failed to toggle user status:", err)
		return
	}

	// 等待数据库操作完成
	err = db.Ping()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Failed to wait for database operation:", err)
		return
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)

}

func handlerupload(w http.ResponseWriter, r *http.Request) {

	// 解析表单数据
	err := r.ParseMultipartForm(32 << 20) // 解析最大32MB的表单数据
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 获取上传的文件
	//file, handler, err := r.FormFile("avatar")
	file, _, err := r.FormFile("avatar")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 读取文件数据
	fileData, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 将文件数据保存到数据库中
	db, err := sql.Open("postgres", "postgres://postgres:lzx@localhost:5432/user_manager?sslmode=disable")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// 准备 SQL 语句
	stmt, err := db.Prepare("INSERT INTO users (image) VALUES ($1)")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	// 执行 SQL 语句，并将文件数据作为参数传递
	_, err = stmt.Exec(fileData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File uploaded successfully"))

}

// 增加一个函数，用于处理用户注销的请求
func handlerlogout(w http.ResponseWriter, r *http.Request) {

	username := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")

	// 打开数据库连接
	db, err := sql.Open("postgres", "postgres://postgres:lzx@localhost:5432/user_manager?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 删除用户信息，包括avatar字段
	deleteQuery := fmt.Sprintf("DELETE FROM users WHERE username='%s' AND password='%s'", username, password)
	_, err = db.Exec(deleteQuery)
	if err != nil {
		log.Fatal(err)
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w)
}

func main() {

	fmt.Println("hello world")
	http.HandleFunc("/api/data", sendData)
	http.HandleFunc("/api/login", handleLogin)
	http.HandleFunc("/api/register", handlerregister)
	http.HandleFunc("/api/change", handlerchange)
	http.HandleFunc("/api/user", handlermanager)
	http.HandleFunc("/api/userdata", handlersearch)
	http.HandleFunc("/api/banuser", handlerban)
	http.HandleFunc("/api/upload", handlerupload)
	http.HandleFunc("/api/logout", handlerlogout)
	http.ListenAndServe(":6166", nil)

}
