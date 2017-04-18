package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/HeavyHorst/knowledgebase/pkg/models"
	"github.com/HeavyHorst/knowledgebase/pkg/ulid"
	"github.com/pressly/chi"
)

func articleCtx(store ArticleGetter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			art, err := store.GetArticle(chi.URLParam(r, "articleID"))
			if err != nil {
				logAndHTTPError(w, r, 404, http.StatusText(404), err)
				return
			}

			if art.ID != "" {
				ctx := context.WithValue(r.Context(), contextKeyArticle, art)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}

func listArticles(store ArticleLister) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		offset := 0
		limit := 20

		err := r.ParseForm()
		if err != nil {
			logAndHTTPError(w, r, 500, err.Error(), err)
			return
		}

		if r.Form.Get("offset") != "" {
			offset, err = strconv.Atoi(r.Form.Get("offset"))
			if err != nil {
				logAndHTTPError(w, r, 500, err.Error(), err)
				return
			}
		}

		if r.Form.Get("limit") != "" {
			limit, err = strconv.Atoi(r.Form.Get("limit"))
			if err != nil {
				logAndHTTPError(w, r, 500, err.Error(), err)
				return
			}
		}

		result, totalCount, err := store.ListArticles(limit, offset)
		if err != nil {
			logAndHTTPError(w, r, 500, err.Error(), err)
			return
		}

		w.Header().Set("X-Total-Count", fmt.Sprintf("%d", totalCount))
		writeJSON(w, r, result)
	}
}

func createArticle(store ArticleCreator) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, ok := ctx.Value(contextKeyCurrentUser).(models.User)
		if !ok {
			http.Error(w, http.StatusText(422), 422)
			return
		}

		var art models.Article

		err := json.NewDecoder(r.Body).Decode(&art)
		defer r.Body.Close()
		if err != nil {
			logAndHTTPError(w, r, 500, err.Error(), err)
			return
		}

		art.ID = ulid.GetULID()

		err = store.CreateArticle(art, user)
		if err != nil {
			logAndHTTPError(w, r, 500, err.Error(), err)
			return
		}

		w.Header().Set("Location", "/api/articles/"+art.ID)
		w.WriteHeader(http.StatusCreated)
	}
}

func searchArticles(store ArticleSearcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		if len(r.Form.Get("q")) < 3 {
			http.Error(w, "query too short", 422)
			return
		}

		articles, err := store.SearchArticles(r.Form.Get("q"))
		if err != nil {
			logAndHTTPError(w, r, 500, err.Error(), err)
			return
		}

		writeJSON(w, r, articles)
	}
}

func listArticlesForCategory(store ArticleLister) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := store.ListArticlesForCategory(chi.URLParam(r, "categoryID"))
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		writeJSON(w, r, result)
	}
}

func getArticleHistory(store ArticleHistoryGetter) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := store.GetArticleHistory(chi.URLParam(r, "articleID"))
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		writeJSON(w, r, result)
	}
}

func getArticle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	art, ok := ctx.Value(contextKeyArticle).(models.Article)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	writeJSON(w, r, art)
}

func updateArticle(store ArticleUpdater) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user, ok := ctx.Value(contextKeyCurrentUser).(models.User)
		if !ok {
			http.Error(w, http.StatusText(422), 422)
			return
		}

		art, ok := ctx.Value(contextKeyArticle).(models.Article)
		id := art.ID
		if !ok {
			http.Error(w, http.StatusText(422), 422)
			return
		}

		err := json.NewDecoder(r.Body).Decode(&art)
		defer r.Body.Close()
		if err != nil {
			logAndHTTPError(w, r, 500, err.Error(), err)
			return
		}

		art.ID = id
		err = store.UpdateArticle(art, user)
		if err != nil {
			logAndHTTPError(w, r, 500, err.Error(), err)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func deleteArticle(store ArticleDeleter) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		art, ok := ctx.Value(contextKeyArticle).(models.Article)
		if !ok {
			http.Error(w, http.StatusText(422), 422)
			return
		}

		err := store.DeleteArticle(art)
		if err != nil {
			logAndHTTPError(w, r, 500, err.Error(), err)
			return
		}

		writeJSON(w, r, art)
	}
}
