package apiserver

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

const ctxKeyRequestID crtKey = iota

type crtKey int8

type Server struct {
	router *mux.Router
	logger *zap.Logger
	config *Config
}

func (s *Server) configureRouter() {
	// 1. Опции CORS
	corsOpts := handlers.CORS(
		handlers.AllowCredentials(),
		handlers.AllowedOrigins([]string{"http://localhost:8080", "http://127.0.0.1:8080", "http://localhost:7000", "http://127.0.0.1:7000"}),
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

	// 3. Лог мидлвары
	s.router.Use(s.setRequestID)
	s.router.Use(s.logRequest)

	// 4. Роуты на AuthServer
	s.router.HandleFunc("/login", s.redirectToAuth()).Methods(http.MethodPost, http.MethodOptions)
	s.router.HandleFunc("/register", s.redirectToAuth()).Methods(http.MethodPost, http.MethodOptions)
	s.router.HandleFunc("/logout", s.redirectToAuth()).Methods(http.MethodPost, http.MethodOptions)

	// 5. Роуты на S3Server
	s.router.HandleFunc("/files", s.redirectToS3()).Methods(http.MethodGet)
	s.router.HandleFunc("/download", s.redirectToS3()).Methods(http.MethodPost)
	s.router.HandleFunc("/upload", s.redirectToS3()).Methods(http.MethodPost)
	s.router.HandleFunc("/delete", s.redirectToS3()).Methods(http.MethodDelete)
	s.router.HandleFunc("/share", s.redirectToS3()).Methods(http.MethodPost)

	// 6. Роуты на FileServer (ну пока что просто на S3Server) TODO: сделать отдельный сервер
	s.router.HandleFunc("/login", s.redirectToFile()).Methods(http.MethodGet)
	s.router.HandleFunc("/register", s.redirectToFile()).Methods(http.MethodGet)
	s.router.HandleFunc("/share/{uuid}", s.redirectToFile()).Methods(http.MethodGet, http.MethodOptions)
	s.router.HandleFunc("/file/{uuid}", s.redirectToFile()).Methods(http.MethodGet)
	s.router.HandleFunc("/", s.redirectToFile()).Methods(http.MethodGet)
}

func (s Server) redirectToFile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("### Redirecting to File Server: http://" + s.config.S3Server.Host + ":" + s.config.S3Server.Port)
		fileUrl := "http://" + s.config.S3Server.Host + ":" + s.config.S3Server.Port
		target, _ := url.Parse(fileUrl)
		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.ServeHTTP(w, r)
	}
}

func (s Server) redirectToAuth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("### Redirecting to Auth Server")
		authUrl := "http://" + s.config.AuthServer.Host + ":" + s.config.AuthServer.Port + s.config.AuthServer.PathPrefix
		target, _ := url.Parse(authUrl)
		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.ServeHTTP(w, r)
	}
}

func (s Server) redirectToS3() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s3Url := "http://" + s.config.S3Server.Host + ":" + s.config.S3Server.Port + s.config.S3Server.PathPrefix
		fmt.Println("### Redirecting to S3:", s3Url)
		target, _ := url.Parse(s3Url)
		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.ServeHTTP(w, r)
	}
}

func (s Server) Start(config *Config) error {
	s.config = config
	s.configureRouter()
	return http.ListenAndServe(config.Gateway.Host+":"+config.Gateway.Port, s.router)
}

type Config struct {
	Gateway    RemoteServer `yaml:"gateway_server"`
	AuthServer RemoteServer `yaml:"auth_server"`
	S3Server   RemoteServer `yaml:"s3_server"`
}

type RemoteServer struct {
	Host       string `yaml:"host"`
	Port       string `yaml:"port"`
	PathPrefix string `yaml:"path_prefix"`
	Timeout    struct {
		Server int `yaml:"server"`
		Write  int `yaml:"write"`
		Read   int `yaml:"read"`
		Idle   int `yaml:"idle"`
	} `yaml:"timeout"`
}

func NewServer() Server {
	logger, _ := zap.NewProduction()
	return Server{
		router: mux.NewRouter(),
		logger: logger,
	}
}
func NewConfig(configPath string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)

	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

func (s *Server) setRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		r.Header.Set("X-Request-ID", id)
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
