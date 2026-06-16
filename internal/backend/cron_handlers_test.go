package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/automation"
	"github.com/timo-42/rayboard/internal/backend/cronjobs"
)

func TestCronEndpointsUseResourceEnvelope(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	authorizer := authz.NewSQLEvaluator(db.SQL)
	handler := NewHandler(
		WithAuthService(auth.NewService(db.SQL)),
		WithAuthorizer(authorizer),
		WithCronService(cronjobs.NewService(db.SQL, authorizer, automation.NewRunStore(db.SQL))),
	)

	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": bootstrap.Username,
		"password": bootstrap.Password,
	}, nil)
	session := responseCookie(t, login.Result(), auth.SessionCookieName)
	csrf := responseCookie(t, login.Result(), csrfCookieName)

	createReq := httptest.NewRequest(http.MethodPost, "/api/cron-jobs", mustJSON(t, map[string]any{
		"spec": map[string]any{
			"name":     "Daily triage",
			"schedule": "0 9 * * *",
			"timezone": "UTC",
			"engine": map[string]any{
				"type":   "lua",
				"script": `rayboard.log("triage")`,
			},
		},
	}))
	addSessionCSRF(createReq, session, csrf)
	create := httptest.NewRecorder()
	handler.ServeHTTP(create, createReq)
	if create.Code != http.StatusCreated {
		t.Fatalf("expected create cron status 201, got %d: %s", create.Code, create.Body.String())
	}
	created := decodeCronResource(t, create.Body.Bytes())
	if created.Metadata.ID == "" || created.Spec.Name != "Daily triage" || created.Spec.Engine.Type != "lua" || created.Status.NextRunAt == nil {
		t.Fatalf("unexpected created cron resource: %#v", created)
	}

	renamed := "Morning triage"
	updateReq := httptest.NewRequest(http.MethodPatch, "/api/cron-jobs/"+created.Metadata.ID, mustJSON(t, map[string]any{
		"spec": map[string]any{
			"name": renamed,
		},
	}))
	addSessionCSRF(updateReq, session, csrf)
	update := httptest.NewRecorder()
	handler.ServeHTTP(update, updateReq)
	if update.Code != http.StatusOK {
		t.Fatalf("expected update cron status 200, got %d: %s", update.Code, update.Body.String())
	}
	updated := decodeCronResource(t, update.Body.Bytes())
	if updated.Spec.Name != renamed || updated.Spec.Schedule != "0 9 * * *" {
		t.Fatalf("unexpected updated cron resource: %#v", updated)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/cron-jobs/"+created.Metadata.ID, nil)
	getReq.AddCookie(session)
	get := httptest.NewRecorder()
	handler.ServeHTTP(get, getReq)
	if get.Code != http.StatusOK {
		t.Fatalf("expected get cron status 200, got %d: %s", get.Code, get.Body.String())
	}
	got := decodeCronResource(t, get.Body.Bytes())
	if got.Metadata.ID != created.Metadata.ID || got.Spec.Name != renamed || got.Status.NextRunAt == nil {
		t.Fatalf("unexpected fetched cron resource: %#v", got)
	}
}

func TestCronEndpointEngineDiscriminatorValidation(t *testing.T) {
	ctx := context.Background()
	db, bootstrap := openBackendTestDB(t, ctx)
	authorizer := authz.NewSQLEvaluator(db.SQL)
	handler := NewHandler(
		WithAuthService(auth.NewService(db.SQL)),
		WithAuthorizer(authorizer),
		WithCronService(cronjobs.NewService(db.SQL, authorizer, automation.NewRunStore(db.SQL))),
	)

	login := postJSON(t, handler, "/api/login", map[string]string{
		"username": bootstrap.Username,
		"password": bootstrap.Password,
	}, nil)
	session := responseCookie(t, login.Result(), auth.SessionCookieName)
	csrf := responseCookie(t, login.Result(), csrfCookieName)

	for name, body := range map[string]map[string]any{
		"lua script required": {
			"spec": map[string]any{
				"name":     "Lua missing script",
				"schedule": "* * * * *",
				"engine": map[string]any{
					"type": "lua",
				},
			},
		},
		"ai provider required": {
			"spec": map[string]any{
				"name":     "AI missing provider",
				"schedule": "* * * * *",
				"engine": map[string]any{
					"type":   "ai",
					"prompt": "Return JSON",
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/cron-jobs", mustJSON(t, body))
			addSessionCSRF(req, session, csrf)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusUnprocessableEntity {
				t.Fatalf("expected discriminator validation status 422, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

type cronResourceBody struct {
	Metadata struct {
		ID        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	} `json:"metadata"`
	Spec struct {
		Name     string `json:"name"`
		Schedule string `json:"schedule"`
		Timezone string `json:"timezone"`
		Engine   struct {
			Type   string `json:"type"`
			Script string `json:"script"`
		} `json:"engine"`
	} `json:"spec"`
	Status struct {
		NextRunAt *time.Time `json:"next_run_at"`
	} `json:"status"`
}

func decodeCronResource(t *testing.T, data []byte) cronResourceBody {
	t.Helper()

	var body cronResourceBody
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("decode cron resource: %v", err)
	}
	return body
}
