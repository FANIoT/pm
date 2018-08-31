/*
 * +===============================================
 * | Author:        Parham Alvani <parham.alvani@gmail.com>
 * |
 * | Creation Date: 18-11-2017
 * |
 * | File Name:     project.go
 * +===============================================
 */

package models

import (
	"context"
	"fmt"

	"github.com/I1820/pm/runner"
)

// Project represents structure of ISRC projects
type Project struct {
	Name   string        `json:"name" bson:"name"`
	User   string        `json:"user" bson:"user"` // project owner username (1995parham)
	Runner runner.Runner `json:"runner" bson:"runner"`
	Things []Thing       `json:"things" bson:"things"`
	Status bool          `json:"status" bson:"status"` // active/inactive

	Inspects interface{} `json:"inspects,omitempty" bson:"-"`
}

// NewProject creates new project with given name
func NewProject(ctx context.Context, user string, name string, envs []runner.Env) (*Project, error) {
	r, err := runner.New(ctx, fmt.Sprintf("%s_%s", name, user), envs)

	if err != nil {
		return nil, err
	}

	return &Project{
		Name:   name,
		User:   user,
		Runner: r,
		Things: make([]Thing, 0),
		Status: true,
	}, nil
}
