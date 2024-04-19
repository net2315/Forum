package Core

import (
	"fmt"
	"net/http"
)

const port = ":3000"

func HandleFunc() {

	http.HandleFunc("/", Home)

	fmt.Println("http://localhost:3000 - Server started on port :3000")
	http.ListenAndServe(port, nil)
}

func Home(w http.ResponseWriter, r *http.Request) { //affiche la page du menu principal
	renderTemplate(w, "src/html/menu_principal")
}

func renderTemplate(w http.ResponseWriter, tmpl string) { //Parse le fichier html et envoi les informations au client
	t, err := template.ParseFiles("./" + tmpl + ".html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, nil)
}
