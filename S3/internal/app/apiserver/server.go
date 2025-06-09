package apiserver

import (
	"S3_project/S3/internal/app/store/filestore"
	"context"
	"database/sql"
	"encoding/base64"
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
	filesNames           = "files"
	fileDates            = "dates"
	authorization        = "Authorization"
	ctxKeyUserId  crtKey = iota
	ctxKeyRequestID
)

var (
	secretKey              = []byte("secret")
	errInternalServerError = errors.New("internal server error")
	errNotAuthenticated    = errors.New("not authenticated")
	errEmptyFile           = errors.New("filename is empty")
	errFileAlreadyExist    = errors.New("file already exist")
	errDataBaseError       = errors.New("database error")
	errFileNotFound        = errors.New("file not found")
)

type crtKey int8

type Server struct {
	router    *mux.Router
	logger    *zap.Logger
	filestore filestore.FileStore
}

func NewServer(filestore *filestore.FileStore, apiGatewayUrl string) *Server {
	logger, _ := zap.NewProduction()
	s := &Server{
		router:    mux.NewRouter(),
		logger:    logger,
		filestore: *filestore,
	}

	s.configureRouter(apiGatewayUrl)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) configureRouter(apiGatewayUrl string) {
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

	// 2. Подключаем CORS-мидлвар самым первым
	s.router.Use(corsOpts)
	s.router.Use(s.setRequestID)
	s.router.Use(s.logRequest)

	s.router.HandleFunc("/login", s.handleLogin()).Methods(http.MethodGet)
	s.router.HandleFunc("/register", s.handleRegister()).Methods(http.MethodGet)
	s.router.HandleFunc("/share/{uuid}", s.handleShared()).Methods(http.MethodGet)
	s.router.HandleFunc("/file/{uuid}", s.handleDownloadFile()).Methods(http.MethodGet)

	api := s.router.PathPrefix("/api").Subrouter()
	api.Use(s.authenticateUser)
	api.HandleFunc("/files", s.handleFiles()).Methods(http.MethodGet)
	api.HandleFunc("/download", s.handleDownload()).Methods(http.MethodPost)
	api.HandleFunc("/upload", s.handleUpload()).Methods(http.MethodPost)
	api.HandleFunc("/delete", s.handleDelete()).Methods(http.MethodDelete)
	api.HandleFunc("/share", s.handleShareFile()).Methods(http.MethodPost)

	s.router.PathPrefix("/").Handler(http.FileServer(http.Dir("S3\\static")))
}

func (s *Server) handleShared() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "S3\\static\\file.html")
		return
	}
}

func (s *Server) handleLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "S3\\static\\login.html")
		return
	}
}

func (s *Server) handleRegister() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "S3\\static\\register.html")
		return
	}
}

func (s *Server) handleDelete() http.HandlerFunc {
	type request struct {
		Filename string `json:"filename"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(ctxKeyUserId).(int)

		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		if req.Filename == "" {
			s.error(w, r, http.StatusBadRequest, errEmptyFile)
			return
		}

		userFiles, err := s.filestore.FindFiles(userID)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, errDataBaseError)
			return
		}

		for i := 0; i < len(userFiles); i++ {
			if userFiles[i] == req.Filename {
				err := s.filestore.Delete(userID, req.Filename)
				if err != nil {
					s.error(w, r, http.StatusInternalServerError, err)
				}

				s.respond(w, r, http.StatusOK, map[string]string{"status": "ok"})
				return
			}
		}
		s.error(w, r, http.StatusBadRequest, errFileNotFound)
		return
	}
}

func (s *Server) handleUpload() http.HandlerFunc {
	type request struct {
		Filename string `json:"filename"`
		File     string `json:"file"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(ctxKeyUserId).(int)

		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		if req.Filename == "" {
			s.error(w, r, http.StatusBadRequest, errEmptyFile)
			return
		}

		userFiles, err := s.filestore.FindFiles(userID)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, errDataBaseError)
			return
		}

		for i := 0; i < len(userFiles); i++ {
			if userFiles[i] == req.Filename {
				s.error(w, r, http.StatusBadRequest, errFileAlreadyExist)
				return
			}
		}

		fileBytes, err := base64.StdEncoding.DecodeString(req.File)
		if err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		if err := s.filestore.Save(userID, req.Filename, fileBytes); err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func (s *Server) handleShareFile() func(http.ResponseWriter, *http.Request) {
	type request struct {
		Filename string `json:"filename"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(ctxKeyUserId).(int)

		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		if req.Filename == "" {
			s.error(w, r, http.StatusBadRequest, errEmptyFile)
			return
		}

		Uuid, err := s.filestore.Share(userID, req.Filename)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				s.error(w, r, http.StatusNotFound, errFileNotFound)
				return
			}
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		s.respond(w, r, http.StatusOK, map[string]string{"status": Uuid})
		return
	}
}

func (s *Server) handleDownloadFile() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		Uuid := vars["uuid"]

		userID, filename, err := s.filestore.FindByUUID(Uuid)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				s.error(w, r, http.StatusNotFound, errFileNotFound)
				return
			}
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		fileBytes, err := s.filestore.GetFileBytes(userID, filename)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
		}
		// http.ServeFile(w, r, fullFilename)
		s.respond(w, r, http.StatusOK, map[string]string{"base64": base64.StdEncoding.EncodeToString(fileBytes), "filename": filename})
		return
	}
}

func (s *Server) handleDownload() http.HandlerFunc {
	type request struct {
		Filename string `json:"filename"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		if req.Filename == "" {
			s.error(w, r, http.StatusBadRequest, errEmptyFile)
			return
		}

		userID := r.Context().Value(ctxKeyUserId).(int)

		userFiles, err := s.filestore.FindFiles(userID)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, errDataBaseError)
			return
		}
		for i := 0; i < len(userFiles); i++ {
			if userFiles[i] == req.Filename {
				fileBytes, err := s.filestore.GetFileBytes(userID, req.Filename)
				if err != nil {
					s.error(w, r, http.StatusInternalServerError, err)
					return
				}
				s.respond(w, r, http.StatusOK, map[string]string{"status": base64.StdEncoding.EncodeToString(fileBytes)})
				return
			}
		}
		s.error(w, r, http.StatusInternalServerError, errFileNotFound)
		return
	}
}

func (s *Server) handleFiles() http.HandlerFunc {
	type FileInfo struct {
		Name string `json:"name"`
		Date int    `json:"date"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(ctxKeyUserId).(int)
		names, dates, err := s.filestore.FindFilesWithDates(userID)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, errDataBaseError)
			return
		}
		numFiles := len(names)
		if numFiles == 0 {
			s.respond(w, r, http.StatusNoContent, map[string]string{filesNames: "", fileDates: ""})
			return
		}
		if numFiles < 0 {
			s.error(w, r, http.StatusInternalServerError, errInternalServerError)
			return
		}

		files := make([]FileInfo, len(names))
		for i := range names {
			files[i] = FileInfo{
				Name: names[i],
				Date: dates[i],
			}
		}
		s.respond(w, r, http.StatusOK, files)
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

func (s *Server) authenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		expirationTime, err := claims.GetExpirationTime()
		if err != nil {
			s.error(w, r, http.StatusUnauthorized, err)
			return
		}

		if expirationTime.Before(time.Now()) {
			s.error(w, r, http.StatusUnauthorized, errNotAuthenticated)
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

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyUserId, userID)))
	})
}

func (s *Server) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	s.logger.Error("error", zap.Error(err))
	s.respond(w, r, code, map[string]string{"error": err.Error()})
}

func (s *Server) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
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
