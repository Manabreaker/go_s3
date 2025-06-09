package apiserver

import (
	"S3_project/auth/internal/app/model"
	"S3_project/auth/internal/app/store"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

const (
	authorization          = "Authorization"
	ctxKeyRequestID crtKey = iota
)

var (
	secretKey                   = []byte("secret")
	errIncorrectEmailOrPassword = errors.New("incorrect email or password")
	errNotAuthenticated         = errors.New("not authenticated")
)

type crtKey int8

type Server struct {
	router *mux.Router
	logger *zap.Logger
	store  store.Store
}

func NewServer(store store.Store, apiGatewayUrl string) *Server {
	logger, _ := zap.NewProduction()
	s := &Server{
		router: mux.NewRouter(),
		logger: logger,
		store:  store,
	}

	s.configureRouter(apiGatewayUrl)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) configureRouter(apiGatewayUrl string) {
	// 1. Опции CORS
	corsOpts := handlers.CORS(
		handlers.AllowCredentials(),
		handlers.AllowedOrigins([]string{apiGatewayUrl}),
		handlers.AllowedMethods([]string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		}),
		handlers.AllowedHeaders([]string{
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"Access-Control-Allow-Origin",
		}),
	)

	// 2. CORS-мидлвар самым первым
	s.router.Use(corsOpts)

	// 3. Лог мидлвары
	s.router.Use(s.setRequestID)
	s.router.Use(s.logRequest)

	// 4. Роуты для /account
	account := s.router.PathPrefix("/account").Subrouter()
	account.HandleFunc("/login", s.handleLogin()).Methods(http.MethodPost, http.MethodOptions)
	account.HandleFunc("/register", s.handleRegister()).Methods(http.MethodPost, http.MethodOptions)
	account.HandleFunc("/logout", s.handleLogout()).Methods(http.MethodPost, http.MethodOptions)
}

func (s *Server) handleLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(authorization)
		if err != nil {
			s.error(w, r, http.StatusUnauthorized, errNotAuthenticated)
			return
		}

		tokenString := cookie.Value
		token, err := parseToken(tokenString)
		if err != nil || !token.Valid {
			s.error(w, r, http.StatusUnauthorized, errNotAuthenticated)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			s.error(w, r, http.StatusUnauthorized, err)
			return
		}

		dataMap, ok := claims["data"].(map[string]interface{})

		if !ok {
			s.error(w, r, http.StatusUnauthorized, err)
			return
		}

		stringUserID, ok := dataMap["user_id"].(string)
		if !ok {
			s.error(w, r, http.StatusUnauthorized, err)
			return
		}

		userID, err := strconv.Atoi(stringUserID)
		if err != nil {
			s.error(w, r, http.StatusUnauthorized, err)
			return
		}

		_, err = s.store.User().Find(userID)
		if err != nil {
			s.error(w, r, http.StatusUnauthorized, errNotAuthenticated)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     authorization,
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
		})
		s.respond(w, r, http.StatusOK, map[string]string{authorization: tokenString})
	}
}

func (s *Server) handleRegister() http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		u := &model.User{
			Email:    req.Email,
			Password: req.Password,
		}
		if err := s.store.User().Create(u); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		u.Sanitize()
		s.respond(w, r, http.StatusCreated, u)
	}
}

func (s *Server) handleLogin() http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}
		u, err := s.store.User().FindByEmail(req.Email)
		if err != nil || !u.ComparePassword(req.Password) {
			s.error(w, r, http.StatusUnauthorized, errIncorrectEmailOrPassword)
			return
		}

		tokenString, err := generateToken(u.ID)

		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     authorization,
			Value:    tokenString,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
		})
		s.respond(w, r, http.StatusOK, map[string]string{authorization: tokenString})
	}
}

// Устанавливаем уникальный идентификатор запроса, работает, только если запрос не содержит его
func (s *Server) setRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyRequestID, id)))
	})
}

func (s *Server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := s.logger.With(
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("request_id", r.Context().Value(ctxKeyRequestID).(string)),
		)
		logger.Info(fmt.Sprintf("started %s %s", r.Method, r.RequestURI))

		start := time.Now()
		rw := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(rw, r)

		logger.Info(
			"completed with",
			zap.Int("status code", rw.code),
			zap.String("status text\n", http.StatusText(rw.code)),
			zap.Duration("duration", time.Since(start)),
		)
	})
}

func (s *Server) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	s.logger.Error("error", zap.Error(err))
	s.respond(w, r, code, map[string]string{"error": err.Error()})
}

func (s *Server) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func parseToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
}

func generateToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"iss": "issuer",
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"data": map[string]string{
			"user_id": strconv.Itoa(userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}
