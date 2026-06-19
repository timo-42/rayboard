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
		!strings.Contains(body, `class="production-dashboard"`) ||
		!strings.Contains(body, `href="/docs"`) ||
		!strings.Contains(body, `href="/api/docs"`) ||
		!strings.Contains(body, `href="/api/docs/redoc"`) ||
		!strings.Contains(body, `href="/profile"`) ||
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
		!strings.Contains(body, `id="rbac-permission-form"`) ||
		!strings.Contains(body, `id="rbac-permissions"`) ||
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
		!strings.Contains(body, `id="engine-result-summary"`) ||
		!strings.Contains(body, `id="cron-job-form"`) ||
		!strings.Contains(body, `id="cron-jobs"`) ||
		!strings.Contains(body, `id="webhook-form"`) ||
		!strings.Contains(body, `id="webhooks"`) ||
		!strings.Contains(body, `id="ticket-hook-form"`) ||
		!strings.Contains(body, `id="ticket-hooks"`) ||
		!strings.Contains(body, `id="ticket-hook-preview-output"`) ||
		!strings.Contains(body, `id="create-page-form"`) ||
		!strings.Contains(body, `id="create-pages"`) ||
		!strings.Contains(body, `id="ticket-filter-form"`) ||
		!strings.Contains(body, `id="ticket-filter-label"`) ||
		!strings.Contains(body, `id="ticket-filter-component"`) ||
		!strings.Contains(body, `id="ticket-filter-version"`) ||
		!strings.Contains(body, `id="ticket-filter-summary"`) ||
		!strings.Contains(body, `id="label-panel"`) ||
		!strings.Contains(body, `id="project-label-form"`) ||
		!strings.Contains(body, `id="project-labels"`) ||
		!strings.Contains(body, `id="notification-inbox"`) ||
		!strings.Contains(body, `id="nav-unread-count"`) ||
		!strings.Contains(body, `id="backlog-panel"`) ||
		!strings.Contains(body, `id="backlog"`) ||
		!strings.Contains(body, `id="workflow-panel"`) ||
		!strings.Contains(body, `id="status-form"`) ||
		!strings.Contains(body, `id="board-form"`) ||
		!strings.Contains(body, `id="board-saved-view-filter"`) ||
		!strings.Contains(body, `id="board-saved-view-status"`) ||
		!strings.Contains(body, `id="boards"`) ||
		!strings.Contains(body, `id="sprint-panel"`) ||
		!strings.Contains(body, `id="sprint-report"`) ||
		!strings.Contains(body, `id="release-panel"`) ||
		!strings.Contains(body, `id="version-report"`) ||
		!strings.Contains(body, `id="roadmap-panel"`) ||
		!strings.Contains(body, `id="roadmap-dependencies"`) ||
		!strings.Contains(body, `id="field-panel"`) ||
		!strings.Contains(body, `id="search-panel"`) ||
		!strings.Contains(body, `id="saved-view-cancel-edit"`) ||
		!strings.Contains(body, `id="account-panel"`) ||
		!strings.Contains(body, `id="ticket-columns"`) {
		t.Fatalf("unexpected body: %s", body)
	}
	body := rec.Body.String()
	for _, unexpected := range []string{"Design Selector", "UI Designs", `class="design-selector-home"`, `class="tool-block design-selector"`, `href="/1"`, `href="/5"`} {
		if strings.Contains(body, unexpected) {
			t.Fatalf("root dashboard should not contain %q: %s", unexpected, body)
		}
	}
	if strings.Contains(body, `href="/settings"`) {
		t.Fatalf("production primary nav should not link to settings: %s", body)
	}
}

func TestDesignPreviewRoutesRenderSelector(t *testing.T) {
	for _, path := range []string{"/1", "/2", "/3", "/4", "/5"} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()

			NewHandler("http://backend.test").ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", rec.Code)
			}
			body := rec.Body.String()
			for _, expected := range []string{"UI Designs", `class="tool-block design-selector"`, `href="/1"`, `href="/5"`, `aria-label="UI design variants"`} {
				if !strings.Contains(body, expected) {
					t.Fatalf("expected design preview body to contain %q: %s", expected, body)
				}
			}
		})
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
		"isDocumentLink",
		"renderIssue",
		"renderCreatePageView",
		"createPageLayout",
		"createPageLayoutNode",
		"createPageGroupNode",
		"createPageCustomFieldKey",
		"renderRBAC",
		"renderSettings",
		"renderAuditLog",
		"renderProjectNotificationPreferences",
		"renderNotificationDeliveries",
		"renderNotificationDeliverySummary",
		"notificationDeliverySummary",
		"deliveryUpdatedAt",
		"queued",
		"sending",
		"renderNotificationBadge",
		"unreadNotificationCount",
		"renderBacklog",
		"renderWorkflowPanel",
		"renderTicketHooks",
		"renderEngineResultSummary",
		"engineBadge",
		"action_previews",
		"previewed",
		"loadDashboardSummaries",
		"loadSelectedIssue",
		"loadCreatePageForRoute",
		"loadRBAC",
		"loadSettingsPage",
		"loadBacklog",
		"loadWorkflowStatuses",
		"loadBoards",
		"loadBoardTickets",
		"reorderBacklogTicket",
		"dataTransferHasType",
		"loadAuditLog",
		"loadOpenRouterProviders",
		"loadNotificationDestinations",
		"loadNotificationPolicies",
		"loadNotificationHooks",
		"renderNotificationPolicyDestinationOptions",
		"renderNotificationHookPreviewPolicyOptions",
		"renderNotificationHookPreviewDestinationOptions",
		"applyNotificationHookPreviewPolicy",
		"notificationHookPreviewDisplay",
		"els.notificationPolicyScope.value === \"project\" ? selectedNotificationPolicyProjectID() : \"\"",
		"els.notificationHookScope.value === \"project\" ? selectedNotificationHookProjectID() : \"\"",
		"availableNotificationDestinations",
		"renderDestinationMultiSelect",
		"selectedFormValues",
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
		"rbacPermissionForm",
		"loadRBACEffectivePermissions",
		"renderRBACPermissions",
		"normalizeEffectivePermissions",
		"/api/users",
		"/api/users/${data.user_id}/effective-permissions?${params.toString()}",
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
		"backlogSummaryNode",
		"backlogSummaryMetrics",
		"backlogSummaryMetricNode",
		"Sprint assigned",
		"data-backlog-move-id",
		"data-backlog-drag-id",
		"data-backlog-sprint-control",
		"data-backlog-assign-sprint-id",
		"data-backlog-remove-sprint-id",
		"/api/tickets/${assignSprint.dataset.backlogAssignSprintId}/sprint",
		"/api/tickets/${removeSprint.dataset.backlogRemoveSprintId}/sprint",
		"/api/projects/${state.selectedProject.id}/statuses",
		"/api/projects/${state.selectedProject.id}/boards",
		"/api/saved-views?project_id=${projectID}&limit=200&offset=0",
		"/api/boards/${form.dataset.boardEditForm}",
		"/api/boards/${boardID}/tickets",
		"selectedBoardSavedViewID",
		"searchAllBoardSavedViewTickets",
		"next_cursor",
		"boardTicketsFromSavedViewMatches",
		"boardSummaryNode",
		"boardSummaryMetrics",
		"boardSummaryMetricNode",
		"filtered_by_saved_view",
		"WIP warnings",
		"parseBoardWIPLimits",
		"over_wip_limit",
		"data-select-board-id",
		"data-board-edit-form",
		"data-board-drop-status",
		"data-board-ticket-id",
		"data-delete-board-id",
		"/api/cron-jobs?${query.toString()}",
		"/api/cron-jobs/${run.dataset.runCronJobId}/run",
		"data-delete-cron-job-id",
		"/api/cron-jobs/${form.dataset.cronJobForm}",
		"renderCronJobEditEngineFields",
		"cronJobEditForm",
		"automationRunSummaryNode",
		"summarizeAutomationRuns",
		"handleAutomationRunFilterChange",
		"filterAutomationRuns",
		"automationRunDurationLabel",
		"automationRunDurationMs",
		"formatDuration",
		"data-automation-run-filter",
		"avg duration",
		"max duration",
		"oldestRunLabel",
		"newestRunLabel",
		"triggerCounts",
		"trigger ${trigger} ${count}",
		"completionRateLabel",
		"failureRateLabel",
		"formatRunRate",
		"completion rate",
		"failure rate",
		"[\"oldest\", summary.oldestRunLabel]",
		"[\"newest\", summary.newestRunLabel]",
		"duration ${formatDuration(durationMs)}",
		"No runs match this filter",
		"latest failure:",
		"data-cron-job-form",
		"data-cron-job-edit-engine-field",
		"/api/projects/${projectID}/webhooks?limit=100",
		"/api/webhook-definitions/${rotate.dataset.rotateWebhookTokenId}/rotate-token",
		"/api/webhook-definitions/${form.dataset.webhookForm}",
		"renderWebhookEditEngineFields",
		"webhookEditForm",
		"data-webhook-form",
		"data-webhook-edit-engine-field",
		"data-delete-webhook-id",
		"/api/projects/${projectID}/ticket-hooks?limit=100",
		"/api/ticket-hooks/${preview.dataset.previewTicketHookId}/preview",
		"/api/ticket-hooks/${hookID}/runs?limit=10",
		"/api/ticket-hooks/${form.dataset.ticketHookForm}",
		"renderTicketHookEditEngineFields",
		"ticketHookEditForm",
		"ticketHookRunListNode",
		"data-load-ticket-hook-runs-id",
		"data-ticket-hook-form",
		"data-ticket-hook-edit-engine-field",
		"data-delete-ticket-hook-id",
		"/api/projects/${projectID}/ticket-create-pages?include_disabled=true&limit=100",
		"/api/projects/${projectID}/ticket-create-pages/${encodeURIComponent(slug)}/schema",
		"/api/ticket-create-pages/${pageID}/runs?limit=10",
		"/api/projects/${route.projectID}/ticket-create-pages/${encodeURIComponent(route.slug)}/submit",
		"/projects/${encodeURIComponent(page.project_id)}/create/${encodeURIComponent(page.slug)}",
		"/api/ticket-create-pages/${toggle.dataset.toggleCreatePageId}",
		"/api/ticket-create-pages/${form.dataset.createPageForm}",
		"renderCreatePageEditLogicFields",
		"createPageEditForm",
		"data-create-page-form",
		"data-create-page-edit-logic-field",
		"createPageRunListNode",
		"data-load-create-page-runs-id",
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
		".nav-badge",
		".dashboard-view",
		".metric-grid",
		".issue-view",
		".create-page-view",
		".create-page-submit-form",
		".create-page-result",
		".create-page-layout-group",
		".create-page-layout-columns",
		".rbac-panel",
		".admin-form",
		".admin-actions",
		".member-list",
		".permission-list",
		".permission-chip",
		".settings-panel",
		".settings-grid",
		".audit-form",
		".audit-log",
		".openrouter-provider-form",
		".openrouter-provider-list",
		".notification-destination-form",
		".notification-destination-list",
		".notification-policy-form",
		".notification-policy-form select[multiple]",
		".notification-policy-list",
		".notification-hook-form",
		".notification-hook-preview-form select[multiple]",
		".notification-hook-list",
		".notification-hook-preview",
		".notification-admin-card",
		".notification-delivery-form",
		".notification-delivery-summary",
		".notification-delivery-metrics",
		".notification-delivery-list",
		".notification-delivery-item",
		".backlog-panel",
		".backlog-list",
		".backlog-summary",
		".backlog-summary-metric",
		".backlog-sprint",
		".backlog-sprint-controls",
		".workflow-panel",
		".status-form",
		".board-form",
		".board-saved-view-filter",
		".board-edit-form",
		".board-list",
		".board-summary",
		".board-summary-metric",
		".column-capacity",
		".is-over-wip",
		".is-dragging",
		".cron-job-form",
		".cron-job-list",
		".cron-job-edit-form",
		".webhook-form",
		".webhook-list",
		".webhook-edit-form",
		".ticket-hook-panel",
		".ticket-hook-list",
		".ticket-hook-edit-form",
		".ticket-hook-run-list",
		".create-page-form",
		".create-page-list",
		".create-page-edit-form",
		".create-page-run-list",
		".automation-run-summary",
		".automation-run-filter",
		".automation-run-summary-error",
		".engine-result-summary",
		".engine-result-badge",
		".engine-action-preview",
	} {
		if !strings.Contains(cssText, expected) {
			t.Fatalf("expected app.css to contain %q", expected)
		}
	}
}

func TestEmbeddedAppSupportsCreatePageLayoutWidgets(t *testing.T) {
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
		"normalizeCreatePageLayoutItem",
		"createPageLayoutHasField",
		"createPageTextLayoutKinds",
		"createPageColumnsLayoutKinds",
		"item.fields.map(normalizeCreatePageLayoutItem)",
		"item.html !== undefined",
		"create-page-layout-heading",
		"create-page-layout-text",
		"custom_fields.",
		"customFields[createPageCustomFieldKey(key)] = value",
	} {
		if !strings.Contains(appText, expected) {
			t.Fatalf("expected app.js to contain %q", expected)
		}
	}
	cssText := string(css)
	for _, expected := range []string{
		".create-page-layout-group",
		".create-page-layout-heading",
		".create-page-layout-text",
		".create-page-layout-fields",
		".create-page-layout-columns > .create-page-layout-fields",
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
		"@username",
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

func TestEmbeddedAppSupportsTicketWatchers(t *testing.T) {
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
		"loadTicketWatchers",
		"normalizeTicketWatcher",
		"watcherNode",
		"/api/tickets/${ticketID}/watchers",
		"/api/tickets/${ticketID}/watchers/me",
		"ticket.watcher_added",
		"watcher_count",
		"watching",
		"data-delete-ticket-id",
		"ticketDeleteButton",
		"ticket.deleted",
	} {
		if !strings.Contains(appText, expected) {
			t.Fatalf("expected app.js to contain %q", expected)
		}
	}
	cssText := string(css)
	for _, expected := range []string{
		".ticket-watchers",
		".ticket-delete",
		".watcher-heading",
		".watcher-list",
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
		"comment_mentioned",
		"notificationTypeLabel",
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
		"loadCustomFields({ renderTickets: false })",
		"renderCustomFieldSearchControls",
		"customFieldSearchExpression",
		"appendSearchFilter",
		"`custom.${field.key}`",
		"normalizeSavedView",
		"savedViewSpecFromForm",
		"savedViewUpdateSpec",
		"savedViewMetadataNode",
		"savedViewMetadataItems",
		"`filter: ${query.filter}`",
		"`columns ${columnCount}`",
		"editSavedView",
		"loadPinnedProjectSavedViews",
		"renderPinnedProjectSavedViews",
		"pinnedProjectViewNode",
		"applySavedView",
		"searchNextCursor",
		"searchCursorStack",
		"renderSearchPagination",
		"renderSavedViewPagination",
		"dataset.searchNext",
		"dataset.savedViewNext",
		"limit=${savedViewPageSize + 1}&offset=${state.savedViewOffset}",
		"/api/search",
		"/api/saved-views",
		"pinned=true&limit=20&offset=0",
		"/api/saved-views/${editingID}",
		"data-apply-saved-view-id",
		"data-apply-pinned-project-view-id",
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
		".custom-field-search-controls",
		".custom-field-search-form",
		".custom-field-search-value",
		".search-results",
		".search-pagination",
		".saved-view-list",
		".saved-view-metadata",
		".saved-view-pagination",
		".pinned-project-views",
		".pinned-project-view-list",
		".pinned-project-view",
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
		"loadSprintReport",
		"refreshSelectedSprintReport",
		"renderSprints",
		"renderSprintFilter",
		"renderSprintReport",
		"sprintReportAnalyticsNode",
		"normalizeSprint",
		"normalizeSprintReport",
		"normalizeSprintAnalytics",
		"selectedSprintReportID",
		"sprintFilterState",
		"burndown",
		"burnup",
		"velocity",
		"story_points_total",
		"formatStoryPoints",
		"/api/projects/${state.selectedProject.id}/sprints",
		"?state=${encodeURIComponent(state.sprintFilterState)}",
		"/api/sprints/${form.dataset.sprintEditForm}",
		"/api/sprints/${sprintID}/report",
		"/api/sprints/${start.dataset.startSprintId}/start",
		"/api/sprints/${complete.dataset.completeSprintId}/complete",
		"/api/tickets/${assignSprint.dataset.assignSprintId}/sprint",
		"/api/tickets/${updateStoryPoints.dataset.updateStoryPointsId}",
		"await refreshTicketViews(updateStoryPoints.dataset.updateStoryPointsId);",
		"data-ticket-sprint-control",
		"data-ticket-story-points-control",
		"data-sprint-edit-form",
		"data-sprint-report-id",
		"completed_snapshot",
		"Live current assignment",
	} {
		if !strings.Contains(appText, expected) {
			t.Fatalf("expected app.js to contain %q", expected)
		}
	}
	cssText := string(css)
	for _, expected := range []string{
		".sprint-panel",
		".sprint-form",
		".sprint-edit-form",
		".sprint-item",
		".sprint-report",
		".sprint-report-analytics",
		".sprint-report-chart",
		".sprint-report-ticket",
		".ticket-sprint",
		".ticket-story-points",
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
		"loadVersionReport",
		"refreshSelectedVersionReport",
		"await refreshSelectedVersionReport();",
		"renderVersionReport",
		"versionReportHealthNode",
		"versionReleaseHealth",
		"versionReportHealthDates",
		"versionReportSummaryNode",
		"versionReportBreakdownNode",
		"versionReportComponentNode",
		"versionReportComponents",
		"versionReportTicketNode",
		"versionReportScopeText",
		"normalizeComponent",
		"normalizeVersion",
		"normalizeVersionReport",
		"released_snapshot",
		"Due soon",
		"Add a target date to track release timing",
		"Component breakdown",
		"Status breakdown",
		"No component",
		"story_points_total",
		"componentUpdateSpec",
		"versionUpdateSpec",
		"/api/projects/${state.selectedProject.id}/components",
		"/api/projects/${state.selectedProject.id}/versions",
		"/api/versions/${versionID}/report",
		"/api/components/${form.dataset.componentEditForm}",
		"/api/versions/${form.dataset.versionEditForm}",
		"/api/tickets/${assignPlanning.dataset.assignPlanningId}",
		"data-component-edit-form",
		"data-version-edit-form",
		"data-version-report-id",
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
		".version-report",
		".version-report-health",
		".version-report-health-dates",
		".version-report-ticket",
		".version-report-summary",
		".version-report-progress",
		".version-report-breakdown",
		".version-report-component",
		".component-edit-form",
		".version-edit-form",
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
		"scheduleRoadmapItem",
		"scheduleRoadmapItemSpec",
		"renderRoadmap",
		"loadRoadmapCapacityTickets",
		"roadmapCapacityNode",
		"roadmapCapacitySummary",
		"roadmapCapacityItemWork",
		"roadmapCapacityBucketNode",
		"roadmapCapacityChildTickets",
		"roadmapTimelineNode",
		"roadmapScheduleFormNode",
		"roadmapQuickScheduleNode",
		"roadmapDependencyFormNode",
		"refreshRoadmapDependencyViews",
		"renderRoadmapDependencies",
		"normalizeRoadmapDependency",
		"roadmapDependencyNode",
		"normalizeRoadmapItem",
		"roadmapEpics",
		"/api/projects/${state.selectedProject.id}/roadmap",
		"/api/projects/${state.selectedProject.id}/tickets?limit=${limit}&offset=${offset}",
		"offset += limit",
		"/api/projects/${state.selectedProject.id}/roadmap/dependencies",
		"/api/projects/${state.selectedProject.id}/roadmap/schedule",
		"ticket-parent-id",
		"data-roadmap-schedule-form",
		"data-roadmap-drag-id",
		"data-roadmap-timeline-track",
		"data-roadmap-quick-schedule-id",
		"data-roadmap-dependency-form",
		"data-delete-roadmap-dependency-id",
		"els.roadmapDependencies.addEventListener(\"submit\"",
		"els.roadmapDependencies.addEventListener(\"click\"",
		"els.roadmapDependencies.addEventListener(\"change\"",
		"application/rayboard-roadmap-epic",
		"Capacity summary",
		"Remaining pts",
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
		".roadmap-capacity",
		".roadmap-capacity-bucket",
		".roadmap-timeline",
		".roadmap-unscheduled",
		".roadmap-dependencies",
		".roadmap-dependency-form",
		".roadmap-dependency",
		".roadmap-item",
		".roadmap-quick-actions",
		".roadmap-schedule-form",
		".roadmap-progress",
	} {
		if !strings.Contains(cssText, expected) {
			t.Fatalf("expected app.css to contain %q", expected)
		}
	}
}

func TestEmbeddedAppSupportsTicketLinks(t *testing.T) {
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
		"ticketLinks",
		"loadTicketLinks",
		"ticketLinksNode",
		"ticketLinkItemNode",
		"normalizeTicketLink",
		"dataset.ticketLinkForm",
		"deleteTicketLinkId",
		"/api/tickets/${ticketID}/links",
		"/api/tickets/${deleteLink.dataset.ticketId}/links/${deleteLink.dataset.deleteTicketLinkId}",
		"ticket.link_created",
		"ticket.link_deleted",
	} {
		if !strings.Contains(appText, expected) {
			t.Fatalf("expected app.js to contain %q", expected)
		}
	}
	cssText := string(css)
	for _, expected := range []string{
		".ticket-links",
		".ticket-link-list",
		".ticket-link-item",
		".ticket-link-form",
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
		"projectLabelNode",
		"loadProjectLabels",
		"normalizeProjectLabel",
		"renderProjectLabels",
		"renderTicketFilters",
		"ticketFilterParams",
		"ticketFiltersFromForm",
		"new URLSearchParams",
		"component_id",
		"version_id",
		"/api/projects/${state.selectedProject.id}/labels",
		"data-project-label-edit-form",
		"data-delete-project-label",
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
		".project-label-form",
		".project-label-list",
		".project-label-item",
		".label-color-swatch",
		".ticket-label-controls",
		".ticket-filter-form",
		".ticket-filter-summary",
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
		"customFieldMetadataNode",
		"customFieldMetadataItems",
		"customFieldOptionsSummary",
		"`type ${field.field_type}`",
		"no options",
		"normalizeCustomField",
		"parseCustomFields",
		"customFieldsFromControls",
		"renderTicketCreateCustomFields",
		"renderCustomFieldInputs",
		"customFieldInputNode",
		"ticketCustomFields",
		"data-custom-field-key",
		"data-custom-field-input",
		"customFieldUpdateSpec",
		"data-custom-field-edit-form",
		"/api/projects/${state.selectedProject.id}/custom-fields",
		"/api/custom-fields/${form.dataset.customFieldEditForm}",
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
		".field-edit-form",
		".field-list",
		".field-metadata",
		".ticket-custom-fields",
		".custom-field-input",
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
