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
	Metadata struct {
		ID string `json:"id"`
	} `json:"metadata"`
	Spec struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
	} `json:"spec"`
}

func (group demoGroup) id() string {
	return group.Metadata.ID
}

type demoCreatedUser struct {
	Metadata struct {
		ID string `json:"id"`
	} `json:"metadata"`
	Spec struct {
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
	} `json:"spec"`
	Status struct {
		Password string `json:"password"`
	} `json:"status"`
}

func (user demoCreatedUser) id() string {
	return user.Metadata.ID
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

type demoTicketResource struct {
	Metadata struct {
		ID        string `json:"id"`
		ProjectID string `json:"project_id"`
	} `json:"metadata"`
	Spec struct {
		Title    string `json:"title"`
		Priority string `json:"priority"`
		Type     string `json:"type"`
		Status   string `json:"status"`
	} `json:"spec"`
	Status struct {
		Key string `json:"key"`
	} `json:"status"`
}

func (ticket demoTicketResource) id() string {
	return ticket.Metadata.ID
}

type demoIDResource struct {
	Metadata struct {
		ID string `json:"id"`
	} `json:"metadata"`
}

func (resource demoIDResource) id() string {
	return resource.Metadata.ID
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
	project, err := s.createProject(ctx, users["lead"].id())
	if err != nil {
		return err
	}
	if err := s.bindProjectAccess(ctx, groups["engineers"].id(), users["lead"].id(), project.ID); err != nil {
		return err
	}
	assets, err := s.createProjectPlanning(ctx, project.ID, users)
	if err != nil {
		return err
	}
	if err := s.createTicketHook(ctx, project.ID); err != nil {
		return err
	}
	if err := s.createTickets(ctx, project.ID, assets); err != nil {
		return err
	}

	fmt.Fprintf(s.stdout, "demo project: key=%s id=%s\n", project.Key, project.ID)
	fmt.Fprintln(s.stdout, "demo seed completed")
	return nil
}

func (s *demoSeeder) login(ctx context.Context, username string, password string) error {
	if err := s.apiJSON(ctx, http.MethodPost, "/api/login", map[string]demoLoginRequest{
		"spec": {Username: username, Password: password},
	}, nil); err != nil {
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

func (s *demoSeeder) createUsers(ctx context.Context) (map[string]demoCreatedUser, error) {
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

	users := make(map[string]demoCreatedUser, len(inputs))
	for key, input := range inputs {
		var created demoCreatedUser
		if err := s.apiJSON(ctx, http.MethodPost, "/api/users", map[string]any{"spec": input}, &created); err != nil {
			return nil, fmt.Errorf("create demo user %s: %w", key, err)
		}
		users[key] = created
		fmt.Fprintf(s.stdout, "demo user: role=%s username=%s password=%s\n", key, created.Spec.Username, created.Status.Password)
	}
	return users, nil
}

func (s *demoSeeder) createGroups(ctx context.Context, users map[string]demoCreatedUser) (map[string]demoGroup, error) {
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
		if err := s.apiJSON(ctx, http.MethodPost, "/api/groups", map[string]any{"spec": input}, &group); err != nil {
			return nil, fmt.Errorf("create demo group %s: %w", key, err)
		}
		groups[key] = group
	}

	if err := s.apiJSON(ctx, http.MethodPost, "/api/groups/"+groups["engineers"].id()+"/members/"+users["lead"].id(), nil, nil); err != nil {
		return nil, fmt.Errorf("add lead to engineers: %w", err)
	}
	if err := s.apiJSON(ctx, http.MethodPost, "/api/groups/"+groups["engineers"].id()+"/members/"+users["engineer"].id(), nil, nil); err != nil {
		return nil, fmt.Errorf("add engineer to engineers: %w", err)
	}
	if err := s.apiJSON(ctx, http.MethodPost, "/api/groups/"+groups["stakeholders"].id()+"/members/"+users["reporter"].id(), nil, nil); err != nil {
		return nil, fmt.Errorf("add reporter to stakeholders: %w", err)
	}
	return groups, nil
}

func (s *demoSeeder) createProject(ctx context.Context, leadUserID string) (tracker.Project, error) {
	input := tracker.CreateProjectInput{
		Key:         "DEMO" + s.suffix,
		Name:        "Demo Project " + s.suffix,
		Description: "Seeded project for demos",
		LeadUserID:  leadUserID,
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

func (s *demoSeeder) bindProjectAccess(ctx context.Context, groupID string, leadUserID string, projectID string) error {
	if err := s.apiJSON(ctx, http.MethodPost, "/api/role-bindings", map[string]demoRoleBinding{
		"spec": {
			RoleName:    authz.RoleProjectOwner,
			SubjectType: authz.BindingTargetUser,
			SubjectID:   leadUserID,
			Scope:       authz.ScopeKindProject,
			ProjectID:   projectID,
		},
	}, nil); err != nil {
		return fmt.Errorf("bind demo project owner: %w", err)
	}
	if err := s.apiJSON(ctx, http.MethodPost, "/api/role-bindings", map[string]demoRoleBinding{
		"spec": {
			RoleName:    authz.RoleProjectMember,
			SubjectType: authz.BindingTargetGroup,
			SubjectID:   groupID,
			Scope:       authz.ScopeKindProject,
			ProjectID:   projectID,
		},
	}, nil); err != nil {
		return fmt.Errorf("bind demo project members: %w", err)
	}
	fmt.Fprintln(s.stdout, "demo role binding: project owner/member")
	return nil
}

type demoPlanningAssets struct {
	ComponentID string
	VersionID   string
	SprintID    string
}

func (s *demoSeeder) createProjectPlanning(ctx context.Context, projectID string, users map[string]demoCreatedUser) (demoPlanningAssets, error) {
	statuses := []tracker.ProjectStatusInput{
		{Slug: "todo", Name: "To Do"},
		{Slug: "in_progress", Name: "In Progress"},
		{Slug: "review", Name: "Review"},
		{Slug: "done", Name: "Done"},
	}
	if err := s.apiJSON(ctx, http.MethodPut, "/api/projects/"+projectID+"/statuses", map[string]any{
		"spec": map[string]any{"statuses": statuses},
	}, nil); err != nil {
		return demoPlanningAssets{}, fmt.Errorf("replace demo workflow statuses: %w", err)
	}
	fmt.Fprintln(s.stdout, "demo workflow: todo,in_progress,review,done")

	if err := s.apiJSON(ctx, http.MethodPost, "/api/projects/"+projectID+"/boards", map[string]any{
		"spec": map[string]any{
			"name":         "Delivery Board",
			"description":  "Seeded board for demos",
			"status_slugs": []string{"todo", "in_progress", "review", "done"},
		},
	}, nil); err != nil {
		return demoPlanningAssets{}, fmt.Errorf("create demo board: %w", err)
	}
	fmt.Fprintln(s.stdout, "demo board: Delivery Board")

	var component demoIDResource
	if err := s.apiJSON(ctx, http.MethodPost, "/api/projects/"+projectID+"/components", map[string]any{
		"spec": map[string]any{
			"name":                "Platform",
			"description":         "Backend and automation foundation",
			"owner_user_id":       users["lead"].id(),
			"default_assignee_id": users["engineer"].id(),
		},
	}, &component); err != nil {
		return demoPlanningAssets{}, fmt.Errorf("create demo component: %w", err)
	}
	fmt.Fprintln(s.stdout, "demo component: Platform")

	var version demoIDResource
	if err := s.apiJSON(ctx, http.MethodPost, "/api/projects/"+projectID+"/versions", map[string]any{
		"spec": map[string]any{
			"name":        "2026.7",
			"description": "Demo release target",
			"target_date": "2026-07-31",
		},
	}, &version); err != nil {
		return demoPlanningAssets{}, fmt.Errorf("create demo version: %w", err)
	}
	fmt.Fprintln(s.stdout, "demo version: 2026.7")

	if err := s.apiJSON(ctx, http.MethodPost, "/api/projects/"+projectID+"/custom-fields", map[string]any{
		"spec": map[string]any{
			"key":        "severity",
			"name":       "Severity",
			"field_type": "single_select",
			"required":   true,
			"options":    []string{"Low", "Medium", "High"},
		},
	}, nil); err != nil {
		return demoPlanningAssets{}, fmt.Errorf("create demo custom field: %w", err)
	}
	fmt.Fprintln(s.stdout, "demo custom field: severity")

	var sprint demoIDResource
	if err := s.apiJSON(ctx, http.MethodPost, "/api/projects/"+projectID+"/sprints", map[string]any{
		"spec": map[string]any{
			"name":       "Sprint Demo",
			"goal":       "Show planning and automation surfaces",
			"start_date": "2026-06-17",
			"end_date":   "2026-07-01",
		},
	}, &sprint); err != nil {
		return demoPlanningAssets{}, fmt.Errorf("create demo sprint: %w", err)
	}
	fmt.Fprintln(s.stdout, "demo sprint: Sprint Demo")

	return demoPlanningAssets{
		ComponentID: component.id(),
		VersionID:   version.id(),
		SprintID:    sprint.id(),
	}, nil
}

func (s *demoSeeder) createTicketHook(ctx context.Context, projectID string) error {
	if err := s.apiJSON(ctx, http.MethodPost, "/api/projects/"+projectID+"/ticket-hooks", map[string]any{
		"spec": map[string]any{
			"name":     "demo-label",
			"event":    tracker.HookEventTicketCreate,
			"phase":    tracker.HookPhaseBefore,
			"enabled":  true,
			"position": 10,
			"engine": map[string]any{
				"type": tracker.HookEngineLua,
				"script": `
ticket.labels = ticket.labels or {}
ticket.labels[#ticket.labels + 1] = "demo-hook"
return { ticket = ticket }
`,
			},
		},
	}, nil); err != nil {
		return fmt.Errorf("create demo ticket hook: %w", err)
	}
	fmt.Fprintln(s.stdout, "demo ticket hook: demo-label")
	return nil
}

func (s *demoSeeder) createTickets(ctx context.Context, projectID string, assets demoPlanningAssets) error {
	epic, err := s.createDemoTicket(ctx, projectID, tracker.CreateTicketInput{
		Title:       "Launch customer onboarding",
		Description: "Epic for the seeded onboarding workflow.",
		Priority:    "High",
		Type:        "Epic",
		Status:      "todo",
		ComponentID: assets.ComponentID,
		VersionID:   assets.VersionID,
		StartDate:   "2026-06-17",
		DueDate:     "2026-07-31",
		Labels:      []string{"demo", "roadmap"},
		CustomFields: map[string]any{
			"severity": "High",
		},
	})
	if err != nil {
		return err
	}

	tickets := []tracker.CreateTicketInput{
		{
			Title:          "Prepare customer onboarding board",
			Description:    "Create the initial backlog and workflow states.",
			Priority:       "High",
			Type:           "Task",
			ParentTicketID: epic.id(),
			SprintID:       assets.SprintID,
			ComponentID:    assets.ComponentID,
			VersionID:      assets.VersionID,
			Labels:         []string{"demo", "planning"},
			CustomFields:   map[string]any{"severity": "High"},
		},
		{
			Title:          "Investigate login audit trail",
			Description:    "Capture activity that helps explain auth changes.",
			Priority:       "Medium",
			Type:           "Bug",
			ParentTicketID: epic.id(),
			SprintID:       assets.SprintID,
			ComponentID:    assets.ComponentID,
			VersionID:      assets.VersionID,
			Labels:         []string{"demo", "auth"},
			CustomFields:   map[string]any{"severity": "Medium"},
		},
		{
			Title:          "Draft roadmap for automation hooks",
			Description:    "Outline Lua and AI hook milestones.",
			Priority:       "Low",
			Type:           "Story",
			ParentTicketID: epic.id(),
			SprintID:       assets.SprintID,
			ComponentID:    assets.ComponentID,
			VersionID:      assets.VersionID,
			Labels:         []string{"demo", "automation"},
			CustomFields:   map[string]any{"severity": "Low"},
		},
	}
	createdTickets := []demoTicketResource{epic}
	for index, ticket := range tickets {
		created, err := s.createDemoTicket(ctx, projectID, ticket)
		if err != nil {
			return fmt.Errorf("create demo ticket %d: %w", index+1, err)
		}
		createdTickets = append(createdTickets, created)
		if index == 1 {
			status := "in_progress"
			update := tracker.UpdateTicketInput{Status: &status}
			if err := s.apiJSON(ctx, http.MethodPatch, "/api/tickets/"+created.Metadata.ID, map[string]tracker.UpdateTicketInput{"spec": update}, nil); err != nil {
				return fmt.Errorf("update demo ticket %s: %w", created.Status.Key, err)
			}
		}
	}
	reordered := make([]string, 0, len(createdTickets))
	for index := len(createdTickets) - 1; index >= 0; index-- {
		reordered = append(reordered, createdTickets[index].id())
	}
	if err := s.apiJSON(ctx, http.MethodPatch, "/api/projects/"+projectID+"/backlog", map[string]any{
		"spec": map[string]any{"ticket_ids": reordered},
	}, nil); err != nil {
		return fmt.Errorf("reorder demo backlog: %w", err)
	}
	fmt.Fprintln(s.stdout, "demo backlog: reordered")
	return nil
}

func (s *demoSeeder) createDemoTicket(ctx context.Context, projectID string, ticket tracker.CreateTicketInput) (demoTicketResource, error) {
	var created demoTicketResource
	if err := s.apiJSON(ctx, http.MethodPost, "/api/projects/"+projectID+"/tickets", map[string]tracker.CreateTicketInput{"spec": ticket}, &created); err != nil {
		return demoTicketResource{}, err
	}
	fmt.Fprintf(s.stdout, "demo ticket: key=%s title=%q\n", created.Status.Key, created.Spec.Title)
	return created, nil
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
