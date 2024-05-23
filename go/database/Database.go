package database

import (
	"Forum/go/mag"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

var db *sql.DB

func InitDB(filepath string) {
	var err error
	db, err = sql.Open("sqlite3", filepath)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
}

func GetCategories() ([]mag.Categorie, error) {
	rows, err := db.Query("SELECT id, nom, description FROM categorie")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []mag.Categorie
	for rows.Next() {
		var category mag.Categorie
		if err := rows.Scan(&category.ID, &category.Nom, &category.Description); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}
