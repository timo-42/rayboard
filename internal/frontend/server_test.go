package frontend

import (
	"net/http"
	"net/http/httptest"
	"os/exec"
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
		!strings.Contains(body, `id="notification-hook-preview-summary"`) ||
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
		"notificationDeliveryAnalytics",
		"notificationDeliveryAnalyticsLabel",
		"notificationDeliverySummaryTime",
		"deliveryUpdatedAt",
		"Oldest",
		"Newest",
		"Retry pressure",
		"attempts exhausted",
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
		"notificationHookPreviewSummary",
		"notificationHookPreviewSummaryItems",
		"notification-hook-preview-summary",
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
		"backlogEstimateCoverage",
		"backlogEstimateCoverageNode",
		"backlogStatusBreakdown",
		"backlogStatusBreakdownNode",
		"backlogPriorityBreakdown",
		"backlogPriorityBreakdownNode",
		"backlogIssueTypeBreakdown",
		"backlogIssueTypeBreakdownNode",
		"backlogLabelBreakdown",
		"backlogLabelBreakdownNode",
		"backlogLabelBreakdownVisibleItems",
		"backlogComponentBreakdown",
		"backlogComponentBreakdownNode",
		"backlogComponentBreakdownLabel",
		"backlogVersionBreakdown",
		"backlogVersionBreakdownNode",
		"backlogVersionBreakdownLabel",
		"backlogEpicBreakdown",
		"backlogEpicBreakdownNode",
		"backlogEpicBreakdownLabel",
		"backlogAssigneeBreakdown",
		"backlogAssigneeBreakdownNode",
		"backlogAssigneeBreakdownLabel",
		"backlogReporterBreakdown",
		"backlogReporterBreakdownNode",
		"backlogReporterBreakdownLabel",
		"backlogSprintWorkloads",
		"backlogSprintWorkloadsNode",
		"backlogSprintWorkloadLabel",
		"backlogStartDateBreakdown",
		"backlogStartDateBreakdownNode",
		"backlogDueDateBreakdown",
		"backlogDueDateBreakdownNode",
		"backlogReadinessSummary",
		"backlogReadinessSummaryNode",
		"backlogRiskSummary",
		"backlogRiskSummaryNode",
		"backlogAttentionSummary",
		"backlogAttentionSummaryNode",
		"backlogAgeBreakdown",
		"backlogAgeBreakdownNode",
		"backlogUpdateFreshness",
		"backlogUpdateFreshnessNode",
		"Risk signals",
		"Attention",
		"No attention data",
		"Start dates",
		"Due dates",
		"Ticket age",
		"Update freshness",
		"Estimated",
		"No component",
		"No version",
		"No epic",
		"Unestimated",
		"Sprint assigned",
		"Unassigned",
		"No status",
		"No priority",
		"No sprint",
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
		"boardMetadataNode",
		"boardMetadataItems",
		"boardColumnSettingsOverviewNode",
		"boardColumnSettingsOverviewItems",
		"boardStatusCoverageOverviewNode",
		"boardStatusCoverageOverviewItems",
		"Column settings",
		"Status coverage",
		"configured column",
		"covered status",
		"covered workflow status",
		"not on board",
		"no workflow statuses off board",
		"no extra board statuses",
		"WIP-limited column",
		"no WIP limits",
		"selectedBoardSavedViewID",
		"searchAllBoardSavedViewTickets",
		"next_cursor",
		"boardTicketsFromSavedViewMatches",
		"boardSummaryNode",
		"boardSummaryMetrics",
		"boardSummaryMetricNode",
		"boardCapacityOverview",
		"boardCapacityOverviewNode",
		"boardCapacityOverviewLabel",
		"boardRiskOverview",
		"boardRiskOverviewNode",
		"boardRiskOverviewLabel",
		"boardFlowBalance",
		"boardFlowBalanceNode",
		"boardFlowBalanceItems",
		"boardPriorityBreakdown",
		"boardPriorityBreakdownNode",
		"boardDueDateBreakdown",
		"boardDueDateBreakdownNode",
		"boardIssueTypeBreakdown",
		"boardIssueTypeBreakdownNode",
		"boardLabelBreakdown",
		"boardLabelBreakdownNode",
		"boardComponentBreakdown",
		"boardComponentBreakdownNode",
		"boardVersionBreakdown",
		"boardVersionBreakdownNode",
		"boardEpicBreakdown",
		"boardEpicBreakdownNode",
		"boardSprintBreakdown",
		"boardSprintBreakdownNode",
		"boardAssigneeWorkloads",
		"boardAssigneeWorkloadsNode",
		"boardReporterBreakdown",
		"boardReporterBreakdownNode",
		"boardEstimateCoverage",
		"boardEstimateCoverageNode",
		"boardAttentionSummary",
		"boardAttentionSummaryNode",
		"Flow balance",
		"Priorities",
		"Due dates",
		"Issue types",
		"Labels",
		"Components",
		"Versions",
		"Parent epics",
		"Sprints",
		"Assignee workload",
		"Reporters",
		"Estimate coverage",
		"Attention (filtered saved view)",
		"No attention data",
		"boardColumnTicketCount",
		"filtered_by_saved_view",
		"Capacity (filtered saved view)",
		"Risk signals (filtered saved view)",
		"No active board risk signals",
		"over limit",
		"unlimited",
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
		"automationRunFailureBreakdown",
		"automationRunFailureLabel",
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
		"failure ",
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
		"bindCreatePageLayoutBuilder",
		"createPageLayoutBuilderNode",
		"createPageLayoutBuilderFieldOptions",
		"createPageLayoutBuilderFieldItem",
		"handleCreatePageLayoutBuilderClick",
		"handleCreatePageLayoutBuilderChange",
		"data-create-page-layout-builder",
		"data-create-page-layout-add",
		"custom_fields.${field.key}",
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
		".notification-hook-preview-summary",
		".notification-admin-card",
		".notification-delivery-form",
		".notification-delivery-summary",
		".notification-delivery-metrics",
		".notification-delivery-analytics",
		".notification-delivery-analytics-group",
		".notification-delivery-analytics-chips",
		".notification-delivery-list",
		".notification-delivery-item",
		".backlog-panel",
		".backlog-list",
		".backlog-summary",
		".backlog-summary-metric",
		".backlog-estimate-coverage",
		".backlog-status-breakdown",
		".backlog-priority-breakdown",
		".backlog-issue-type-breakdown",
		".backlog-label-breakdown",
		".backlog-component-breakdown",
		".backlog-version-breakdown",
		".backlog-epic-breakdown",
		".backlog-assignee-breakdown",
		".backlog-reporter-breakdown",
		".backlog-sprint-workloads",
		".backlog-readiness-summary",
		".backlog-risk-summary",
		".backlog-attention-summary",
		".backlog-start-date-breakdown",
		".backlog-due-date-breakdown",
		".backlog-planning-chips",
		".backlog-sprint",
		".backlog-sprint-controls",
		".workflow-panel",
		".status-form",
		".board-form",
		".board-saved-view-filter",
		".board-edit-form",
		".board-list",
		".board-metadata",
		".board-column-settings-overview",
		".board-column-settings-chips",
		".board-status-coverage-overview",
		".board-status-coverage-chips",
		".board-summary",
		".board-summary-metric",
		".board-flow-balance",
		".board-flow-chips",
		".board-priority-breakdown",
		".board-priority-chips",
		".board-due-date-breakdown",
		".board-due-date-chips",
		".board-issue-type-breakdown",
		".board-issue-type-chips",
		".board-label-breakdown",
		".board-label-chips",
		".board-component-breakdown",
		".board-component-chips",
		".board-version-breakdown",
		".board-version-chips",
		".board-epic-breakdown",
		".board-epic-chips",
		".board-sprint-breakdown",
		".board-sprint-chips",
		".board-assignee-workloads",
		".board-assignee-chips",
		".board-reporter-breakdown",
		".board-reporter-chips",
		".board-estimate-coverage",
		".board-estimate-chips",
		".board-attention-summary",
		".board-attention-chips",
		".board-capacity-overview",
		".board-capacity-chips",
		".board-risk-overview",
		".board-risk-chips",
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
		".create-page-layout-builder",
		".create-page-layout-builder-item",
		".create-page-layout-builder-actions",
		".create-page-run-list",
		".automation-run-summary",
		".automation-run-filter",
		".automation-run-summary-error",
		".automation-run-summary-failure",
		".engine-result-summary",
		".engine-result-badge",
		".engine-action-preview",
	} {
		if !strings.Contains(cssText, expected) {
			t.Fatalf("expected app.css to contain %q", expected)
		}
	}
}

func TestBacklogReadinessRiskSummaries(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "backlog_readiness_risk_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("backlog readiness risk node test failed: %v\n%s", err, output)
	}
}

func TestBoardCapacityOverview(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "board_capacity_overview_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("board capacity overview node test failed: %v\n%s", err, output)
	}
}

func TestNotificationDeliveryAnalytics(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "notification_delivery_analytics_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("notification delivery analytics node test failed: %v\n%s", err, output)
	}
}

func TestAutomationRunSummary(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "automation_run_summary_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("automation run summary node test failed: %v\n%s", err, output)
	}
}

func TestNotificationHookPreviewSummary(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "notification_hook_preview_summary_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("notification hook preview summary node test failed: %v\n%s", err, output)
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

func TestCreatePageLayoutBuilderHelpers(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "create_page_layout_builder_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("create page layout builder test failed: %v\n%s", err, output)
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
		"savedViewOverviewNode",
		"savedViewOverviewSummary",
		"savedViewScopeBreakdownNode",
		"savedViewDisplayModeSummaryNode",
		"savedViewDisplayModeSummary",
		"savedViewDisplayModeSummaryItems",
		"Saved-view overview",
		"Scopes",
		"Display modes",
		"scope_type",
		"savedViewMetadataNode",
		"savedViewMetadataItems",
		"`filter: ${query.filter}`",
		"`columns ${columnCount}`",
		"editSavedView",
		"loadPinnedProjectSavedViews",
		"renderPinnedProjectSavedViews",
		"pinnedProjectViewNode",
		"pinned-project-view-title",
		"pinned-project-view-metadata",
		"applySavedView",
		"activeSearchPresentation",
		"savedViewSearchPresentation",
		"savedViewOverviewSummary",
		"savedViewConfigurationInsightsNode",
		"savedViewConfigurationInsightItems",
		"savedViewFieldUsageNode",
		"savedViewFieldUsageSummary",
		"savedViewFieldUsageInsightItems",
		"Field usage",
		"clearSearchPresentation",
		"activeSearchResultColumns",
		"searchResultColumns",
		"searchResultColumnLabel",
		"searchResultColumnValue",
		"groupedSearchResults",
		"searchResultGroupNode",
		"search-result-columns",
		"search-result-group",
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
		"Configuration",
		"text queries",
		"CEL filters",
		"board mode",
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
		".saved-view-overview",
		".saved-view-overview-chips",
		".search-result-columns",
		".search-result-column",
		".search-result-group",
		".search-result-group-list",
		".saved-view-scope-breakdown",
		".saved-view-scope-chips",
		".saved-view-display-modes",
		".saved-view-display-mode-chips",
		".saved-view-configuration-insights",
		".saved-view-configuration-chips",
		".saved-view-field-usage",
		".saved-view-field-usage-chips",
		".saved-view-metadata",
		".saved-view-pagination",
		".pinned-project-views",
		".pinned-project-view-list",
		".pinned-project-view",
		".pinned-project-view-title",
		".pinned-project-view-metadata",
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
		"sprintReportHealthNode",
		"sprintReportHealth",
		"sprintReportHealthDates",
		"todayLocalISODate",
		"sprintReportAnalyticsNode",
		"sprintReportChipSectionNode",
		"sprintReportStatusBreakdownNode",
		"sprintReportStatusBreakdown",
		"sprintReportStartDateBreakdownNode",
		"sprintReportStartDateBreakdown",
		"sprintReportDueDateBreakdownNode",
		"sprintReportDueDateBreakdown",
		"sprintReportAgeBreakdownNode",
		"sprintReportAgeBreakdown",
		"Ticket age breakdown",
		"No age data",
		"sprintReportUpdateFreshnessNode",
		"sprintReportUpdateFreshness",
		"Update freshness",
		"No update data",
		"sprintReportReadinessSummaryNode",
		"sprintReportReadinessSummary",
		"Readiness summary",
		"No readiness data",
		"sprintReportRiskSummaryNode",
		"sprintReportRiskSummary",
		"Risk summary",
		"No risk data",
		"sprintReportAttentionSummaryNode",
		"sprintReportAttentionSummary",
		"Attention summary",
		"No attention data",
		"sprintReportScopeChangesNode",
		"sprintReportScopeChangeItems",
		"sprintReportPriorityBreakdownNode",
		"sprintReportPriorityBreakdown",
		"sprintReportTypeBreakdownNode",
		"sprintReportTypeBreakdown",
		"sprintReportLabelBreakdownNode",
		"sprintReportLabelBreakdown",
		"sprintReportEstimateCoverageNode",
		"sprintReportEstimateCoverage",
		"sprintReportComponentNode",
		"sprintReportComponents",
		"sprintReportComponentItemNode",
		"sprintReportVersionNode",
		"sprintReportVersions",
		"sprintReportVersionItemNode",
		"sprintReportEpicBreakdownNode",
		"sprintReportEpics",
		"sprintReportEpicItemNode",
		"sprintReportReporterBreakdown",
		"sprintReportReporterBreakdownNode",
		"sprintReportAssigneeWorkloads",
		"sprintReportAssigneeWorkloadsNode",
		"normalizeSprint",
		"normalizeSprintReport",
		"normalizeSprintReportScopeChanges",
		"normalizeSprintAnalytics",
		"selectedSprintReportID",
		"sprintFilterState",
		"burndown",
		"burnup",
		"velocity",
		"scope_changes",
		"Scope changes",
		"unchanged",
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
		"Sprint ends today",
		"Add sprint dates to track schedule health",
		"Live current assignment",
		"Status breakdown",
		"No status data",
		"Start date breakdown",
		"No start date data",
		"Due date breakdown",
		"No due date data",
		"Priority breakdown",
		"No priority",
		"Issue type breakdown",
		"No issue type",
		"Label breakdown",
		"No label data",
		"Estimate coverage",
		"No sprint tickets",
		"Component breakdown",
		"No component assignments",
		"Version breakdown",
		"No version assignments",
		"Epic breakdown",
		"No epic assignments",
		"Reporter breakdown",
		"No reporter data",
		"Assignee workload",
		"Unassigned",
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
		".sprint-report-health",
		".sprint-report-health-dates",
		".sprint-report-analytics",
		".sprint-report-chart",
		".sprint-report-statuses",
		".sprint-report-status-list",
		".sprint-report-start-dates",
		".sprint-report-start-date-list",
		".sprint-report-due-dates",
		".sprint-report-due-date-list",
		".sprint-report-ages",
		".sprint-report-age-list",
		".sprint-report-updates",
		".sprint-report-update-list",
		".sprint-report-readiness",
		".sprint-report-readiness-list",
		".sprint-report-risks",
		".sprint-report-risk-list",
		".sprint-report-attention",
		".sprint-report-attention-list",
		".sprint-report-scope-changes",
		".sprint-report-scope-change-list",
		".sprint-report-priorities",
		".sprint-report-priority-list",
		".sprint-report-types",
		".sprint-report-type-list",
		".sprint-report-labels",
		".sprint-report-label-list",
		".sprint-report-estimate-coverage",
		".sprint-report-estimate-coverage-list",
		".sprint-report-components",
		".sprint-report-component-list",
		".sprint-report-component",
		".sprint-report-versions",
		".sprint-report-version-list",
		".sprint-report-version",
		".sprint-report-epics",
		".sprint-report-epic-list",
		".sprint-report-epic",
		".sprint-report-reporters",
		".sprint-report-reporter-list",
		".sprint-report-assignees",
		".sprint-report-assignee-list",
		".sprint-report-ticket",
		".ticket-sprint",
		".ticket-story-points",
	} {
		if !strings.Contains(cssText, expected) {
			t.Fatalf("expected app.css to contain %q", expected)
		}
	}
}

func TestSprintReportLabelBreakdown(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "sprint_label_breakdown_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sprint label breakdown node test failed: %v\n%s", err, output)
	}
}

func TestSprintReportStatusBreakdown(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "sprint_status_breakdown_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sprint status breakdown node test failed: %v\n%s", err, output)
	}
}

func TestSprintReportDueDateBreakdown(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "sprint_due_date_breakdown_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sprint due date breakdown node test failed: %v\n%s", err, output)
	}
}

func TestSprintReportStartDateBreakdown(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "sprint_start_date_breakdown_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sprint start date breakdown node test failed: %v\n%s", err, output)
	}
}

func TestSprintReportAgeBreakdown(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "sprint_age_breakdown_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sprint age breakdown node test failed: %v\n%s", err, output)
	}
}

func TestSprintReportUpdateFreshness(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "sprint_update_freshness_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sprint update freshness node test failed: %v\n%s", err, output)
	}
}

func TestSprintReportReadinessSummary(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "sprint_readiness_summary_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sprint readiness summary node test failed: %v\n%s", err, output)
	}
}

func TestSprintReportRiskSummary(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "sprint_risk_summary_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sprint risk summary node test failed: %v\n%s", err, output)
	}
}

func TestSprintReportAttentionSummary(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "sprint_attention_summary_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sprint attention summary node test failed: %v\n%s", err, output)
	}
}

func TestSprintReportVersionBreakdown(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "sprint_version_breakdown_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sprint version breakdown node test failed: %v\n%s", err, output)
	}
}

func TestSprintReportComponentBreakdown(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "sprint_component_breakdown_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sprint component breakdown node test failed: %v\n%s", err, output)
	}
}

func TestSprintReportEpicBreakdown(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "sprint_epic_breakdown_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sprint epic breakdown node test failed: %v\n%s", err, output)
	}
}

func TestSprintReportReporterBreakdown(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "sprint_reporter_breakdown_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sprint reporter breakdown node test failed: %v\n%s", err, output)
	}
}

func TestSprintReportEstimateCoverage(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "sprint_estimate_coverage_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sprint estimate coverage node test failed: %v\n%s", err, output)
	}
}

func TestSprintReportTypeBreakdown(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "sprint_type_breakdown_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sprint type breakdown node test failed: %v\n%s", err, output)
	}
}

func TestSprintReportPriorityBreakdown(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "sprint_priority_breakdown_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sprint priority breakdown node test failed: %v\n%s", err, output)
	}
}

func TestSprintReportHealthDateBoundaries(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "sprint_health_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sprint health node test failed: %v\n%s", err, output)
	}
}

func TestVersionReportTimelineItems(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "release_timeline_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("release timeline node test failed: %v\n%s", err, output)
	}
}

func TestVersionTargetHealthSummary(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_target_health_summary_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version target health summary node test failed: %v\n%s", err, output)
	}
}

func TestVersionLifecycleSummary(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_lifecycle_summary_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version lifecycle summary node test failed: %v\n%s", err, output)
	}
}

func TestVersionDateCoverageSummary(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_date_coverage_summary_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version date coverage summary node test failed: %v\n%s", err, output)
	}
}

func TestVersionTimingVarianceSummary(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_timing_variance_summary_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version timing variance summary node test failed: %v\n%s", err, output)
	}
}

func TestVersionStatusSummary(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_status_summary_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version status summary node test failed: %v\n%s", err, output)
	}
}

func TestVersionPrioritySummary(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_priority_summary_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version priority summary node test failed: %v\n%s", err, output)
	}
}

func TestVersionIssueTypeSummary(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_issue_type_summary_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version issue-type summary node test failed: %v\n%s", err, output)
	}
}

func TestVersionLabelSummary(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_label_summary_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version label summary node test failed: %v\n%s", err, output)
	}
}

func TestVersionAssigneeSummary(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_assignee_summary_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version assignee summary node test failed: %v\n%s", err, output)
	}
}

func TestVersionReporterSummary(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_reporter_summary_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version reporter summary node test failed: %v\n%s", err, output)
	}
}

func TestVersionDueDateSummary(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_due_date_summary_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version due date summary node test failed: %v\n%s", err, output)
	}
}

func TestVersionStartDateSummary(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_start_date_summary_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version start date summary node test failed: %v\n%s", err, output)
	}
}

func TestVersionReportAssigneeWorkloads(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_assignee_workload_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version assignee workload node test failed: %v\n%s", err, output)
	}
}

func TestVersionReportReporterBreakdown(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_reporter_breakdown_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version reporter breakdown node test failed: %v\n%s", err, output)
	}
}

func TestVersionReportDueDateBreakdown(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_due_date_breakdown_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version due-date breakdown node test failed: %v\n%s", err, output)
	}
}

func TestVersionReportStartDateBreakdown(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_start_date_breakdown_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version start-date breakdown node test failed: %v\n%s", err, output)
	}
}

func TestVersionReportAgeBreakdown(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_age_breakdown_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version age breakdown node test failed: %v\n%s", err, output)
	}
}

func TestVersionReportUpdateFreshness(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_update_freshness_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version update freshness node test failed: %v\n%s", err, output)
	}
}

func TestVersionReportReadinessSummary(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_readiness_summary_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version readiness summary node test failed: %v\n%s", err, output)
	}
}

func TestVersionReportRiskSummary(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_risk_summary_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version risk summary node test failed: %v\n%s", err, output)
	}
}

func TestVersionReportAttentionSummary(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_attention_summary_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version attention summary node test failed: %v\n%s", err, output)
	}
}

func TestVersionReportEpicBreakdown(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_epic_breakdown_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version epic breakdown node test failed: %v\n%s", err, output)
	}
}

func TestVersionReportSprintBreakdown(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_sprint_breakdown_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version sprint breakdown node test failed: %v\n%s", err, output)
	}
}

func TestVersionReportAnalyticsSummary(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_analytics_summary_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version analytics summary node test failed: %v\n%s", err, output)
	}
}

func TestVersionReportComponentBreakdown(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_component_breakdown_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version component breakdown node test failed: %v\n%s", err, output)
	}
}

func TestComponentStatusSummary(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "component_status_summary_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("component status summary node test failed: %v\n%s", err, output)
	}
}

func TestComponentOwnershipSummary(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "component_ownership_summary_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("component ownership summary node test failed: %v\n%s", err, output)
	}
}

func TestVersionReportLabelBreakdown(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_label_breakdown_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version label breakdown node test failed: %v\n%s", err, output)
	}
}

func TestVersionReportScopeChanges(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_scope_change_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version scope-change node test failed: %v\n%s", err, output)
	}
}

func TestVersionReportPriorityBreakdown(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_priority_breakdown_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version priority breakdown node test failed: %v\n%s", err, output)
	}
}

func TestVersionReportTypeBreakdown(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_type_breakdown_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version type breakdown node test failed: %v\n%s", err, output)
	}
}

func TestVersionReportEstimateCoverage(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "version_estimate_coverage_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("version estimate coverage node test failed: %v\n%s", err, output)
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
		"componentOwnershipSummary",
		"componentOwnershipSummaryItems",
		"componentOwnershipSummaryNode",
		"componentOwnerMetadataItems",
		"componentOwnerMetadataNode",
		"Ownership coverage",
		"missing default",
		"componentStatusSummary",
		"componentStatusSummaryItems",
		"componentStatusSummaryNode",
		"Visible ticket status",
		"No visible tickets",
		"versionReportHealthNode",
		"versionReleaseHealth",
		"versionLifecycleSummaryNode",
		"versionLifecycleSummary",
		"versionLifecycleSummaryItems",
		"Lifecycle states",
		"unmodeled",
		"versionDateCoverageSummaryNode",
		"versionDateCoverageSummary",
		"versionDateCoverageSummaryItems",
		"Date coverage",
		"missing dates",
		"versionTargetHealthSummaryNode",
		"versionTargetHealthSummary",
		"versionTimingVarianceSummaryNode",
		"versionTimingVarianceSummary",
		"versionStatusSummary",
		"versionStatusSummaryItems",
		"versionStatusSummaryNode",
		"versionPrioritySummary",
		"versionPrioritySummaryItems",
		"versionPrioritySummaryNode",
		"Visible ticket priorities",
		"versionIssueTypeSummary",
		"versionIssueTypeSummaryItems",
		"versionIssueTypeSummaryNode",
		"Visible ticket issue types",
		"versionLabelSummary",
		"versionLabelSummaryItems",
		"versionLabelSummaryNode",
		"Visible ticket labels",
		"versionAssigneeSummary",
		"versionAssigneeSummaryItems",
		"versionAssigneeSummaryNode",
		"Visible ticket assignees",
		"versionReporterSummary",
		"versionReporterSummaryItems",
		"versionReporterSummaryNode",
		"Visible ticket reporters",
		"versionDueDateSummary",
		"versionDueDateSummaryItems",
		"versionDueDateSummaryNode",
		"Visible ticket due dates",
		"versionStartDateSummary",
		"versionStartDateSummaryItems",
		"versionStartDateSummaryNode",
		"Visible ticket start dates",
		"versionReportHealthDates",
		"versionReportTimelineNode",
		"versionReportTimelineItems",
		"versionReportSummaryNode",
		"versionReportAnalyticsNode",
		"versionReportAnalyticsSummary",
		"versionReportEstimateCoverageNode",
		"versionReportEstimateCoverage",
		"versionReportScopeChangesNode",
		"versionReportScopeChangeItems",
		"versionReportBreakdownNode",
		"versionReportComponentNode",
		"versionReportComponents",
		"versionReportEpicBreakdownNode",
		"versionReportEpics",
		"versionReportEpicItemNode",
		"versionReportSprintBreakdownNode",
		"versionReportSprints",
		"versionReportSprintItemNode",
		"versionReportLabelBreakdownNode",
		"versionReportLabelBreakdown",
		"versionReportAssigneeWorkloadsNode",
		"versionReportAssigneeWorkloads",
		"versionReportAssigneeItemNode",
		"versionReportReporterBreakdownNode",
		"versionReportReporterBreakdown",
		"versionReportDueDateBreakdownNode",
		"versionReportDueDateBreakdown",
		"versionReportStartDateBreakdownNode",
		"versionReportStartDateBreakdown",
		"versionReportAgeBreakdownNode",
		"versionReportAgeBreakdown",
		"versionReportUpdateFreshnessNode",
		"versionReportUpdateFreshness",
		"versionReportReadinessSummaryNode",
		"versionReportReadinessSummary",
		"versionReportRiskSummaryNode",
		"versionReportRiskSummary",
		"versionReportAttentionSummaryNode",
		"versionReportAttentionSummary",
		"versionReportPriorityBreakdownNode",
		"versionReportPriorityBreakdown",
		"versionReportTypeBreakdownNode",
		"versionReportTypeBreakdown",
		"Assignee workload",
		"Reporter breakdown",
		"No reporter data",
		"No reporter",
		"Due date breakdown",
		"No due date data",
		"No due date",
		"Start date breakdown",
		"No start date data",
		"No start date",
		"Ticket age breakdown",
		"No age data",
		"No created date",
		"Update freshness",
		"No update data",
		"No update date",
		"Readiness summary",
		"No readiness data",
		"Missing estimate",
		"Risk summary",
		"No risk data",
		"Open overdue",
		"Attention summary",
		"No attention data",
		"Blocked open",
		"Estimate coverage",
		"No release tickets",
		"Priority breakdown",
		"Issue type breakdown",
		"Label breakdown",
		"No label data",
		"No issue type",
		"No priority",
		"Unassigned",
		"versionReportTicketNode",
		"versionReportScopeText",
		"normalizeComponent",
		"normalizeVersion",
		"normalizeVersionReport",
		"normalizeVersionReportScopeChanges",
		"released_snapshot",
		"scope_changes",
		"Scope changes",
		"unchanged",
		"Due soon",
		"Target-date health",
		"Scheduled later",
		"Unscheduled",
		"Release timing",
		"Released early",
		"Released late",
		"Released without target date",
		"Add a target date to track release timing",
		"Release timeline",
		"released on target",
		"release date missing",
		"Release analytics will appear when the report is loaded",
		"Component breakdown",
		"Epic breakdown",
		"No epic assignments",
		"No epic",
		"Sprint breakdown",
		"No sprint assignments",
		"No sprint",
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
		".component-ownership-summary",
		".component-ownership-chips",
		".component-owner-metadata",
		".component-status-summary",
		".component-status-chips",
		".version-form",
		".version-report",
		".version-status-summary",
		".version-status-chips",
		".version-priority-summary",
		".version-priority-chips",
		".version-issue-type-summary",
		".version-issue-type-chips",
		".version-label-summary",
		".version-label-chips",
		".version-assignee-summary",
		".version-assignee-chips",
		".version-reporter-summary",
		".version-reporter-chips",
		".version-due-date-summary",
		".version-due-date-chips",
		".version-start-date-summary",
		".version-start-date-chips",
		".version-lifecycle-summary",
		".version-lifecycle-list",
		".version-date-coverage",
		".version-date-coverage-list",
		".version-target-health",
		".version-target-health-list",
		".version-timing-variance",
		".version-timing-variance-list",
		".version-report-health",
		".version-report-health-dates",
		".version-report-timeline",
		".version-report-timeline-chips",
		".version-report-ticket",
		".version-report-summary",
		".version-report-progress",
		".version-report-analytics",
		".version-report-chart",
		".version-report-estimate-coverage",
		".version-report-estimate-coverage-list",
		".version-report-scope-changes",
		".version-report-scope-change-list",
		".version-report-assignees",
		".version-report-assignee-list",
		".version-report-reporters",
		".version-report-reporter-list",
		".version-report-due-dates",
		".version-report-due-date-list",
		".version-report-start-dates",
		".version-report-start-date-list",
		".version-report-ages",
		".version-report-age-list",
		".version-report-updates",
		".version-report-update-list",
		".version-report-readiness",
		".version-report-readiness-list",
		".version-report-risks",
		".version-report-risk-list",
		".version-report-attention",
		".version-report-attention-list",
		".version-report-types",
		".version-report-priorities",
		".version-report-labels",
		".version-report-breakdown",
		".version-report-component",
		".version-report-epic-list",
		".version-report-epic",
		".version-report-sprint-list",
		".version-report-sprint",
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
		"loadRoadmapCapacityTargets",
		"loadRoadmapCapacityTickets",
		"saveRoadmapCapacityTarget",
		"roadmapCapacityNode",
		"roadmapCapacitySummary",
		"roadmapCapacityInsightItems",
		"roadmapCapacityItemWork",
		"roadmapCapacityBucketNode",
		"roadmapCapacityDrilldown",
		"roadmapCapacityDrilldownNode",
		"roadmapCapacityDrilldownRowNode",
		"roadmapCapacityTargetNode",
		"roadmapCapacityTargetForBucket",
		"roadmapCapacityTargetMap",
		"roadmapCapacityTargetValue",
		"roadmapCapacityBucketTargetStatus",
		"roadmapCapacityBucketAtRisk",
		"roadmapCapacityChildTickets",
		"roadmapTimelineNode",
		"roadmapScheduleFormNode",
		"roadmapQuickScheduleNode",
		"roadmapDependencyFormNode",
		"roadmapDependencyOverviewNode",
		"roadmapDependencyOverviewSummary",
		"roadmapDependencyGraph",
		"roadmapDependencyGraphNode",
		"roadmapDependencyGraphEdgeNode",
		"Dependency graph",
		"refreshRoadmapDependencyViews",
		"renderRoadmapDependencies",
		"normalizeRoadmapDependency",
		"roadmapDependencyNode",
		"Dependency overview",
		"cross-epic",
		"incomplete",
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
		"Capacity drilldown",
		"Monthly point target",
		"over target by",
		"Remaining pts",
		"at-risk months",
		"data-roadmap-capacity-bucket",
		"data-roadmap-capacity-target",
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
		".roadmap-capacity-target",
		".roadmap-capacity-insights",
		".roadmap-capacity-bucket",
		".roadmap-capacity-bucket.is-over-target",
		".roadmap-capacity-drilldown",
		".roadmap-capacity-drilldown-item",
		".roadmap-timeline",
		".roadmap-unscheduled",
		".roadmap-dependencies",
		".roadmap-dependency-form",
		".roadmap-dependency-overview",
		".roadmap-dependency-overview-chips",
		".roadmap-dependency-graph",
		".roadmap-dependency-graph-node",
		".roadmap-dependency-graph-edge",
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

func TestRoadmapDependencyGraph(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "roadmap_dependency_graph_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("roadmap dependency graph node test failed: %v\n%s", err, output)
	}
}

func TestRoadmapCapacityDrilldown(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "roadmap_capacity_drilldown_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("roadmap capacity drilldown node test failed: %v\n%s", err, output)
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
		"ticketLinkDependencyOverviewNode",
		"ticketLinkDependencySummary",
		"ticketLinkItemNode",
		"normalizeTicketLink",
		"No dependencies",
		"Blocked by",
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
		".ticket-link-dependency-overview",
		".ticket-link-list",
		".ticket-link-item",
		".ticket-link-form",
	} {
		if !strings.Contains(cssText, expected) {
			t.Fatalf("expected app.css to contain %q", expected)
		}
	}
}

func TestTicketDependencySummary(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "ticket_dependency_summary_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ticket dependency summary node test failed: %v\n%s", err, output)
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
		"loadCustomFieldUsageTickets",
		"renderCustomFields",
		"customFieldLayoutOverviewNode",
		"customFieldLayoutSummary",
		"customFieldTypeBreakdownNode",
		"customFieldRequirementInsightsNode",
		"customFieldRequirementInsights",
		"customFieldRequirementInsightItems",
		"customFieldUsageSummaryNode",
		"customFieldUsageSummary",
		"customFieldOptionUsageSummaryNode",
		"customFieldOptionUsageSummary",
		"customFieldUnmodeledValueSummaryNode",
		"customFieldUnmodeledValueSummary",
		"customFieldUsageTickets",
		"customFieldValuePresent",
		"Field layout",
		"Field types",
		"Requirements",
		"Usage coverage",
		"Option usage",
		"Unmodeled values",
		"no unmodeled values",
		"no select fields",
		"unconfigured",
		"required missing",
		"selects missing options",
		"search-ready",
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
		".field-layout-overview",
		".field-layout-chips",
		".field-type-breakdown",
		".field-type-chips",
		".field-requirement-insights",
		".field-requirement-chips",
		".field-usage-summary",
		".field-usage-chips",
		".field-option-usage",
		".field-option-usage-chips",
		".field-unmodeled-values",
		".field-unmodeled-value-chips",
		".field-metadata",
		".ticket-custom-fields",
		".custom-field-input",
	} {
		if !strings.Contains(cssText, expected) {
			t.Fatalf("expected app.css to contain %q", expected)
		}
	}
}

func TestCustomFieldRequirementInsights(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "custom_field_requirement_insights_node_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("custom-field requirement insights node test failed: %v\n%s", err, output)
	}
}

func TestSearchSavedViewPresentationHelpers(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	cmd := exec.Command("node", "search_saved_view_presentation_test.js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("search saved-view presentation helper test failed: %v\n%s", err, output)
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
