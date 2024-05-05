package server

import (
	"fmt"
	"net/http"
	"text/template"
)

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

func Home(w http.ResponseWriter, r *http.Request) { //affiche la page du menu principal
	renderTemplate(w, "assets/html/Accueil")
}

func Login(w http.ResponseWriter, r *http.Request) {
    renderTemplate(w, "assets/html/login")
}

func Register(w http.ResponseWriter, r *http.Request) {
    renderTemplate(w, "assets/html/register")
}


func renderTemplate(w http.ResponseWriter, tmpl string) { //Parse le fichier html et envoi les informations au client
	t, err := template.ParseFiles("./" + tmpl + ".html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, nil)
}
