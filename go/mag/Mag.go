package mag

type Categorie struct {
	ID          int
	Nom         string
	Description string
}

type Post struct {
	ID          int
	CategorieID int
	Texte       string
	DateHeure   string
	Photo       []byte
	Comments    []Comment
	Likes       int
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
	ID        int
	UserID    int
	PostID    int
	Texte     string
	DateHeure string
	Likes     int
}

type MessagesCree struct {
	ID           int    `json:"id"`
	PostID       int    `json:"post_id"`
	UserID       int    `json:"user_id"`
	DateCreation string `json:"date_creation"` 
}

type MessagesAime struct {
	ID        int    `json:"id"`
	PostID    int    `json:"post_id"`
	UserID    int    `json:"user_id"`
	DateAimee string `json:"date_aimee"` 
}