package app

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/timo-42/rayboard/internal/backend/auth"
	"github.com/timo-42/rayboard/internal/backend/authz"
	"github.com/timo-42/rayboard/internal/backend/tracker"
	"github.com/timo-42/rayboard/internal/config"
)

func runDemoSeed(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "seed" {
		fmt.Fprintln(stderr, "usage: rayboard demo seed --backend-url http://host:port --admin-user admin --admin-password <password> --fresh-reset")
		return 2
	}

	flags := flag.NewFlagSet("demo seed", flag.ContinueOnError)
	flags.SetOutput(stderr)

	var backendURL string
	var adminUser string
	var adminPassword string
	var freshReset bool
	flags.StringVar(&backendURL, "backend-url", config.DefaultBackendURL, "backend API base URL")
	flags.StringVar(&adminUser, "admin-user", "admin", "admin username")
	flags.StringVar(&adminPassword, "admin-password", "", "admin password")
	flags.BoolVar(&freshReset, "fresh-reset", false, "confirm demo data should be seeded")

	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	if backendURL == "" || adminUser == "" || adminPassword == "" || !freshReset {
		fmt.Fprintln(stderr, "demo seed requires --backend-url, --admin-user, --admin-password, and --fresh-reset")
		return 2
	}

	seeder, err := newDemoSeeder(backendURL, stdout)
	if err != nil {
		fmt.Fprintf(stderr, "demo seed: %v\n", err)
		return 2
	}
	if err := seeder.seed(ctx, adminUser, adminPassword); err != nil {
		fmt.Fprintf(stderr, "demo seed: %v\n", err)
		return 1
	}
	return 0
}

type demoSeeder struct {
	baseURL *url.URL
	client  *http.Client
	stdout  io.Writer
	csrf    string
	suffix  string
}

type demoLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type demoGroup struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

type demoRoleBinding struct {
	RoleName    authz.RoleName          `json:"role_name"`
	SubjectType authz.BindingTargetKind `json:"subject_type"`
	SubjectID   string                  `json:"subject_id"`
	Scope       authz.ScopeKind         `json:"scope"`
	ProjectID   string                  `json:"project_id,omitempty"`
}

type demoProjectResource struct {
	Metadata struct {
		ID string `json:"id"`
	} `json:"metadata"`
	Spec struct {
		Key         string `json:"key"`
		Name        string `json:"name"`
		Description string `json:"description"`
		LeadUserID  string `json:"lead_user_id"`
	} `json:"spec"`
}

func newDemoSeeder(rawBackendURL string, stdout io.Writer) (*demoSeeder, error) {
	parsed, err := url.Parse(strings.TrimRight(rawBackendURL, "/"))
	if err != nil {
		return nil, fmt.Errorf("parse backend URL: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("backend URL must include scheme and host")
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("create cookie jar: %w", err)
	}
	suffix, err := randomDemoSuffix()
	if err != nil {
		return nil, err
	}
	return &demoSeeder{
		baseURL: parsed,
		client: &http.Client{
			Jar:     jar,
			Timeout: 15 * time.Second,
		},
		stdout: stdout,
		suffix: suffix,
	}, nil
}

func (s *demoSeeder) seed(ctx context.Context, adminUser string, adminPassword string) error {
	if err := s.login(ctx, adminUser, adminPassword); err != nil {
		return err
	}

	fmt.Fprintf(s.stdout, "demo seed suffix: %s\n", s.suffix)
	fmt.Fprintln(s.stdout, "demo admin login succeeded")

	users, err := s.createUsers(ctx)
	if err != nil {
		return err
	}
	groups, err := s.createGroups(ctx, users)
	if err != nil {
		return err
	}
	project, err := s.createProject(ctx)
	if err != nil {
		return err
	}
	if err := s.bindProjectMembers(ctx, groups["engineers"].ID, project.ID); err != nil {
		return err
	}
	if err := s.createTickets(ctx, project.ID); err != nil {
		return err
	}

	fmt.Fprintf(s.stdout, "demo project: key=%s id=%s\n", project.Key, project.ID)
	fmt.Fprintln(s.stdout, "demo seed completed")
	return nil
}

func (s *demoSeeder) login(ctx context.Context, username string, password string) error {
	if err := s.apiJSON(ctx, http.MethodPost, "/api/login", demoLoginRequest{Username: username, Password: password}, nil); err != nil {
		return err
	}
	for _, cookie := range s.client.Jar.Cookies(s.baseURL) {
		if cookie.Name == "rayboard_csrf" {
			s.csrf = cookie.Value
			return nil
		}
	}
	return fmt.Errorf("login succeeded but CSRF cookie was not set")
}

func (s *demoSeeder) createUsers(ctx context.Context) (map[string]auth.CreatedUser, error) {
	inputs := map[string]map[string]any{
		"lead": {
			"username":     "demo_lead_" + strings.ToLower(s.suffix),
			"display_name": "Demo Lead " + s.suffix,
		},
		"engineer": {
			"username":     "demo_engineer_" + strings.ToLower(s.suffix),
			"display_name": "Demo Engineer " + s.suffix,
		},
		"reporter": {
			"username":     "demo_reporter_" + strings.ToLower(s.suffix),
			"display_name": "Demo Reporter " + s.suffix,
		},
	}

	users := make(map[string]auth.CreatedUser, len(inputs))
	for key, input := range inputs {
		var created auth.CreatedUser
		if err := s.apiJSON(ctx, http.MethodPost, "/api/users", input, &created); err != nil {
			return nil, fmt.Errorf("create demo user %s: %w", key, err)
		}
		users[key] = created
		fmt.Fprintf(s.stdout, "demo user: role=%s username=%s password=%s\n", key, created.Username, created.Password)
	}
	return users, nil
}

func (s *demoSeeder) createGroups(ctx context.Context, users map[string]auth.CreatedUser) (map[string]demoGroup, error) {
	inputs := map[string]map[string]any{
		"engineers": {
			"name":         "demo_engineers_" + strings.ToLower(s.suffix),
			"display_name": "Demo Engineers " + s.suffix,
		},
		"stakeholders": {
			"name":         "demo_stakeholders_" + strings.ToLower(s.suffix),
			"display_name": "Demo Stakeholders " + s.suffix,
		},
	}

	groups := make(map[string]demoGroup, len(inputs))
	for key, input := range inputs {
		var group demoGroup
		if err := s.apiJSON(ctx, http.MethodPost, "/api/groups", input, &group); err != nil {
			return nil, fmt.Errorf("create demo group %s: %w", key, err)
		}
		groups[key] = group
	}

	if err := s.apiJSON(ctx, http.MethodPost, "/api/groups/"+groups["engineers"].ID+"/members/"+users["lead"].ID, nil, nil); err != nil {
		return nil, fmt.Errorf("add lead to engineers: %w", err)
	}
	if err := s.apiJSON(ctx, http.MethodPost, "/api/groups/"+groups["engineers"].ID+"/members/"+users["engineer"].ID, nil, nil); err != nil {
		return nil, fmt.Errorf("add engineer to engineers: %w", err)
	}
	if err := s.apiJSON(ctx, http.MethodPost, "/api/groups/"+groups["stakeholders"].ID+"/members/"+users["reporter"].ID, nil, nil); err != nil {
		return nil, fmt.Errorf("add reporter to stakeholders: %w", err)
	}
	return groups, nil
}

func (s *demoSeeder) createProject(ctx context.Context) (tracker.Project, error) {
	input := tracker.CreateProjectInput{
		Key:         "DEMO" + s.suffix,
		Name:        "Demo Project " + s.suffix,
		Description: "Seeded project for demos",
	}
	var project demoProjectResource
	if err := s.apiJSON(ctx, http.MethodPost, "/api/projects", map[string]tracker.CreateProjectInput{"spec": input}, &project); err != nil {
		return tracker.Project{}, fmt.Errorf("create demo project: %w", err)
	}
	return tracker.Project{
		ID:          project.Metadata.ID,
		Key:         project.Spec.Key,
		Name:        project.Spec.Name,
		Description: project.Spec.Description,
		LeadUserID:  project.Spec.LeadUserID,
	}, nil
}

func (s *demoSeeder) bindProjectMembers(ctx context.Context, groupID string, projectID string) error {
	return s.apiJSON(ctx, http.MethodPost, "/api/role-bindings", demoRoleBinding{
		RoleName:    authz.RoleProjectMember,
		SubjectType: authz.BindingTargetGroup,
		SubjectID:   groupID,
		Scope:       authz.ScopeKindProject,
		ProjectID:   projectID,
	}, nil)
}

func (s *demoSeeder) createTickets(ctx context.Context, projectID string) error {
	tickets := []tracker.CreateTicketInput{
		{
			Title:       "Prepare customer onboarding board",
			Description: "Create the initial backlog and workflow states.",
			Priority:    "High",
			Type:        "Task",
		},
		{
			Title:       "Investigate login audit trail",
			Description: "Capture activity that helps explain auth changes.",
			Priority:    "Medium",
			Type:        "Bug",
		},
		{
			Title:       "Draft roadmap for automation hooks",
			Description: "Outline Lua and AI hook milestones.",
			Priority:    "Low",
			Type:        "Story",
		},
	}
	for index, ticket := range tickets {
		var created tracker.Ticket
		if err := s.apiJSON(ctx, http.MethodPost, "/api/projects/"+projectID+"/tickets", ticket, &created); err != nil {
			return fmt.Errorf("create demo ticket %d: %w", index+1, err)
		}
		fmt.Fprintf(s.stdout, "demo ticket: key=%s title=%q\n", created.Key, created.Title)
		if index == 1 {
			status := "in_progress"
			if err := s.apiJSON(ctx, http.MethodPatch, "/api/tickets/"+created.ID, tracker.UpdateTicketInput{Status: &status}, nil); err != nil {
				return fmt.Errorf("update demo ticket %s: %w", created.Key, err)
			}
		}
	}
	return nil
}

func (s *demoSeeder) apiJSON(ctx context.Context, method string, path string, input any, output any) error {
	var body io.Reader
	if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		body = bytes.NewReader(encoded)
	}

	target := s.baseURL.ResolveReference(&url.URL{Path: path})
	request, err := http.NewRequestWithContext(ctx, method, target.String(), body)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if input != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if s.csrf != "" && mutatingMethod(method) {
		request.Header.Set("X-CSRF-Token", s.csrf)
	}

	response, err := s.client.Do(request)
	if err != nil {
		return fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("%s %s returned %d: %s", method, path, response.StatusCode, strings.TrimSpace(string(data)))
	}
	if output == nil || response.StatusCode == http.StatusNoContent {
		return nil
	}
	if err := json.NewDecoder(response.Body).Decode(output); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func randomDemoSuffix() (string, error) {
	var raw [3]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate demo suffix: %w", err)
	}
	return strings.TrimRight(base32.StdEncoding.EncodeToString(raw[:]), "="), nil
}

func mutatingMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	default:
		return true
	}
}
