package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

type SnippetModel struct {
	DB *sql.DB
}

func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {
	stmt := fmt.Sprintf(`
    INSERT INTO snippets (title, content, created, expires)
    VALUES ($1, $2, NOW() AT TIME ZONE 'UTC', (NOW() AT TIME ZONE 'UTC') + INTERVAL '%d days')
		RETURNING id
`, expires)

	row := m.DB.QueryRow(stmt, title, content)

	var id int
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (m *SnippetModel) Get(id int) (Snippet, error) {
	stmt :=
		`SELECT id, title, content, created, expires FROM snippets
			WHERE expires > (NOW() AT TIME ZONE 'UTC')
			AND id = $1
		`

	row := m.DB.QueryRow(stmt, id)

	var s Snippet

	err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return Snippet{}, ErrNoRecord
		default:
			return Snippet{}, err
		}
	}

	return s, nil
}

func (m *SnippetModel) Latest() ([]Snippet, error) {
	stmt :=
		`SELECT id, title, content, created, expires FROM snippets
			WHERE expires > (NOW() AT TIME ZONE 'UTC')
			ORDER BY id DESC
			LIMIT 10
		`
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var snippets []Snippet

	for rows.Next() {
		var s Snippet
		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}
		snippets = append(snippets, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return snippets, nil

}

func (s Snippet) FormattedCreated() string {
	return s.Created.UTC().Format("2006-01-02 15:04:05.00")
}

func (s Snippet) FormattedExpires() string {
	return s.Expires.UTC().Format("2006-01-02 15:04:05.00")
}
