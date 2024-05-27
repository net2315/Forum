package server

import (
	"Forum/go/database"
	"Forum/go/mag"
	"fmt"
	"net/http"
	"text/template"
)

var data mag.Categorie

const port = ":3000"

func HandleFunc() {
	http.HandleFunc("/", Home)
	http.HandleFunc("/login", Login)
	http.HandleFunc("/register", Register)
	http.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir("./media"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./assets/css/"))))

	fmt.Println("http://localhost:3000 - Server started on port :3000")
	err := http.ListenAndServe(port, nil)
	if err != nil {
		return
	}
}

func Home(w http.ResponseWriter, r *http.Request) {
	categories, err := database.GetCategories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, category := range categories {
		fmt.Printf("ID: %d, Nom: %s, Description: %s\n", category.ID, category.Nom, category.Description)
	}

	renderTemplate(w, "assets/html/Accueil")
}

func Login(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "assets/html/login")
}

func Register(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "assets/html/register")
}

func renderTemplate(w http.ResponseWriter, tmpl string) {
	t, err := template.ParseFiles("./" + tmpl + ".html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, nil)
}
