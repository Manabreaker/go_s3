package tests

import (
	"net/http"
	"testing"
)

func TestAPIEndpoints(t *testing.T) {
	runTests(t, getAPITests())
}

func getAPITests() []TestCase {
	return []TestCase{
		// Auth Service Tests (Port 8000)
		// Register Tests
		{
			Name:   "Auth: Valid register",
			Method: http.MethodPost,
			URL:    "http://localhost:8000/account/register",
			Body: map[string]string{
				"email":    "test1@example.com",
				"password": "password123",
			},
		},
		{
			Name:   "Auth: Invalid register (email already exists)",
			Method: http.MethodPost,
			URL:    "http://localhost:8000/account/register",
			Body: map[string]string{
				"email":    "test1@example.com",
				"password": "password123",
			},
		},
		{
			Name:   "Auth: Invalid register (invalid email format)",
			Method: http.MethodPost,
			URL:    "http://localhost:8000/account/register",
			Body: map[string]string{
				"email":    "invalid-email",
				"password": "password123",
			},
		},
		{
			Name:   "Auth: Invalid register (short password)",
			Method: http.MethodPost,
			URL:    "http://localhost:8000/account/register",
			Body: map[string]string{
				"email":    "test2@example.com",
				"password": "123",
			},
		},
		// Login Tests
		{
			Name:   "Auth: Valid login",
			Method: http.MethodPost,
			URL:    "http://localhost:8000/account/login",
			Body: map[string]string{
				"email":    "test1@example.com",
				"password": "password123",
			},
		},
		{
			Name:   "Auth: Invalid login (wrong password)",
			Method: http.MethodPost,
			URL:    "http://localhost:8000/account/login",
			Body: map[string]string{
				"email":    "test1@example.com",
				"password": "wrongpassword",
			},
		},
		{
			Name:   "Auth: Invalid login (non-existent user)",
			Method: http.MethodPost,
			URL:    "http://localhost:8000/account/login",
			Body: map[string]string{
				"email":    "nonexistent@example.com",
				"password": "password123",
			},
		},
		// Logout Test
		{
			Name:   "Auth: Valid logout",
			Method: http.MethodPost,
			URL:    "http://localhost:8000/account/logout",
			Body:   nil,
		},

		// S3 Service Tests (Port 8080)
		// File Operations Tests
		{
			Name:   "S3: Get files list",
			Method: http.MethodGet,
			URL:    "http://localhost:8080/api/files",
			Body:   nil,
		},
		{
			Name:   "S3: Upload file",
			Method: http.MethodPost,
			URL:    "http://localhost:8080/api/upload",
			Body: map[string]string{
				"filename": "test_api.txt",
				"file":     "VGhpcyBpcyBhIHRlc3QgZmlsZSBmb3IgQVBJIHRlc3Rpbmc=",
			},
		},
		{
			Name:   "S3: Download file",
			Method: http.MethodPost,
			URL:    "http://localhost:8080/api/download",
			Body: map[string]string{
				"filename": "test_api.txt",
			},
		},
		{
			Name:   "S3: Share file",
			Method: http.MethodPost,
			URL:    "http://localhost:8080/api/share",
			Body: map[string]string{
				"filename": "test_api.txt",
			},
		},
		{
			Name:   "S3: Access shared file",
			Method: http.MethodGet,
			URL:    "http://localhost:8080/file/00000000-0000-0000-0000-000000000000",
			Body:   nil,
		},
		{
			Name:   "S3: Delete file",
			Method: http.MethodDelete,
			URL:    "http://localhost:8080/api/delete",
			Body: map[string]string{
				"filename": "test_api.txt",
			},
		},

		// APIGateway Tests (Port 8000)
		{
			Name:   "Gateway: Register",
			Method: http.MethodPost,
			URL:    "http://localhost:8000/register",
			Body: map[string]string{
				"email":    "gateway_test@example.com",
				"password": "password123",
			},
		},
		{
			Name:   "Gateway: Login",
			Method: http.MethodPost,
			URL:    "http://localhost:8000/login",
			Body: map[string]string{
				"email":    "gateway_test@example.com",
				"password": "password123",
			},
		},
		{
			Name:   "Gateway: Get files",
			Method: http.MethodGet,
			URL:    "http://localhost:8000/files",
			Body:   nil,
		},
		{
			Name:   "Gateway: Upload file",
			Method: http.MethodPost,
			URL:    "http://localhost:8000/upload",
			Body: map[string]string{
				"filename": "gateway_test.txt",
				"file":     "VGhpcyBpcyBhIHRlc3QgZmlsZSB1cGxvYWRlZCB2aWEgZ2F0ZXdheQ==",
			},
		},
		{
			Name:   "Gateway: Download file",
			Method: http.MethodPost,
			URL:    "http://localhost:8000/download",
			Body: map[string]string{
				"filename": "gateway_test.txt",
			},
		},
		{
			Name:   "Gateway: Share file",
			Method: http.MethodPost,
			URL:    "http://localhost:8000/share",
			Body: map[string]string{
				"filename": "gateway_test.txt",
			},
		},
		{
			Name:   "Gateway: Delete file",
			Method: http.MethodDelete,
			URL:    "http://localhost:8000/delete",
			Body: map[string]string{
				"filename": "gateway_test.txt",
			},
		},
		{
			Name:   "Gateway: Logout",
			Method: http.MethodPost,
			URL:    "http://localhost:8000/logout",
			Body:   nil,
		},
	}
}
