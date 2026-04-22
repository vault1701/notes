package models

import (
	"database/sql"
	"errors"
	"html/template"
)

type NoteModelInterface interface {
	Insert(title string, content string) (int, error)
	Get(id int) (Note, error)
	Modify(id int, title, content string) error
	Delete(id int) error
	All() ([]Note, error)
}

type Note struct {
	ID        int
	Title     string
	Content   string
	ContentMD template.HTML
}

type NoteModel struct {
	DB *sql.DB
}

func (m *NoteModel) Insert(title string, content string) (int, error) {
	stmt := `INSERT INTO notes (title, content) VALUES(?, ?)`

	result, err := m.DB.Exec(stmt, title, content)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *NoteModel) Get(id int) (Note, error) {
	var n Note

	err := m.DB.QueryRow("SELECT id, title, content from notes where id = ?", id).Scan(&n.ID, &n.Title, &n.Content)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Note{}, ErrNoRecord
		} else {
			return Note{}, err
		}
	}

	return n, nil
}

func (m *NoteModel) Modify(id int, title, content string) error {
	stmt := `UPDATE notes set title = ?, content = ? where id = ?`

	_, err := m.DB.Exec(stmt, title, content, id)
	if err != nil {
		return err
	}

	return nil
}

func (m *NoteModel) Delete(id int) error {
	stmt := `DELETE FROM notes where id = ?`

	_, err := m.DB.Exec(stmt, id)
	if err != nil {
		return err
	}

	return nil
}

func (m *NoteModel) All() ([]Note, error) {
	stmt := `SELECT id, title, content FROM notes ORDER BY id ASC;`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var notes []Note

	for rows.Next() {
		var n Note

		err = rows.Scan(&n.ID, &n.Title, &n.Content)
		if err != nil {
			return nil, err
		}

		notes = append(notes, n)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}
