package lagoondb

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

// Client is a Lagoon API-DB client
type Client struct {
	db *sqlx.DB
}

// Project is a Lagoon project.
type Project struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

// ErrNoResult is returned by client methods if there is no result.
var ErrNoResult = errors.New("no rows in result set")

// NewClient returns a new Lagoon DB Client.
func NewClient(ctx context.Context, dsn string) (*Client, error) {
	db, err := sqlx.ConnectContext(ctx, "mysql", dsn)
	if err != nil {
		return nil, err
	}
	// https://github.com/go-sql-driver/mysql#important-settings
	db.SetConnMaxLifetime(4 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	return &Client{db: db}, nil
}

// Projects returns the Environment associated with the given
// Namespace name (on Openshift this is the project name).
func (c *Client) Projects(ctx context.Context) ([]Project, error) {
	// run query
	var projects []Project
	err := c.db.SelectContext(ctx, &projects, `
	SELECT id, name
	FROM project`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoResult
		}
		return nil, err
	}
	return projects, nil
}
