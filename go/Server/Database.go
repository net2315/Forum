package server

import (
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

func GetCategories() ([]Categorie, error) {
    rows, err := db.Query("SELECT id, nom, description FROM categorie")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var categories []Categorie
    for rows.Next() {
        var category Categorie
        if err := rows.Scan(&category.ID, &category.Nom, &category.Description); err != nil {
            return nil, err
        }
        categories = append(categories, category)
    }
    return categories, nil
}

type Categorie struct {
    ID          int
    Nom         string
    Description string
}
