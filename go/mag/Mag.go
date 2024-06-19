package mag

import "time"

type Categorie struct {
	ID          int
	Nom         string
	Description string
}

type Post struct {
	ID           int
	CategorieID  int
	CategorieNom string
	Texte        string
	DateHeure    time.Time
	Photo        string
	Likes        int
	Comments     []Comment
}

type Comment struct {
	ID        int
	UserID    int
	Texte     string
	DateHeure time.Time
	Likes     int
	PostID    int
}

type User struct {
	ID            int
	Pseudo        string
	Mail          string
	MotDePasse    string
	IDPost        int
	IDCommentaire int
}

type MessagesCree struct {
	ID           int
	PostID       int
	UserID       int
	DateCreation string
}

type MessagesAime struct {
	ID        int
	PostID    int
	UserID    int
	DateAimee string
}
