package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type Pet struct {
	ID        string
	Name      string
	Species   string
	Rarity    string
	Shiny     bool
	Level     int
	XP        int
	Evolution int
	Prestige  int
	Stats     Stats
	Active    bool
	HatchedAt time.Time
	UpdatedAt time.Time
}

type Stats struct {
	Wit     int `json:"wit"`
	Depth   int `json:"depth"`
	Stamina int `json:"stamina"`
	Luck    int `json:"luck"`
	Attune  int `json:"attune"`
}

type Commit struct {
	ID           int
	PetID        string
	Repo         string
	LinesChanged int
	FilesTouched int
	XPEarned     int
	DropItem     string
	CreatedAt    time.Time
}

type InventoryItem struct {
	Item     string
	Quantity int
}

type Store struct {
	db *sql.DB
}

func dbPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".petd")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "pets.db"), nil
}

func New() (*Store, error) {
	path, err := dbPath()
	if err != nil {
		return nil, fmt.Errorf("finding db path: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening db: %w", err)
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrating: %w", err)
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS pets (
			id          TEXT PRIMARY KEY,
			name        TEXT NOT NULL,
			species     TEXT NOT NULL,
			rarity      TEXT NOT NULL,
			shiny       INTEGER DEFAULT 0,
			level       INTEGER DEFAULT 1,
			xp          INTEGER DEFAULT 0,
			evolution   INTEGER DEFAULT 1,
			prestige    INTEGER DEFAULT 0,
			wit         INTEGER DEFAULT 0,
			depth       INTEGER DEFAULT 0,
			stamina     INTEGER DEFAULT 0,
			luck        INTEGER DEFAULT 0,
			attune      INTEGER DEFAULT 0,
			active      INTEGER DEFAULT 0,
			hatched_at  DATETIME,
			updated_at  DATETIME
		);

		CREATE TABLE IF NOT EXISTS commits (
			id              INTEGER PRIMARY KEY AUTOINCREMENT,
			pet_id          TEXT REFERENCES pets(id),
			repo            TEXT,
			lines_changed   INTEGER,
			files_touched   INTEGER,
			xp_earned       INTEGER,
			drop_item       TEXT,
			created_at      DATETIME
		);

		CREATE TABLE IF NOT EXISTS inventory (
			item        TEXT PRIMARY KEY,
			quantity    INTEGER DEFAULT 1
		);
	`)
	return err
}

func (s *Store) CreatePet(p *Pet) error {
	_, err := s.db.Exec(`
		INSERT INTO pets (id, name, species, rarity, shiny, level, xp, evolution, prestige,
			wit, depth, stamina, luck, attune, active, hatched_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.Species, p.Rarity, p.Shiny,
		p.Level, p.XP, p.Evolution, p.Prestige,
		p.Stats.Wit, p.Stats.Depth, p.Stats.Stamina, p.Stats.Luck, p.Stats.Attune,
		p.Active, p.HatchedAt, p.UpdatedAt,
	)
	return err
}

func (s *Store) GetActivePet() (*Pet, error) {
	row := s.db.QueryRow(`SELECT id, name, species, rarity, shiny, level, xp, evolution, prestige,
		wit, depth, stamina, luck, attune, active, hatched_at, updated_at
		FROM pets WHERE active = 1 LIMIT 1`)
	return scanPet(row)
}

func (s *Store) GetPetByName(name string) (*Pet, error) {
	row := s.db.QueryRow(`SELECT id, name, species, rarity, shiny, level, xp, evolution, prestige,
		wit, depth, stamina, luck, attune, active, hatched_at, updated_at
		FROM pets WHERE name = ? LIMIT 1`, name)
	return scanPet(row)
}

func (s *Store) HasAnyPet() (bool, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM pets`).Scan(&count)
	return count > 0, err
}

func (s *Store) SetActivePet(id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE pets SET active = 0`); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE pets SET active = 1 WHERE id = ?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) UpdatePetXP(id string, xp, level int) error {
	_, err := s.db.Exec(`UPDATE pets SET xp = ?, level = ?, updated_at = ? WHERE id = ?`,
		xp, level, time.Now(), id)
	return err
}

func (s *Store) RecordCommit(c *Commit) error {
	_, err := s.db.Exec(`
		INSERT INTO commits (pet_id, repo, lines_changed, files_touched, xp_earned, drop_item, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		c.PetID, c.Repo, c.LinesChanged, c.FilesTouched, c.XPEarned, c.DropItem, c.CreatedAt,
	)
	return err
}

func (s *Store) GetStreakDays() (int, error) {
	rows, err := s.db.Query(`
		SELECT DISTINCT date(created_at) as d FROM commits
		ORDER BY d DESC LIMIT 30`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var dates []string
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			return 0, err
		}
		dates = append(dates, d)
	}

	if len(dates) == 0 {
		return 0, nil
	}

	today := time.Now().Format("2006-01-02")
	streak := 0

	for i, d := range dates {
		expected := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		if i == 0 && d != today {
			expected = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
			if d != expected {
				return 0, nil
			}
			streak = 1
			continue
		}
		if i == 0 {
			streak = 1
			continue
		}
		if d == expected || (streak == 1 && i == 1 && dates[0] != today) {
			streak++
		} else {
			break
		}
	}

	return streak, nil
}

func (s *Store) GetTodayCommitCount() (int, error) {
	today := time.Now().Format("2006-01-02")
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM commits WHERE date(created_at) = ?`, today).Scan(&count)
	return count, err
}

func scanPet(row *sql.Row) (*Pet, error) {
	p := &Pet{}
	var shiny, active int
	err := row.Scan(
		&p.ID, &p.Name, &p.Species, &p.Rarity, &shiny,
		&p.Level, &p.XP, &p.Evolution, &p.Prestige,
		&p.Stats.Wit, &p.Stats.Depth, &p.Stats.Stamina, &p.Stats.Luck, &p.Stats.Attune,
		&active, &p.HatchedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	p.Shiny = shiny == 1
	p.Active = active == 1
	return p, nil
}
