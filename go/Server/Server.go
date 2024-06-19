package server

import (
	"Forum/go/mag"
	"database/sql"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"io"
	"net/http"
	"text/template"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var data []mag.Categorie

type HomePageData struct {
	Posts      []mag.Post
	Categories []mag.Categorie
}

const port = ":3000"

func HandleFunc() {

	http.HandleFunc("/", Home)
	http.HandleFunc("/login", Login)
	http.HandleFunc("/register", Register)
	http.HandleFunc("/categories", categoriesHandler)
	http.HandleFunc("/nouv_cat", Nouv_Cat)
	http.HandleFunc("/addCategory", AddCategory)
	http.HandleFunc("/nouv_post", Nouv_Post)
	http.HandleFunc("/addPost", AddPostHandler)
	http.HandleFunc("/addComment", AddCommentHandler)
	http.HandleFunc("/categories-page", categoriesPageHandler)
	http.HandleFunc("/cat-posts", postsHandler)
	http.HandleFunc("/messagesCrees", messagesCreesHandler)
	http.HandleFunc("/messagesAimes", messagesAimesHandler)
	http.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir("./media"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./assets/css/"))))

	fmt.Println("http://localhost:3000 - Serveur démarré sur le port :3000")
	err := http.ListenAndServe(port, nil)
	if err != nil {
		return
	}
}

func Home(w http.ResponseWriter, r *http.Request) {
	// Récupérer toutes les catégories
	categories, err := GetCategories()
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des catégories: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var allPosts []mag.Post

	// Pour chaque catégorie, récupérer les posts et les ajouter à la liste
	for _, cat := range categories {
		posts, err := GetPostsByCategory(fmt.Sprint(cat.ID))
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des posts pour la catégorie "+cat.Nom+": "+err.Error(), http.StatusInternalServerError)
			return
		}
		allPosts = append(allPosts, posts...)
	}

	// Créer la structure de données pour le template
	data := HomePageData{
		Categories: categories, // Ajouter les catégories à la structure de données
		Posts:      allPosts,
	}

	renderTemplate(w, "assets/html/Accueil", data)
}

func Nouv_Post(w http.ResponseWriter, r *http.Request) {
	categories, err := GetCategories()
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des catégories: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := HomePageData{
		Categories: categories,
	}

	renderTemplate(w, "assets/html/Nouv-Post", data)
}

func Nouv_Cat(w http.ResponseWriter, r *http.Request) {
	categories, err := GetCategories()
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des catégories: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := HomePageData{
		Categories: categories,
	}

	renderTemplate(w, "assets/html/Nouv-Cat", data)
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

func AddPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		categorieID := r.FormValue("categorie_id")
		texte := r.FormValue("texte")
		dateHeure := time.Now()

		file, _, err := r.FormFile("photo")
		var imgData []byte
		if err == nil && file != nil {
			defer file.Close()
			imgData, err = io.ReadAll(file)
			if err != nil {
				http.Error(w, "Erreur lors de la lecture de l'image: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		err = InsertPost(categorieID, texte, dateHeure, imgData)
		if err != nil {
			http.Error(w, "Erreur lors de l'ajout du post: "+err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

func InsertPost(categorieID, texte string, dateHeure time.Time, photo []byte) error {
	db, err := sql.Open("sqlite3", "./db/database.db")
	if err != nil {
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO post (categorie_id, texte, date_heure, photo) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(categorieID, texte, dateHeure, photo)
	return err
}

func InsertCategory(nom, description string) error {
	db, err := sql.Open("sqlite3", "./db/database.db")
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO categorie (nom, description) VALUES (?, ?)", nom, description)
	return err
}

func AddCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		nom := r.FormValue("nom")
		description := r.FormValue("description")

		if nom == "" {
			http.Error(w, "Le nom de la catégorie est requis", http.StatusBadRequest)
			return
		}

		err := InsertCategory(nom, description)
		if err != nil {
			http.Error(w, "Erreur lors de l'ajout de la catégorie: "+err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/categories-page", http.StatusSeeOther)
	} else {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

func categoriesPageHandler(w http.ResponseWriter, r *http.Request) {
	categories, err := GetCategories()
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des catégories", http.StatusInternalServerError)
		return
	}

	renderTemplate(w, "assets/html/Categories", categories)
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
		return nil, fmt.Errorf("erreur d'ouverture de la base de données: %w", err)
	}
	defer db.Close()

	query := `
    SELECT p.id, p.categorie_id, c.nom, p.texte, p.date_heure, p.photo
    FROM post p
    JOIN categorie c ON p.categorie_id = c.id
    WHERE p.categorie_id = ?
    `
	rows, err := db.Query(query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des posts: %w", err)
	}
	defer rows.Close()

	var posts []mag.Post
	for rows.Next() {
		var post mag.Post
		var photoData []byte
		if err := rows.Scan(&post.ID, &post.CategorieID, &post.CategorieNom, &post.Texte, &post.DateHeure, &photoData); err != nil {
			return nil, fmt.Errorf("erreur lors du scan des posts: %w", err)
		}
		post.Photo = base64.StdEncoding.EncodeToString(photoData)
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur lors du parcours des posts: %w", err)
	}

	return posts, nil
}

func AddComment(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Parse form values
		postID := r.FormValue("post_id")
		userID := r.FormValue("user_id")
		texte := r.FormValue("texte")

		if postID == "" || userID == "" || texte == "" {
			http.Error(w, "Tous les champs sont obligatoires", http.StatusBadRequest)
			return
		}

		// Insert comment into the database
		err := InsertComment(postID, userID, texte)
		if err != nil {
			http.Error(w, "Erreur lors de l'ajout du commentaire", http.StatusInternalServerError)
			return
		}

		// Redirect to the post page
		http.Redirect(w, r, fmt.Sprintf("/cat-posts?id=%s", postID), http.StatusSeeOther)
	} else {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

func AddCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Récupérer les données du formulaire
		userID := r.FormValue("user_id")
		postID := r.FormValue("post_id")
		texte := r.FormValue("texte")

		// Insérer le commentaire dans la base de données
		err := InsertComment(userID, postID, texte)
		if err != nil {
			http.Error(w, "Erreur lors de l'insertion du commentaire", http.StatusInternalServerError)
			return
		}

		// Rediriger ou renvoyer une réponse appropriée après l'insertion réussie
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Pour les méthodes autres que POST, retourner une erreur HTTP appropriée
	http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
}

func InsertComment(userID, postID, texte string) error {
	db, err := sql.Open("sqlite3", "./db/database.db")
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO comments (UserID, PostID, Texte) VALUES (?, ?, ?)", userID, postID, texte)
	if err != nil {
		return err
	}

	return nil
}

func postsHandler(w http.ResponseWriter, r *http.Request) {
	categoryID := r.URL.Query().Get("id")
	if categoryID == "" {
		http.Error(w, "ID de la catégorie est manquant", http.StatusBadRequest)
		return
	}

	category, err := GetCategoryByID(categoryID)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération de la catégorie: "+err.Error(), http.StatusInternalServerError)
		return
	}

	posts, err := GetPostsByCategory(categoryID)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des posts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Les posts sont déjà préparés avec les photos encodées en base64, donc pas besoin de modifier

	data := struct {
		Categorie mag.Categorie
		Posts     []mag.Post
	}{
		Categorie: category,
		Posts:     posts,
	}

	renderTemplate(w, "assets/html/Cat-Posts", data)
}

func GetPosts() ([]mag.Post, error) {
	db, err := sql.Open("sqlite3", "./db/database.db")
	if err != nil {
		return nil, fmt.Errorf("erreur d'ouverture de la base de données: %w", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, categorie_id, texte, date_heure, photo, likes FROM post")
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des posts: %w", err)
	}
	defer rows.Close()

	var posts []mag.Post
	for rows.Next() {
		var post mag.Post
		if err := rows.Scan(&post.ID, &post.CategorieID, &post.Texte, &post.DateHeure, &post.Photo, &post.Likes); err != nil {
			return nil, fmt.Errorf("erreur lors du scan des posts: %w", err)
		}

		comments, err := GetCommentsByPost(fmt.Sprint(post.ID))
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la récupération des commentaires: %w", err)
		}
		post.Comments = comments

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur lors du parcours des posts: %w", err)
	}

	return posts, nil
}

func GetCommentsByPost(postID string) ([]mag.Comment, error) {
	db, err := sql.Open("sqlite3", "./db/database.db")
	if err != nil {
		return nil, fmt.Errorf("erreur d'ouverture de la base de données: %w", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT ID, UserID, Texte, DateHeure, Likes, PostID FROM comments WHERE PostID = ?", postID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des commentaires: %w", err)
	}
	defer rows.Close()

	var comments []mag.Comment
	for rows.Next() {
		var comment mag.Comment
		if err := rows.Scan(&comment.ID, &comment.UserID, &comment.Texte, &comment.DateHeure, &comment.Likes, &comment.PostID); err != nil {
			return nil, fmt.Errorf("erreur lors du scan des commentaires: %w", err)
		}
		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur lors du parcours des commentaires: %w", err)
	}

	return comments, nil
}

func InsertLike(postID, userID int, likeType string, sticker []byte) error {
	db, err := sql.Open("sqlite3", "./db/database.db")
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	res, err := tx.Exec("INSERT INTO Likes (post_id, user_id, type) VALUES (?, ?, ?)", postID, userID, likeType)
	if err != nil {
		tx.Rollback()
		return err
	}

	likeID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("INSERT INTO Stickers (like_id, sticker) VALUES (?, ?)", likeID, sticker)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
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
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func GetMessagesCrees(userID int) ([]mag.MessagesCree, error) {
	db, err := sql.Open("sqlite3", "./db/database.db")
	if err != nil {
		return nil, fmt.Errorf("erreur d'ouverture de la base de données: %w", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, post_id, user_id, date_creation FROM MessagesCree WHERE user_id = ?", userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des messages créés: %w", err)
	}
	defer rows.Close()

	var messagesCrees []mag.MessagesCree
	for rows.Next() {
		var message mag.MessagesCree
		if err := rows.Scan(&message.ID, &message.PostID, &message.UserID, &message.DateCreation); err != nil {
			return nil, fmt.Errorf("erreur lors du scan des messages créés: %w", err)
		}
		messagesCrees = append(messagesCrees, message)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur lors du parcours des messages créés: %w", err)
	}

	return messagesCrees, nil
}

func GetMessagesAimes(userID int) ([]mag.MessagesAime, error) {
	db, err := sql.Open("sqlite3", "./db/database.db")
	if err != nil {
		return nil, fmt.Errorf("erreur d'ouverture de la base de données: %w", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, post_id, user_id, date_aimee FROM MessagesAime WHERE user_id = ?", userID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des messages aimés: %w", err)
	}
	defer rows.Close()

	var messagesAimes []mag.MessagesAime
	for rows.Next() {
		var message mag.MessagesAime
		if err := rows.Scan(&message.ID, &message.PostID, &message.UserID, &message.DateAimee); err != nil {
			return nil, fmt.Errorf("erreur lors du scan des messages aimés: %w", err)
		}
		messagesAimes = append(messagesAimes, message)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur lors du parcours des messages aimés: %w", err)
	}

	return messagesAimes, nil
}

func messagesCreesHandler(w http.ResponseWriter, r *http.Request) {

	userID := 1

	messagesCrees, err := GetMessagesCrees(userID)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des messages créés: "+err.Error(), http.StatusInternalServerError)
		return
	}

	renderTemplate(w, "assets/html/MessagesCrees", messagesCrees)
}

func messagesAimesHandler(w http.ResponseWriter, r *http.Request) {

	userID := 1

	messagesAimes, err := GetMessagesAimes(userID)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des messages aimés: "+err.Error(), http.StatusInternalServerError)
		return
	}

	renderTemplate(w, "assets/html/MessagesAimes", messagesAimes)
}
