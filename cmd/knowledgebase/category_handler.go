package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/HeavyHorst/knowledgebase/pkg/models"
	"github.com/HeavyHorst/knowledgebase/pkg/ulid"
	"github.com/go-chi/chi/v5"
)

func categoryCtx(store CategoryGetter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cat, err := store.GetCategory(chi.URLParam(r, "categoryID"))
			if err != nil {
				logAndHTTPError(w, r, 404, http.StatusText(404), err)
				return
			}

			if cat.ID != "" {
				ctx := context.WithValue(r.Context(), contextKeyCategory, cat)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}

func listCategories(store CategoryLister) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			result []models.Category
			err    error
		)

		result, err = store.ListCategories(!isLoggedIn(r.Context()))
		if err != nil {
			logAndHTTPError(w, r, 500, err.Error(), err)
			return
		}

		writeJSON(w, r, result)
	}
}

func searchCategories(store CategorySearcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		if len(r.Form.Get("q")) < 3 {
			http.Error(w, "query too short", 422)
			return
		}

		categories, err := store.SearchCategories(r.Form.Get("q"), !isLoggedIn(r.Context()))
		if err != nil {
			logAndHTTPError(w, r, 500, err.Error(), err)
			return
		}

		writeJSON(w, r, categories)
	}
}

func listCategoriesForCategory(store CategoryLister) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := store.ListCategoriesForCategory(chi.URLParam(r, "categoryID"), !isLoggedIn(r.Context()))
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}

		writeJSON(w, r, result)
	}
}

func getCategory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cat, ok := ctx.Value(contextKeyCategory).(models.Category)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	if !cat.Public && !isLoggedIn(r.Context()) {
		http.Error(w, "Unauthorized", 401)
		return
	}

	writeJSON(w, r, cat)
}

func createCategory(store CategoryCreator) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var cat models.Category

		err := json.NewDecoder(r.Body).Decode(&cat)
		if err != nil {
			logAndHTTPError(w, r, 500, err.Error(), err)
			return
		}

		cat.ID = ulid.GetULID()

		err = store.CreateCategory(cat)
		if err != nil {
			logAndHTTPError(w, r, 500, err.Error(), err)
			return
		}

		w.Header().Set("Location", "/api/categories/"+cat.ID)
		w.WriteHeader(http.StatusCreated)
	}
}

func updateCategory(store CategoryUpdater) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		cat, ok := ctx.Value(contextKeyCategory).(models.Category)
		id := cat.ID
		if !ok {
			http.Error(w, http.StatusText(422), 422)
			return
		}

		err := json.NewDecoder(r.Body).Decode(&cat)
		if err != nil {
			logAndHTTPError(w, r, 500, err.Error(), err)
			return
		}

		cat.ID = id
		// update store
		err = store.UpdateCategory(cat)
		if err != nil {
			logAndHTTPError(w, r, 500, err.Error(), err)
			return
		}

		w.WriteHeader(http.StatusNotModified)
	}
}

func deleteCategory(store CategoryDeleter) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		cat, ok := ctx.Value(contextKeyCategory).(models.Category)
		if !ok {
			http.Error(w, http.StatusText(422), 422)
			return
		}

		err := store.DeleteCategory(cat)
		if err != nil {
			logAndHTTPError(w, r, 500, err.Error(), err)
			return
		}

		writeJSON(w, r, cat)
	}
}
