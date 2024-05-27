package server

import (
	"Forum/go/mag"
	"database/sql"
	"fmt"
	"net/http"
	"text/template"

	_ "github.com/mattn/go-sqlite3"
)

var data []mag.Categorie

const port = ":3000"

// func HandleFunc() {
// 	// Initialize data when server starts
// 	_, err := GetCategories()
// 	if err != nil {
// 		fmt.Println("Error initializing categories:", err)
// 		return
// 	}

// 	http.HandleFunc("/", Home)
// 	http.HandleFunc("/login", Login)
// 	http.HandleFunc("/register", Register)
// 	http.HandleFunc("/categories", categoriesHandler)
// 	http.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir("./media"))))
// 	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./assets/css/"))))

// 	fmt.Println("http://localhost:3000 - Server started on port :3000")
// 	err = http.ListenAndServe(port, nil)
// 	if err != nil {
// 		return
// 	}
// }
func HandleFunc() {
	// Initialiser les données au démarrage du serveur
	_, err := GetCategories()
	if err != nil {
		fmt.Println("Erreur lors de l'initialisation des catégories:", err)
		return
	}

	http.HandleFunc("/", Home)
	http.HandleFunc("/login", Login)
	http.HandleFunc("/register", Register)
	http.HandleFunc("/categories", categoriesHandler)
	http.HandleFunc("/addCategory", AddCategory)
	http.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir("./media"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./assets/css/"))))

	fmt.Println("http://localhost:3000 - Serveur démarré sur le port :3000")
	err = http.ListenAndServe(port, nil)
	if err != nil {
		return
	}
}


func GetCategories() ([]mag.Categorie, error) {
	//Open database
	db, err := sql.Open("sqlite3", "./db/database.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, nom FROM categorie")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []mag.Categorie
	for rows.Next() {
		var category mag.Categorie
		if err := rows.Scan(&category.ID, &category.Nom); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	data = categories 

	return categories, nil
}


func categoriesHandler(w http.ResponseWriter, r *http.Request) {
	categories, err := GetCategories()
	if err != nil {
		http.Error(w, "Error retrieving categories", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	for _, category := range categories {
		fmt.Fprintf(w, "ID: %d\nNom: %s\n\n", category.ID, category.Nom)
	}
}

func Home(w http.ResponseWriter, r *http.Request) {
	categories, err := GetCategories()
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des catégories", http.StatusInternalServerError)
		return
	}

	renderTemplate(w, "assets/html/Accueil", categories)
}

func InsertCategory(nom string) error {
	db, err := sql.Open("sqlite3", "./db/database.db")
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO categorie (nom) VALUES (?)", nom)
	return err
}

func AddCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		nom := r.FormValue("nom")
		if nom == "" {
			http.Error(w, "Le nom de la catégorie est requis", http.StatusBadRequest)
			return
		}

		err := InsertCategory(nom)
		if err != nil {
			http.Error(w, "Erreur lors de l'ajout de la catégorie", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}


func Login(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "assets/html/login", nil)
}

func Register(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "assets/html/register", nil)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t, err := template.ParseFiles("./" + tmpl + ".html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, data)
}
