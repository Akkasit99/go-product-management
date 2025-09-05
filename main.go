package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
)

var (
	db    *sql.DB
	tpl   = template.Must(template.ParseGlob("templates/*"))
	store = sessions.NewCookieStore([]byte("secret-key"))
)

func initDB() {
	var err error
	db, err = sql.Open("mysql", "root:@/db_go")
	if err != nil {
		log.Fatal("Cannot connect to database: ", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("Cannot ping database: ", err)
	}
}

func main() {
	initDB()
	defer db.Close()

	// Static file serving for uploaded images
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	http.HandleFunc("/", checkLogin(indexHandler))
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/add", checkLogin(addHandler))
	http.HandleFunc("/edit", checkLogin(editHandler))
	http.HandleFunc("/delete", checkLogin(deleteHandler))
	http.HandleFunc("/logout", logoutHandler)

	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// ตรวจสอบการล็อกอิน
func checkLogin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "session-name")
		if err != nil || session.Values["loggedIn"] != true {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	}
}

// แสดงข้อมูลจากฐานข้อมูล
func indexHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	role := session.Values["role"].(string)

	rows, err := db.Query("SELECT id, name, COALESCE(description, ''), price, COALESCE(image, '') FROM good")
	if err != nil {
		http.Error(w, "Database query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var goods []struct {
		ID          int
		Name        string
		Description string
		Price       float64
		Image       string
	}
	for rows.Next() {
		var g struct {
			ID          int
			Name        string
			Description string
			Price       float64
			Image       string
		}
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.Price, &g.Image); err != nil {
			http.Error(w, "Error scanning row", http.StatusInternalServerError)
			return
		}
		goods = append(goods, g)
	}

	// ตรวจสอบข้อความแจ้งเตือนจาก URL parameters
	var message, messageType string
	if r.URL.Query().Get("login") == "success" {
		message = "Welcome back! You have successfully logged in."
		messageType = "success"
	}

	data := struct {
		Goods []struct {
			ID          int
			Name        string
			Description string
			Price       float64
			Image       string
		}
		Role        string
		Message     string
		MessageType string
	}{
		Goods:       goods,
		Role:        role,
		Message:     message,
		MessageType: messageType,
	}

	tpl.ExecuteTemplate(w, "index.html", data)
}

// ล็อกอิน
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		var dbUsername, dbPassword, role string
		err := db.QueryRow("SELECT username, password, role FROM users WHERE username = ?", username).Scan(&dbUsername, &dbPassword, &role)
		if err != nil || dbPassword != password {
			// ส่งข้อความแจ้งเตือนเมื่อล็อกอินไม่สำเร็จ
			data := struct {
				Message     string
				MessageType string
			}{
				Message:     "Invalid username or password. Please try again.",
				MessageType: "error",
			}
			tpl.ExecuteTemplate(w, "login.html", data)
			return
		}

		session, _ := store.Get(r, "session-name")
		session.Values["loggedIn"] = true
		session.Values["username"] = username
		session.Values["role"] = role
		session.Save(r, w)

		// ส่งข้อความแจ้งเตือนเมื่อล็อกอินสำเร็จ (redirect ไปหน้าหลักพร้อมข้อความ)
		http.Redirect(w, r, "/?login=success", http.StatusSeeOther)
		return
	}
	// ตรวจสอบข้อความแจ้งเตือนจาก URL parameters
	var message, messageType string
	if r.URL.Query().Get("register") == "success" {
		message = "Registration successful! Please login with your new account."
		messageType = "success"
	}

	data := struct {
		Message     string
		MessageType string
	}{
		Message:     message,
		MessageType: messageType,
	}

	tpl.ExecuteTemplate(w, "login.html", data)
}

// เพิ่มข้อมูล (ป้องกัน SQL injection)
func addHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	role := session.Values["role"].(string)

	if role != "admin" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		description := r.FormValue("description")
		price := r.FormValue("price")

		// Handle image upload
		var imagePath string
		file, header, err := r.FormFile("image")
		if err == nil {
			defer file.Close()

			// Generate unique filename using timestamp
			timestamp := strconv.FormatInt(time.Now().Unix(), 10)
			extension := filepath.Ext(header.Filename)
			filename := timestamp + extension
			imagePath = "uploads/" + filename

			// Create the file
			dst, err := os.Create(imagePath)
			if err != nil {
				http.Error(w, "Error creating image file", http.StatusInternalServerError)
				return
			}
			defer dst.Close()

			// Copy the uploaded file to the destination
			_, err = io.Copy(dst, file)
			if err != nil {
				http.Error(w, "Error saving image file", http.StatusInternalServerError)
				return
			}
		} else {
			imagePath = "" // No image uploaded
		}

		stmt, err := db.Prepare("INSERT INTO good (name, description, price, image) VALUES (?, ?, ?, ?)") //(ป้องกัน SQL injection)
		if err != nil {
			http.Error(w, "Error preparing query", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(name, description, price, imagePath)
		if err != nil {
			http.Error(w, "Error executing query", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	tpl.ExecuteTemplate(w, "add.html", nil)
}

// แก้ไขข้อมูล (ป้องกัน SQL injection)
func editHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	role := session.Values["role"].(string)

	if role != "admin" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodPost {
		id := r.FormValue("id")
		name := r.FormValue("name")
		description := r.FormValue("description")
		price := r.FormValue("price")

		// Get current image path
		var currentImage string
		err := db.QueryRow("SELECT COALESCE(image, '') FROM good WHERE id = ?", id).Scan(&currentImage)
		if err != nil {
			http.Error(w, "Error getting current image", http.StatusInternalServerError)
			return
		}

		// Handle image upload
		imagePath := currentImage // Keep current image by default
		file, header, err := r.FormFile("image")
		if err == nil {
			defer file.Close()

			// Delete old image if exists
			if currentImage != "" {
				os.Remove(currentImage)
			}

			// Generate unique filename
			timestamp := strconv.FormatInt(time.Now().Unix(), 10)
			extension := filepath.Ext(header.Filename)
			filename := timestamp + "_" + id + extension
			imagePath = "uploads/" + filename

			// Create the file
			dst, err := os.Create(imagePath)
			if err != nil {
				http.Error(w, "Error creating image file", http.StatusInternalServerError)
				return
			}
			defer dst.Close()

			// Copy the uploaded file to the destination
			_, err = io.Copy(dst, file)
			if err != nil {
				http.Error(w, "Error saving image file", http.StatusInternalServerError)
				return
			}
		}

		stmt, err := db.Prepare("UPDATE good SET name = ?, description = ?, price = ?, image = ? WHERE id = ?") //(ป้องกัน SQL injection)
		if err != nil {
			http.Error(w, "Error preparing query", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(name, description, price, imagePath, id)
		if err != nil {
			http.Error(w, "Error executing query", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	id := r.URL.Query().Get("id")
	row := db.QueryRow("SELECT id, name, COALESCE(description, ''), price, COALESCE(image, '') FROM good WHERE id = ?", id)

	var g struct {
		ID          int
		Name        string
		Description string
		Price       float64
		Image       string
	}
	if err := row.Scan(&g.ID, &g.Name, &g.Description, &g.Price, &g.Image); err != nil {
		http.Error(w, "Good not found", http.StatusNotFound)
		return
	}

	tpl.ExecuteTemplate(w, "edit.html", g)
}

// ลบข้อมูล (ป้องกัน SQL injection)
func deleteHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	role := session.Values["role"].(string)

	if role != "admin" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.URL.Query().Get("id")

	stmt, err := db.Prepare("DELETE FROM good WHERE id = ?") //(ป้องกัน SQL injection)
	if err != nil {
		http.Error(w, "Error preparing query", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		http.Error(w, "Error executing query", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// การสมัครสมาชิก
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		// ตรวจสอบว่าผู้ใช้มีอยู่แล้วหรือไม่
		var existingUser string
		err := db.QueryRow("SELECT username FROM users WHERE username = ?", username).Scan(&existingUser)
		if err != nil && err != sql.ErrNoRows {
			// ส่งข้อความแจ้งเตือนเมื่อเกิดข้อผิดพลาดฐานข้อมูล
			data := struct {
				Message     string
				MessageType string
			}{
				Message:     "Database error occurred. Please try again later.",
				MessageType: "error",
			}
			tpl.ExecuteTemplate(w, "register.html", data)
			return
		}

		if existingUser != "" {
			// ส่งข้อความแจ้งเตือนเมื่อชื่อผู้ใช้มีอยู่แล้ว
			data := struct {
				Message     string
				MessageType string
			}{
				Message:     "Username already exists. Please choose a different username.",
				MessageType: "error",
			}
			tpl.ExecuteTemplate(w, "register.html", data)
			return
		}

		stmt, err := db.Prepare("INSERT INTO users (username, password, role) VALUES (?, ?, 'user')") //(ป้องกัน SQL injection)
		if err != nil {
			// ส่งข้อความแจ้งเตือนเมื่อเกิดข้อผิดพลาดในการเตรียม query
			data := struct {
				Message     string
				MessageType string
			}{
				Message:     "Database error occurred. Please try again later.",
				MessageType: "error",
			}
			tpl.ExecuteTemplate(w, "register.html", data)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(username, password)
		if err != nil {
			// ส่งข้อความแจ้งเตือนเมื่อเกิดข้อผิดพลาดในการสร้างผู้ใช้
			data := struct {
				Message     string
				MessageType string
			}{
				Message:     "Error creating user account. Please try again.",
				MessageType: "error",
			}
			tpl.ExecuteTemplate(w, "register.html", data)
			return
		}

		// ส่งข้อความแจ้งเตือนเมื่อสมัครสมาชิกสำเร็จ
		http.Redirect(w, r, "/login?register=success", http.StatusSeeOther)
		return
	}
	tpl.ExecuteTemplate(w, "register.html", nil)
}

// ล็อกเอาท์
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	session.Values["loggedIn"] = false
	session.Values["username"] = ""
	session.Values["role"] = ""
	session.Save(r, w)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}