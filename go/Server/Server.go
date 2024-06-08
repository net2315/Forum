package server

import (
	"Forum/go/mag"
	"database/sql"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"text/template"

	_ "github.com/mattn/go-sqlite3"
)

var data []mag.Categorie

const port = ":3000"

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
	http.HandleFunc("/categories-page", categoriesPageHandler)
	http.HandleFunc("/posts", postsHandler)
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
        http.Error(w, "Erreur lors de la récupération des catégories", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "text/html")
    tmpl := template.Must(template.ParseFiles("assets/html/categories.html"))
    err = tmpl.Execute(w, categories)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func postsHandler(w http.ResponseWriter, r *http.Request) {
	categoryID := r.URL.Query().Get("id")
	if categoryID == "" {
		http.Error(w, "ID de la catégorie est manquant", http.StatusBadRequest)
		return
	}

	category, err := GetCategoryByID(categoryID)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération de la catégorie", http.StatusInternalServerError)
		return
	}

	posts, err := GetPostsByCategory(categoryID)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des posts", http.StatusInternalServerError)
		return
	}

	for i, post := range posts {
		if len(post.Photo) > 0 {
			posts[i].Photo = []byte(base64.StdEncoding.EncodeToString(post.Photo))
		}
	}

	data := struct {
		Categorie mag.Categorie
		Posts     []mag.Post
	}{
		Categorie: category,
		Posts:     posts,
	}

	renderTemplate(w, "assets/html/Posts", data)
}

func GetCategoryByID(categoryID string) (mag.Categorie, error) {
	db, err := sql.Open("sqlite3", "./db/database.db")
	if err != nil {
		return mag.Categorie{}, err
	}
	defer db.Close()

	var category mag.Categorie
	err = db.QueryRow("SELECT id, nom, description FROM categorie WHERE id = ?", categoryID).Scan(&category.ID, &category.Nom, &category.Description)
	if err != nil {
		return mag.Categorie{}, err
	}

	return category, nil
}

func GetPostsByCategory(categoryID string) ([]mag.Post, error) {
	db, err := sql.Open("sqlite3", "./db/database.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, categorie_id, texte, date_heure, photo FROM post WHERE categorie_id = ?", categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []mag.Post
	for rows.Next() {
		var post mag.Post
		if err := rows.Scan(&post.ID, &post.CategorieID, &post.Texte, &post.DateHeure, &post.Photo); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func categoriesPageHandler(w http.ResponseWriter, r *http.Request) {
	categories, err := GetCategories()
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des catégories", http.StatusInternalServerError)
		return
	}

	renderTemplate(w, "assets/html/Categories", categories)
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

func AuthenticateUser(mail, password string) (mag.User, error) {
	db, err := sql.Open("sqlite3", "./db/database.db")
	if err != nil {
		return mag.User{}, err
	}
	defer db.Close()

	var user mag.User
	err = db.QueryRow("SELECT ID, Pseudo, Mail, MotDePasse, IDPost, IDCommentaire FROM Users WHERE Mail = ?", mail).Scan(&user.ID, &user.Pseudo, &user.Mail, &user.MotDePasse, &user.IDPost, &user.IDCommentaire)
	if err != nil {
		return mag.User{}, fmt.Errorf("utilisateur non trouvé")
	}

	// Compare the stored hashed password with the provided password
	err = bcrypt.CompareHashAndPassword([]byte(user.MotDePasse), []byte(password))
	if err != nil {
		return mag.User{}, fmt.Errorf("mot de passe incorrect")
	}

	return user, nil
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		mail := r.FormValue("mail")
		password := r.FormValue("password")

		user, err := AuthenticateUser(mail, password)
		if err != nil {
			http.Error(w, "Identifiants invalides", http.StatusUnauthorized)
			return
		}

		// Set a session cookie or token here if needed
		// Example using a cookie:
		cookie := http.Cookie{
			Name:  "user",
			Value: base64.StdEncoding.EncodeToString([]byte(user.Mail)),
			Path:  "/",
		}
		http.SetCookie(w, &cookie)

		// Redirect the user after successful login
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// For GET request to /login, render the login form
	renderTemplate(w, "assets/html/login", nil)
}

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		mail := r.FormValue("mail")
		password := r.FormValue("password")
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Erreur lors de la création du mot de passe", http.StatusInternalServerError)
			return
		}

		err = InsertUser(mail, string(hashedPassword))
		if err != nil {
			http.Error(w, "Erreur lors de l'enregistrement de l'utilisateur", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	renderTemplate(w, "assets/html/register", nil)
}

func InsertUser(mail, hashedPassword string) error {
	db, err := sql.Open("sqlite3", "./db/database.db")
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO Users (Mail, MotDePasse) VALUES (?, ?)", mail, hashedPassword)
	return err
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t, err := template.ParseFiles("./" + tmpl + ".html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, data)
}