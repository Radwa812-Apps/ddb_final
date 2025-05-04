package main

/*import (
	"log"
	"net/http"
	"os"
	"database/sql"
	"fmt"
	"text/template"
	"database/sql"
	"github.com/go-sql-driver/mysql"
)

var db *sql.DB
var cfg *mysql.Config
var tmpl *template.Template

type PageData struct {
	Title string
}

func main() {
	// Load templates
	tmpl = template.Must(template.ParseGlob("templates/*.html"))
	// Capture connection properties.
	cfg = mysql.NewConfig()
	cfg.User = os.Getenv("DBUSER")
	cfg.Passwd = os.Getenv("DBPASS")
	cfg.Net = "tcp"
	cfg.Addr = "127.0.0.1:3306"
	cfg.DBName = "ecommerce_db"

	// Get a database handle.
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected!")

	// التعامل مع التحديثات القادمة من الـ Master
	http.HandleFunc("/replicate", replicateHandler)

	http.HandleFunc("/", dashboardHandler)

	http.HandleFunc("/dashboard", dashboardHandler)
	http.HandleFunc("/products", productsHandler)
	http.HandleFunc("/orders", ordersHandler)
	http.HandleFunc("/customers", customersHandler)
	http.HandleFunc("/reports", reportsHandler)

	http.HandleFunc("/customer/add", addCustomerHandler)
	http.HandleFunc("/customers/add", addCustomerSubmitHandler)
	// تشغيل السيرفر على البورت 8081 (مثلاً)
	fmt.Println("Slave Node running on http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

// وظيفة استقبال التحديثات من الـ Master
func replicateHandler(w http.ResponseWriter, r *http.Request) {
	// في هذا الجزء، سنقوم بتنفيذ التحديثات القادمة من الـ Master
	if r.Method == http.MethodPost {
		// عند وصول التحديثات، يمكن تنفيذ الاستعلامات هنا
		// هنا على سبيل المثال، لو الـ Master أرسل استعلام INSERT أو UPDATE
		// سننفذه على قاعدة بيانات الـ Slave

		// مثال: إضافة منتج في قاعدة البيانات
		_, err := db.Exec("INSERT INTO products (name, price) VALUES ('Sample Product', 100.0)")
		if err != nil {
			log.Printf("Error executing replication: %v", err)
			http.Error(w, "Replication failed", http.StatusInternalServerError)
			return
		}

		// إرجاع استجابة ناجحة
		fmt.Fprintf(w, "Replication Successful")
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Dashboard",
	}
	err := tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func productsHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Products",
	}

	err := tmpl.ExecuteTemplate(w, "products.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ordersHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Orders",
	}

	err := tmpl.ExecuteTemplate(w, "orders.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func customersHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Customers",
	}

	err := tmpl.ExecuteTemplate(w, "customers.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func reportsHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Reports",
	}

	err := tmpl.ExecuteTemplate(w, "reports.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func addCustomerHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Add Customer",
	}

	err := tmpl.ExecuteTemplate(w, "add-customer.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func addCustomerSubmitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		firstName := r.FormValue("firstName")
		email := r.FormValue("email")
		password := r.FormValue("password")

		var err error
		_, err = db.Exec("INSERT INTO users (name, email, password) VALUES (?, ?, ?)", firstName, email, password)
		if err != nil {
			log.Fatal(err)
		}

		http.Redirect(w, r, "/customers", http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFiles("templates/add-customer.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}
*/