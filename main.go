package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Motor struct {
	ID             uint `gorm:"primaryKey"`
	Name           string
	Brand          string
	EngineCapacity int
	Transmission   string
	Color          string
	Year           int
	Price          int
	Stock          int
	Description    string
}

func connectDB() (*gorm.DB, error) {
	dsn := "root:@tcp(localhost:3306)/uts?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func migrasiDB() {
	db, err := connectDB()
	if err != nil {
		log.Fatal("Gagal koneksi:", err)
	}
	db.AutoMigrate(&Motor{})
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	db, err := connectDB()
	if err != nil {
		http.Error(w, "Gagal terhubung ke database", http.StatusInternalServerError)
		return
	}

	var motors []Motor
	db.Find(&motors)

	tmpl, err := template.ParseFiles("template/index.html")
	if err != nil {
		http.Error(w, "Gagal memuat template", http.StatusInternalServerError)
		log.Println("Error loading template:", err)
		return
	}

	err = tmpl.Execute(w, motors)
	if err != nil {
		http.Error(w, "Gagal menampilkan data motor", http.StatusInternalServerError)
		log.Println("Error executing template:", err)
		return
	}
}

func tambahHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		engine, _ := strconv.Atoi(r.FormValue("engine_capacity"))
		year, _ := strconv.Atoi(r.FormValue("year"))
		price, _ := strconv.Atoi(r.FormValue("price"))
		stock, _ := strconv.Atoi(r.FormValue("stock"))

		motor := Motor{
			Name:           r.FormValue("name"),
			Brand:          r.FormValue("brand"),
			EngineCapacity: engine,
			Transmission:   r.FormValue("transmission"),
			Color:          r.FormValue("color"),
			Year:           year,
			Price:          price,
			Stock:          stock,
			Description:    r.FormValue("description"),
		}

		db, err := connectDB()
		if err != nil {
			http.Error(w, "Gagal terhubung ke database", http.StatusInternalServerError)
			log.Println("Error connecting to DB:", err)
			return
		}

		db.Create(&motor)

		// Redirect ke halaman utama setelah berhasil menambah data
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Handle GET request untuk menampilkan form
	var years []int
	for i := 2010; i <= 2025; i++ {
		years = append(years, i)
	}

	tmpl, err := template.ParseFiles("template/tambah.html")
	if err != nil {
		http.Error(w, "Gagal memuat template tambah.html", http.StatusInternalServerError)
		log.Println("Error loading tambah.html:", err)
		return
	}

	err = tmpl.Execute(w, map[string]interface{}{
		"Years": years,
	})
	if err != nil {
		http.Error(w, "Gagal menampilkan form tambah motor", http.StatusInternalServerError)
		log.Println("Error executing template:", err)
		return
	}
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id == 0 {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}

	db, err := connectDB()
	if err != nil {
		http.Error(w, "Gagal terhubung ke database", http.StatusInternalServerError)
		log.Println("Error connecting to DB:", err)
		return
	}

	var motor Motor
	if err := db.First(&motor, id).Error; err != nil {
		http.Error(w, "Motor tidak ditemukan", http.StatusNotFound)
		log.Println("Motor not found:", err)
		return
	}

	// Menangani POST data dari form edit
	if r.Method == http.MethodPost {
		engine, err := strconv.Atoi(r.FormValue("engine_capacity"))
		if err != nil {
			http.Error(w, "Kapasitas mesin tidak valid", http.StatusBadRequest)
			return
		}

		year, err := strconv.Atoi(r.FormValue("year"))
		if err != nil {
			http.Error(w, "Tahun tidak valid", http.StatusBadRequest)
			return
		}

		price, err := strconv.Atoi(r.FormValue("price"))
		if err != nil {
			http.Error(w, "Harga tidak valid", http.StatusBadRequest)
			return
		}

		stock, err := strconv.Atoi(r.FormValue("stock"))
		if err != nil {
			http.Error(w, "Stok tidak valid", http.StatusBadRequest)
			return
		}

		motor.Name = r.FormValue("name")
		motor.Brand = r.FormValue("brand")
		motor.EngineCapacity = engine
		motor.Transmission = r.FormValue("transmission")
		motor.Color = r.FormValue("color")
		motor.Year = year
		motor.Price = price
		motor.Stock = stock
		motor.Description = r.FormValue("description")

		if err := db.Save(&motor).Error; err != nil {
			http.Error(w, "Gagal mengupdate data motor", http.StatusInternalServerError)
			log.Println("Error updating motor:", err)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFiles("template/edit.html")
	if err != nil {
		http.Error(w, "Gagal memuat template edit.html", http.StatusInternalServerError)
		log.Println("Error loading edit.html:", err)
		return
	}

	err = tmpl.Execute(w, motor)
	if err != nil {
		http.Error(w, "Gagal menampilkan form edit motor", http.StatusInternalServerError)
		log.Println("Error executing template:", err)
		return
	}
}

func hapusHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))

	db, err := connectDB()
	if err != nil {
		http.Error(w, "Gagal terhubung ke database", http.StatusInternalServerError)
		log.Println("Error connecting to DB:", err)
		return
	}

	db.Delete(&Motor{}, id)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	migrasiDB()

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/tambah", tambahHandler)
	http.HandleFunc("/edit", editHandler)
	http.HandleFunc("/hapus", hapusHandler)

	log.Println("Server berjalan di :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
