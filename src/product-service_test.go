package main

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

const validBody = `{"product_id":1,"sku":"ABC","manufacturer":"Updated","category_id":10,"weight":0,"some_other_id":20}`

func TestProductAPI(t *testing.T) {
	router := newRouter("test-key", true)
	cases := []struct {
		name, method, path, body string
		status                   int
		fault                    bool
	}{
		{"get", "GET", "/v1/products/1", "", 200, false},
		{"missing", "GET", "/v1/products/99", "", 404, false},
		{"bad path", "GET", "/v1/products/abc", "", 404, false},
		{"overflow path", "GET", "/v1/products/2147483648", "", 404, false},
		{"update zero weight", "POST", "/v1/products/1/details", validBody, 204, false},
		{"post missing", "POST", "/v1/products/99/details", strings.Replace(validBody, `"product_id":1`, `"product_id":99`, 1), 404, false},
		{"missing weight", "POST", "/v1/products/1/details", strings.Replace(validBody, `"weight":0,`, "", 1), 400, false},
		{"null weight", "POST", "/v1/products/1/details", strings.Replace(validBody, `"weight":0`, `"weight":null`, 1), 400, false},
		{"negative weight", "POST", "/v1/products/1/details", strings.Replace(validBody, `"weight":0`, `"weight":-1`, 1), 400, false},
		{"overflow weight", "POST", "/v1/products/1/details", strings.Replace(validBody, `"weight":0`, `"weight":2147483648`, 1), 400, false},
		{"fractional weight", "POST", "/v1/products/1/details", strings.Replace(validBody, `"weight":0`, `"weight":1.5`, 1), 400, false},
		{"empty sku", "POST", "/v1/products/1/details", strings.Replace(validBody, `"ABC"`, `""`, 1), 400, false},
		{"long sku", "POST", "/v1/products/1/details", strings.Replace(validBody, "ABC", strings.Repeat("x", 101), 1), 400, false},
		{"long manufacturer", "POST", "/v1/products/1/details", strings.Replace(validBody, "Updated", strings.Repeat("x", 201), 1), 400, false},
		{"null body", "POST", "/v1/products/1/details", "null", 400, false},
		{"malformed", "POST", "/v1/products/1/details", "{", 400, false},
		{"extra JSON", "POST", "/v1/products/1/details", validBody + " {}", 400, false},
		{"mismatched ID", "POST", "/v1/products/2/details", validBody, 400, false},
		{"bad post path", "POST", "/v1/products/0/details", validBody, 400, false},
		{"get failure", "GET", "/v1/products/1", "", 500, true},
		{"post failure", "POST", "/v1/products/1/details", validBody, 500, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.Header.Set("X-API-Key", "test-key")
			req.Header.Set("Content-Type", "application/json")
			if tc.fault {
				req.Header.Set("X-Test-Error", "500")
			}
			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)
			if res.Code != tc.status {
				t.Fatalf("got %d %s, want %d", res.Code, res.Body.String(), tc.status)
			}
			if tc.status == 204 && res.Body.Len() != 0 {
				t.Fatal("204 must be empty")
			}
			if tc.status >= 400 {
				var e APIError
				if json.Unmarshal(res.Body.Bytes(), &e) != nil || e.Error == "" || e.Message == "" {
					t.Fatal("invalid Error schema")
				}
			}
		})
	}
	// GET must show the update and all rejected writes must leave it unchanged.
	req := httptest.NewRequest("GET", "/v1/products/1", nil)
	req.Header.Set("X-API-Key", "test-key")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	var p Product
	if json.Unmarshal(res.Body.Bytes(), &p) != nil || p.Manufacturer != "Updated" || p.Weight != 0 {
		t.Fatal("update not persisted correctly")
	}
}

func TestAuthAndFaultSwitch(t *testing.T) {
	router := newRouter("test-key", false)
	for _, key := range []string{"", "wrong", "test-key"} {
		req := httptest.NewRequest("GET", "/v1/products/1", nil)
		req.Header.Set("X-API-Key", key)
		req.Header.Set("X-Test-Error", "500")
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		want := 401
		if key == "test-key" {
			want = 200
		}
		if res.Code != want {
			t.Fatalf("got %d, want %d", res.Code, want)
		}
	}
}
