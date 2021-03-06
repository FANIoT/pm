package actions

import (
	"context"
	"strconv"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	contenttype "github.com/gobuffalo/mw-contenttype"
	paramlogger "github.com/gobuffalo/mw-paramlogger"
	mgo "github.com/mongodb/mongo-go-driver/mongo"
	validator "gopkg.in/go-playground/validator.v9"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gobuffalo/x/sessions"
	"github.com/rs/cors"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")
var app *buffalo.App
var db *mgo.Database
var validate *validator.Validate

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
func App() *buffalo.App {
	if app == nil {
		app = buffalo.New(buffalo.Options{
			Env:          ENV,
			SessionStore: sessions.Null{},
			PreWares: []buffalo.PreWare{
				cors.Default().Handler,
			},
			SessionName: "_pm_session",
		})

		// If no content type is sent by the client
		// the application/json will be set, otherwise the client's
		// content type will be used.
		app.Use(contenttype.Add("application/json"))

		// create mongodb connection
		url := envy.Get("DB_URL", "mongodb://172.18.0.1:27017")
		client, err := mgo.NewClient(url)
		if err != nil {
			buffalo.NewLogger("fatal").Fatalf("DB new client error: %s", err)
		}
		if err := client.Connect(context.Background()); err != nil {
			buffalo.NewLogger("fatal").Fatalf("DB connection error: %s", err)
		}
		db = client.Database("i1820")

		// validator
		validate = validator.New()

		if ENV == "development" {
			app.Use(paramlogger.ParameterLogger)
		}

		// prometheus collector
		rds := prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "pm",
				Name:      "request_duration_seconds",
				Help:      "A histogram of latencies for requests.",
			},
			[]string{"path", "method", "code"},
		)

		prometheus.NewGoCollector()
		prometheus.MustRegister(rds)

		app.Use(func(next buffalo.Handler) buffalo.Handler {
			return func(c buffalo.Context) error {
				now := time.Now()

				defer func() {
					ws := c.Response().(*buffalo.Response)
					req := c.Request()

					rds.With(prometheus.Labels{
						"path":   req.URL.String(),
						"code":   strconv.Itoa(ws.Status),
						"method": req.Method,
					}).Observe(time.Since(now).Seconds())
				}()

				return next(c)
			}
		})

		// Routes
		app.GET("/about", AboutHandler)
		api := app.Group("/api")
		{

			// /projects
			pr := ProjectsResource{}
			api.Resource("/projects", pr)
			api.GET("/projects/{project_id}/logs", pr.Logs)
			api.GET("/projects/{project_id}/recreate", pr.Recreate)

			pg := api.Group("/projects/{project_id}")
			{
				// /projects/{project_id}/things
				tr := ThingsResource{}
				pg.Resource("/things", tr)
				pg.POST("/things/geo", tr.GeoWithin)
				pg.POST("/things/tags", tr.HaveTags)
				pg.GET("/things/{thing_id}/{t:(?:activate|deactivate)}", tr.Activation)

				// /projects/{project_id}/things/{thing_id}/tokens
				kr := TokensResource{}
				pg.GET("/things/{thing_id}/tokens", kr.Create)
				pg.DELETE("/things/{thing_id}/tokens/{token}", kr.Destroy)

				// projects/{project_id}/things/{thing_id}/assets
				pg.Resource("/things/{thing_id}/assets", AssetsResource{})

				// /projects/{project_id}/things/{thing_id}/connectivities
				pg.Resource("/things/{thing_id}/connectivities", ConnectivitiesResource{})

				// /projects/{project_id}/things/{thing_id}/tags
				gr := TagsResource{}
				pg.POST("/things/{thing_id}/tags", gr.Create)
				pg.GET("/things/{thing_id}/tags", gr.List)
			}

			// /runners
			api.GET("/runners/pull", PullHandler)
			api.ANY("/runners/{project_id}/{path:.+}", RunnersHandler)
		}
		app.GET("/metrics", buffalo.WrapHandler(promhttp.Handler()))
	}

	return app
}
