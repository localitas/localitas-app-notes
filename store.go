package notes

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	_ "github.com/localitas/localitas-go"
)

const DatabaseName = "notes"

type Store struct {
	db           *sql.DB
	PythonRunner *PythonRunner
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func OpenStore(coreURL, dbID, token string) (*Store, error) {
	dsn := fmt.Sprintf("%s?database_id=%s&token=%s", coreURL, dbID, token)
	db, err := sql.Open("localitas", dsn)
	if err != nil {
		return nil, fmt.Errorf("open localitas db: %w", err)
	}
	return NewStore(db), nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Create(ctx context.Context, userID, title, content string) (*Note, error) {
	id := newNoteID()
	now := time.Now().UTC().Unix()
	summary := extractSummary(content)

	_, err := s.db.ExecContext(ctx,
		"INSERT INTO notes (id, user_id, title, content, tags, summary, created_at, updated_at) VALUES (?, ?, ?, ?, '[]', ?, ?, ?)",
		id, userID, title, content, summary, now, now)
	if err != nil {
		return nil, fmt.Errorf("insert note: %w", err)
	}

	return &Note{
		ID:        id,
		Title:     title,
		Content:   content,
		Tags:      "[]",
		Summary:   summary,
		CreatedAt: time.Unix(now, 0),
		UpdatedAt: time.Unix(now, 0),
	}, nil
}

func (s *Store) Get(ctx context.Context, id string) (*Note, error) {
	var n Note
	var createdAt, updatedAt int64
	err := s.db.QueryRowContext(ctx,
		"SELECT id, title, content, tags, summary, created_at, updated_at FROM notes WHERE id = ?", id,
	).Scan(&n.ID, &n.Title, &n.Content, &n.Tags, &n.Summary, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("note %s not found", id)
	}
	n.CreatedAt = time.Unix(createdAt, 0)
	n.UpdatedAt = time.Unix(updatedAt, 0)
	return &n, nil
}

func (s *Store) List(ctx context.Context, userID string, limit, offset int) ([]*Note, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, title, content, tags, summary, created_at, updated_at FROM notes WHERE user_id = ? ORDER BY updated_at DESC LIMIT ? OFFSET ?",
		userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	defer rows.Close()

	out := make([]*Note, 0)
	for rows.Next() {
		var n Note
		var createdAt, updatedAt int64
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.Tags, &n.Summary, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		n.CreatedAt = time.Unix(createdAt, 0)
		n.UpdatedAt = time.Unix(updatedAt, 0)
		out = append(out, &n)
	}
	return out, nil
}

func (s *Store) Update(ctx context.Context, id, title, content string) error {
	now := time.Now().UTC().Unix()
	summary := extractSummary(content)
	_, err := s.db.ExecContext(ctx,
		"UPDATE notes SET title = ?, content = ?, summary = ?, updated_at = ? WHERE id = ?",
		title, content, summary, now, id)
	return err
}

func (s *Store) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM notes WHERE id = ?", id)
	return err
}

func (s *Store) Search(ctx context.Context, userID, query string, limit int) ([]*Note, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT n.id, n.title, n.content, n.tags, n.summary, n.created_at, n.updated_at
		FROM notes n
		JOIN notes_fts ON n.rowid = notes_fts.rowid
		WHERE notes_fts MATCH ? AND n.user_id = ?
		ORDER BY rank
		LIMIT ?
	`, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("search notes: %w", err)
	}
	defer rows.Close()

	out := make([]*Note, 0)
	for rows.Next() {
		var n Note
		var createdAt, updatedAt int64
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.Tags, &n.Summary, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		n.CreatedAt = time.Unix(createdAt, 0)
		n.UpdatedAt = time.Unix(updatedAt, 0)
		out = append(out, &n)
	}
	return out, nil
}

func newNoteID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

func extractSummary(content string) string {
	lines := strings.SplitN(content, "\n", 4)
	var parts []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" && !strings.HasPrefix(l, "#") {
			parts = append(parts, l)
		}
	}
	s := strings.Join(parts, " ")
	if len(s) > 100 {
		s = s[:100] + "..."
	}
	return s
}
