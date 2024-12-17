package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/bxcodec/faker/v3"
	_ "github.com/go-sql-driver/mysql"
)

type Product struct {
	ID          int
	Name        string
	Description string
	Price       float64
	ImageURL    string
}

var db *sql.DB
var tmpl *template.Template

func main() {
	var err error
	db, err = sql.Open("mysql", "root@tcp(127.0.0.1:3306)/products_db")

	if err != nil {
		log.Fatal("Erreur de connexion à la base de données :", err)
	}
	defer db.Close()
	
	rand.Seed(time.Now().UnixNano())

	insertProducts(1000)

funcMap := template.FuncMap{
	"add": func(a, b int) int { return a + b },
	"sub": func(a, b int) int { return a - b },
	"sequence": func(start, end int) []int {
		var s []int
		for i := start; i <= end; i++ {
			s = append(s, i)
		}
		return s
	},
	"min": func(a, b int) int {
		if a < b {
			return a
		}
		return b
	},
}
	
	tmpl = template.Must(template.New("index.html").Funcs(funcMap).ParseFiles("templates/index.html"))

	http.HandleFunc("/", HomeHandler)

	fmt.Println("Serveur démarré sur http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 20
	offset := (page - 1) * limit
	
	rows, err := db.Query("SELECT id, name, description, price, image_url FROM products LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		http.Error(w, "Erreur de récupération des produits", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.ImageURL); err != nil {
			log.Println(err)
			continue
		}
		products = append(products, p)
	}

	
	var total int
	db.QueryRow("SELECT COUNT(*) FROM products").Scan(&total)
	totalPages := (total + limit - 1) / limit
	
	startPage := (page / 10) * 10 + 1
	endPage := startPage + 9
	if endPage > totalPages {
		endPage = totalPages
	}
	
	data := map[string]interface{}{
		"Products":   products,
		"Page":       page,
		"TotalPages": totalPages,
		"StartPage":  startPage,
		"EndPage":    endPage,
	}
	tmpl.Execute(w, data)
}


func insertProducts(n int) {
	for i := 0; i < n; i++ {
		name := faker.Word()            
		description := faker.Sentence()  
		price := rand.Float64() * 100   
		imageURL := fmt.Sprintf("https://via.placeholder.com/200?text=Produit+%d", i+1) 
		
		_, err := db.Exec("INSERT INTO products (name, description, price, image_url) VALUES (?, ?, ?, ?)", name, description, price, imageURL)
		if err != nil {
			log.Println("Erreur lors de l'insertion du produit:", err)
			continue
		}

		if i%100 == 0 {
			fmt.Printf("Produit %d inséré\n", i+1) 
		}
	}

	fmt.Println("1000 produits insérés avec succès !")
}
