package mag

type Categorie struct {
	ID          int
	Nom         string
	Description string
}

type Post struct {
	ID         int
	CategorieID int
	Texte      string
	DateHeure  string 
	Photo      []byte
	Comments   []Comment 
}

type User struct {
	ID            int
	Pseudo        string
	Mail          string
	MotDePasse    string
	IDPost        int
	IDCommentaire int
}

type Comment struct {
	ID         int
	UserID     int
	PostID    int  
	Texte      string
	DateHeure  string
	Likes      int
}