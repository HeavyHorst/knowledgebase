package main

import (
	"flag"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/HeavyHorst/knowledgebase/pkg/auth"
	"github.com/HeavyHorst/knowledgebase/pkg/log"
	"github.com/HeavyHorst/knowledgebase/pkg/storage"
	"github.com/Sirupsen/logrus"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/goware/cors"
	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

const jwtSecret = "thisIsSuperSecret"

var printVersionAndExit bool

func init() {
	flag.BoolVar(&printVersionAndExit, "version", false, "print version and exit")
}

func main() {
	flag.Parse()
	if printVersionAndExit {
		printVersion()
		return
	}

	logger := logrus.New()
	//logger.Formatter = &logrus.JSONFormatter{}
	logger.Formatter = &prefixed.TextFormatter{DisableSorting: false}

	tokenGenerator := &auth.JWTTokenGenerator{
		Method: jwt.SigningMethodHS256,
		Exp:    time.Hour * 24,
		Secret: []byte(jwtSecret),
	}

	store, err := storage.NewBoltHoldClient("data/kb.db")
	if err != nil {
		logger.Fatalln(err)
	}

	r := chi.NewRouter()
	rta := requireTokenAuthentication(store, tokenGenerator)

	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link", "Location"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	r.Use(cors.Handler)

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(log.NewStructuredLogger(logger))
	r.Use(middleware.Recoverer)

	// static FileServer
	workDir, _ := os.Getwd()
	filesDir := filepath.Join(workDir, "static")
	r.FileServer("/static", http.Dir(filesDir))

	// admin and index
	r.Get("/", fileHandler("static/templates/index.html"))
	r.Get("/categories", fileHandler("static/templates/index.html"))
	r.Get("/articles/*", fileHandler("static/templates/index.html"))
	r.Get("/image/:imageHash", imageHandler(store))

	r.Route("/admin", func(r chi.Router) {
		r.Get("/", fileHandler("static/templates/admin.html"))
		r.With(rta, requireAdmin).Get("/backup", backupDBHandler(store))
	})

	r.Route("/api", func(r chi.Router) {
		r.Route("/authorize", func(r chi.Router) {
			r.Post("/", authenticate(store, tokenGenerator))
			r.With(rta).Get("/refresh", refreshToken(tokenGenerator))
		})

		r.Route("/users", func(r chi.Router) {
			r.Use(rta)
			r.Get("/", listUsers(store))
			r.With(requireAdmin).Post("/", createUser(store))
			r.Route("/:username", func(r chi.Router) {
				r.Use(userCtx(store))
				r.Get("/", getUser)
				r.With(requireAdmin).Put("/", updateUser(store))
				r.With(requireAdmin).Delete("/", deleteUser(store))
			})
		})

		r.Route("/categories", func(r chi.Router) {
			r.Use(rta)
			r.Get("/", listCategories(store))
			r.With(requireUser).Post("/", createCategory(store))
			r.Get("/search", searchCategories(store))
			r.Get("/category/:categoryID", listCategoriesForCategory(store))
			r.Route("/:categoryID", func(r chi.Router) {
				r.Use(categoryCtx(store))
				r.Get("/", getCategory)
				r.With(requireUser).Put("/", updateCategory(store))
				r.With(requireUser).Delete("/", deleteCategory(store))
			})
		})

		r.Route("/articles", func(r chi.Router) {
			r.Use(rta)
			r.Get("/", listArticles(store))
			r.Post("/", createArticle(store))
			r.Get("/search", searchArticles(store))
			r.Get("/category/:categoryID", listArticlesForCategory(store))
			r.Route("/:articleID", func(r chi.Router) {
				r.Use(articleCtx(store))
				r.Get("/", getArticle)
				r.Get("/history", getArticleHistory(store))
				r.Put("/", updateArticle(store))
				r.Delete("/", deleteArticle(store))
			})
		})
	})

	http.ListenAndServe("127.0.0.1:3000", r)
}
