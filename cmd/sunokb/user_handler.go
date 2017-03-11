package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/HeavyHorst/sunoKB/pkg/auth"
	"github.com/HeavyHorst/sunoKB/pkg/models"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/pressly/chi"
	"github.com/timshannon/bolthold"
)

func userCtx(store UserGetter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := store.GetUser(chi.URLParam(r, "username"))
			if err != nil {
				logAndHTTPError(w, r, 404, http.StatusText(404), err)
				return
			}
			user.Password = ""

			if user.Username != "" {
				ctx := context.WithValue(r.Context(), contextKeyUser, user)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}

func requireTokenAuthentication(store UserGetter, tokenGenerator auth.TokenGenerator) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var token string
			var user models.User

			auth := r.Header.Get("Authorization")

			if len(auth) > 7 && auth[:6] == "Bearer" {
				token = auth[7:]
			}

			_, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
				// Check the signing method
				var err error
				if t.Method.Alg() != "HS256" {
					return nil, fmt.Errorf("Unexpected jwt signing method=%v", t.Header["alg"])
				}

				username, ok := t.Claims.(jwt.MapClaims)["sub"].(string)
				if ok {
					user, err = store.GetUser(username)
					if err != nil {
						return nil, errors.Wrap(err, "couldn't get key for validating the token")
					}
				}

				return append([]byte(user.Password), tokenGenerator.GetSecret()...), nil
			})
			if err != nil {
				logAndHTTPError(w, r, 401, err.Error(), err)
				return
			}

			ctx := context.WithValue(r.Context(), contextKeyCurrentUser, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, ok := ctx.Value(contextKeyCurrentUser).(models.User)
		if !ok {
			http.Error(w, http.StatusText(422), 422)
			return
		}

		if !user.IsAdmin {
			http.Error(w, "Unauthorized", 401)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func authenticate(store Authenticator, tokenGenerator auth.TokenGenerator) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.PostFormValue("username")
		password := r.PostFormValue("password")

		user, err := store.Authenticate(username, password)
		if err != nil {
			logAndHTTPError(w, r, 401, "Unauthorized", err)
			return
		}

		if user == nil {
			http.Error(w, "Unauthorized", 401)
			return
		}

		token, err := tokenGenerator.GenerateToken(user.Username, user.Password, user.IsAdmin)
		if err != nil {
			logAndHTTPError(w, r, 500, err.Error(), err)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err = json.NewEncoder(w).Encode(map[string]string{"token": token}); err != nil {
			logAndHTTPError(w, r, 500, err.Error(), err)
			return
		}
	}
}

func refreshToken(store auth.TokenGenerator) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, ok := ctx.Value(contextKeyCurrentUser).(models.User)
		if !ok {
			http.Error(w, http.StatusText(422), 422)
			return
		}

		token, err := store.GenerateToken(user.Username, user.Password, user.IsAdmin)
		if err != nil {
			logAndHTTPError(w, r, 500, err.Error(), err)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err = json.NewEncoder(w).Encode(map[string]string{"token": token}); err != nil {
			logAndHTTPError(w, r, 500, err.Error(), err)
			return
		}
	}
}

func listUsers(store UserLister) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := store.ListUsers()
		if err != nil {
			logAndHTTPError(w, r, 500, err.Error(), err)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err = json.NewEncoder(w).Encode(result); err != nil {
			logAndHTTPError(w, r, 500, err.Error(), err)
			return
		}
	}
}

func getUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value(contextKeyUser).(models.User)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		logAndHTTPError(w, r, 500, err.Error(), err)
		return
	}
}

func createUser(store UserCreator) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.User

		err := json.NewDecoder(r.Body).Decode(&user)
		defer r.Body.Close()
		if err != nil {
			logAndHTTPError(w, r, 500, err.Error(), err)
			return
		}

		err = store.CreateUser(user, user.Password)
		if err != nil {
			switch errors.Cause(err) {
			case bolthold.ErrKeyExists:
				logAndHTTPError(w, r, 409, err.Error(), err)
				return
			default:
				logAndHTTPError(w, r, 500, err.Error(), err)
				return
			}
		}

		w.Header().Set("Location", "/api/users/"+user.Username)
		w.WriteHeader(http.StatusCreated)
	}
}

func updateUser(store UserUpdater) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, ok := ctx.Value(contextKeyUser).(models.User)
		if !ok {
			http.Error(w, http.StatusText(422), 422)
			return
		}

		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			logAndHTTPError(w, r, 500, err.Error(), err)
			return
		}

		// update store
		err = store.UpdateUser(user, user.Password)
		if err != nil {
			logAndHTTPError(w, r, 500, err.Error(), err)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func deleteUser(store UserDeleter) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, ok := ctx.Value(contextKeyUser).(models.User)
		if !ok {
			http.Error(w, http.StatusText(422), 422)
			return
		}

		err := store.DeleteUser(user)
		if err != nil {
			logAndHTTPError(w, r, 500, err.Error(), err)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err = json.NewEncoder(w).Encode(user); err != nil {
			logAndHTTPError(w, r, 500, err.Error(), err)
			return
		}
	}
}
