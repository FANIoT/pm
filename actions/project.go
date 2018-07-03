/*
 * +===============================================
 * | Author:        Parham Alvani <parham.alvani@gmail.com>
 * |
 * | Creation Date: 02-07-2018
 * |
 * | File Name:     actions/project.go
 * +===============================================
 */

package actions

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aiotrc/pm/project"
	"github.com/aiotrc/pm/runner"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/mongodb/mongo-go-driver/bson"
	mgo "github.com/mongodb/mongo-go-driver/mongo"
)

// ProjectsResource manages existing projects
type ProjectsResource struct {
	buffalo.Resource
}

// project request payload
type projectReq struct {
	Name string `json:"name" binding:"required"`
	// TODO adds docker constraints and envs
}

// List gets all projects. This function is mapped to the path
// GET /projects
func (v ProjectsResource) List(c buffalo.Context) error {
	ps := make([]project.Project, 0)

	cur, err := db.Collection("pm").Find(c, bson.NewDocument())
	if err != nil {
		return c.Error(http.StatusInternalServerError, err)
	}

	for cur.Next(context.Background()) {
		var p project.Project

		if err := cur.Decode(&p); err != nil {
			return c.Error(http.StatusInternalServerError, err)
		}

		ps = append(ps, p)
	}
	if err := cur.Close(context.Background()); err != nil {
		return c.Error(http.StatusInternalServerError, err)
	}

	return c.Render(http.StatusOK, r.JSON(ps))
}

// Create adds a project to the DB and creates its docker. This function is mapped to the
// path POST /projects
func (v ProjectsResource) Create(c buffalo.Context) error {
	var rq projectReq
	if err := c.Bind(&rq); err != nil {
		return c.Error(http.StatusBadRequest, err)
	}

	name := rq.Name

	p, err := project.New(name, []runner.Env{
		{Name: "MONGO_URL", Value: envy.Get("DB_URL", "mongodb://172.18.0.1:27017")},
	})
	if err != nil {
		return c.Error(http.StatusInternalServerError, err)
	}

	if _, err := db.Collection("pm").InsertOne(c, p); err != nil {
		return c.Error(http.StatusInternalServerError, err)
	}

	// numberOfCreatedProjects.Inc()

	return c.Render(http.StatusOK, r.JSON(p))
}

// Show gets the data for one project. This function is mapped to
// the path GET /projects/{project_id}
func (v ProjectsResource) Show(c buffalo.Context) error {
	name := c.Param("project_id")

	var p project.Project

	dr := db.Collection("pm").FindOne(context.Background(), bson.NewDocument(
		bson.EC.String("name", name),
	))

	if err := dr.Decode(&p); err != nil {
		if err == mgo.ErrNoDocuments {
			return c.Error(http.StatusNotFound, fmt.Errorf("Project %s not found", name))
		}
		return c.Error(http.StatusInternalServerError, err)
	}

	return c.Render(http.StatusOK, r.JSON(p))
}

// Destroy deletes a project from the DB and its docker. This function is mapped
// to the path DELETE /projects/{project_id}
func (v ProjectsResource) Destroy(c buffalo.Context) error {
	name := c.Param("project_id")

	var p project.Project

	dr := db.Collection("pm").FindOne(c, bson.NewDocument(
		bson.EC.String("name", name),
	))

	if err := dr.Decode(&p); err != nil {
		if err == mgo.ErrNoDocuments {
			return c.Error(http.StatusNotFound, fmt.Errorf("Project %s not found", name))
		}
		return c.Error(http.StatusInternalServerError, err)
	}

	if err := p.Runner.Remove(); err != nil {
		return c.Error(http.StatusInternalServerError, err)
	}

	if _, err := db.Collection("pm").DeleteOne(c, bson.NewDocument(
		bson.EC.String("name", name),
	)); err != nil {
		return c.Error(http.StatusInternalServerError, err)
	}

	return c.Render(http.StatusOK, r.JSON(p))
}
