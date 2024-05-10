// Package lagoondb implements a client for the Lagoon API database.
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

// groupProjectMapping maps Lagoon group ID to project ID.
// This type is only used for database unmarshalling.
type groupProjectMapping struct {
	GroupID   string `db:"group_id"`
	ProjectID int    `db:"project_id"`
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

// Projects returns a slice of all Projects in the Lagoon API DB.
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

// GroupProjectsMap returns a map of Group (UU)IDs to Project IDs.
// This denotes Project Group membership in Lagoon.
func (c *Client) GroupProjectsMap(
	ctx context.Context,
) (map[string][]int, error) {
	var gpms []groupProjectMapping
	err := c.db.SelectContext(ctx, &gpms, `
	SELECT group_id, project_id
	FROM kc_group_projects`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoResult
		}
		return nil, err
	}
	groupProjectsMap := map[string][]int{}
	// no need to check for duplicates here since the table has:
	// UNIQUE KEY `group_project` (`group_id`,`project_id`)
	for _, gpm := range gpms {
		groupProjectsMap[gpm.GroupID] =
			append(groupProjectsMap[gpm.GroupID], gpm.ProjectID)
	}
	return groupProjectsMap, nil
}
