package mocks

import "notes.fritz.box/internal/models"

var mockNote1 = models.Note{
	ID:      1,
	Title:   "Note 1",
	Content: "Note 1 Content",
}

var mockNote2 = models.Note{
	ID:      2,
	Title:   "Note 2",
	Content: "Note 2 Content",
}

type NoteModel struct{}

func (m *NoteModel) Insert(title string, content string) (int, error) {
	return 2, nil
}

func (m *NoteModel) Get(id int) (models.Note, error) {
	switch id {
	case 1:
		return mockNote1, nil
	case 2:
		return mockNote2, nil
	default:
		return models.Note{}, models.ErrNoRecord
	}
}

func (m *NoteModel) Modify(id int, title, content string) error {
	return nil
}

func (m *NoteModel) Delete(id int) error {
	return nil
}

func (m *NoteModel) All() ([]models.Note, error) {
	return []models.Note{mockNote1, mockNote2}, nil
}
