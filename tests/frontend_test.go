package tests

import (
	"net/http"
	"testing"
)

func TestFrontendEndpoints(t *testing.T) {
	runTests(t, getFrontendTests())
}

func getFrontendTests() []TestCase {
	return []TestCase{
		// Authentication Tests
		// Register Tests
		{
			Name:   "Frontend: Valid register",
			Method: http.MethodPost,
			URL:    "http://localhost:7000/register",
			Body: map[string]string{
				"email":    "frontend_user@example.com",
				"password": "password123",
			},
		},
		{
			Name:   "Frontend: Invalid register (email already exists)",
			Method: http.MethodPost,
			URL:    "http://localhost:7000/register",
			Body: map[string]string{
				"email":    "frontend_user@example.com",
				"password": "password123",
			},
		},
		{
			Name:   "Frontend: Invalid register (invalid email format)",
			Method: http.MethodPost,
			URL:    "http://localhost:7000/register",
			Body: map[string]string{
				"email":    "invalid-email",
				"password": "password123",
			},
		},
		{
			Name:   "Frontend: Invalid register (short password)",
			Method: http.MethodPost,
			URL:    "http://localhost:7000/register",
			Body: map[string]string{
				"email":    "another_user@example.com",
				"password": "123",
			},
		},
		// Login Tests
		{
			Name:   "Frontend: Valid login",
			Method: http.MethodPost,
			URL:    "http://localhost:7000/login",
			Body: map[string]string{
				"email":    "frontend_user@example.com",
				"password": "password123",
			},
		},
		{
			Name:   "Frontend: Invalid login (wrong password)",
			Method: http.MethodPost,
			URL:    "http://localhost:7000/login",
			Body: map[string]string{
				"email":    "frontend_user@example.com",
				"password": "wrongpassword",
			},
		},
		{
			Name:   "Frontend: Invalid login (non-existent user)",
			Method: http.MethodPost,
			URL:    "http://localhost:7000/login",
			Body: map[string]string{
				"email":    "nonexistent@example.com",
				"password": "password123",
			},
		},
		// UI Page Access Tests
		{
			Name:   "Frontend: Access login page",
			Method: http.MethodGet,
			URL:    "http://localhost:7000/login",
			Body:   nil,
		},
		{
			Name:   "Frontend: Access register page",
			Method: http.MethodGet,
			URL:    "http://localhost:7000/register",
			Body:   nil,
		},
		{
			Name:   "Frontend: Access main page",
			Method: http.MethodGet,
			URL:    "http://localhost:7000/",
			Body:   nil,
		},
		// File Operations Tests
		{
			Name:   "Frontend: Get files list",
			Method: http.MethodGet,
			URL:    "http://localhost:7000/files",
			Body:   nil,
		},
		{
			Name:   "Frontend: Upload file",
			Method: http.MethodPost,
			URL:    "http://localhost:7000/upload",
			Body: map[string]string{
				"filename": "frontend_test.txt",
				"file":     "VGhpcyBpcyBhIHRlc3QgZmlsZSBmb3IgZnJvbnRlbmQgdGVzdGluZw==",
			},
		},
		{
			Name:   "Frontend: Download file",
			Method: http.MethodPost,
			URL:    "http://localhost:7000/download",
			Body: map[string]string{
				"filename": "frontend_test.txt",
			},
		},
		{
			Name:   "Frontend: Share file",
			Method: http.MethodPost,
			URL:    "http://localhost:7000/share",
			Body: map[string]string{
				"filename": "frontend_test.txt",
			},
		},
		{
			Name:   "Frontend: Access shared file page",
			Method: http.MethodGet,
			URL:    "http://localhost:7000/share/00000000-0000-0000-0000-000000000000",
			Body:   nil,
		},
		{
			Name:   "Frontend: Access shared file content",
			Method: http.MethodGet,
			URL:    "http://localhost:7000/file/00000000-0000-0000-0000-000000000000",
			Body:   nil,
		},
		{
			Name:   "Frontend: Delete file",
			Method: http.MethodDelete,
			URL:    "http://localhost:7000/delete",
			Body: map[string]string{
				"filename": "frontend_test.txt",
			},
		},
		{
			Name:   "Frontend: Logout",
			Method: http.MethodPost,
			URL:    "http://localhost:7000/logout",
			Body:   nil,
		},
		// Edge Cases
		{
			Name:   "Frontend: Access non-existent file",
			Method: http.MethodPost,
			URL:    "http://localhost:7000/download",
			Body: map[string]string{
				"filename": "nonexistent.txt",
			},
		},
		{
			Name:   "Frontend: Access invalid shared file UUID",
			Method: http.MethodGet,
			URL:    "http://localhost:7000/file/invalid-uuid",
			Body:   nil,
		},
		{
			Name:   "Frontend: Upload file with empty content",
			Method: http.MethodPost,
			URL:    "http://localhost:7000/upload",
			Body: map[string]string{
				"filename": "empty.txt",
				"file":     "",
			},
		},
		{
			Name:   "Frontend: Upload file with no filename",
			Method: http.MethodPost,
			URL:    "http://localhost:7000/upload",
			Body: map[string]string{
				"filename": "",
				"file":     "VGhpcyBpcyBhIHRlc3QgZmlsZQ==",
			},
		},
	}
}
