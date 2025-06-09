package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

type TestCase struct {
	Name   string
	Method string
	URL    string
	Body   interface{}
}

// runTests executes all the test cases and reports results to the testing framework
func runTests(t *testing.T, tests []TestCase) {
	client := &http.Client{}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			var bodyReader io.Reader
			if tc.Body != nil {
				data, err := json.Marshal(tc.Body)
				if err != nil {
					t.Fatalf("Failed to marshal body: %v", err)
				}
				bodyReader = bytes.NewReader(data)
			}

			req, err := http.NewRequest(tc.Method, tc.URL, bodyReader)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			if tc.Body != nil {
				req.Header.Set("Content-Type", "application/json")
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to execute request: %v", err)
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			t.Logf("Status: %s", resp.Status)
			t.Logf("Body: %s", respBody)
		})
	}
}
