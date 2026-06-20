package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_SendRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}

		if r.URL.Path == "/test" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "success"})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token")

	t.Run("Success Path", func(t *testing.T) {
		var resp map[string]string
		err := c.sendRequest(context.Background(), "GET", "/test", nil, &resp)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if resp["message"] != "success" {
			t.Errorf("Expected message 'success', got %q", resp["message"])
		}
	})

	t.Run("Unauthorized Path", func(t *testing.T) {
		badClient := NewClient(ts.URL, "bad-token")
		err := badClient.sendRequest(context.Background(), "GET", "/test", nil, nil)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
	})
}

func TestClient_CRUD(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v1/config/certificates" {
			if r.Method == "GET" {
				_ = json.NewEncoder(w).Encode([]CertConfig{{ID: "test-cert-id", Primary: "test.com"}})
			} else if r.Method == "POST" {
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(CertConfig{ID: "new-cert-id", Primary: "new.com"})
			}
			return
		}
		if r.URL.Path == "/api/v1/config/certificates/test-cert-id" {
			if r.Method == "PUT" {
				w.WriteHeader(http.StatusOK)
			} else if r.Method == "DELETE" {
				w.WriteHeader(http.StatusNoContent)
			}
			return
		}
		if r.URL.Path == "/api/v1/config/api_keys" {
			if r.Method == "GET" {
				_ = json.NewEncoder(w).Encode([]APIKeyConfig{{ID: "test-key-id", Description: "test-key"}})
			} else if r.Method == "POST" {
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(APIKeyConfig{
					ID:             "new-key-id",
					Description:    "test-key",
					CleartextToken: "generated-token",
				})
			}
			return
		}
		if r.URL.Path == "/api/v1/config/api_keys/test-key-id" {
			if r.Method == "PUT" {
				w.WriteHeader(http.StatusOK)
			} else if r.Method == "DELETE" {
				w.WriteHeader(http.StatusNoContent)
			}
			return
		}
		if r.URL.Path == "/api/v1/config/teams" {
			if r.Method == "GET" {
				_ = json.NewEncoder(w).Encode([]TeamConfig{{ID: "test-team-id", Name: "test-team"}})
			} else if r.Method == "POST" {
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(TeamConfig{ID: "new-team-id", Name: "new-team"})
			}
			return
		}
		if r.URL.Path == "/api/v1/config/teams/test-team-id" {
			if r.Method == "PUT" {
				w.WriteHeader(http.StatusOK)
			} else if r.Method == "DELETE" {
				w.WriteHeader(http.StatusNoContent)
			}
			return
		}
		if r.URL.Path == "/api/v1/certificates" && r.Method == "GET" {
			_ = json.NewEncoder(w).Encode([]CertificateData{{Domain: "test.com", Certificate: "PEM"}})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "token")

	t.Run("Certificates", func(t *testing.T) {
		certs, err := c.GetCertificates(context.Background())
		if err != nil || len(certs) != 1 || certs[0].Primary != "test.com" {
			t.Errorf("GetCertificates failed: %v, certs: %+v", err, certs)
		}

		createdCert, err := c.CreateCertificate(context.Background(), CertConfig{Primary: "new.com"})
		if err != nil || createdCert.ID != "new-cert-id" {
			t.Errorf("CreateCertificate failed: %v, createdCert: %+v", err, createdCert)
		}

		err = c.UpdateCertificate(context.Background(), "test-cert-id", CertConfig{Sans: []string{"www.test.com"}})
		if err != nil {
			t.Errorf("UpdateCertificate failed: %v", err)
		}

		err = c.DeleteCertificate(context.Background(), "test-cert-id")
		if err != nil {
			t.Errorf("DeleteCertificate failed: %v", err)
		}
	})

	t.Run("APIKeys", func(t *testing.T) {
		keys, err := c.GetAPIKeys(context.Background())
		if err != nil || len(keys) != 1 || keys[0].ID != "test-key-id" {
			t.Errorf("GetAPIKeys failed: %v, keys: %+v", err, keys)
		}

		createdKey, err := c.CreateAPIKey(context.Background(), APIKeyConfig{Description: "test-key"})
		if err != nil || createdKey.CleartextToken != "generated-token" || createdKey.ID != "new-key-id" {
			t.Errorf("CreateAPIKey failed: %v, createdKey: %+v", err, createdKey)
		}

		err = c.UpdateAPIKey(context.Background(), APIKeyConfig{ID: "test-key-id", Description: "test-key", Admin: true})
		if err != nil {
			t.Errorf("UpdateAPIKey failed: %v", err)
		}

		err = c.DeleteAPIKey(context.Background(), "test-key-id")
		if err != nil {
			t.Errorf("DeleteAPIKey failed: %v", err)
		}
	})

	t.Run("CertificateData", func(t *testing.T) {
		data, err := c.GetCertificateData(context.Background())
		if err != nil || len(data) != 1 || data[0].Certificate != "PEM" {
			t.Errorf("GetCertificateData failed: %v, data: %+v", err, data)
		}
	})

	t.Run("Teams", func(t *testing.T) {
		teams, err := c.GetTeams(context.Background())
		if err != nil || len(teams) != 1 || teams[0].ID != "test-team-id" {
			t.Errorf("GetTeams failed: %v, teams: %+v", err, teams)
		}

		createdTeam, err := c.CreateTeam(context.Background(), TeamConfig{Name: "new-team"})
		if err != nil || createdTeam.ID != "new-team-id" {
			t.Errorf("CreateTeam failed: %v, createdTeam: %+v", err, createdTeam)
		}

		err = c.UpdateTeam(context.Background(), "test-team-id", TeamConfig{Name: "updated-team"})
		if err != nil {
			t.Errorf("UpdateTeam failed: %v", err)
		}

		err = c.DeleteTeam(context.Background(), "test-team-id")
		if err != nil {
			t.Errorf("DeleteTeam failed: %v", err)
		}
	})
}

