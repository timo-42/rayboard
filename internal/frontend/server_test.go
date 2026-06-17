package frontend

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIndex(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	NewHandler("http://backend.test").ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, "Rayboard") ||
		!strings.Contains(body, "http://backend.test") ||
		!strings.Contains(body, "/static/app.js") ||
		!strings.Contains(body, `class="design-selector-home"`) ||
		!strings.Contains(body, "Design Selector") ||
		!strings.Contains(body, `href="/docs"`) ||
		!strings.Contains(body, `href="/api/docs"`) ||
		!strings.Contains(body, `href="/api/docs/redoc"`) ||
		!strings.Contains(body, "Engine Workbench") ||
		!strings.Contains(body, `id="app-nav"`) ||
		!strings.Contains(body, `id="dashboard-view"`) ||
		!strings.Contains(body, `id="issue-view"`) ||
		!strings.Contains(body, `id="create-page-view"`) ||
		!strings.Contains(body, `id="create-page-submit-form"`) ||
		!strings.Contains(body, `id="rbac-panel"`) ||
		!strings.Contains(body, `id="rbac-user-form"`) ||
		!strings.Contains(body, `id="rbac-group-form"`) ||
		!strings.Contains(body, `id="rbac-member-form"`) ||
		!strings.Contains(body, `id="rbac-binding-form"`) ||
		!strings.Contains(body, `id="settings-panel"`) ||
		!strings.Contains(body, `id="audit-form"`) ||
		!strings.Contains(body, `id="audit-log"`) ||
		!strings.Contains(body, `id="openrouter-provider-form"`) ||
		!strings.Contains(body, `id="openrouter-providers"`) ||
		!strings.Contains(body, `id="notification-destination-form"`) ||
		!strings.Contains(body, `id="notification-destinations"`) ||
		!strings.Contains(body, `id="notification-policy-form"`) ||
		!strings.Contains(body, `id="notification-policies"`) ||
		!strings.Contains(body, `id="notification-hook-form"`) ||
		!strings.Contains(body, `id="notification-hooks"`) ||
		!strings.Contains(body, `id="notification-hook-preview-output"`) ||
		!strings.Contains(body, `id="project-preference-form"`) ||
		!strings.Contains(body, `id="notification-delivery-form"`) ||
		!strings.Contains(body, `id="notification-deliveries"`) ||
		!strings.Contains(body, `id="engine-form"`) ||
		!strings.Contains(body, `id="cron-job-form"`) ||
		!strings.Contains(body, `id="cron-jobs"`) ||
		!strings.Contains(body, `id="webhook-form"`) ||
		!strings.Contains(body, `id="webhooks"`) ||
		!strings.Contains(body, `id="ticket-hook-form"`) ||
		!strings.Contains(body, `id="ticket-hooks"`) ||
		!strings.Contains(body, `id="ticket-hook-preview-output"`) ||
		!strings.Contains(body, `id="create-page-form"`) ||
		!strings.Contains(body, `id="create-pages"`) ||
		!strings.Contains(body, `id="notification-inbox"`) ||
		!strings.Contains(body, `id="backlog-panel"`) ||
		!strings.Contains(body, `id="backlog"`) ||
		!strings.Contains(body, `id="workflow-panel"`) ||
		!strings.Contains(body, `id="status-form"`) ||
		!strings.Contains(body, `id="board-form"`) ||
		!strings.Contains(body, `id="boards"`) ||
		!strings.Contains(body, `id="sprint-panel"`) ||
		!strings.Contains(body, `id="release-panel"`) ||
		!strings.Contains(body, `id="roadmap-panel"`) ||
		!strings.Contains(body, `id="field-panel"`) ||
		!strings.Contains(body, `id="search-panel"`) ||
		!strings.Contains(body, `id="saved-view-cancel-edit"`) ||
		!strings.Contains(body, `id="account-panel"`) ||
		!strings.Contains(body, `id="ticket-columns"`) ||
		!strings.Contains(body, `href="/1"`) ||
		!strings.Contains(body, `href="/5"`) {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestAppPageRoutesRenderShell(t *testing.T) {
	for _, path := range []string{
		"/projects",
		"/projects/project_demo",
		"/issues/ticket_demo",
		"/projects/project_demo/create/bug-intake",
		"/profile",
		"/rbac",
		"/admin/rbac",
		"/settings",
		"/search",
		"/automation",
	} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()

			NewHandler("http://backend.test").ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", rec.Code)
			}
			body := rec.Body.String()
			for _, expected := range []string{"Rayboard", `id="app-nav"`, `id="dashboard-view"`, `id="issue-view"`} {
				if !strings.Contains(body, expected) {
					t.Fatalf("expected body to contain %q: %s", expected, body)
				}
			}
		})
	}
}

func TestEmbeddedAppSupportsWebsitePages(t *testing.T) {
	app, err := assets.ReadFile("static/app.js")
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	css, err := assets.ReadFile("static/app.css")
	if err != nil {
		t.Fatalf("read app.css: %v", err)
	}
	appText := string(app)
	for _, expected := range []string{
		"currentRoute",
		"navigate",
		"renderDashboard",
		"renderIssue",
		"renderCreatePageView",
		"renderRBAC",
		"renderSettings",
		"renderAuditLog",
		"renderProjectNotificationPreferences",
		"renderNotificationDeliveries",
		"renderBacklog",
		"renderWorkflowPanel",
		"renderTicketHooks",
		"loadDashboardSummaries",
		"loadSelectedIssue",
		"loadCreatePageForRoute",
		"loadRBAC",
		"loadSettingsPage",
		"loadBacklog",
		"loadWorkflowStatuses",
		"loadBoards",
		"loadBoardTickets",
		"loadAuditLog",
		"loadOpenRouterProviders",
		"loadNotificationDestinations",
		"loadNotificationPolicies",
		"loadNotificationHooks",
		"loadProjectNotificationPreferences",
		"loadNotificationDeliveries",
		"loadCronJobs",
		"loadWebhooks",
		"loadTicketHooks",
		"loadCreatePages",
		"normalizeAuditEntry",
		"normalizeOpenRouterProvider",
		"normalizeNotificationDestination",
		"normalizeNotificationPolicy",
		"normalizeNotificationHook",
		"normalizeNotificationHookRun",
		"normalizeNotificationDelivery",
		"normalizeCronJob",
		"normalizeWebhook",
		"normalizeTicketHook",
		"normalizeCreatePage",
		"normalizeCreatePageSchema",
		"normalizeWorkflowStatus",
		"normalizeBoard",
		"normalizeBoardTickets",
		"cronJobSpec",
		"webhookSpec",
		"ticketHookSpec",
		"ticketHookPreviewSpec",
		"notificationPreferenceSpec",
		"createPageSpec",
		"rbacUserForm",
		"rbacGroupForm",
		"rbacMemberForm",
		"rbacBindingForm",
		"/api/users",
		"/api/groups/${data.group_id}/members/${data.user_id}",
		"data-rbac-user-disabled",
		"data-remove-group-member",
		"/api/role-bindings",
		"/api/role-bindings/${deleteBinding.dataset.deleteBindingId}",
		"/api/settings",
		"/api/audit-log",
		"/api/openrouter-providers",
		"/api/openrouter-providers/${form.dataset.openrouterProviderForm}",
		"data-delete-openrouter-provider-id",
		"/api/notification-destinations",
		"/api/projects/${projectID}/notification-destinations",
		"data-test-notification-destination-id",
		"data-delete-notification-destination-id",
		"/api/notification-policies",
		"/api/projects/${projectID}/notification-policies",
		"data-delete-notification-policy-id",
		"/api/notification-hooks",
		"/api/projects/${projectID}/notification-hooks",
		"/api/notification-hooks/${preview.dataset.previewNotificationHookId}/preview",
		"data-delete-notification-hook-id",
		"/api/projects/${projectID}/notification-preferences",
		"/api/notification-deliveries${query}",
		"/api/projects/${projectID}/notification-deliveries${query}",
		"/api/notification-deliveries/${retry.dataset.retryNotificationDeliveryId}/retry",
		"data-retry-notification-delivery-id",
		"/api/projects/${state.selectedProject.id}/backlog",
		"data-backlog-move-id",
		"/api/projects/${state.selectedProject.id}/statuses",
		"/api/projects/${state.selectedProject.id}/boards",
		"/api/boards/${boardID}/tickets",
		"data-select-board-id",
		"data-delete-board-id",
		"/api/cron-jobs?${query.toString()}",
		"/api/cron-jobs/${run.dataset.runCronJobId}/run",
		"data-delete-cron-job-id",
		"/api/projects/${projectID}/webhooks?limit=100",
		"/api/webhook-definitions/${rotate.dataset.rotateWebhookTokenId}/rotate-token",
		"data-delete-webhook-id",
		"/api/projects/${projectID}/ticket-hooks?limit=100",
		"/api/ticket-hooks/${preview.dataset.previewTicketHookId}/preview",
		"data-delete-ticket-hook-id",
		"/api/projects/${projectID}/ticket-create-pages?include_disabled=true&limit=100",
		"/api/projects/${projectID}/ticket-create-pages/${encodeURIComponent(slug)}/schema",
		"/api/projects/${route.projectID}/ticket-create-pages/${encodeURIComponent(route.slug)}/submit",
		"/projects/${encodeURIComponent(page.project_id)}/create/${encodeURIComponent(page.slug)}",
		"/api/ticket-create-pages/${toggle.dataset.toggleCreatePageId}",
		"data-delete-create-page-id",
		"/api/me/notification-preferences",
	} {
		if !strings.Contains(appText, expected) {
			t.Fatalf("expected app.js to contain %q", expected)
		}
	}
	cssText := string(css)
	for _, expected := range []string{
		".app-nav",
		".dashboard-view",
		".metric-grid",
		".issue-view",
		".create-page-view",
		".create-page-submit-form",
		".create-page-result",
		".rbac-panel",
		".admin-form",
		".admin-actions",
		".member-list",
		".settings-panel",
		".settings-grid",
		".audit-form",
		".audit-log",
		".openrouter-provider-form",
		".openrouter-provider-list",
		".notification-destination-form",
		".notification-destination-list",
		".notification-policy-form",
		".notification-policy-list",
		".notification-hook-form",
		".notification-hook-list",
		".notification-hook-preview",
		".notification-admin-card",
		".notification-delivery-form",
		".notification-delivery-list",
		".notification-delivery-item",
		".backlog-panel",
		".backlog-list",
		".workflow-panel",
		".status-form",
		".board-form",
		".board-list",
		".cron-job-form",
		".cron-job-list",
		".webhook-form",
		".webhook-list",
		".ticket-hook-panel",
		".ticket-hook-list",
		".create-page-form",
		".create-page-list",
	} {
		if !strings.Contains(cssText, expected) {
			t.Fatalf("expected app.css to contain %q", expected)
		}
	}
}

func TestEmbeddedAppSupportsAttachments(t *testing.T) {
	app, err := assets.ReadFile("static/app.js")
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	css, err := assets.ReadFile("static/app.css")
	if err != nil {
		t.Fatalf("read app.css: %v", err)
	}
	appText := string(app)
	for _, expected := range []string{
		"loadAttachments",
		"normalizeAttachment",
		"/api/tickets/${ticketID}/attachments",
		"/api/attachments/${attachment.id}/download",
		"data-delete-attachment-id",
		"new FormData()",
	} {
		if !strings.Contains(appText, expected) {
			t.Fatalf("expected app.js to contain %q", expected)
		}
	}
	cssText := string(css)
	for _, expected := range []string{
		".ticket-attachments",
		".attachment-item",
		".attachment-form",
	} {
		if !strings.Contains(cssText, expected) {
			t.Fatalf("expected app.css to contain %q", expected)
		}
	}
}

func TestEmbeddedAppSupportsComments(t *testing.T) {
	app, err := assets.ReadFile("static/app.js")
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	css, err := assets.ReadFile("static/app.css")
	if err != nil {
		t.Fatalf("read app.css: %v", err)
	}
	appText := string(app)
	for _, expected := range []string{
		"loadComments",
		"normalizeComment",
		"/api/tickets/${ticketID}/comments",
		"/api/comments/${deleteComment.dataset.deleteCommentId}",
		"data-delete-comment-id",
		"data-comment-form",
	} {
		if !strings.Contains(appText, expected) {
			t.Fatalf("expected app.js to contain %q", expected)
		}
	}
	cssText := string(css)
	for _, expected := range []string{
		".ticket-comments",
		".comment-item",
		".comment-form",
	} {
		if !strings.Contains(cssText, expected) {
			t.Fatalf("expected app.css to contain %q", expected)
		}
	}
}

func TestEmbeddedAppSupportsTicketActivity(t *testing.T) {
	app, err := assets.ReadFile("static/app.js")
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	css, err := assets.ReadFile("static/app.css")
	if err != nil {
		t.Fatalf("read app.css: %v", err)
	}
	appText := string(app)
	for _, expected := range []string{
		"loadActivity",
		"normalizeActivity",
		"activityNode",
		"/api/tickets/${ticketID}/activity",
		"ticket.updated",
		"activityDataLabel",
	} {
		if !strings.Contains(appText, expected) {
			t.Fatalf("expected app.js to contain %q", expected)
		}
	}
	cssText := string(css)
	for _, expected := range []string{
		".ticket-activity",
		".activity-heading",
		".activity-list",
		".activity-item",
	} {
		if !strings.Contains(cssText, expected) {
			t.Fatalf("expected app.css to contain %q", expected)
		}
	}
}

func TestEmbeddedAppSupportsNotifications(t *testing.T) {
	app, err := assets.ReadFile("static/app.js")
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	css, err := assets.ReadFile("static/app.css")
	if err != nil {
		t.Fatalf("read app.css: %v", err)
	}
	appText := string(app)
	for _, expected := range []string{
		"loadNotifications",
		"normalizeNotification",
		"/api/notifications${query}",
		"/api/notifications/read-all",
		"/api/notifications/${button.dataset.notificationId}/${action}",
		"data-notification-read-state",
	} {
		if !strings.Contains(appText, expected) {
			t.Fatalf("expected app.js to contain %q", expected)
		}
	}
	cssText := string(css)
	for _, expected := range []string{
		".notification-inbox",
		".notification-item",
		".notification-item.is-unread",
	} {
		if !strings.Contains(cssText, expected) {
			t.Fatalf("expected app.css to contain %q", expected)
		}
	}
}

func TestEmbeddedAppSupportsSearchSavedViews(t *testing.T) {
	app, err := assets.ReadFile("static/app.js")
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	css, err := assets.ReadFile("static/app.css")
	if err != nil {
		t.Fatalf("read app.css: %v", err)
	}
	appText := string(app)
	for _, expected := range []string{
		"loadSavedViews",
		"runSearch",
		"normalizeSavedView",
		"savedViewSpecFromForm",
		"savedViewUpdateSpec",
		"editSavedView",
		"/api/search",
		"/api/saved-views",
		"/api/saved-views/${editingID}",
		"data-apply-saved-view-id",
		"data-edit-saved-view-id",
		"data-delete-saved-view-id",
	} {
		if !strings.Contains(appText, expected) {
			t.Fatalf("expected app.js to contain %q", expected)
		}
	}
	cssText := string(css)
	for _, expected := range []string{
		".search-panel",
		".search-results",
		".saved-view-list",
	} {
		if !strings.Contains(cssText, expected) {
			t.Fatalf("expected app.css to contain %q", expected)
		}
	}
}

func TestEmbeddedAppSupportsSprints(t *testing.T) {
	app, err := assets.ReadFile("static/app.js")
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	css, err := assets.ReadFile("static/app.css")
	if err != nil {
		t.Fatalf("read app.css: %v", err)
	}
	appText := string(app)
	for _, expected := range []string{
		"loadSprints",
		"renderSprints",
		"normalizeSprint",
		"/api/projects/${state.selectedProject.id}/sprints",
		"/api/sprints/${start.dataset.startSprintId}/start",
		"/api/sprints/${complete.dataset.completeSprintId}/complete",
		"/api/tickets/${assignSprint.dataset.assignSprintId}/sprint",
		"data-ticket-sprint-control",
	} {
		if !strings.Contains(appText, expected) {
			t.Fatalf("expected app.js to contain %q", expected)
		}
	}
	cssText := string(css)
	for _, expected := range []string{
		".sprint-panel",
		".sprint-form",
		".sprint-item",
		".ticket-sprint",
	} {
		if !strings.Contains(cssText, expected) {
			t.Fatalf("expected app.css to contain %q", expected)
		}
	}
}

func TestEmbeddedAppSupportsComponentsVersions(t *testing.T) {
	app, err := assets.ReadFile("static/app.js")
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	css, err := assets.ReadFile("static/app.css")
	if err != nil {
		t.Fatalf("read app.css: %v", err)
	}
	appText := string(app)
	for _, expected := range []string{
		"loadComponents",
		"loadVersions",
		"normalizeComponent",
		"normalizeVersion",
		"/api/projects/${state.selectedProject.id}/components",
		"/api/projects/${state.selectedProject.id}/versions",
		"/api/tickets/${assignPlanning.dataset.assignPlanningId}",
		"data-ticket-planning-control",
	} {
		if !strings.Contains(appText, expected) {
			t.Fatalf("expected app.js to contain %q", expected)
		}
	}
	cssText := string(css)
	for _, expected := range []string{
		".release-panel",
		".component-form",
		".version-form",
		".ticket-planning",
	} {
		if !strings.Contains(cssText, expected) {
			t.Fatalf("expected app.css to contain %q", expected)
		}
	}
}

func TestEmbeddedAppSupportsRoadmap(t *testing.T) {
	app, err := assets.ReadFile("static/app.js")
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	css, err := assets.ReadFile("static/app.css")
	if err != nil {
		t.Fatalf("read app.css: %v", err)
	}
	appText := string(app)
	for _, expected := range []string{
		"loadRoadmap",
		"renderRoadmap",
		"normalizeRoadmapItem",
		"roadmapEpics",
		"/api/projects/${state.selectedProject.id}/roadmap",
		"ticket-parent-id",
		"roadmap-progress",
	} {
		if !strings.Contains(appText, expected) {
			t.Fatalf("expected app.js to contain %q", expected)
		}
	}
	cssText := string(css)
	for _, expected := range []string{
		".roadmap-panel",
		".roadmap-list",
		".roadmap-item",
		".roadmap-progress",
	} {
		if !strings.Contains(cssText, expected) {
			t.Fatalf("expected app.css to contain %q", expected)
		}
	}
}

func TestEmbeddedAppSupportsTicketLabels(t *testing.T) {
	app, err := assets.ReadFile("static/app.js")
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	css, err := assets.ReadFile("static/app.css")
	if err != nil {
		t.Fatalf("read app.css: %v", err)
	}
	appText := string(app)
	for _, expected := range []string{
		"parseLabels",
		"labelControlNode",
		"data-ticket-label-control",
		"data-update-labels-id",
		"labels: parseLabels",
	} {
		if !strings.Contains(appText, expected) {
			t.Fatalf("expected app.js to contain %q", expected)
		}
	}
	cssText := string(css)
	for _, expected := range []string{
		".ticket-labels",
		".label-chips",
		".ticket-label-controls",
	} {
		if !strings.Contains(cssText, expected) {
			t.Fatalf("expected app.css to contain %q", expected)
		}
	}
}

func TestEmbeddedAppSupportsCustomFields(t *testing.T) {
	app, err := assets.ReadFile("static/app.js")
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	css, err := assets.ReadFile("static/app.css")
	if err != nil {
		t.Fatalf("read app.css: %v", err)
	}
	appText := string(app)
	for _, expected := range []string{
		"loadCustomFields",
		"renderCustomFields",
		"normalizeCustomField",
		"parseCustomFields",
		"/api/projects/${state.selectedProject.id}/custom-fields",
		"/api/custom-fields/${remove.dataset.deleteFieldId}",
		"data-ticket-custom-field-control",
		"data-update-custom-fields-id",
	} {
		if !strings.Contains(appText, expected) {
			t.Fatalf("expected app.js to contain %q", expected)
		}
	}
	cssText := string(css)
	for _, expected := range []string{
		".field-panel",
		".field-form",
		".field-list",
		".ticket-custom-fields",
	} {
		if !strings.Contains(cssText, expected) {
			t.Fatalf("expected app.css to contain %q", expected)
		}
	}
}

func TestEmbeddedAppSupportsAPITokens(t *testing.T) {
	app, err := assets.ReadFile("static/app.js")
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	css, err := assets.ReadFile("static/app.css")
	if err != nil {
		t.Fatalf("read app.css: %v", err)
	}
	appText := string(app)
	for _, expected := range []string{
		"loadTokens",
		"renderTokens",
		"normalizeToken",
		"/api/tokens",
		"data-revoke-token-id",
		"createdToken",
	} {
		if !strings.Contains(appText, expected) {
			t.Fatalf("expected app.js to contain %q", expected)
		}
	}
	cssText := string(css)
	for _, expected := range []string{
		".account-panel",
		".token-list",
		".token-item",
		".created-token",
	} {
		if !strings.Contains(cssText, expected) {
			t.Fatalf("expected app.css to contain %q", expected)
		}
	}
}

func TestDesignVariantRoute(t *testing.T) {
	tests := []struct {
		path      string
		bodyClass string
		name      string
	}{
		{path: "/1", bodyClass: "design-operations", name: "Operations"},
		{path: "/2", bodyClass: "design-planning", name: "Planning"},
		{path: "/3", bodyClass: "design-automation", name: "Automation"},
		{path: "/4", bodyClass: "design-triage", name: "Triage"},
		{path: "/5", bodyClass: "design-executive", name: "Executive"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			NewHandler("http://backend.test").ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", rec.Code)
			}
			body := rec.Body.String()
			for _, expected := range []string{
				`class="` + tt.bodyClass + `"`,
				tt.name,
				`href="` + tt.path + `" aria-current="page"`,
				`href="/1"`,
				`href="/5"`,
			} {
				if !strings.Contains(body, expected) {
					t.Fatalf("expected body to contain %q: %s", expected, body)
				}
			}
		})
	}
}

func TestAPIProxy(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/health" {
			t.Fatalf("unexpected proxied path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT proxy, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	t.Cleanup(backend.Close)

	req := httptest.NewRequest(http.MethodPut, "/api/health", nil)
	rec := httptest.NewRecorder()

	NewHandler(backend.URL).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, `"status":"ok"`) {
		t.Fatalf("unexpected proxy body: %s", body)
	}
}

func TestDocsHTML(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rec := httptest.NewRecorder()

	NewHandler("http://backend.test").ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if contentType := rec.Header().Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		t.Fatalf("expected HTML docs content type, got %q", contentType)
	}
	body := rec.Body.String()
	for _, expected := range []string{"Rayboard Docs", "Rayboard Documentation", "/docs/api", "/docs/auth-rbac"} {
		if !strings.Contains(body, expected) {
			t.Fatalf("expected docs page to contain %q", expected)
		}
	}
}

func TestDocsNamedPage(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/docs/api", nil)
	rec := httptest.NewRecorder()

	NewHandler("http://backend.test").ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, "API Guide") || !strings.Contains(body, "/api/openapi.json") {
		t.Fatalf("unexpected docs body: %s", body)
	}
}
