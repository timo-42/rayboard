const searchPageSize = 20;
const savedViewPageSize = 20;

const state = {
  user: null,
  projects: [],
  selectedProject: null,
  selectedIssue: null,
  selectedCreatePageSchema: null,
  createPageSubmission: null,
  tickets: [],
  ticketFilters: { label: "", component_id: "", version_id: "" },
  projectLabels: [],
  projectSummaries: [],
  recentTickets: [],
  activeSprints: [],
  backlog: [],
  workflowStatuses: [],
  boards: [],
  selectedBoardID: "",
  boardTickets: null,
  boardSavedViews: [],
  selectedBoardSavedViewID: "",
  boardSavedViewsError: "",
  sprints: [],
  sprintFilterState: "",
  selectedSprintReportID: "",
  sprintReport: null,
  components: [],
  versions: [],
  selectedVersionReportID: "",
  versionReport: null,
  customFields: [],
  roadmap: [],
  roadmapCapacityTickets: [],
  roadmapDependencies: [],
  attachments: {},
  comments: {},
  ticketLinks: {},
  ticketWatchers: {},
  activities: {},
  notifications: [],
  unreadNotificationsOnly: false,
  searchResults: [],
  searchNextCursor: "",
  searchCursorStack: [""],
  searchCursorIndex: 0,
  savedViews: [],
  savedViewOffset: 0,
  savedViewHasMore: false,
  pinnedProjectSavedViews: [],
  pinnedProjectSavedViewsLoading: false,
  pinnedProjectSavedViewsError: "",
  lastSearchSpec: { text: "", filter: "" },
  tokens: [],
  createdToken: null,
  rbac: { users: [], groups: [], roles: [], bindings: [], members: {}, effectivePermissions: null, effectivePermissionsError: "" },
  settings: null,
  notificationPreferences: null,
  projectNotificationPreferences: null,
  notificationDeliveries: [],
  auditLog: [],
  openRouterProviders: [],
  notificationDestinations: [],
  notificationPolicies: [],
  notificationHooks: [],
  notificationHookRuns: {},
  notificationHookPreview: null,
  settingsError: "",
  auditLogError: "",
  openRouterProvidersError: "",
  notificationDestinationsError: "",
  notificationPoliciesError: "",
  notificationHooksError: "",
  projectNotificationPreferencesError: "",
  notificationDeliveriesError: "",
  cronJobs: [],
  cronRuns: {},
  cronJobsError: "",
  webhooks: [],
  webhookRuns: {},
  webhookDeliveries: {},
  webhookTokens: {},
  webhooksError: "",
  ticketHooks: [],
  ticketHookRuns: {},
  ticketHooksError: "",
  ticketHookPreview: null,
  createPages: [],
  createPageRuns: {},
  createPagesError: "",
  automationRunFilters: {},
  engineResult: null
};

const els = {
  appNav: document.querySelector("#app-nav"),
  loginForm: document.querySelector("#login-form"),
  projectForm: document.querySelector("#project-form"),
  ticketForm: document.querySelector("#ticket-form"),
  engineForm: document.querySelector("#engine-form"),
  engineType: document.querySelector("#engine-type"),
  engineProjectID: document.querySelector("#engine-project-id"),
  engineWorkbench: document.querySelector("#engine-workbench"),
  engineResultSummary: document.querySelector("#engine-result-summary"),
  engineOutput: document.querySelector("#engine-output"),
  cronJobProject: document.querySelector("#cron-job-project"),
  cronJobForm: document.querySelector("#cron-job-form"),
  cronJobEngineType: document.querySelector("#cron-job-engine-type"),
  cronJobStatus: document.querySelector("#cron-job-status"),
  cronJobs: document.querySelector("#cron-jobs"),
  webhookProject: document.querySelector("#webhook-project"),
  webhookForm: document.querySelector("#webhook-form"),
  webhookEngineType: document.querySelector("#webhook-engine-type"),
  webhookStatus: document.querySelector("#webhook-status"),
  webhooks: document.querySelector("#webhooks"),
  ticketHookProject: document.querySelector("#ticket-hook-project"),
  ticketHookForm: document.querySelector("#ticket-hook-form"),
  ticketHookEngineType: document.querySelector("#ticket-hook-engine-type"),
  ticketHookPreviewForm: document.querySelector("#ticket-hook-preview-form"),
  ticketHookStatus: document.querySelector("#ticket-hook-status"),
  ticketHooks: document.querySelector("#ticket-hooks"),
  ticketHookPreviewOutput: document.querySelector("#ticket-hook-preview-output"),
  createPageProject: document.querySelector("#create-page-project"),
  createPageForm: document.querySelector("#create-page-form"),
  createPageLogicType: document.querySelector("#create-page-logic-type"),
  createPageStatus: document.querySelector("#create-page-status"),
  createPages: document.querySelector("#create-pages"),
  dashboardView: document.querySelector("#dashboard-view"),
  metricProjects: document.querySelector("#metric-projects"),
  metricOpenTickets: document.querySelector("#metric-open-tickets"),
  metricDoneTickets: document.querySelector("#metric-done-tickets"),
  metricUnread: document.querySelector("#metric-unread"),
  recentTickets: document.querySelector("#recent-tickets"),
  bigProjects: document.querySelector("#big-projects"),
  dashboardSprints: document.querySelector("#dashboard-sprints"),
  issueView: document.querySelector("#issue-view"),
  issueTitle: document.querySelector("#issue-title"),
  issueProjectLink: document.querySelector("#issue-project-link"),
  issueDetail: document.querySelector("#issue-detail"),
  createPageView: document.querySelector("#create-page-view"),
  createPageTitle: document.querySelector("#create-page-title"),
  createPageProjectLink: document.querySelector("#create-page-project-link"),
  createPageSubmitForm: document.querySelector("#create-page-submit-form"),
  rbacPanel: document.querySelector("#rbac-panel"),
  rbacRefresh: document.querySelector("#rbac-refresh"),
  rbacUserForm: document.querySelector("#rbac-user-form"),
  rbacGroupForm: document.querySelector("#rbac-group-form"),
  rbacMemberForm: document.querySelector("#rbac-member-form"),
  rbacBindingForm: document.querySelector("#rbac-binding-form"),
  rbacPermissionForm: document.querySelector("#rbac-permission-form"),
  rbacUsers: document.querySelector("#rbac-users"),
  rbacGroups: document.querySelector("#rbac-groups"),
  rbacRoles: document.querySelector("#rbac-roles"),
  rbacBindings: document.querySelector("#rbac-bindings"),
  rbacPermissions: document.querySelector("#rbac-permissions"),
  settingsPanel: document.querySelector("#settings-panel"),
  settingsRefresh: document.querySelector("#settings-refresh"),
  settingsForm: document.querySelector("#settings-form"),
  settingsStatus: document.querySelector("#settings-status"),
  auditForm: document.querySelector("#audit-form"),
  auditStatus: document.querySelector("#audit-status"),
  auditLog: document.querySelector("#audit-log"),
  openRouterProviderForm: document.querySelector("#openrouter-provider-form"),
  openRouterProviderStatus: document.querySelector("#openrouter-provider-status"),
  openRouterProviders: document.querySelector("#openrouter-providers"),
  notificationDestinationForm: document.querySelector("#notification-destination-form"),
  notificationDestinationScope: document.querySelector("#notification-destination-scope"),
  notificationDestinationProject: document.querySelector("#notification-destination-project"),
  notificationDestinationStatus: document.querySelector("#notification-destination-status"),
  notificationDestinations: document.querySelector("#notification-destinations"),
  notificationPolicyForm: document.querySelector("#notification-policy-form"),
  notificationPolicyScope: document.querySelector("#notification-policy-scope"),
  notificationPolicyProject: document.querySelector("#notification-policy-project"),
  notificationPolicyDestinations: document.querySelector("#notification-policy-destinations"),
  notificationPolicyStatus: document.querySelector("#notification-policy-status"),
  notificationPolicies: document.querySelector("#notification-policies"),
  notificationHookForm: document.querySelector("#notification-hook-form"),
  notificationHookScope: document.querySelector("#notification-hook-scope"),
  notificationHookProject: document.querySelector("#notification-hook-project"),
  notificationHookEngineType: document.querySelector("#notification-hook-engine-type"),
  notificationHookPreviewForm: document.querySelector("#notification-hook-preview-form"),
  notificationHookPreviewPolicy: document.querySelector("#notification-hook-preview-policy"),
  notificationHookPreviewDestinations: document.querySelector("#notification-hook-preview-destinations"),
  notificationHookStatus: document.querySelector("#notification-hook-status"),
  notificationHooks: document.querySelector("#notification-hooks"),
  notificationHookPreviewOutput: document.querySelector("#notification-hook-preview-output"),
  preferenceForm: document.querySelector("#preference-form"),
  preferenceStatus: document.querySelector("#preference-status"),
  projectPreferenceForm: document.querySelector("#project-preference-form"),
  projectPreferenceProject: document.querySelector("#project-preference-project"),
  projectPreferenceStatus: document.querySelector("#project-preference-status"),
  notificationDeliveryForm: document.querySelector("#notification-delivery-form"),
  notificationDeliveryProject: document.querySelector("#notification-delivery-project"),
  notificationDeliverySummary: document.querySelector("#notification-delivery-summary"),
  notificationDeliveries: document.querySelector("#notification-deliveries"),
  notificationInbox: document.querySelector("#notification-inbox"),
  notificationCount: document.querySelector("#notification-count"),
  navUnreadCount: document.querySelector("#nav-unread-count"),
  notificationUnreadOnly: document.querySelector("#notifications-unread-only"),
  notificationRefresh: document.querySelector("#notifications-refresh"),
  notificationReadAll: document.querySelector("#notifications-read-all"),
  notifications: document.querySelector("#notifications"),
  sprintPanel: document.querySelector("#sprint-panel"),
  sprintForm: document.querySelector("#sprint-form"),
  sprints: document.querySelector("#sprints"),
  sprintReport: document.querySelector("#sprint-report"),
  labelPanel: document.querySelector("#label-panel"),
  projectLabelForm: document.querySelector("#project-label-form"),
  projectLabels: document.querySelector("#project-labels"),
  backlogPanel: document.querySelector("#backlog-panel"),
  backlog: document.querySelector("#backlog"),
  workflowPanel: document.querySelector("#workflow-panel"),
  statusForm: document.querySelector("#status-form"),
  workflowStatuses: document.querySelector("#workflow-statuses"),
  boardForm: document.querySelector("#board-form"),
  boardSavedViewFilter: document.querySelector("#board-saved-view-filter"),
  boardSavedViewStatus: document.querySelector("#board-saved-view-status"),
  boards: document.querySelector("#boards"),
  releasePanel: document.querySelector("#release-panel"),
  componentForm: document.querySelector("#component-form"),
  versionForm: document.querySelector("#version-form"),
  components: document.querySelector("#components"),
  versions: document.querySelector("#versions"),
  versionReport: document.querySelector("#version-report"),
  fieldPanel: document.querySelector("#field-panel"),
  fieldForm: document.querySelector("#field-form"),
  customFields: document.querySelector("#custom-fields"),
  roadmapPanel: document.querySelector("#roadmap-panel"),
  roadmap: document.querySelector("#roadmap"),
  roadmapDependencies: document.querySelector("#roadmap-dependencies"),
  ticketParentID: document.querySelector("#ticket-parent-id"),
  ticketCustomFields: document.querySelector("#ticket-custom-fields"),
  ticketComponentID: document.querySelector("#ticket-component-id"),
  ticketVersionID: document.querySelector("#ticket-version-id"),
  searchPanel: document.querySelector("#search-panel"),
  searchForm: document.querySelector("#search-form"),
  customFieldSearchControls: document.querySelector("#custom-field-search-controls"),
  savedViewForm: document.querySelector("#saved-view-form"),
  savedViewCancelEdit: document.querySelector("#saved-view-cancel-edit"),
  searchResults: document.querySelector("#search-results"),
  searchResultCount: document.querySelector("#search-result-count"),
  searchPagination: document.querySelector("#search-pagination"),
  savedViews: document.querySelector("#saved-views"),
  savedViewPagination: document.querySelector("#saved-view-pagination"),
  accountPanel: document.querySelector("#account-panel"),
  accountUser: document.querySelector("#account-user"),
  tokenForm: document.querySelector("#token-form"),
  createdToken: document.querySelector("#created-token"),
  apiTokens: document.querySelector("#api-tokens"),
  projectCreate: document.querySelector("#project-create"),
  logoutButton: document.querySelector("#logout-button"),
  signedOut: document.querySelector("#signed-out"),
  boardView: document.querySelector("#board-view"),
  sessionState: document.querySelector("#session-state"),
  projects: document.querySelector("#projects"),
  pinnedProjectViews: document.querySelector("#pinned-project-views"),
  selectedProject: document.querySelector("#selected-project"),
  ticketFilterForm: document.querySelector("#ticket-filter-form"),
  ticketFilterLabel: document.querySelector("#ticket-filter-label"),
  ticketFilterComponent: document.querySelector("#ticket-filter-component"),
  ticketFilterVersion: document.querySelector("#ticket-filter-version"),
  ticketFilterClear: document.querySelector("#ticket-filter-clear"),
  ticketFilterSummary: document.querySelector("#ticket-filter-summary"),
  ticketColumns: document.querySelector("#ticket-columns"),
  notice: document.querySelector("#notice")
};

document.addEventListener("DOMContentLoaded", () => {
  bindEvents();
  refreshSession();
});

function bindEvents() {
  window.addEventListener("popstate", () => {
    handleRouteChange();
  });

  els.appNav.addEventListener("click", (event) => {
    const link = event.target.closest("a[href]");
    if (!link || link.origin !== window.location.origin) {
      return;
    }
    if (isDocumentLink(link.pathname)) {
      return;
    }
    event.preventDefault();
    navigate(link.pathname);
  });

  els.loginForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const data = formData(form);
    await runAction(async () => {
      await api("/api/login", { method: "POST", body: { spec: data } });
      form.reset();
      await refreshSession();
    }, "Signed in");
  });

  els.logoutButton.addEventListener("click", async () => {
    await runAction(async () => {
      await api("/api/logout", { method: "POST" });
      state.user = null;
      state.projects = [];
      state.selectedProject = null;
      state.selectedIssue = null;
      state.selectedCreatePageSchema = null;
      state.createPageSubmission = null;
      state.tickets = [];
      state.ticketFilters = emptyTicketFilters();
      state.projectLabels = [];
      state.projectSummaries = [];
      state.recentTickets = [];
      state.activeSprints = [];
      state.backlog = [];
      state.workflowStatuses = [];
      state.boards = [];
      state.selectedBoardID = "";
      state.boardTickets = null;
      state.sprints = [];
      state.sprintFilterState = "";
      state.selectedSprintReportID = "";
      state.sprintReport = null;
      state.components = [];
      state.versions = [];
      state.customFields = [];
      state.roadmap = [];
      state.roadmapCapacityTickets = [];
      state.attachments = {};
      state.comments = {};
      state.ticketLinks = {};
      state.ticketWatchers = {};
      state.activities = {};
      state.notifications = [];
      state.unreadNotificationsOnly = false;
      state.searchResults = [];
      resetSearchPagination();
      state.savedViews = [];
      resetSavedViewPagination();
      state.pinnedProjectSavedViews = [];
      state.pinnedProjectSavedViewsLoading = false;
      state.pinnedProjectSavedViewsError = "";
      state.lastSearchSpec = { text: "", filter: "" };
      state.tokens = [];
      state.createdToken = null;
      state.rbac = { users: [], groups: [], roles: [], bindings: [], members: {}, effectivePermissions: null, effectivePermissionsError: "" };
      state.settings = null;
      state.notificationPreferences = null;
      state.projectNotificationPreferences = null;
      state.notificationDeliveries = [];
      state.auditLog = [];
      state.openRouterProviders = [];
      state.notificationDestinations = [];
      state.notificationPolicies = [];
      state.notificationHooks = [];
      state.notificationHookRuns = {};
      state.notificationHookPreview = null;
      state.settingsError = "";
      state.auditLogError = "";
      state.openRouterProvidersError = "";
      state.notificationDestinationsError = "";
      state.notificationPoliciesError = "";
      state.notificationHooksError = "";
      state.projectNotificationPreferencesError = "";
      state.notificationDeliveriesError = "";
      state.cronJobs = [];
      state.cronRuns = {};
      state.cronJobsError = "";
      state.webhooks = [];
      state.webhookRuns = {};
      state.webhookDeliveries = {};
      state.webhookTokens = {};
      state.webhooksError = "";
      state.ticketHooks = [];
      state.ticketHookRuns = {};
      state.ticketHooksError = "";
      state.ticketHookPreview = null;
      state.createPages = [];
      state.createPageRuns = {};
      state.createPagesError = "";
      render();
    }, "Signed out");
  });

  els.projectForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const data = formData(form);
    await runAction(async () => {
      const project = normalizeProject(await api("/api/projects", { method: "POST", body: { spec: data } }));
      form.reset();
      await loadProjects(project.id);
      await navigate(`/projects/${project.id}`);
    }, "Project created");
  });

  els.ticketForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    if (!state.selectedProject) {
      return;
    }
    const form = event.currentTarget;
    const data = formData(form);
    data.labels = parseLabels(data.labels);
    data.custom_fields = customFieldsFromControls(els.ticketCustomFields);
    applyStoryPointsSpec(data, data.story_points);
    await runAction(async () => {
      await api(`/api/projects/${state.selectedProject.id}/tickets`, { method: "POST", body: { spec: data } });
      form.reset();
      renderTicketCreateCustomFields();
      await loadRoadmap({ renderTickets: false });
      await loadProjectLabels({ renderTickets: false });
      await loadTickets();
    }, "Ticket created");
  });

  els.ticketFilterForm.addEventListener("change", () => {
    state.ticketFilters = ticketFiltersFromForm();
    loadTickets();
  });

  els.ticketFilterClear.addEventListener("click", () => {
    state.ticketFilters = emptyTicketFilters();
    renderTicketFilters();
    loadTickets();
  });

  els.engineType.addEventListener("change", () => {
    renderEngineFields();
  });

  els.ticketHookEngineType.addEventListener("change", () => {
    renderTicketHookEngineFields();
  });

  els.cronJobEngineType.addEventListener("change", () => {
    renderCronJobEngineFields();
  });

  els.cronJobProject.addEventListener("change", async () => {
    const projectID = els.cronJobProject.value;
    const project = state.projects.find((item) => item.id === projectID);
    if (project) {
      state.selectedProject = project;
    }
    await Promise.all([loadCronJobs(projectID), loadWebhooks(projectID), loadTicketHooks(projectID), loadCreatePages(projectID)]);
    renderEngineFields();
  });

  els.cronJobForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    await runAction(async () => {
      await api("/api/cron-jobs", {
        method: "POST",
        body: { spec: cronJobSpec(form) }
      });
      form.reset();
      setFormValue(form, "timezone", "UTC");
      renderCronJobEngineFields();
      await loadCronJobs();
    }, "Cron job created");
  });

  els.cronJobs.addEventListener("click", async (event) => {
    const showRuns = event.target.closest("[data-load-cron-runs-id]");
    if (showRuns) {
      await runAction(async () => {
        await loadCronRuns(showRuns.dataset.loadCronRunsId);
      }, "Cron runs loaded");
      return;
    }

    const run = event.target.closest("[data-run-cron-job-id]");
    if (run) {
      await runAction(async () => {
        const result = normalizeCronRun(await api(`/api/cron-jobs/${run.dataset.runCronJobId}/run`, { method: "POST" }));
        state.cronRuns[run.dataset.runCronJobId] = [result];
        await loadCronJobs();
      }, "Cron job started");
      return;
    }

    const toggle = event.target.closest("[data-toggle-cron-job-id]");
    if (toggle) {
      await runAction(async () => {
        await api(`/api/cron-jobs/${toggle.dataset.toggleCronJobId}`, {
          method: "PATCH",
          body: { spec: { enabled: toggle.dataset.cronJobEnabled === "true" } }
        });
        await loadCronJobs();
      }, "Cron job updated");
      return;
    }

    const remove = event.target.closest("[data-delete-cron-job-id]");
    if (remove) {
      await runAction(async () => {
        await api(`/api/cron-jobs/${remove.dataset.deleteCronJobId}`, { method: "DELETE" });
        delete state.cronRuns[remove.dataset.deleteCronJobId];
        delete state.automationRunFilters[automationRunFilterKey("cron", remove.dataset.deleteCronJobId)];
        await loadCronJobs();
      }, "Cron job deleted");
    }
  });

  els.cronJobs.addEventListener("change", (event) => {
    if (handleAutomationRunFilterChange(event, renderCronJobs)) {
      return;
    }
    if (!event.target.matches("select[name='engine_type']")) {
      return;
    }
    const form = event.target.closest("[data-cron-job-form]");
    if (form) {
      renderCronJobEditEngineFields(form);
    }
  });

  els.cronJobs.addEventListener("submit", async (event) => {
    const form = event.target.closest("[data-cron-job-form]");
    if (!form) {
      return;
    }
    event.preventDefault();
    await runAction(async () => {
      await api(`/api/cron-jobs/${form.dataset.cronJobForm}`, {
        method: "PATCH",
        body: { spec: cronJobSpec(form) }
      });
      await loadCronJobs();
    }, "Cron job saved");
  });

  els.webhookEngineType.addEventListener("change", () => {
    renderWebhookEngineFields();
  });

  els.webhookProject.addEventListener("change", async () => {
    const projectID = els.webhookProject.value;
    const project = state.projects.find((item) => item.id === projectID);
    if (project) {
      state.selectedProject = project;
    }
    await Promise.all([loadCronJobs(projectID), loadWebhooks(projectID), loadTicketHooks(projectID), loadCreatePages(projectID)]);
    renderEngineFields();
  });

  els.webhookForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const projectID = selectedWebhookProjectID();
    if (!projectID) {
      setNotice("Choose a project for webhooks");
      return;
    }
    const form = event.currentTarget;
    await runAction(async () => {
      const created = normalizeWebhook(await api(`/api/projects/${projectID}/webhooks`, {
        method: "POST",
        body: { spec: webhookSpec(form) }
      }));
      if (created && created.token) {
        state.webhookTokens[created.id] = created.token;
      }
      form.reset();
      renderWebhookEngineFields();
      await loadWebhooks(projectID);
    }, "Webhook created");
  });

  els.webhooks.addEventListener("click", async (event) => {
    const runs = event.target.closest("[data-load-webhook-runs-id]");
    if (runs) {
      await runAction(async () => {
        await loadWebhookRuns(runs.dataset.loadWebhookRunsId);
      }, "Webhook runs loaded");
      return;
    }

    const deliveries = event.target.closest("[data-load-webhook-deliveries-id]");
    if (deliveries) {
      await runAction(async () => {
        await loadWebhookDeliveries(deliveries.dataset.loadWebhookDeliveriesId);
      }, "Webhook deliveries loaded");
      return;
    }

    const retry = event.target.closest("[data-retry-webhook-delivery-id]");
    if (retry) {
      await runAction(async () => {
        await api(`/api/webhook-deliveries/${retry.dataset.retryWebhookDeliveryId}/retry`, { method: "POST" });
        await loadWebhookDeliveries(retry.dataset.webhookId);
      }, "Webhook delivery requeued");
      return;
    }

    const rotate = event.target.closest("[data-rotate-webhook-token-id]");
    if (rotate) {
      await runAction(async () => {
        const rotated = normalizeWebhook(await api(`/api/webhook-definitions/${rotate.dataset.rotateWebhookTokenId}/rotate-token`, { method: "POST" }));
        if (rotated && rotated.token) {
          state.webhookTokens[rotated.id] = rotated.token;
        }
        await loadWebhooks();
      }, "Webhook token rotated");
      return;
    }

    const toggle = event.target.closest("[data-toggle-webhook-id]");
    if (toggle) {
      await runAction(async () => {
        await api(`/api/webhook-definitions/${toggle.dataset.toggleWebhookId}`, {
          method: "PATCH",
          body: { spec: { enabled: toggle.dataset.webhookEnabled === "true" } }
        });
        await loadWebhooks();
      }, "Webhook updated");
      return;
    }

    const remove = event.target.closest("[data-delete-webhook-id]");
    if (remove) {
      await runAction(async () => {
        await api(`/api/webhook-definitions/${remove.dataset.deleteWebhookId}`, { method: "DELETE" });
        delete state.webhookRuns[remove.dataset.deleteWebhookId];
        delete state.webhookDeliveries[remove.dataset.deleteWebhookId];
        delete state.webhookTokens[remove.dataset.deleteWebhookId];
        delete state.automationRunFilters[automationRunFilterKey("webhook", remove.dataset.deleteWebhookId)];
        await loadWebhooks();
      }, "Webhook deleted");
    }
  });

  els.webhooks.addEventListener("change", (event) => {
    if (handleAutomationRunFilterChange(event, renderWebhooks)) {
      return;
    }
    if (!event.target.matches("select[name='engine_type']")) {
      return;
    }
    const form = event.target.closest("[data-webhook-form]");
    if (form) {
      renderWebhookEditEngineFields(form);
    }
  });

  els.webhooks.addEventListener("submit", async (event) => {
    const form = event.target.closest("[data-webhook-form]");
    if (!form) {
      return;
    }
    event.preventDefault();
    await runAction(async () => {
      await api(`/api/webhook-definitions/${form.dataset.webhookForm}`, {
        method: "PATCH",
        body: { spec: webhookSpec(form) }
      });
      await loadWebhooks();
    }, "Webhook saved");
  });

  els.ticketHookProject.addEventListener("change", async () => {
    const projectID = els.ticketHookProject.value;
    state.selectedProject = state.projects.find((project) => project.id === projectID) || state.selectedProject;
    await Promise.all([loadCronJobs(projectID), loadWebhooks(projectID), loadTicketHooks(projectID), loadCreatePages(projectID)]);
  });

  els.ticketHookForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const projectID = selectedTicketHookProjectID();
    if (!projectID) {
      setNotice("Choose a project for ticket hooks");
      return;
    }
    const form = event.currentTarget;
    await runAction(async () => {
      await api(`/api/projects/${projectID}/ticket-hooks`, {
        method: "POST",
        body: { spec: ticketHookSpec(form) }
      });
      form.reset();
      setFormChecked(form, "enabled", true);
      renderTicketHookEngineFields();
      await loadTicketHooks(projectID);
    }, "Ticket hook created");
  });

  els.ticketHooks.addEventListener("click", async (event) => {
    const showRuns = event.target.closest("[data-load-ticket-hook-runs-id]");
    if (showRuns) {
      await runAction(async () => {
        await loadTicketHookRuns(showRuns.dataset.loadTicketHookRunsId);
      }, "Ticket hook runs loaded");
      return;
    }

    const preview = event.target.closest("[data-preview-ticket-hook-id]");
    if (preview) {
      await runAction(async () => {
        const result = await api(`/api/ticket-hooks/${preview.dataset.previewTicketHookId}/preview`, {
          method: "POST",
          body: { spec: ticketHookPreviewSpec() }
        });
        state.ticketHookPreview = result;
        renderTicketHookPreview();
      }, "Ticket hook previewed");
      return;
    }

    const toggle = event.target.closest("[data-toggle-ticket-hook-id]");
    if (toggle) {
      await runAction(async () => {
        await api(`/api/ticket-hooks/${toggle.dataset.toggleTicketHookId}`, {
          method: "PATCH",
          body: { spec: { enabled: toggle.dataset.ticketHookEnabled === "true" } }
        });
        await loadTicketHooks();
      }, "Ticket hook updated");
      return;
    }

    const remove = event.target.closest("[data-delete-ticket-hook-id]");
    if (remove) {
      await runAction(async () => {
        await api(`/api/ticket-hooks/${remove.dataset.deleteTicketHookId}`, { method: "DELETE" });
        if (state.ticketHookPreview && state.ticketHookPreview.metadata && state.ticketHookPreview.metadata.hook_id === remove.dataset.deleteTicketHookId) {
          state.ticketHookPreview = null;
        }
        delete state.ticketHookRuns[remove.dataset.deleteTicketHookId];
        delete state.automationRunFilters[automationRunFilterKey("ticket-hook", remove.dataset.deleteTicketHookId)];
        await loadTicketHooks();
      }, "Ticket hook deleted");
    }
  });

  els.ticketHooks.addEventListener("change", (event) => {
    if (handleAutomationRunFilterChange(event, renderTicketHooks)) {
      return;
    }
    if (!event.target.matches("select[name='engine_type']")) {
      return;
    }
    const form = event.target.closest("[data-ticket-hook-form]");
    if (form) {
      renderTicketHookEditEngineFields(form);
    }
  });

  els.ticketHooks.addEventListener("submit", async (event) => {
    const form = event.target.closest("[data-ticket-hook-form]");
    if (!form) {
      return;
    }
    event.preventDefault();
    await runAction(async () => {
      await api(`/api/ticket-hooks/${form.dataset.ticketHookForm}`, {
        method: "PATCH",
        body: { spec: ticketHookSpec(form) }
      });
      await loadTicketHooks();
    }, "Ticket hook saved");
  });

  els.createPageLogicType.addEventListener("change", () => {
    renderCreatePageLogicFields();
  });

  els.createPages.addEventListener("change", (event) => {
    if (handleAutomationRunFilterChange(event, renderCreatePages)) {
      return;
    }
    if (!event.target.matches("select[name='logic_type']")) {
      return;
    }
    const form = event.target.closest("[data-create-page-form]");
    if (form) {
      renderCreatePageEditLogicFields(form);
    }
  });

  els.createPageProject.addEventListener("change", async () => {
    const projectID = els.createPageProject.value;
    state.selectedProject = state.projects.find((project) => project.id === projectID) || state.selectedProject;
    await Promise.all([loadCronJobs(projectID), loadWebhooks(projectID), loadTicketHooks(projectID), loadCreatePages(projectID)]);
  });

  els.createPageForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const projectID = selectedCreatePageProjectID();
    if (!projectID) {
      setNotice("Choose a project for create pages");
      return;
    }
    const form = event.currentTarget;
    await runAction(async () => {
      await api(`/api/projects/${projectID}/ticket-create-pages`, {
        method: "POST",
        body: { spec: createPageSpec(form) }
      });
      form.reset();
      setFormChecked(form, "enabled", true);
      setFormValue(form, "field_layout", `[{"key":"title","type":"text","required":true}]`);
      setFormValue(form, "defaults", `{"priority":"High"}`);
      renderCreatePageLogicFields();
      await loadCreatePages(projectID);
    }, "Create page saved");
  });

  els.createPages.addEventListener("click", async (event) => {
    const showRuns = event.target.closest("[data-load-create-page-runs-id]");
    if (showRuns) {
      await runAction(async () => {
        await loadCreatePageRuns(showRuns.dataset.loadCreatePageRunsId);
      }, "Create page runs loaded");
      return;
    }

    const schema = event.target.closest("[data-load-create-page-schema-id]");
    if (schema) {
      await runAction(async () => {
        await loadCreatePageSchema(schema.dataset.loadCreatePageSchemaId, schema.dataset.createPageProjectId, schema.dataset.createPageSlug);
      }, "Create page schema loaded");
      return;
    }

    const toggle = event.target.closest("[data-toggle-create-page-id]");
    if (toggle) {
      await runAction(async () => {
        await api(`/api/ticket-create-pages/${toggle.dataset.toggleCreatePageId}`, {
          method: "PATCH",
          body: { spec: { enabled: toggle.dataset.createPageEnabled === "true" } }
        });
        await loadCreatePages();
      }, "Create page updated");
      return;
    }

    const remove = event.target.closest("[data-delete-create-page-id]");
    if (remove) {
      if (!window.confirm("Delete this create page?")) {
        return;
      }
      await runAction(async () => {
        await api(`/api/ticket-create-pages/${remove.dataset.deleteCreatePageId}`, { method: "DELETE" });
        delete state.createPageRuns[remove.dataset.deleteCreatePageId];
        delete state.automationRunFilters[automationRunFilterKey("create-page", remove.dataset.deleteCreatePageId)];
        await loadCreatePages();
      }, "Create page deleted");
    }
  });

  els.createPages.addEventListener("submit", async (event) => {
    const form = event.target.closest("[data-create-page-form]");
    if (!form) {
      return;
    }
    event.preventDefault();
    await runAction(async () => {
      await api(`/api/ticket-create-pages/${form.dataset.createPageForm}`, {
        method: "PATCH",
        body: { spec: createPageSpec(form, { includeEmptyOptionals: true, clearUnselectedLogic: true }) }
      });
      await loadCreatePages();
    }, "Create page saved");
  });

  els.engineForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    await runAction(async () => {
      const spec = engineTestSpec(form);
      const result = await api("/api/engines/test", { method: "POST", body: { spec } });
      state.engineResult = result;
      renderEngineResult();
    }, "Engine tested");
  });

  els.notificationUnreadOnly.addEventListener("change", async () => {
    state.unreadNotificationsOnly = els.notificationUnreadOnly.checked;
    await loadNotifications();
  });

  els.notificationRefresh.addEventListener("click", async () => {
    await runAction(async () => {
      await loadNotifications();
    }, "Notifications refreshed");
  });

  els.notificationReadAll.addEventListener("click", async () => {
    await runAction(async () => {
      await api("/api/notifications/read-all", { method: "POST" });
      await loadNotifications();
    }, "Notifications marked read");
  });

  els.sprintForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    if (!state.selectedProject) {
      setNotice("Select a project before creating a sprint");
      return;
    }
    const form = event.currentTarget;
    const data = formData(form);
    await runAction(async () => {
      await api(`/api/projects/${state.selectedProject.id}/sprints`, {
        method: "POST",
        body: {
          spec: {
            name: data.name || "",
            goal: data.goal || "",
            start_date: data.start_date || "",
            end_date: data.end_date || ""
          }
        }
      });
      form.reset();
      renderSprintFilter();
      await loadSprints();
    }, "Sprint created");
  });

  els.sprintForm.elements.state_filter.addEventListener("change", async (event) => {
    state.sprintFilterState = event.target.value || "";
    await loadSprints();
  });

  els.sprints.addEventListener("submit", async (event) => {
    const form = event.target.closest("[data-sprint-edit-form]");
    if (!form) {
      return;
    }
    event.preventDefault();
    const data = formData(form);
    await runAction(async () => {
      await api(`/api/sprints/${form.dataset.sprintEditForm}`, {
        method: "PATCH",
        body: {
          spec: {
            name: data.name || "",
            goal: data.goal || "",
            start_date: data.start_date || "",
            end_date: data.end_date || ""
          }
        }
      });
      await loadSprints();
      if (state.selectedSprintReportID === form.dataset.sprintEditForm) {
        await loadSprintReport(state.selectedSprintReportID);
      }
    }, "Sprint updated");
  });

  els.sprints.addEventListener("click", async (event) => {
    const report = event.target.closest("[data-sprint-report-id]");
    if (report) {
      state.selectedSprintReportID = report.dataset.sprintReportId;
      await runAction(async () => {
        await loadSprintReport(state.selectedSprintReportID);
      }, "Sprint report loaded");
      return;
    }

    const start = event.target.closest("[data-start-sprint-id]");
    if (start) {
      await runAction(async () => {
        await api(`/api/sprints/${start.dataset.startSprintId}/start`, { method: "POST" });
        await loadSprints();
      }, "Sprint started");
      return;
    }

    const complete = event.target.closest("[data-complete-sprint-id]");
    if (complete) {
      await runAction(async () => {
        await api(`/api/sprints/${complete.dataset.completeSprintId}/complete`, { method: "POST" });
        await loadSprints();
        if (state.selectedSprintReportID === complete.dataset.completeSprintId) {
          await loadSprintReport(state.selectedSprintReportID);
        }
      }, "Sprint completed");
      return;
    }

    const remove = event.target.closest("[data-delete-sprint-id]");
    if (remove) {
      await runAction(async () => {
        await api(`/api/sprints/${remove.dataset.deleteSprintId}`, { method: "DELETE" });
        await loadSprints();
        await loadTickets();
      }, "Sprint deleted");
    }
  });

  els.projectLabelForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    if (!state.selectedProject) {
      setNotice("Select a project before creating a label");
      return;
    }
    const form = event.currentTarget;
    const data = formData(form);
    await runAction(async () => {
      await api(`/api/projects/${state.selectedProject.id}/labels`, {
        method: "POST",
        body: {
          spec: {
            label: data.label || "",
            description: data.description || "",
            color: data.color || ""
          }
        }
      });
      form.reset();
      setFormValue(form, "color", "#5b7cfa");
      await loadProjectLabels();
    }, "Label created");
  });

  els.projectLabels.addEventListener("click", async (event) => {
    const remove = event.target.closest("[data-delete-project-label]");
    if (!remove || !state.selectedProject) {
      return;
    }
    await runAction(async () => {
      await api(`/api/projects/${state.selectedProject.id}/labels/${encodeURIComponent(remove.dataset.deleteProjectLabel)}`, { method: "DELETE" });
      await loadProjectLabels();
    }, "Label deleted");
  });

  els.projectLabels.addEventListener("submit", async (event) => {
    const form = event.target.closest("[data-project-label-edit-form]");
    if (!form || !state.selectedProject) {
      return;
    }
    event.preventDefault();
    const data = formData(form);
    await runAction(async () => {
      if (form.dataset.projectLabelCatalog === "true") {
        await api(`/api/projects/${state.selectedProject.id}/labels/${encodeURIComponent(form.dataset.projectLabelEditForm)}`, {
          method: "PATCH",
          body: {
            spec: {
              description: data.description || "",
              color: data.color || ""
            }
          }
        });
      } else {
        await api(`/api/projects/${state.selectedProject.id}/labels`, {
          method: "POST",
          body: {
            spec: {
              label: form.dataset.projectLabelEditForm || "",
              description: data.description || "",
              color: data.color || ""
            }
          }
        });
      }
      await loadProjectLabels();
    }, "Label updated");
  });

  els.backlog.addEventListener("click", async (event) => {
    const assignSprint = event.target.closest("[data-backlog-assign-sprint-id]");
    if (assignSprint) {
      const control = assignSprint.closest("[data-backlog-sprint-control]");
      const select = control ? control.querySelector("select") : null;
      const sprintID = select ? select.value : "";
      if (!sprintID) {
        setNotice("Choose a sprint first");
        return;
      }
      await runAction(async () => {
        await api(`/api/tickets/${assignSprint.dataset.backlogAssignSprintId}/sprint`, {
          method: "PUT",
          body: { spec: { sprint_id: sprintID } }
        });
        await refreshBacklogSprintViews(assignSprint.dataset.backlogAssignSprintId);
      }, "Backlog ticket assigned to sprint");
      return;
    }

    const removeSprint = event.target.closest("[data-backlog-remove-sprint-id]");
    if (removeSprint) {
      await runAction(async () => {
        await api(`/api/tickets/${removeSprint.dataset.backlogRemoveSprintId}/sprint`, { method: "DELETE" });
        await refreshBacklogSprintViews(removeSprint.dataset.backlogRemoveSprintId);
      }, "Backlog ticket removed from sprint");
      return;
    }

    const move = event.target.closest("[data-backlog-move-id]");
    if (!move || !state.selectedProject) {
      return;
    }
    const direction = move.dataset.backlogMoveDirection;
    const index = state.backlog.findIndex((ticket) => ticket.id === move.dataset.backlogMoveId);
    const targetIndex = direction === "up" ? index - 1 : index + 1;
    if (index < 0 || targetIndex < 0 || targetIndex >= state.backlog.length) {
      return;
    }
    const reordered = state.backlog.slice();
    const [ticket] = reordered.splice(index, 1);
    reordered.splice(targetIndex, 0, ticket);
    await runAction(async () => {
      const data = await api(`/api/projects/${state.selectedProject.id}/backlog`, {
        method: "PATCH",
        body: { spec: { ticket_ids: reordered.map((item) => item.id) } }
      });
      state.backlog = listItems(data).map(normalizeTicket);
      renderBacklog();
    }, "Backlog reordered");
  });

  els.backlog.addEventListener("dragstart", (event) => {
    const item = event.target.closest("[data-backlog-drag-id]");
    if (!item || event.target.closest("a, button, input, textarea, select")) {
      event.preventDefault();
      return;
    }
    event.dataTransfer.effectAllowed = "move";
    event.dataTransfer.setData("application/rayboard-backlog-ticket", item.dataset.backlogDragId);
    event.dataTransfer.setData("text/plain", item.dataset.backlogDragId);
    item.classList.add("is-dragging");
  });

  els.backlog.addEventListener("dragend", (event) => {
    const item = event.target.closest("[data-backlog-drag-id]");
    if (item) {
      item.classList.remove("is-dragging");
    }
  });

  els.backlog.addEventListener("dragover", (event) => {
    if (dataTransferHasType(event, "application/rayboard-backlog-ticket")) {
      event.preventDefault();
      event.dataTransfer.dropEffect = "move";
    }
  });

  els.backlog.addEventListener("drop", async (event) => {
    const ticketID = event.dataTransfer.getData("application/rayboard-backlog-ticket");
    if (!ticketID || !state.selectedProject) {
      return;
    }
    event.preventDefault();
    const target = event.target.closest("[data-backlog-drag-id]");
    await runAction(async () => {
      await reorderBacklogTicket(ticketID, target ? target.dataset.backlogDragId : "");
    }, "Backlog reordered");
  });

  els.statusForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    if (!state.selectedProject) {
      setNotice("Select a project before replacing statuses");
      return;
    }
    const form = event.currentTarget;
    const data = formData(form);
    const statuses = parseJSONArrayField(data.statuses, "Workflow statuses JSON")
      .map((status) => ({
        slug: String(status.slug || "").trim(),
        name: String(status.name || "").trim()
      }))
      .filter((status) => status.slug && status.name);
    await runAction(async () => {
      await api(`/api/projects/${state.selectedProject.id}/statuses`, {
        method: "PUT",
        body: { spec: { statuses } }
      });
      await loadWorkflowStatuses();
      await loadBoards();
      await loadTickets();
      await loadBacklog();
    }, "Workflow statuses replaced");
  });

  els.boardForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    if (!state.selectedProject) {
      setNotice("Select a project before creating a board");
      return;
    }
    const form = event.currentTarget;
    const data = formData(form);
    await runAction(async () => {
      const board = normalizeBoard(await api(`/api/projects/${state.selectedProject.id}/boards`, {
        method: "POST",
        body: { spec: {
          name: data.name || "",
          description: data.description || "",
          status_slugs: parseCommaList(data.status_slugs),
          wip_limits: parseBoardWIPLimits(data.wip_limits)
        } }
      }));
      form.reset();
      state.selectedBoardID = board ? board.id : "";
      await loadBoards();
    }, "Board created");
  });

  els.boardSavedViewFilter.addEventListener("change", async (event) => {
    state.selectedBoardSavedViewID = event.currentTarget.value || "";
    await loadBoardTickets(state.selectedBoardID);
    renderWorkflowPanel();
    renderTickets();
  });

  els.boards.addEventListener("click", async (event) => {
    const select = event.target.closest("[data-select-board-id]");
    if (select) {
      state.selectedBoardID = select.dataset.selectBoardId;
      await loadBoardTickets(state.selectedBoardID);
      renderWorkflowPanel();
      renderTickets();
      return;
    }
    const remove = event.target.closest("[data-delete-board-id]");
    if (!remove) {
      return;
    }
    await runAction(async () => {
      await api(`/api/boards/${remove.dataset.deleteBoardId}`, { method: "DELETE" });
      if (state.selectedBoardID === remove.dataset.deleteBoardId) {
        state.selectedBoardID = "";
      }
      await loadBoards();
    }, "Board deleted");
  });

  els.boards.addEventListener("submit", async (event) => {
    const form = event.target.closest("[data-board-edit-form]");
    if (!form) {
      return;
    }
    event.preventDefault();
    const data = formData(form);
    await runAction(async () => {
      await api(`/api/boards/${form.dataset.boardEditForm}`, {
        method: "PATCH",
        body: {
          spec: {
            name: data.name || "",
            description: data.description || "",
            status_slugs: parseCommaList(data.status_slugs),
            wip_limits: parseBoardWIPLimits(data.wip_limits)
          }
        }
      });
      await loadBoards();
    }, "Board updated");
  });

  els.ticketColumns.addEventListener("dragstart", (event) => {
    const ticket = event.target.closest("[data-board-ticket-id]");
    if (!ticket || event.target.closest("a, button, input, textarea, select")) {
      event.preventDefault();
      return;
    }
    event.dataTransfer.effectAllowed = "move";
    event.dataTransfer.setData("application/rayboard-board-ticket", ticket.dataset.boardTicketId);
    event.dataTransfer.setData("text/plain", ticket.dataset.boardTicketId);
    ticket.classList.add("is-dragging");
  });

  els.ticketColumns.addEventListener("dragend", (event) => {
    const ticket = event.target.closest("[data-board-ticket-id]");
    if (ticket) {
      ticket.classList.remove("is-dragging");
    }
  });

  els.ticketColumns.addEventListener("dragover", (event) => {
    const list = event.target.closest("[data-board-drop-status]");
    if (list && dataTransferHasType(event, "application/rayboard-board-ticket")) {
      event.preventDefault();
      event.dataTransfer.dropEffect = "move";
    }
  });

  els.ticketColumns.addEventListener("drop", async (event) => {
    const list = event.target.closest("[data-board-drop-status]");
    const ticketID = event.dataTransfer.getData("application/rayboard-board-ticket");
    if (!list || !ticketID) {
      return;
    }
    event.preventDefault();
    const status = list.dataset.boardDropStatus;
    const ticket = state.tickets.find((item) => item.id === ticketID);
    if (!status || (ticket && ticket.status === status)) {
      return;
    }
    await runAction(async () => {
      await api(`/api/tickets/${ticketID}`, {
        method: "PATCH",
        body: { spec: { status } }
      });
      await refreshTicketViews(ticketID);
      await refreshSelectedSprintReport();
      await refreshSelectedVersionReport();
    }, "Ticket moved");
  });

  els.componentForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    if (!state.selectedProject) {
      setNotice("Select a project before creating a component");
      return;
    }
    const form = event.currentTarget;
    const data = formData(form);
    await runAction(async () => {
      await api(`/api/projects/${state.selectedProject.id}/components`, {
        method: "POST",
        body: { spec: { name: data.name || "", description: data.description || "" } }
      });
      form.reset();
      await loadComponents();
    }, "Component created");
  });

  els.versionForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    if (!state.selectedProject) {
      setNotice("Select a project before creating a version");
      return;
    }
    const form = event.currentTarget;
    const data = formData(form);
    await runAction(async () => {
      await api(`/api/projects/${state.selectedProject.id}/versions`, {
        method: "POST",
        body: {
          spec: {
            name: data.name || "",
            description: data.description || "",
            target_date: data.target_date || "",
            release_date: ""
          }
        }
      });
      form.reset();
      await loadVersions();
    }, "Version created");
  });

  els.components.addEventListener("click", async (event) => {
    const remove = event.target.closest("[data-delete-component-id]");
    if (!remove) {
      return;
    }
    await runAction(async () => {
      await api(`/api/components/${remove.dataset.deleteComponentId}`, { method: "DELETE" });
      await loadComponents();
      await loadTickets();
    }, "Component deleted");
  });

  els.components.addEventListener("submit", async (event) => {
    const form = event.target.closest("[data-component-edit-form]");
    if (!form) {
      return;
    }
    event.preventDefault();
    await runAction(async () => {
      await api(`/api/components/${form.dataset.componentEditForm}`, {
        method: "PATCH",
        body: { spec: componentUpdateSpec(form) }
      });
      await loadComponents();
      await loadTickets();
    }, "Component updated");
  });

  els.versions.addEventListener("click", async (event) => {
    const report = event.target.closest("[data-version-report-id]");
    if (report) {
      state.selectedVersionReportID = report.dataset.versionReportId;
      await runAction(async () => {
        await loadVersionReport(state.selectedVersionReportID);
      }, "Version report loaded");
      return;
    }

    const status = event.target.closest("[data-version-status]");
    if (status) {
      await runAction(async () => {
        await api(`/api/versions/${status.dataset.versionId}`, {
          method: "PATCH",
          body: { spec: { status: status.dataset.versionStatus } }
        });
        await loadVersions();
        if (state.selectedVersionReportID === status.dataset.versionId) {
          await loadVersionReport(state.selectedVersionReportID);
        }
      }, "Version updated");
      return;
    }

    const remove = event.target.closest("[data-delete-version-id]");
    if (!remove) {
      return;
    }
    await runAction(async () => {
      await api(`/api/versions/${remove.dataset.deleteVersionId}`, { method: "DELETE" });
      if (state.selectedVersionReportID === remove.dataset.deleteVersionId) {
        state.selectedVersionReportID = "";
        state.versionReport = null;
      }
      await loadVersions();
      await loadTickets();
    }, "Version deleted");
  });

  els.versions.addEventListener("submit", async (event) => {
    const form = event.target.closest("[data-version-edit-form]");
    if (!form) {
      return;
    }
    event.preventDefault();
    await runAction(async () => {
      await api(`/api/versions/${form.dataset.versionEditForm}`, {
        method: "PATCH",
        body: { spec: versionUpdateSpec(form) }
      });
      await loadVersions();
      await loadTickets();
      await loadRoadmap();
      if (state.selectedVersionReportID === form.dataset.versionEditForm) {
        await loadVersionReport(state.selectedVersionReportID);
      }
    }, "Version updated");
  });

  els.roadmap.addEventListener("submit", async (event) => {
    const form = event.target.closest("[data-roadmap-schedule-form]");
    if (!form || !state.selectedProject) {
      return;
    }
    event.preventDefault();
    await runAction(async () => {
      await scheduleRoadmapItem(form);
      await loadTickets();
    }, "Roadmap scheduled");
  });

  els.roadmap.addEventListener("click", async (event) => {
    const quickSchedule = event.target.closest("[data-roadmap-quick-schedule-id]");
    if (!quickSchedule || !state.selectedProject) {
      return;
    }
    const start = todayISODate();
    await runAction(async () => {
      await scheduleRoadmapItemSpec(quickSchedule.dataset.roadmapQuickScheduleId, start, addDaysISO(start, 6));
      await loadTickets();
    }, "Roadmap scheduled");
  });

  els.roadmap.addEventListener("dragstart", (event) => {
    const item = event.target.closest("[data-roadmap-drag-id]");
    if (!item || event.target.closest("a, button, input, textarea, select")) {
      event.preventDefault();
      return;
    }
    const track = item.closest("[data-roadmap-timeline-track]");
    const start = dateToUTC(item.dataset.roadmapStartDate);
    const due = dateToUTC(item.dataset.roadmapDueDate);
    if (!track || !start || !due) {
      event.preventDefault();
      return;
    }
    const bounds = {
      start: item.dataset.roadmapBoundsStart,
      end: item.dataset.roadmapBoundsEnd
    };
    const spanDays = Math.max(daysBetween(dateToUTC(bounds.start), dateToUTC(bounds.end)) + 1, 1);
    const rect = track.getBoundingClientRect();
    event.dataTransfer.effectAllowed = "move";
    event.dataTransfer.setData("application/rayboard-roadmap-epic", JSON.stringify({
      id: item.dataset.roadmapDragId,
      start: item.dataset.roadmapStartDate,
      due: item.dataset.roadmapDueDate,
      startX: event.clientX,
      trackWidth: rect.width,
      spanDays
    }));
    event.dataTransfer.setData("text/plain", item.dataset.roadmapDragId);
    item.classList.add("is-dragging");
  });

  els.roadmap.addEventListener("dragend", (event) => {
    const item = event.target.closest("[data-roadmap-drag-id]");
    if (item) {
      item.classList.remove("is-dragging");
    }
  });

  els.roadmap.addEventListener("dragover", (event) => {
    const track = event.target.closest("[data-roadmap-timeline-track]");
    if (track && dataTransferHasType(event, "application/rayboard-roadmap-epic")) {
      event.preventDefault();
      event.dataTransfer.dropEffect = "move";
    }
  });

  els.roadmap.addEventListener("drop", async (event) => {
    const track = event.target.closest("[data-roadmap-timeline-track]");
    const payload = event.dataTransfer.getData("application/rayboard-roadmap-epic");
    if (!track || !payload || !state.selectedProject) {
      return;
    }
    event.preventDefault();
    let details;
    try {
      details = JSON.parse(payload);
    } catch {
      return;
    }
    const start = dateToUTC(details.start);
    const due = dateToUTC(details.due);
    const trackWidth = Number(details.trackWidth || 0);
    const spanDays = Number(details.spanDays || 0);
    if (!details.id || !start || !due || trackWidth <= 0 || spanDays <= 0) {
      return;
    }
    const deltaDays = Math.round(((event.clientX - Number(details.startX || 0)) / trackWidth) * spanDays);
    if (deltaDays === 0) {
      return;
    }
    await runAction(async () => {
      await scheduleRoadmapItemSpec(details.id, formatISODate(addDays(start, deltaDays)), formatISODate(addDays(due, deltaDays)));
      await loadTickets();
    }, "Roadmap rescheduled");
  });

  els.roadmapDependencies.addEventListener("submit", async (event) => {
    const dependencyForm = event.target.closest("[data-roadmap-dependency-form]");
    if (!dependencyForm) {
      return;
    }
    event.preventDefault();
    const data = formData(dependencyForm);
    if (!data.source_ticket_id || !data.target_ticket_id) {
      setNotice("Choose source and target issues first");
      return;
    }
    if (data.source_ticket_id === data.target_ticket_id) {
      setNotice("Choose two different roadmap issues");
      return;
    }
    await runAction(async () => {
      await api(`/api/tickets/${data.source_ticket_id}/links`, {
        method: "POST",
        body: {
          spec: {
            target_ticket_id: data.target_ticket_id || "",
            link_type: data.link_type || "relates_to"
          }
        }
      });
      dependencyForm.reset();
      await refreshRoadmapDependencyViews(data.source_ticket_id, data.target_ticket_id);
    }, "Roadmap dependency added");
  });

  els.roadmapDependencies.addEventListener("click", async (event) => {
    const deleteDependency = event.target.closest("[data-delete-roadmap-dependency-id]");
    if (!deleteDependency) {
      return;
    }
    await runAction(async () => {
      await api(`/api/tickets/${deleteDependency.dataset.sourceTicketId}/links/${deleteDependency.dataset.deleteRoadmapDependencyId}`, { method: "DELETE" });
      await refreshRoadmapDependencyViews(deleteDependency.dataset.sourceTicketId, deleteDependency.dataset.targetTicketId);
    }, "Roadmap dependency removed");
  });

  els.roadmapDependencies.addEventListener("change", (event) => {
    const form = event.target.closest("[data-roadmap-dependency-form]");
    if (form) {
      syncRoadmapDependencyTargetOptions(form);
    }
  });

  els.fieldForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    if (!state.selectedProject) {
      setNotice("Select a project before creating a custom field");
      return;
    }
    const form = event.currentTarget;
    const data = formData(form);
    await runAction(async () => {
      await api(`/api/projects/${state.selectedProject.id}/custom-fields`, {
        method: "POST",
        body: {
          spec: {
            key: data.key || "",
            name: data.name || "",
            field_type: data.field_type || "text",
            required: Boolean(data.required),
            options: parseOptions(data.options)
          }
        }
      });
      form.reset();
      await loadCustomFields();
    }, "Custom field created");
  });

  els.customFields.addEventListener("click", async (event) => {
    const remove = event.target.closest("[data-delete-field-id]");
    if (!remove) {
      return;
    }
    await runAction(async () => {
      await api(`/api/custom-fields/${remove.dataset.deleteFieldId}`, { method: "DELETE" });
      await loadCustomFields();
      await loadTickets();
    }, "Custom field deleted");
  });

  els.customFields.addEventListener("submit", async (event) => {
    const form = event.target.closest("[data-custom-field-edit-form]");
    if (!form) {
      return;
    }
    event.preventDefault();
    await runAction(async () => {
      await api(`/api/custom-fields/${form.dataset.customFieldEditForm}`, {
        method: "PATCH",
        body: { spec: customFieldUpdateSpec(form) }
      });
      await loadCustomFields();
      await loadTickets();
    }, "Custom field updated");
  });

  els.notifications.addEventListener("click", async (event) => {
    const button = event.target.closest("[data-notification-read-state]");
    if (!button) {
      return;
    }
    await runAction(async () => {
      const action = button.dataset.notificationReadState === "read" ? "read" : "unread";
      await api(`/api/notifications/${button.dataset.notificationId}/${action}`, { method: "POST" });
      await loadNotifications();
    }, "Notification updated");
  });

  els.searchForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    await runAction(async () => {
      await runSearch(searchSpecFromForm(form), { reset: true });
    }, "Search complete");
  });

  els.customFieldSearchControls.addEventListener("change", (event) => {
    if (!event.target.matches("select[name='field_key'], select[name='operator']")) {
      return;
    }
    const form = event.target.closest("[data-custom-field-search-form]");
    if (form) {
      renderCustomFieldSearchValueControl(form);
    }
  });

  els.customFieldSearchControls.addEventListener("submit", (event) => {
    const form = event.target.closest("[data-custom-field-search-form]");
    if (!form) {
      return;
    }
    event.preventDefault();
    const expression = customFieldSearchExpression(form);
    if (!expression || !els.searchForm) {
      return;
    }
    const input = els.searchForm.elements.filter;
    input.value = appendSearchFilter(input.value, expression);
    input.focus();
  });

  els.searchPagination.addEventListener("click", async (event) => {
    const next = event.target.closest("[data-search-next]");
    if (next) {
      if (!state.searchNextCursor) {
        return;
      }
      await runAction(async () => {
        const nextIndex = state.searchCursorIndex + 1;
        state.searchCursorStack[nextIndex] = state.searchNextCursor;
        state.searchCursorIndex = nextIndex;
        await runSearch(state.lastSearchSpec, { cursor: state.searchCursorStack[state.searchCursorIndex], reset: false });
      }, "Search page loaded");
      return;
    }

    const previous = event.target.closest("[data-search-previous]");
    if (previous) {
      if (state.searchCursorIndex <= 0) {
        return;
      }
      await runAction(async () => {
        state.searchCursorIndex -= 1;
        await runSearch(state.lastSearchSpec, { cursor: state.searchCursorStack[state.searchCursorIndex] || "", reset: false });
      }, "Search page loaded");
    }
  });

  els.savedViewForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    if (!state.selectedProject && form.elements.scope_type.value === "project") {
      setNotice("Select a project before saving a project view");
      return;
    }
    await runAction(async () => {
      const editingID = form.dataset.savedViewEditId || "";
      const spec = savedViewSpecFromForm(form);
      if (editingID) {
        await api(`/api/saved-views/${editingID}`, { method: "PATCH", body: { spec: savedViewUpdateSpec(spec) } });
      } else {
        await api("/api/saved-views", { method: "POST", body: { spec } });
      }
      resetSavedViewForm();
      resetSavedViewPagination();
      await loadSavedViews();
      await loadBoardSavedViews();
    }, form.dataset.savedViewEditId ? "Saved view updated" : "Saved view created");
  });

  els.savedViewCancelEdit.addEventListener("click", () => {
    resetSavedViewForm();
  });

  els.savedViews.addEventListener("click", async (event) => {
    const apply = event.target.closest("[data-apply-saved-view-id]");
    if (apply) {
      const view = state.savedViews.find((item) => item.id === apply.dataset.applySavedViewId);
      if (!view) {
        return;
      }
      await runAction(async () => {
        await applySavedView(view);
      }, "Saved view applied");
      return;
    }

    const edit = event.target.closest("[data-edit-saved-view-id]");
    if (edit) {
      const view = state.savedViews.find((item) => item.id === edit.dataset.editSavedViewId);
      if (view) {
        editSavedView(view);
      }
      return;
    }

    const remove = event.target.closest("[data-delete-saved-view-id]");
    if (remove) {
      await runAction(async () => {
        await api(`/api/saved-views/${remove.dataset.deleteSavedViewId}`, { method: "DELETE" });
        await loadSavedViews();
        await loadBoardSavedViews();
      }, "Saved view deleted");
    }
  });

  els.savedViewPagination.addEventListener("click", async (event) => {
    const next = event.target.closest("[data-saved-view-next]");
    if (next) {
      await runAction(async () => {
        state.savedViewOffset += savedViewPageSize;
        await loadSavedViews();
      }, "Saved view page loaded");
      return;
    }

    const previous = event.target.closest("[data-saved-view-previous]");
    if (previous) {
      await runAction(async () => {
        state.savedViewOffset = Math.max(0, state.savedViewOffset - savedViewPageSize);
        await loadSavedViews();
      }, "Saved view page loaded");
    }
  });

  els.pinnedProjectViews.addEventListener("click", async (event) => {
    const apply = event.target.closest("[data-apply-pinned-project-view-id]");
    if (!apply) {
      return;
    }
    const view = state.pinnedProjectSavedViews.find((item) => item.id === apply.dataset.applyPinnedProjectViewId);
    if (!view) {
      return;
    }
    await runAction(async () => {
      await applySavedView(view, { navigateToSearch: true });
    }, "Pinned view applied");
  });

  els.tokenForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const data = formData(form);
    await runAction(async () => {
      const created = await api("/api/tokens", { method: "POST", body: { spec: { name: data.name || "api-token" } } });
      state.createdToken = normalizeToken(created);
      form.reset();
      await loadTokens();
      renderTokens();
    }, "API token created");
  });

  els.apiTokens.addEventListener("click", async (event) => {
    const button = event.target.closest("[data-revoke-token-id]");
    if (!button) {
      return;
    }
    await runAction(async () => {
      await api(`/api/tokens/${button.dataset.revokeTokenId}`, { method: "DELETE" });
      state.createdToken = null;
      await loadTokens();
    }, "API token revoked");
  });

  els.rbacRefresh.addEventListener("click", async () => {
    await runAction(async () => {
      await loadRBAC();
    }, "RBAC refreshed");
  });

  els.rbacUserForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const data = formData(form);
    await runAction(async () => {
      const created = normalizeUser(await api("/api/users", {
        method: "POST",
        body: {
          spec: {
            username: data.username || "",
            display_name: data.display_name || "",
            password: data.password || "",
            disabled: false
          }
        }
      }));
      form.reset();
      await loadRBAC();
      if (created && created.generated_password) {
        setNotice(`User created. Temporary password: ${created.generated_password}`);
      }
    }, "User created");
  });

  els.rbacGroupForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const data = formData(form);
    await runAction(async () => {
      await api("/api/groups", {
        method: "POST",
        body: { spec: { name: data.name || "", display_name: data.display_name || "" } }
      });
      form.reset();
      await loadRBAC();
    }, "Group created");
  });

  els.rbacMemberForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const data = formData(event.currentTarget);
    if (!data.group_id || !data.user_id) {
      setNotice("Choose a group and user");
      return;
    }
    await runAction(async () => {
      await api(`/api/groups/${data.group_id}/members/${data.user_id}`, { method: "POST" });
      await loadRBAC();
    }, "Group member added");
  });

  els.rbacBindingForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const data = formData(event.currentTarget);
    await runAction(async () => {
      await api("/api/role-bindings", {
        method: "POST",
        body: {
          spec: {
            role_name: data.role_name || "",
            subject_type: data.subject_type || "user",
            subject_id: data.subject_id || "",
            scope: data.scope || "global",
            project_id: data.scope === "project" ? data.project_id || "" : ""
          }
        }
      });
      await loadRBAC();
    }, "Role bound");
  });

  els.rbacPermissionForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    await runAction(async () => {
      await loadRBACEffectivePermissions();
    }, "Effective permissions loaded");
  });

  els.rbacBindingForm.elements.subject_type.addEventListener("change", () => {
    renderBindingSubjectOptions();
  });

  els.rbacPermissionForm.elements.scope.addEventListener("change", () => {
    renderRBACFormOptions();
  });

  els.rbacPanel.addEventListener("click", async (event) => {
    const userState = event.target.closest("[data-rbac-user-disabled]");
    if (userState) {
      await runAction(async () => {
        await api(`/api/users/${userState.dataset.rbacUserId}`, {
          method: "PATCH",
          body: { spec: { disabled: userState.dataset.rbacUserDisabled === "true" } }
        });
        await loadRBAC();
      }, "User updated");
      return;
    }

    const deleteUser = event.target.closest("[data-delete-rbac-user-id]");
    if (deleteUser) {
      await runAction(async () => {
        await api(`/api/users/${deleteUser.dataset.deleteRbacUserId}`, { method: "DELETE" });
        await loadRBAC();
      }, "User deleted");
      return;
    }

    const removeMember = event.target.closest("[data-remove-group-member]");
    if (removeMember) {
      await runAction(async () => {
        await api(`/api/groups/${removeMember.dataset.groupId}/members/${removeMember.dataset.userId}`, { method: "DELETE" });
        await loadRBAC();
      }, "Group member removed");
      return;
    }

    const deleteBinding = event.target.closest("[data-delete-binding-id]");
    if (deleteBinding) {
      await runAction(async () => {
        await api(`/api/role-bindings/${deleteBinding.dataset.deleteBindingId}`, { method: "DELETE" });
        await loadRBAC();
      }, "Role binding deleted");
    }
  });

  els.settingsRefresh.addEventListener("click", async () => {
    await runAction(async () => {
      await loadSettingsPage();
    }, "Settings refreshed");
  });

  els.settingsForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const data = formData(form);
    await runAction(async () => {
      await api("/api/settings", {
        method: "PATCH",
        body: {
          spec: {
            attachment_max_size_bytes: Number(data.attachment_max_size_bytes || 0),
            attachment_allowed_content_types: parseCommaList(data.attachment_allowed_content_types),
            webhook_allowed_base_urls: parseCommaList(data.webhook_allowed_base_urls),
            demo_warning_enabled: Boolean(data.demo_warning_enabled),
            backup_enabled: Boolean(data.backup_enabled),
            system_health_note: data.system_health_note || ""
          }
        }
      });
      await loadGlobalSettings();
    }, "Global settings saved");
  });

  els.auditForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    await runAction(async () => {
      await loadAuditLog();
    }, "Audit log refreshed");
  });

  els.preferenceForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    await runAction(async () => {
      await api("/api/me/notification-preferences", {
        method: "PATCH",
        body: { spec: notificationPreferenceSpec(form) }
      });
      await loadNotificationPreferences();
    }, "Notification preferences saved");
  });

  els.projectPreferenceProject.addEventListener("change", async () => {
    const project = state.projects.find((item) => item.id === els.projectPreferenceProject.value);
    if (project) {
      state.selectedProject = project;
    }
    await runAction(async () => {
      await Promise.all([loadProjectNotificationPreferences(), loadNotificationDeliveries()]);
    }, "Project notification settings loaded");
  });

  els.projectPreferenceForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const projectID = selectedProjectPreferenceProjectID();
    if (!projectID) {
      setNotice("Choose a project for notification defaults");
      return;
    }
    await runAction(async () => {
      await api(`/api/projects/${projectID}/notification-preferences`, {
        method: "PATCH",
        body: { spec: notificationPreferenceSpec(form) }
      });
      await loadProjectNotificationPreferences(projectID);
    }, "Project notification defaults saved");
  });

  els.notificationDeliveryProject.addEventListener("change", async () => {
    const project = state.projects.find((item) => item.id === els.notificationDeliveryProject.value);
    if (project) {
      state.selectedProject = project;
    }
    await runAction(async () => {
      await loadNotificationDeliveries();
    }, "Notification deliveries loaded");
  });

  els.notificationDeliveryForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    await runAction(async () => {
      await loadNotificationDeliveries();
    }, "Notification deliveries loaded");
  });

  els.notificationDeliveries.addEventListener("click", async (event) => {
    const retry = event.target.closest("[data-retry-notification-delivery-id]");
    if (!retry) {
      return;
    }
    await runAction(async () => {
      await api(`/api/notification-deliveries/${retry.dataset.retryNotificationDeliveryId}/retry`, { method: "POST" });
      await loadNotificationDeliveries();
    }, "Notification delivery requeued");
  });

  els.openRouterProviderForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    await runAction(async () => {
      await api("/api/openrouter-providers", {
        method: "POST",
        body: { spec: openRouterProviderSpec(form) }
      });
      form.reset();
      setFormChecked(form, "enabled", true);
      setFormValue(form, "default_timeout_seconds", "30");
      setFormValue(form, "max_output_tokens", "2048");
      await loadOpenRouterProviders();
    }, "OpenRouter provider created");
  });

  els.openRouterProviders.addEventListener("click", async (event) => {
    const remove = event.target.closest("[data-delete-openrouter-provider-id]");
    if (remove) {
      if (!window.confirm("Delete this OpenRouter provider?")) {
        return;
      }
      await runAction(async () => {
        await api(`/api/openrouter-providers/${remove.dataset.deleteOpenrouterProviderId}`, { method: "DELETE" });
        await loadOpenRouterProviders();
      }, "OpenRouter provider deleted");
    }
  });

  els.openRouterProviders.addEventListener("submit", async (event) => {
    const form = event.target.closest("[data-openrouter-provider-form]");
    if (!form) {
      return;
    }
    event.preventDefault();
    await runAction(async () => {
      await api(`/api/openrouter-providers/${form.dataset.openrouterProviderForm}`, {
        method: "PATCH",
        body: { spec: openRouterProviderUpdateSpec(form) }
      });
      await loadOpenRouterProviders();
    }, "OpenRouter provider saved");
  });

  els.notificationDestinationScope.addEventListener("change", async () => {
    renderNotificationDestinationProjectOptions();
    await runAction(async () => {
      await loadNotificationDestinations();
    }, "Notification destinations refreshed");
  });

  els.notificationDestinationProject.addEventListener("change", async () => {
    const project = state.projects.find((item) => item.id === els.notificationDestinationProject.value);
    if (project) {
      state.selectedProject = project;
    }
    await runAction(async () => {
      await loadNotificationDestinations();
    }, "Notification destinations refreshed");
  });

  els.notificationDestinationForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const data = formData(form);
    const scopeType = data.scope_type || "global";
    const projectID = data.project_id || selectedNotificationDestinationProjectID();
    if (scopeType === "project" && !projectID) {
      setActionStatus("Choose a project for project destinations");
      return;
    }
    await runAction(async () => {
      await api(notificationDestinationCollectionPath(scopeType, projectID), {
        method: "POST",
        body: { spec: notificationDestinationSpec(form) }
      });
      form.reset();
      setFormChecked(form, "enabled", true);
      renderNotificationDestinationProjectOptions();
      await loadNotificationDestinations();
    }, "Notification destination created");
  });

  els.notificationDestinations.addEventListener("click", async (event) => {
    const test = event.target.closest("[data-test-notification-destination-id]");
    if (test) {
      const form = test.closest("[data-notification-destination-form]");
      const data = form ? formData(form) : {};
      await runAction(async () => {
        await api(`/api/notification-destinations/${test.dataset.testNotificationDestinationId}/test-send`, {
          method: "POST",
          body: { spec: { message: data.test_message || "" } }
        });
        await loadNotificationDestinations();
      }, "Notification test sent");
      return;
    }

    const remove = event.target.closest("[data-delete-notification-destination-id]");
    if (remove) {
      if (!window.confirm("Delete this notification destination?")) {
        return;
      }
      await runAction(async () => {
        await api(`/api/notification-destinations/${remove.dataset.deleteNotificationDestinationId}`, { method: "DELETE" });
        await loadNotificationDestinations();
      }, "Notification destination deleted");
    }
  });

  els.notificationDestinations.addEventListener("submit", async (event) => {
    const form = event.target.closest("[data-notification-destination-form]");
    if (!form) {
      return;
    }
    event.preventDefault();
    await runAction(async () => {
      await api(`/api/notification-destinations/${form.dataset.notificationDestinationForm}`, {
        method: "PATCH",
        body: { spec: notificationDestinationUpdateSpec(form) }
      });
      await loadNotificationDestinations();
    }, "Notification destination saved");
  });

  els.notificationPolicyScope.addEventListener("change", async () => {
    renderNotificationPolicyProjectOptions();
    await runAction(async () => {
      await loadNotificationDestinations(
        els.notificationPolicyScope.value === "project" ? selectedNotificationPolicyProjectID() : ""
      );
      await loadNotificationPolicies();
    }, "Notification policies refreshed");
  });

  els.notificationPolicyProject.addEventListener("change", async () => {
    const project = state.projects.find((item) => item.id === els.notificationPolicyProject.value);
    if (project) {
      state.selectedProject = project;
    }
    await runAction(async () => {
      await loadNotificationDestinations(project ? project.id : "");
      await loadNotificationPolicies();
    }, "Notification policies refreshed");
  });

  els.notificationPolicyForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const data = formData(form);
    const scopeType = data.scope_type || "global";
    const projectID = data.project_id || selectedNotificationPolicyProjectID();
    if (scopeType === "project" && !projectID) {
      setActionStatus("Choose a project for project policies");
      return;
    }
    await runAction(async () => {
      await api(notificationPolicyCollectionPath(scopeType, projectID), {
        method: "POST",
        body: { spec: notificationPolicySpec(form) }
      });
      form.reset();
      setFormChecked(form, "enabled", true);
      renderNotificationPolicyProjectOptions();
      await loadNotificationPolicies();
    }, "Notification policy created");
  });

  els.notificationPolicies.addEventListener("click", async (event) => {
    const remove = event.target.closest("[data-delete-notification-policy-id]");
    if (remove) {
      if (!window.confirm("Delete this notification policy?")) {
        return;
      }
      await runAction(async () => {
        await api(`/api/notification-policies/${remove.dataset.deleteNotificationPolicyId}`, { method: "DELETE" });
        await loadNotificationPolicies();
      }, "Notification policy deleted");
    }
  });

  els.notificationPolicies.addEventListener("submit", async (event) => {
    const form = event.target.closest("[data-notification-policy-form]");
    if (!form) {
      return;
    }
    event.preventDefault();
    await runAction(async () => {
      await api(`/api/notification-policies/${form.dataset.notificationPolicyForm}`, {
        method: "PATCH",
        body: { spec: notificationPolicyUpdateSpec(form) }
      });
      await loadNotificationPolicies();
    }, "Notification policy saved");
  });

  els.notificationHookEngineType.addEventListener("change", () => {
    renderNotificationHookEngineFields();
  });

  els.notificationHookScope.addEventListener("change", async () => {
    renderNotificationHookProjectOptions();
    await runAction(async () => {
      await loadNotificationDestinations(
        els.notificationHookScope.value === "project" ? selectedNotificationHookProjectID() : ""
      );
      await loadNotificationHooks();
    }, "Notification hooks refreshed");
  });

  els.notificationHookProject.addEventListener("change", async () => {
    const project = state.projects.find((item) => item.id === els.notificationHookProject.value);
    if (project) {
      state.selectedProject = project;
    }
    await runAction(async () => {
      await loadNotificationDestinations(project ? project.id : "");
      await loadNotificationHooks();
    }, "Notification hooks refreshed");
  });

  els.notificationHookPreviewForm.addEventListener("change", (event) => {
    if (event.target.matches("select[name='policy_id']")) {
      applyNotificationHookPreviewPolicy(event.target.value);
    }
  });

  els.notificationHookForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const data = formData(form);
    const scopeType = data.scope_type || "global";
    const projectID = data.project_id || selectedNotificationHookProjectID();
    if (scopeType === "project" && !projectID) {
      setActionStatus("Choose a project for project notification hooks");
      return;
    }
    await runAction(async () => {
      await api(notificationHookCollectionPath(scopeType, projectID), {
        method: "POST",
        body: { spec: notificationHookSpec(form) }
      });
      form.reset();
      setFormChecked(form, "enabled", true);
      renderNotificationHookProjectOptions();
      renderNotificationHookEngineFields();
      await loadNotificationHooks();
    }, "Notification hook created");
  });

  els.notificationHooks.addEventListener("click", async (event) => {
    const preview = event.target.closest("[data-preview-notification-hook-id]");
    if (preview) {
      await runAction(async () => {
        state.notificationHookPreview = await api(`/api/notification-hooks/${preview.dataset.previewNotificationHookId}/preview`, {
          method: "POST",
          body: { spec: notificationHookPreviewSpec() }
        });
        renderNotificationHookPreview();
      }, "Notification hook previewed");
      return;
    }

    const runs = event.target.closest("[data-load-notification-hook-runs-id]");
    if (runs) {
      await runAction(async () => {
        await loadNotificationHookRuns(runs.dataset.loadNotificationHookRunsId);
      }, "Notification hook runs loaded");
      return;
    }

    const toggle = event.target.closest("[data-toggle-notification-hook-id]");
    if (toggle) {
      await runAction(async () => {
        await api(`/api/notification-hooks/${toggle.dataset.toggleNotificationHookId}`, {
          method: "PATCH",
          body: { spec: { enabled: toggle.dataset.notificationHookEnabled === "true" } }
        });
        await loadNotificationHooks();
      }, "Notification hook updated");
      return;
    }

    const remove = event.target.closest("[data-delete-notification-hook-id]");
    if (remove) {
      if (!window.confirm("Delete this notification hook?")) {
        return;
      }
      await runAction(async () => {
        await api(`/api/notification-hooks/${remove.dataset.deleteNotificationHookId}`, { method: "DELETE" });
        delete state.notificationHookRuns[remove.dataset.deleteNotificationHookId];
        if (state.notificationHookPreview && state.notificationHookPreview.metadata && state.notificationHookPreview.metadata.hook_id === remove.dataset.deleteNotificationHookId) {
          state.notificationHookPreview = null;
        }
        await loadNotificationHooks();
      }, "Notification hook deleted");
    }
  });

  document.addEventListener("click", async (event) => {
    if (!event.target.closest("#ticket-columns, #issue-detail")) {
      return;
    }
    const updateLabels = event.target.closest("[data-update-labels-id]");
    if (updateLabels) {
      const control = updateLabels.closest("[data-ticket-label-control]");
      const input = control ? control.querySelector("input[name='labels']") : null;
      await runAction(async () => {
        await api(`/api/tickets/${updateLabels.dataset.updateLabelsId}`, {
          method: "PATCH",
          body: { spec: { labels: parseLabels(input ? input.value : "") } }
        });
        await loadProjectLabels({ renderTickets: false });
        await refreshTicketViews(updateLabels.dataset.updateLabelsId);
      }, "Ticket labels updated");
      return;
    }

    const updateCustomFields = event.target.closest("[data-update-custom-fields-id]");
    if (updateCustomFields) {
      const control = updateCustomFields.closest("[data-ticket-custom-field-control]");
      await runAction(async () => {
        await api(`/api/tickets/${updateCustomFields.dataset.updateCustomFieldsId}`, {
          method: "PATCH",
          body: { spec: { custom_fields: customFieldsFromControls(control) } }
        });
        await refreshTicketViews(updateCustomFields.dataset.updateCustomFieldsId);
      }, "Ticket custom fields updated");
      return;
    }

    const assignPlanning = event.target.closest("[data-assign-planning-id]");
    if (assignPlanning) {
      const control = assignPlanning.closest("[data-ticket-planning-control]");
      const component = control ? control.querySelector("[data-ticket-component-select]") : null;
      const version = control ? control.querySelector("[data-ticket-version-select]") : null;
      await runAction(async () => {
        await api(`/api/tickets/${assignPlanning.dataset.assignPlanningId}`, {
          method: "PATCH",
          body: {
            spec: {
              component_id: component ? component.value : "",
              version_id: version ? version.value : ""
            }
          }
        });
        await refreshTicketViews(assignPlanning.dataset.assignPlanningId);
        await refreshSelectedVersionReport();
      }, "Ticket planning fields updated");
      return;
    }

    const updateStoryPoints = event.target.closest("[data-update-story-points-id]");
    if (updateStoryPoints) {
      const control = updateStoryPoints.closest("[data-ticket-story-points-control]");
      const input = control ? control.querySelector("input[name='story_points']") : null;
      const spec = {};
      applyStoryPointsSpec(spec, input ? input.value : "");
      await runAction(async () => {
        await api(`/api/tickets/${updateStoryPoints.dataset.updateStoryPointsId}`, {
          method: "PATCH",
          body: { spec }
        });
        await refreshTicketViews(updateStoryPoints.dataset.updateStoryPointsId);
        await refreshSelectedSprintReport();
      }, "Ticket story points updated");
      return;
    }

    const assignSprint = event.target.closest("[data-assign-sprint-id]");
    if (assignSprint) {
      const control = assignSprint.closest("[data-ticket-sprint-control]");
      const select = control ? control.querySelector("select") : null;
      const sprintID = select ? select.value : "";
      if (!sprintID) {
        setNotice("Choose a sprint first");
        return;
      }
      await runAction(async () => {
        await api(`/api/tickets/${assignSprint.dataset.assignSprintId}/sprint`, {
          method: "PUT",
          body: { spec: { sprint_id: sprintID } }
        });
        await refreshTicketViews(assignSprint.dataset.assignSprintId, { roadmap: false });
        await refreshSelectedSprintReport();
      }, "Ticket assigned to sprint");
      return;
    }

    const removeSprint = event.target.closest("[data-remove-sprint-id]");
    if (removeSprint) {
      await runAction(async () => {
        await api(`/api/tickets/${removeSprint.dataset.removeSprintId}/sprint`, { method: "DELETE" });
        await refreshTicketViews(removeSprint.dataset.removeSprintId, { roadmap: false });
        await refreshSelectedSprintReport();
      }, "Ticket removed from sprint");
      return;
    }

    const deleteComment = event.target.closest("[data-delete-comment-id]");
    if (deleteComment) {
      await runAction(async () => {
        await api(`/api/comments/${deleteComment.dataset.deleteCommentId}`, { method: "DELETE" });
        await loadComments(deleteComment.dataset.ticketId);
        await loadActivity(deleteComment.dataset.ticketId);
      }, "Comment deleted");
      return;
    }

    const deleteAttachment = event.target.closest("[data-delete-attachment-id]");
    if (deleteAttachment) {
      await runAction(async () => {
        await api(`/api/attachments/${deleteAttachment.dataset.deleteAttachmentId}`, { method: "DELETE" });
        await loadAttachments(deleteAttachment.dataset.ticketId);
        await loadActivity(deleteAttachment.dataset.ticketId);
      }, "Attachment deleted");
      return;
    }

    const deleteLink = event.target.closest("[data-delete-ticket-link-id]");
    if (deleteLink) {
      await runAction(async () => {
        await api(`/api/tickets/${deleteLink.dataset.ticketId}/links/${deleteLink.dataset.deleteTicketLinkId}`, { method: "DELETE" });
        await loadTicketLinks(deleteLink.dataset.ticketId);
        await loadRoadmapDependencies();
        await loadActivity(deleteLink.dataset.ticketId);
      }, "Ticket link removed");
      return;
    }

    const deleteTicket = event.target.closest("[data-delete-ticket-id]");
    if (deleteTicket) {
      const ticketID = deleteTicket.dataset.deleteTicketId;
      const projectID = deleteTicket.dataset.projectId || (state.selectedProject ? state.selectedProject.id : "");
      if (!window.confirm("Delete this ticket?")) {
        return;
      }
      await runAction(async () => {
        await api(`/api/tickets/${ticketID}`, { method: "DELETE" });
        delete state.attachments[ticketID];
        delete state.comments[ticketID];
        delete state.ticketLinks[ticketID];
        delete state.ticketWatchers[ticketID];
        delete state.activities[ticketID];
        await loadProjectLabels({ renderTickets: false });
        await loadRoadmap({ renderTickets: false });
        await loadRoadmapDependencies();
        await loadBacklog();
        if (state.selectedBoardID) {
          await loadBoardTickets(state.selectedBoardID, { renderAfter: false });
        }
        await loadTickets();
        await refreshSelectedSprintReport();
        await refreshSelectedVersionReport();
        if (state.selectedIssue && state.selectedIssue.id === ticketID) {
          state.selectedIssue = null;
          await navigate(`/projects/${encodeURIComponent(projectID)}`);
        }
      }, "Ticket deleted");
      return;
    }

    const watch = event.target.closest("[data-watch-ticket-id]");
    if (watch) {
      const ticketID = watch.dataset.watchTicketId;
      await runAction(async () => {
        await api(`/api/tickets/${ticketID}/watchers/me`, { method: "PUT" });
        await refreshTicketViews(ticketID, { roadmap: false });
      }, "Watching ticket");
      return;
    }

    const unwatch = event.target.closest("[data-unwatch-ticket-id]");
    if (unwatch) {
      const ticketID = unwatch.dataset.unwatchTicketId;
      await runAction(async () => {
        await api(`/api/tickets/${ticketID}/watchers/me`, { method: "DELETE" });
        await refreshTicketViews(ticketID, { roadmap: false });
      }, "Stopped watching ticket");
      return;
    }

    const button = event.target.closest("[data-ticket-status]");
    if (!button) {
      return;
    }
    await runAction(async () => {
      await api(`/api/tickets/${button.dataset.ticketId}`, {
        method: "PATCH",
        body: { spec: { status: button.dataset.ticketStatus } }
      });
      await refreshTicketViews(button.dataset.ticketId);
      await refreshSelectedVersionReport();
    }, "Ticket updated");
  });

  document.addEventListener("submit", async (event) => {
    const createPageForm = event.target.closest("#create-page-submit-form");
    if (createPageForm) {
      event.preventDefault();
      const route = currentRoute();
      if (route.page !== "create-page" || !state.selectedCreatePageSchema) {
        setNotice("Create page is not loaded");
        return;
      }
      const ticket = createPageTicketSpec(createPageForm, state.selectedCreatePageSchema);
      await runAction(async () => {
        const created = normalizeTicket(await api(`/api/projects/${route.projectID}/ticket-create-pages/${encodeURIComponent(route.slug)}/submit`, {
          method: "POST",
          body: { spec: { ticket } }
        }));
        state.createPageSubmission = created;
        if (state.selectedProject && state.selectedProject.id === route.projectID) {
          await loadBacklog();
          await loadTickets();
        }
        renderCreatePageView();
      }, "Ticket created");
      return;
    }

    if (!event.target.closest("#ticket-columns, #issue-detail")) {
      return;
    }
    const commentForm = event.target.closest("[data-comment-form]");
    if (commentForm) {
      event.preventDefault();
      const ticketID = commentForm.dataset.ticketId;
      const textarea = commentForm.querySelector("textarea[name='body']");
      const body = textarea ? textarea.value.trim() : "";
      if (!ticketID || !body) {
        setNotice("Write a comment first");
        return;
      }
      await runAction(async () => {
        await api(`/api/tickets/${ticketID}/comments`, { method: "POST", body: { spec: { body } } });
        commentForm.reset();
        await loadComments(ticketID);
        await loadActivity(ticketID);
      }, "Comment added");
      return;
    }

    const linkForm = event.target.closest("[data-ticket-link-form]");
    if (linkForm) {
      event.preventDefault();
      const ticketID = linkForm.dataset.ticketId;
      const data = formData(linkForm);
      if (!ticketID || !data.target_ticket_id) {
        setNotice("Choose a linked issue first");
        return;
      }
      await runAction(async () => {
        await api(`/api/tickets/${ticketID}/links`, {
          method: "POST",
          body: {
            spec: {
              target_ticket_id: data.target_ticket_id || "",
              link_type: data.link_type || "relates_to"
            }
          }
        });
        linkForm.reset();
        await loadTicketLinks(ticketID);
        await loadRoadmapDependencies();
        await loadActivity(ticketID);
      }, "Ticket link added");
      return;
    }

    const form = event.target.closest("[data-attachment-form]");
    if (!form) {
      return;
    }
    event.preventDefault();
    const ticketID = form.dataset.ticketId;
    const fileInput = form.querySelector('input[type="file"]');
    if (!ticketID || !fileInput || !fileInput.files.length) {
      setNotice("Choose a file to attach");
      return;
    }
    await runAction(async () => {
      const body = new FormData();
      body.append("file", fileInput.files[0]);
      await api(`/api/tickets/${ticketID}/attachments`, { method: "POST", body });
      form.reset();
      await loadAttachments(ticketID);
      await loadActivity(ticketID);
    }, "Attachment uploaded");
  });
}

async function navigate(path) {
  if (window.location.pathname !== path) {
    window.history.pushState({}, "", path);
  }
  await handleRouteChange();
}

async function handleRouteChange() {
  if (!state.user) {
    render();
    return;
  }
  await loadRouteData();
  render();
}

function currentRoute() {
  const path = window.location.pathname;
  if (path === "/profile") {
    return { page: "profile" };
  }
  if (path === "/rbac" || path === "/admin/rbac") {
    return { page: "rbac" };
  }
  if (path === "/settings") {
    return { page: "settings" };
  }
  if (path === "/search") {
    return { page: "search" };
  }
  if (path === "/automation") {
    return { page: "automation" };
  }
  if (path === "/projects") {
    return { page: "projects" };
  }
  const createPageMatch = path.match(/^\/projects\/([^/]+)\/create\/([^/]+)$/);
  if (createPageMatch) {
    return {
      page: "create-page",
      projectID: decodeURIComponent(createPageMatch[1]),
      slug: decodeURIComponent(createPageMatch[2])
    };
  }
  if (path.startsWith("/projects/")) {
    return { page: "projects", projectID: decodeURIComponent(path.slice("/projects/".length)) };
  }
  if (path.startsWith("/issues/")) {
    return { page: "issue", ticketID: decodeURIComponent(path.slice("/issues/".length)) };
  }
  return { page: "dashboard" };
}

async function loadRouteData() {
  const route = currentRoute();
  if (route.page === "projects") {
    if (route.projectID) {
      state.selectedProject = state.projects.find((project) => project.id === route.projectID) || state.selectedProject;
    } else if (!state.selectedProject && state.projects.length) {
      state.selectedProject = state.projects[0];
    }
    if (state.selectedProject) {
      await loadProjectDetails();
    }
    return;
  }
  if (route.page === "create-page" && route.projectID && route.slug) {
    state.selectedProject = state.projects.find((project) => project.id === route.projectID) || state.selectedProject;
    await loadCreatePageForRoute(route.projectID, route.slug);
    return;
  }
  if (route.page === "issue" && route.ticketID) {
    await loadSelectedIssue(route.ticketID);
    return;
  }
  if (route.page === "rbac") {
    await loadRBAC();
    return;
  }
  if (route.page === "settings") {
    await loadSettingsPage();
    return;
  }
  if (route.page === "search") {
    if (state.selectedProject) {
      await loadCustomFields({ renderTickets: false });
    }
    await loadSavedViews();
    return;
  }
  if (route.page === "automation") {
    if (!state.selectedProject && state.projects.length) {
      state.selectedProject = state.projects[0];
    }
    await Promise.all([loadCronJobs(), loadWebhooks(), loadTicketHooks(), loadCreatePages()]);
  }
}

async function loadCreatePageForRoute(projectID, slug) {
  state.selectedCreatePageSchema = normalizeCreatePageSchema(
    await api(`/api/projects/${projectID}/ticket-create-pages/${encodeURIComponent(slug)}/schema`)
  );
  state.createPageSubmission = null;
}

async function refreshSession() {
  try {
    const data = await api("/api/me");
    state.user = {
      id: data.metadata.user_id,
      username: data.spec.username,
      display_name: data.spec.display_name
    };
    await loadProjects();
    await loadNotifications();
    await loadTokens();
    await loadRouteData();
  } catch (error) {
    state.user = null;
    render();
  }
}

async function loadNotifications() {
  if (!state.user) {
    state.notifications = [];
    renderNotificationBadge();
    renderDashboard();
    renderNotifications();
    return;
  }
  const query = state.unreadNotificationsOnly ? "?unread=true&limit=20" : "?limit=20";
  const data = await api(`/api/notifications${query}`);
  state.notifications = listItems(data).map(normalizeNotification);
  renderNotificationBadge();
  renderDashboard();
  renderNotifications();
}

async function loadTokens() {
  if (!state.user) {
    state.tokens = [];
    state.createdToken = null;
    renderTokens();
    return;
  }
  const data = await api("/api/tokens");
  state.tokens = listItems(data).map(normalizeToken);
  renderTokens();
}

async function loadSavedViews() {
  if (!state.user) {
    state.savedViews = [];
    resetSavedViewPagination();
    renderSavedViews();
    return;
  }
  const projectPart = state.selectedProject ? `project_id=${encodeURIComponent(state.selectedProject.id)}&` : "";
  const data = await api(`/api/saved-views?${projectPart}limit=${savedViewPageSize + 1}&offset=${state.savedViewOffset}`);
  const items = listItems(data).map(normalizeSavedView).filter(Boolean);
  state.savedViewHasMore = items.length > savedViewPageSize;
  state.savedViews = items.slice(0, savedViewPageSize);
  if (!state.savedViews.length && state.savedViewOffset > 0) {
    state.savedViewOffset = Math.max(0, state.savedViewOffset - savedViewPageSize);
    await loadSavedViews();
    return;
  }
  renderSavedViews();
}

async function loadBoardSavedViews() {
  if (!state.user || !state.selectedProject) {
    state.boardSavedViews = [];
    state.selectedBoardSavedViewID = "";
    state.boardSavedViewsError = "";
    renderBoardSavedViewFilter();
    return;
  }
  try {
    const projectID = encodeURIComponent(state.selectedProject.id);
    const data = await api(`/api/saved-views?project_id=${projectID}&limit=200&offset=0`);
    state.boardSavedViews = listItems(data)
      .map(normalizeSavedView)
      .filter(isApplicableBoardSavedView)
      .sort(compareBoardSavedViews);
    if (state.selectedBoardSavedViewID && !state.boardSavedViews.some((view) => view.id === state.selectedBoardSavedViewID)) {
      state.selectedBoardSavedViewID = "";
    }
    state.boardSavedViewsError = "";
  } catch (error) {
    state.boardSavedViews = [];
    state.selectedBoardSavedViewID = "";
    state.boardSavedViewsError = error.message || "Unable to load board filters";
  }
  renderBoardSavedViewFilter();
}

async function loadPinnedProjectSavedViews() {
  if (!state.user || !state.selectedProject) {
    state.pinnedProjectSavedViews = [];
    state.pinnedProjectSavedViewsLoading = false;
    state.pinnedProjectSavedViewsError = "";
    renderPinnedProjectSavedViews();
    return;
  }
  state.pinnedProjectSavedViewsLoading = true;
  state.pinnedProjectSavedViewsError = "";
  renderPinnedProjectSavedViews();
  try {
    const projectID = encodeURIComponent(state.selectedProject.id);
    const data = await api(`/api/saved-views?project_id=${projectID}&pinned=true&limit=20&offset=0`);
    state.pinnedProjectSavedViews = listItems(data).map(normalizeSavedView).filter((view) =>
      view && view.pinned && view.scope_type === "project" && view.project_id === state.selectedProject.id
    );
  } catch (error) {
    state.pinnedProjectSavedViews = [];
    state.pinnedProjectSavedViewsError = error.message || "Unable to load pinned views";
  } finally {
    state.pinnedProjectSavedViewsLoading = false;
    renderPinnedProjectSavedViews();
  }
}

async function loadSprints(options = {}) {
  if (!state.user || !state.selectedProject) {
    state.sprints = [];
    state.selectedSprintReportID = "";
    state.sprintReport = null;
    renderSprints();
    renderSprintReport();
    if (options.renderTickets !== false) {
      renderTickets();
    }
    return;
  }
  const query = state.sprintFilterState ? `?state=${encodeURIComponent(state.sprintFilterState)}` : "";
  const data = await api(`/api/projects/${state.selectedProject.id}/sprints${query}`);
  state.sprints = listItems(data).map(normalizeSprint);
  if (state.selectedSprintReportID && !state.sprints.some((sprint) => sprint.id === state.selectedSprintReportID)) {
    state.selectedSprintReportID = "";
    state.sprintReport = null;
  }
  renderSprints();
  renderSprintReport();
  if (options.renderTickets !== false) {
    renderTickets();
  }
}

async function loadSprintReport(sprintID) {
  if (!state.user || !state.selectedProject || !sprintID) {
    state.sprintReport = null;
    renderSprintReport();
    return;
  }
  const data = await api(`/api/sprints/${sprintID}/report`);
  state.sprintReport = normalizeSprintReport(data);
  renderSprints();
  renderSprintReport();
}

async function refreshSelectedSprintReport() {
  if (state.selectedSprintReportID) {
    await loadSprintReport(state.selectedSprintReportID);
  }
}

async function loadBacklog() {
  if (!state.user || !state.selectedProject) {
    state.backlog = [];
    renderBacklog();
    return;
  }
  const data = await api(`/api/projects/${state.selectedProject.id}/backlog`);
  state.backlog = listItems(data).map(normalizeTicket);
  renderBacklog();
}

async function reorderBacklogTicket(ticketID, targetTicketID) {
  if (ticketID === targetTicketID) {
    return;
  }
  const sourceIndex = state.backlog.findIndex((ticket) => ticket.id === ticketID);
  if (sourceIndex < 0) {
    return;
  }
  const reordered = state.backlog.slice();
  const [ticket] = reordered.splice(sourceIndex, 1);
  let targetIndex = targetTicketID ? reordered.findIndex((item) => item.id === targetTicketID) : reordered.length;
  if (targetIndex < 0) {
    targetIndex = reordered.length;
  }
  reordered.splice(targetIndex, 0, ticket);
  const data = await api(`/api/projects/${state.selectedProject.id}/backlog`, {
    method: "PATCH",
    body: { spec: { ticket_ids: reordered.map((item) => item.id) } }
  });
  state.backlog = listItems(data).map(normalizeTicket);
  renderBacklog();
}

async function loadWorkflowStatuses() {
  if (!state.user || !state.selectedProject) {
    state.workflowStatuses = [];
    renderWorkflowPanel();
    renderTickets();
    return;
  }
  const data = await api(`/api/projects/${state.selectedProject.id}/statuses`);
  state.workflowStatuses = listItems(data).map(normalizeWorkflowStatus).filter(Boolean);
  renderWorkflowPanel();
  renderTickets();
}

async function loadBoards() {
  if (!state.user || !state.selectedProject) {
    state.boards = [];
    state.selectedBoardID = "";
    state.boardTickets = null;
    state.selectedBoardSavedViewID = "";
    renderWorkflowPanel();
    renderTickets();
    return;
  }
  const data = await api(`/api/projects/${state.selectedProject.id}/boards`);
  state.boards = listItems(data).map(normalizeBoard).filter(Boolean);
  if (!state.boards.some((board) => board.id === state.selectedBoardID)) {
    state.selectedBoardID = state.boards.length ? state.boards[0].id : "";
  }
  if (state.selectedBoardID) {
    await loadBoardTickets(state.selectedBoardID, { renderAfter: false });
  } else {
    state.boardTickets = null;
  }
  renderWorkflowPanel();
  renderTickets();
}

async function loadBoardTickets(boardID, options = {}) {
  if (!boardID) {
    state.boardTickets = null;
    if (options.renderAfter !== false) {
      renderTickets();
    }
    return;
  }
  const data = await api(`/api/boards/${boardID}/tickets`);
  state.boardTickets = normalizeBoardTickets(data);
  if (state.selectedBoardSavedViewID) {
    await applyBoardSavedViewFilter();
  }
  if (options.renderAfter !== false) {
    renderTickets();
  }
}

async function applyBoardSavedViewFilter() {
  if (!state.selectedProject || !state.boardTickets) {
    return;
  }
  const view = selectedBoardSavedView();
  if (!view) {
    state.selectedBoardSavedViewID = "";
    return;
  }
  const query = view.query || {};
  const spec = {
    project_id: state.selectedProject.id,
    text: query.text || "",
    filter: query.filter || "",
    sort: view.sort && view.sort.length ? view.sort : [{ field: "updated_at", direction: "desc" }],
    limit: 200
  };
  const matches = await searchAllBoardSavedViewTickets(spec);
  state.boardTickets = boardTicketsFromSavedViewMatches(state.boardTickets, matches);
}

async function searchAllBoardSavedViewTickets(spec) {
  const matches = [];
  let cursor = "";
  do {
    const data = await api("/api/search", {
      method: "POST",
      body: { spec: { ...spec, cursor } }
    });
    matches.push(...listItems(data).map(normalizeTicket).filter(Boolean));
    cursor = data && data.status ? data.status.next_cursor || "" : "";
  } while (cursor);
  return matches;
}

async function loadComponents(options = {}) {
  if (!state.user || !state.selectedProject) {
    state.components = [];
    pruneTicketFilters();
    renderComponents();
    renderTicketFormOptions();
    renderTicketFilters();
    if (options.renderTickets !== false) {
      renderTickets();
    }
    return;
  }
  const data = await api(`/api/projects/${state.selectedProject.id}/components`);
  state.components = listItems(data).map(normalizeComponent);
  pruneTicketFilters();
  renderComponents();
  renderTicketFormOptions();
  renderTicketFilters();
  if (options.renderTickets !== false) {
    renderTickets();
  }
}

async function loadVersions(options = {}) {
  if (!state.user || !state.selectedProject) {
    state.versions = [];
    state.selectedVersionReportID = "";
    state.versionReport = null;
    pruneTicketFilters();
    renderVersions();
    renderVersionReport();
    renderTicketFormOptions();
    renderTicketFilters();
    if (options.renderTickets !== false) {
      renderTickets();
    }
    return;
  }
  const data = await api(`/api/projects/${state.selectedProject.id}/versions`);
  state.versions = listItems(data).map(normalizeVersion);
  if (state.selectedVersionReportID && !state.versions.some((version) => version.id === state.selectedVersionReportID)) {
    state.selectedVersionReportID = "";
    state.versionReport = null;
  }
  pruneTicketFilters();
  renderVersions();
  renderVersionReport();
  renderTicketFormOptions();
  renderTicketFilters();
  if (options.renderTickets !== false) {
    renderTickets();
  }
}

async function loadVersionReport(versionID) {
  if (!versionID) {
    state.versionReport = null;
    renderVersionReport();
    return;
  }
  const data = await api(`/api/versions/${versionID}/report`);
  state.versionReport = normalizeVersionReport(data);
  renderVersions();
  renderVersionReport();
}

async function refreshSelectedVersionReport() {
  if (state.selectedVersionReportID) {
    await loadVersionReport(state.selectedVersionReportID);
  }
}

async function loadCustomFields(options = {}) {
  if (!state.user || !state.selectedProject) {
    state.customFields = [];
    renderCustomFields();
    if (options.renderTickets !== false) {
      renderTickets();
    }
    return;
  }
  const data = await api(`/api/projects/${state.selectedProject.id}/custom-fields`);
  state.customFields = listItems(data).map(normalizeCustomField);
  renderCustomFields();
  renderTicketCreateCustomFields();
  if (options.renderTickets !== false) {
    renderTickets();
  }
}

async function loadRoadmap(options = {}) {
  if (!state.user || !state.selectedProject) {
    state.roadmap = [];
    state.roadmapCapacityTickets = [];
    state.roadmapDependencies = [];
    renderRoadmap();
    renderRoadmapDependencies();
    renderTicketFormOptions();
    if (options.renderTickets !== false) {
      renderTickets();
    }
    return;
  }
  const data = await api(`/api/projects/${state.selectedProject.id}/roadmap`);
  state.roadmap = listItems(data).map(normalizeRoadmapItem);
  await loadRoadmapCapacityTickets();
  renderRoadmap();
  renderTicketFormOptions();
  if (options.renderTickets !== false) {
    renderTickets();
  }
}

async function loadRoadmapCapacityTickets() {
  if (!state.selectedProject) {
    state.roadmapCapacityTickets = [];
    return;
  }
  try {
    const tickets = [];
    const limit = 100;
    let offset = 0;
    while (true) {
      const data = await api(`/api/projects/${state.selectedProject.id}/tickets?limit=${limit}&offset=${offset}`);
      const page = listItems(data).map(normalizeTicket).filter(Boolean);
      tickets.push(...page);
      if (page.length < limit) {
        break;
      }
      offset += limit;
    }
    state.roadmapCapacityTickets = tickets;
  } catch (_) {
    state.roadmapCapacityTickets = [];
  }
}

async function loadRoadmapDependencies() {
  if (!state.user || !state.selectedProject) {
    state.roadmapDependencies = [];
    renderRoadmapDependencies();
    return;
  }
  const data = await api(`/api/projects/${state.selectedProject.id}/roadmap/dependencies`);
  state.roadmapDependencies = listItems(data).map(normalizeRoadmapDependency).filter(Boolean);
  renderRoadmapDependencies();
}

async function refreshRoadmapDependencyViews(sourceTicketID, targetTicketID) {
  await loadRoadmapDependencies();
  for (const ticketID of [sourceTicketID, targetTicketID].filter(Boolean)) {
    await loadTicketLinks(ticketID, { renderAfter: false });
    await loadActivity(ticketID, { renderAfter: false });
  }
  renderTickets();
  if (state.selectedIssue && (state.selectedIssue.id === sourceTicketID || state.selectedIssue.id === targetTicketID)) {
    renderIssue();
  }
}

async function scheduleRoadmapItem(form) {
  const data = formData(form);
  await scheduleRoadmapItemSpec(form.dataset.roadmapScheduleForm, data.start_date || "", data.due_date || "");
}

async function scheduleRoadmapItemSpec(ticketID, startDate, dueDate) {
  const response = await api(`/api/projects/${state.selectedProject.id}/roadmap/schedule`, {
    method: "PATCH",
    body: {
      spec: {
        ticket_id: ticketID,
        start_date: startDate || "",
        due_date: dueDate || ""
      }
    }
  });
  state.roadmap = listItems(response).map(normalizeRoadmapItem);
  renderRoadmap();
  renderTicketFormOptions();
}

async function loadProjectLabels(options = {}) {
  if (!state.user || !state.selectedProject) {
    state.projectLabels = [];
    renderTicketFilters();
    renderProjectLabels();
    if (options.renderTickets !== false) {
      renderTickets();
    }
    return;
  }
  const data = await api(`/api/projects/${state.selectedProject.id}/labels`);
  state.projectLabels = listItems(data).map(normalizeProjectLabel);
  pruneTicketFilters();
  renderTicketFilters();
  renderProjectLabels();
  if (options.renderTickets !== false) {
    renderTickets();
  }
}

async function loadProjectDetails() {
  if (!state.selectedProject) {
    return;
  }
  await loadWorkflowStatuses();
  await loadBoardSavedViews();
  await loadBoards();
  await loadBacklog();
  await loadSprints({ renderTickets: false });
  await loadComponents({ renderTickets: false });
  await loadVersions({ renderTickets: false });
  await loadCustomFields({ renderTickets: false });
  await loadRoadmap({ renderTickets: false });
  await loadRoadmapDependencies();
  await loadProjectLabels({ renderTickets: false });
  await loadTickets();
  await loadPinnedProjectSavedViews();
  await loadSavedViews();
}

async function runSearch(spec, options = {}) {
  const reset = options.reset !== false;
  const cursor = reset ? "" : options.cursor || "";
  const normalized = {
    project_id: spec.project_id || (state.selectedProject ? state.selectedProject.id : ""),
    text: spec.text || "",
    filter: spec.filter || "",
    sort: spec.sort && spec.sort.length ? spec.sort : [{ field: "updated_at", direction: "desc" }],
    limit: searchPageSize,
    cursor
  };
  const data = await api("/api/search", { method: "POST", body: { spec: normalized } });
  state.lastSearchSpec = {
    project_id: normalized.project_id,
    text: normalized.text,
    filter: normalized.filter,
    sort: normalized.sort
  };
  if (reset) {
    state.searchCursorStack = [""];
    state.searchCursorIndex = 0;
  }
  state.searchNextCursor = data && data.status ? data.status.next_cursor || "" : "";
  state.searchResults = listItems(data).map(normalizeTicket).filter(Boolean);
  renderSearchResults();
}

async function loadProjects(selectedID = "") {
  const data = await api("/api/projects");
  state.projects = listItems(data).map(normalizeProject);
  const route = currentRoute();
  const previousProjectID = state.selectedProject ? state.selectedProject.id : "";
  if (route.projectID) {
    const nextProject = state.projects.find((project) => project.id === route.projectID) || null;
    if (!state.selectedProject || !nextProject || state.selectedProject.id !== nextProject.id) {
      state.ticketFilters = emptyTicketFilters();
    }
    state.selectedProject = nextProject;
  } else if (selectedID) {
    const nextProject = state.projects.find((project) => project.id === selectedID) || null;
    if (!state.selectedProject || !nextProject || state.selectedProject.id !== nextProject.id) {
      state.ticketFilters = emptyTicketFilters();
    }
    state.selectedProject = nextProject;
  } else if (!state.selectedProject && state.projects.length > 0) {
    state.selectedProject = state.projects[0];
  } else if (state.selectedProject) {
    state.selectedProject = state.projects.find((project) => project.id === state.selectedProject.id) || null;
  }
  const nextProjectID = state.selectedProject ? state.selectedProject.id : "";
  if (previousProjectID !== nextProjectID) {
    resetSavedViewPagination();
    state.selectedBoardSavedViewID = "";
  }
  await loadDashboardSummaries();
  if (state.selectedProject && route.page === "projects") {
    await loadProjectDetails();
  } else {
    state.tickets = [];
    state.backlog = [];
    state.workflowStatuses = [];
    state.boards = [];
    state.selectedBoardID = "";
    state.boardTickets = null;
    state.boardSavedViews = [];
    state.selectedBoardSavedViewID = "";
    state.boardSavedViewsError = "";
    state.ticketFilters = emptyTicketFilters();
    state.sprints = [];
    state.sprintFilterState = "";
    state.selectedSprintReportID = "";
    state.sprintReport = null;
    state.components = [];
    state.versions = [];
    state.projectLabels = [];
    state.customFields = [];
    state.roadmap = [];
    state.roadmapCapacityTickets = [];
    state.roadmapDependencies = [];
    state.attachments = {};
    state.comments = {};
    state.ticketLinks = {};
    state.activities = {};
    state.searchResults = [];
    state.savedViews = [];
    state.pinnedProjectSavedViews = [];
    state.pinnedProjectSavedViewsLoading = false;
    state.pinnedProjectSavedViewsError = "";
  }
  render();
}

async function loadDashboardSummaries() {
  const summaries = await Promise.all(state.projects.map(async (project) => {
    try {
      const [ticketData, sprintData] = await Promise.all([
        api(`/api/projects/${project.id}/tickets?limit=50`),
        api(`/api/projects/${project.id}/sprints`)
      ]);
      const tickets = listItems(ticketData).map(normalizeTicket);
      const sprints = listItems(sprintData).map(normalizeSprint);
      return {
        project,
        tickets,
        sprints,
        count: tickets.length,
        open: tickets.filter((ticket) => ticket.status !== "done").length,
        done: tickets.filter((ticket) => ticket.status === "done").length
      };
    } catch (_) {
      return { project, tickets: [], sprints: [], count: 0, open: 0, done: 0 };
    }
  }));
  state.projectSummaries = summaries;
  state.recentTickets = summaries
    .flatMap((summary) => summary.tickets)
    .sort((a, b) => String(b.updated_at || "").localeCompare(String(a.updated_at || "")))
    .slice(0, 3);
  state.activeSprints = summaries
    .flatMap((summary) => summary.sprints.map((sprint) => ({ ...sprint, project: summary.project })))
    .filter((sprint) => sprint.state === "active")
    .slice(0, 3);
}

async function loadTickets() {
  if (!state.selectedProject) {
    state.tickets = [];
    render();
    return;
  }
  const params = ticketFilterParams();
  const suffix = params.toString() ? `?${params.toString()}` : "";
  const data = await api(`/api/projects/${state.selectedProject.id}/tickets${suffix}`);
  state.tickets = listItems(data).map(normalizeTicket);
  state.attachments = {};
  state.comments = {};
  state.ticketLinks = {};
  state.ticketWatchers = {};
  state.activities = {};
  await Promise.all(state.tickets.flatMap((ticket) => [
    loadAttachments(ticket.id, { renderAfter: false }),
    loadComments(ticket.id, { renderAfter: false }),
    loadTicketLinks(ticket.id, { renderAfter: false }),
    loadTicketWatchers(ticket.id, { renderAfter: false })
  ]));
  render();
}

async function loadSelectedIssue(ticketID) {
  const ticket = normalizeTicket(await api(`/api/tickets/${ticketID}`));
  state.selectedIssue = ticket;
  const previousProjectID = state.selectedProject ? state.selectedProject.id : "";
  state.selectedProject = state.projects.find((project) => project.id === ticket.project_id) || state.selectedProject;
  if (previousProjectID !== (state.selectedProject ? state.selectedProject.id : "")) {
    resetSavedViewPagination();
  }
  if (state.selectedProject) {
    await Promise.all([
      loadSprints({ renderTickets: false }),
      loadComponents({ renderTickets: false }),
      loadVersions({ renderTickets: false }),
      loadCustomFields({ renderTickets: false })
    ]);
    await loadRoadmapDependencies();
    if (!state.tickets.length || state.tickets.some((item) => item.project_id !== ticket.project_id)) {
      const ticketsData = await api(`/api/projects/${state.selectedProject.id}/tickets`);
      state.tickets = listItems(ticketsData).map(normalizeTicket).filter(Boolean);
    }
  }
  await Promise.all([
    loadAttachments(ticket.id, { renderAfter: false }),
    loadComments(ticket.id, { renderAfter: false }),
    loadTicketLinks(ticket.id, { renderAfter: false }),
    loadTicketWatchers(ticket.id, { renderAfter: false }),
    loadActivity(ticket.id, { renderAfter: false })
  ]);
}

async function refreshTicketViews(ticketID, options = {}) {
  if (options.roadmap !== false) {
    await loadRoadmap({ renderTickets: false });
    await loadRoadmapDependencies();
  }
  await loadBacklog();
  if (state.selectedBoardID) {
    await loadBoardTickets(state.selectedBoardID, { renderAfter: false });
  }
  await loadTickets();
  if (state.selectedIssue && state.selectedIssue.id === ticketID) {
    await loadSelectedIssue(ticketID);
    renderIssue();
  }
}

async function refreshBacklogSprintViews(ticketID) {
  await refreshTicketViews(ticketID, { roadmap: false });
  await loadSprints({ renderTickets: false });
  await refreshSelectedSprintReport();
}

async function loadAttachments(ticketID, options = {}) {
  const data = await api(`/api/tickets/${ticketID}/attachments`);
  state.attachments[ticketID] = listItems(data).map(normalizeAttachment);
  if (options.renderAfter !== false) {
    renderTickets();
  }
}

async function loadComments(ticketID, options = {}) {
  const data = await api(`/api/tickets/${ticketID}/comments`);
  state.comments[ticketID] = listItems(data).map(normalizeComment);
  if (options.renderAfter !== false) {
    renderTickets();
  }
}

async function loadTicketLinks(ticketID, options = {}) {
  const data = await api(`/api/tickets/${ticketID}/links`);
  state.ticketLinks[ticketID] = listItems(data).map(normalizeTicketLink).filter(Boolean);
  if (options.renderAfter !== false) {
    renderTickets();
    if (state.selectedIssue && state.selectedIssue.id === ticketID) {
      renderIssue();
    }
  }
}

async function loadTicketWatchers(ticketID, options = {}) {
  const data = await api(`/api/tickets/${ticketID}/watchers`);
  state.ticketWatchers[ticketID] = listItems(data).map(normalizeTicketWatcher).filter(Boolean);
  if (options.renderAfter !== false) {
    renderTickets();
    if (state.selectedIssue && state.selectedIssue.id === ticketID) {
      renderIssue();
    }
  }
}

async function loadActivity(ticketID, options = {}) {
  const data = await api(`/api/tickets/${ticketID}/activity`);
  state.activities[ticketID] = listItems(data).map(normalizeActivity);
  if (options.renderAfter !== false) {
    renderIssue();
  }
}

async function loadRBAC() {
  const [users, groups, roles, bindings] = await Promise.all([
    api("/api/users").catch(() => null),
    api("/api/groups").catch(() => null),
    api("/api/roles").catch(() => null),
    api("/api/role-bindings").catch(() => null)
  ]);
  state.rbac = {
    users: listItems(users).map(normalizeUser),
    groups: listItems(groups).map(normalizeGroup),
    roles: listItems(roles).map(normalizeRole),
    bindings: listItems(bindings).map(normalizeRoleBinding),
    members: {},
    effectivePermissions: state.rbac.effectivePermissions,
    effectivePermissionsError: state.rbac.effectivePermissionsError
  };
  const memberEntries = await Promise.all(state.rbac.groups.map(async (group) => {
    const members = await api(`/api/groups/${group.id}/members`).catch(() => null);
    return [group.id, listItems(members).map(normalizeUser)];
  }));
  state.rbac.members = Object.fromEntries(memberEntries);
  if (state.rbac.effectivePermissions && state.rbac.users.some((user) => user.id === state.rbac.effectivePermissions.user_id)) {
    await loadRBACEffectivePermissions({ render: false });
  }
  renderRBAC();
}

async function loadRBACEffectivePermissions(options = {}) {
  const form = els.rbacPermissionForm;
  if (!form) {
    return;
  }
  const data = formData(form);
  if (!data.user_id) {
    state.rbac.effectivePermissions = null;
    state.rbac.effectivePermissionsError = "Choose a user";
    renderRBACPermissions();
    return;
  }
  const params = new URLSearchParams();
  const scope = data.scope || "global";
  params.set("scope", scope);
  if (scope === "project") {
    if (!data.project_id) {
      state.rbac.effectivePermissions = null;
      state.rbac.effectivePermissionsError = "Choose a project scope";
      renderRBACPermissions();
      return;
    }
    params.set("project_id", data.project_id);
  }
  try {
    state.rbac.effectivePermissions = normalizeEffectivePermissions(await api(`/api/users/${data.user_id}/effective-permissions?${params.toString()}`));
    state.rbac.effectivePermissionsError = "";
  } catch (error) {
    state.rbac.effectivePermissions = null;
    state.rbac.effectivePermissionsError = error.message || "Effective permissions are not available";
  }
  if (options.render !== false) {
    renderRBACPermissions();
  }
}

async function loadSettingsPage() {
  await Promise.all([
    loadGlobalSettings(),
    loadNotificationPreferences(),
    loadProjectNotificationPreferences(),
    loadNotificationDeliveries(),
    loadAuditLog(),
    loadOpenRouterProviders(),
    loadNotificationDestinations(),
    loadNotificationPolicies(),
    loadNotificationHooks()
  ]);
}

async function loadGlobalSettings() {
  try {
    state.settings = normalizeSettings(await api("/api/settings"));
    state.settingsError = "";
  } catch (error) {
    state.settings = null;
    state.settingsError = error.message || "Global settings are not available";
  }
  renderSettings();
}

async function loadNotificationPreferences() {
  try {
    state.notificationPreferences = normalizePreferences(await api("/api/me/notification-preferences"));
  } catch (_) {
    state.notificationPreferences = null;
  }
  renderSettings();
}

async function loadProjectNotificationPreferences(projectID = selectedProjectPreferenceProjectID()) {
  if (!projectID) {
    state.projectNotificationPreferences = null;
    state.projectNotificationPreferencesError = "Choose a project for notification defaults";
    renderSettings();
    return;
  }
  try {
    state.projectNotificationPreferences = normalizePreferences(await api(`/api/projects/${projectID}/notification-preferences`));
    state.projectNotificationPreferencesError = "";
  } catch (error) {
    state.projectNotificationPreferences = null;
    state.projectNotificationPreferencesError = error.message || "Project notification defaults are not available";
  }
  renderSettings();
}

async function loadNotificationDeliveries(projectID = selectedNotificationDeliveryProjectID()) {
  const deliveries = [];
  const errors = [];
  const query = notificationDeliveryQuery();
  try {
    deliveries.push(...listItems(await api(`/api/notification-deliveries${query}`)).map(normalizeNotificationDelivery));
  } catch (error) {
    errors.push(error.message || "Global delivery history is not available");
  }
  if (projectID) {
    try {
      deliveries.push(...listItems(await api(`/api/projects/${projectID}/notification-deliveries${query}`)).map(normalizeNotificationDelivery));
    } catch (error) {
      errors.push(error.message || "Project delivery history is not available");
    }
  }
  state.notificationDeliveries = deliveries.filter(Boolean);
  state.notificationDeliveriesError = errors.length && !state.notificationDeliveries.length ? errors.join(" / ") : "";
  renderSettings();
}

async function loadAuditLog() {
  try {
    const query = auditQuery();
    state.auditLog = listItems(await api(`/api/audit-log${query}`)).map(normalizeAuditEntry);
    state.auditLogError = "";
  } catch (error) {
    state.auditLog = [];
    state.auditLogError = error.message || "Audit log is not available";
  }
  renderSettings();
}

async function loadOpenRouterProviders() {
  try {
    state.openRouterProviders = listItems(await api("/api/openrouter-providers")).map(normalizeOpenRouterProvider);
    state.openRouterProvidersError = "";
  } catch (error) {
    state.openRouterProviders = [];
    state.openRouterProvidersError = error.message || "OpenRouter providers are not available";
  }
  renderSettings();
}

async function loadNotificationDestinations(projectID = selectedNotificationDestinationProjectID()) {
  const destinations = [];
  const errors = [];
  try {
    destinations.push(...listItems(await api("/api/notification-destinations")).map(normalizeNotificationDestination));
  } catch (error) {
    errors.push(error.message || "Global destinations are not available");
  }
  if (projectID) {
    try {
      destinations.push(...listItems(await api(`/api/projects/${projectID}/notification-destinations`)).map(normalizeNotificationDestination));
    } catch (error) {
      errors.push(error.message || "Project destinations are not available");
    }
  }
  state.notificationDestinations = destinations.filter(Boolean);
  state.notificationDestinationsError = errors.length && !destinations.length ? errors.join(" / ") : "";
  renderSettings();
}

async function loadNotificationPolicies(projectID = selectedNotificationPolicyProjectID()) {
  const policies = [];
  const errors = [];
  try {
    policies.push(...listItems(await api("/api/notification-policies")).map(normalizeNotificationPolicy));
  } catch (error) {
    errors.push(error.message || "Global notification policies are not available");
  }
  if (projectID) {
    try {
      policies.push(...listItems(await api(`/api/projects/${projectID}/notification-policies`)).map(normalizeNotificationPolicy));
    } catch (error) {
      errors.push(error.message || "Project notification policies are not available");
    }
  }
  state.notificationPolicies = policies.filter(Boolean);
  state.notificationPoliciesError = errors.length && !policies.length ? errors.join(" / ") : "";
  renderSettings();
}

async function loadNotificationHooks(projectID = selectedNotificationHookProjectID()) {
  const hooks = [];
  const errors = [];
  try {
    hooks.push(...listItems(await api("/api/notification-hooks")).map(normalizeNotificationHook));
  } catch (error) {
    errors.push(error.message || "Global notification hooks are not available");
  }
  if (projectID) {
    try {
      hooks.push(...listItems(await api(`/api/projects/${projectID}/notification-hooks`)).map(normalizeNotificationHook));
    } catch (error) {
      errors.push(error.message || "Project notification hooks are not available");
    }
  }
  state.notificationHooks = hooks.filter(Boolean);
  state.notificationHooksError = errors.length && !hooks.length ? errors.join(" / ") : "";
  renderSettings();
}

async function loadNotificationHookRuns(hookID) {
  state.notificationHookRuns[hookID] = listItems(await api(`/api/notification-hooks/${hookID}/runs?limit=10`)).map(normalizeNotificationHookRun);
  renderNotificationHooks();
}

async function loadCronJobs(projectID = selectedCronJobProjectID()) {
  const query = new URLSearchParams();
  if (projectID) {
    query.set("project_id", projectID);
  }
  query.set("limit", "100");
  try {
    state.cronJobs = listItems(await api(`/api/cron-jobs?${query.toString()}`)).map(normalizeCronJob);
    state.cronJobsError = "";
  } catch (error) {
    state.cronJobs = [];
    state.cronJobsError = error.message || "Cron jobs are not available";
  }
  renderCronJobs();
}

async function loadCronRuns(jobID) {
  state.cronRuns[jobID] = listItems(await api(`/api/cron-jobs/${jobID}/runs?limit=10`)).map(normalizeCronRun);
  renderCronJobs();
}

async function loadWebhooks(projectID = selectedWebhookProjectID()) {
  if (!projectID) {
    state.webhooks = [];
    state.webhooksError = "Choose a project to manage webhooks";
    renderWebhooks();
    return;
  }
  try {
    state.webhooks = listItems(await api(`/api/projects/${projectID}/webhooks?limit=100`)).map(normalizeWebhook);
    state.webhooksError = "";
  } catch (error) {
    state.webhooks = [];
    state.webhooksError = error.message || "Webhooks are not available";
  }
  renderWebhooks();
}

async function loadWebhookRuns(webhookID) {
  state.webhookRuns[webhookID] = listItems(await api(`/api/webhook-definitions/${webhookID}/runs?limit=10`)).map(normalizeWebhookRun);
  renderWebhooks();
}

async function loadWebhookDeliveries(webhookID) {
  state.webhookDeliveries[webhookID] = listItems(await api(`/api/webhook-definitions/${webhookID}/deliveries?limit=10`)).map(normalizeWebhookDelivery);
  renderWebhooks();
}

async function loadTicketHooks(projectID = selectedTicketHookProjectID()) {
  if (!projectID) {
    state.ticketHooks = [];
    state.ticketHooksError = "Choose a project to manage ticket hooks";
    renderTicketHooks();
    return;
  }
  try {
    state.ticketHooks = listItems(await api(`/api/projects/${projectID}/ticket-hooks?limit=100`)).map(normalizeTicketHook);
    state.ticketHooksError = "";
  } catch (error) {
    state.ticketHooks = [];
    state.ticketHooksError = error.message || "Ticket hooks are not available";
  }
  renderTicketHooks();
}

async function loadTicketHookRuns(hookID) {
  state.ticketHookRuns[hookID] = listItems(await api(`/api/ticket-hooks/${hookID}/runs?limit=10`)).map(normalizeTicketHookRun);
  renderTicketHooks();
}

async function loadCreatePages(projectID = selectedCreatePageProjectID()) {
  if (!projectID) {
    state.createPages = [];
    state.createPagesError = "Choose a project to manage create pages";
    renderCreatePages();
    return;
  }
  try {
    state.createPages = listItems(await api(`/api/projects/${projectID}/ticket-create-pages?include_disabled=true&limit=100`)).map(normalizeCreatePage);
    state.createPagesError = "";
  } catch (error) {
    state.createPages = [];
    state.createPagesError = error.message || "Create pages are not available";
  }
  renderCreatePages();
}

async function loadCreatePageSchema(pageID, projectID, slug) {
  const schema = await api(`/api/projects/${projectID}/ticket-create-pages/${encodeURIComponent(slug)}/schema`);
  state.createPages = state.createPages.map((page) => page.id === pageID ? { ...page, schema } : page);
  renderCreatePages();
}

async function loadCreatePageRuns(pageID) {
  state.createPageRuns[pageID] = listItems(await api(`/api/ticket-create-pages/${pageID}/runs?limit=10`)).map(normalizeCreatePageRun);
  renderCreatePages();
}

async function api(path, options = {}) {
  const request = {
    method: options.method || "GET",
    credentials: "same-origin",
    headers: {
      Accept: "application/json"
    }
  };
  if (options.body !== undefined && options.body !== null) {
    if (options.body instanceof FormData) {
      request.body = options.body;
    } else {
      request.body = JSON.stringify(options.body);
      request.headers["Content-Type"] = "application/json";
    }
  }
  if (mutates(request.method)) {
    const csrf = cookieValue("rayboard_csrf");
    if (csrf) {
      request.headers["X-CSRF-Token"] = csrf;
    }
  }

  const response = await fetch(path, request);
  if (!response.ok) {
    let message = `${request.method} ${path} failed`;
    try {
      const body = await response.json();
      if (body.error && body.error.message) {
        message = body.error.message;
      } else if (body.detail) {
        message = body.detail;
      } else if (body.title) {
        message = body.title;
      }
    } catch (_) {
      message = response.statusText || message;
    }
    throw new Error(message);
  }
  if (response.status === 204) {
    return null;
  }
  return response.json();
}

async function runAction(action, success) {
  try {
    setNotice("");
    await action();
    setNotice(success);
  } catch (error) {
    setNotice(error.message || "Request failed");
  }
}

function render() {
  const signedIn = Boolean(state.user);
  const route = currentRoute();
  els.loginForm.hidden = signedIn;
  els.logoutButton.hidden = !signedIn;
  els.appNav.hidden = !signedIn;
  els.projectCreate.hidden = !signedIn;
  els.dashboardView.hidden = !signedIn || route.page !== "dashboard";
  els.notificationInbox.hidden = !signedIn || route.page !== "dashboard";
  els.sprintPanel.hidden = !signedIn || route.page !== "projects" || !state.selectedProject;
  els.labelPanel.hidden = !signedIn || route.page !== "projects" || !state.selectedProject;
  els.backlogPanel.hidden = !signedIn || route.page !== "projects" || !state.selectedProject;
  els.workflowPanel.hidden = !signedIn || route.page !== "projects" || !state.selectedProject;
  els.releasePanel.hidden = !signedIn || route.page !== "projects" || !state.selectedProject;
  els.fieldPanel.hidden = !signedIn || route.page !== "projects" || !state.selectedProject;
  els.roadmapPanel.hidden = !signedIn || route.page !== "projects" || !state.selectedProject;
  els.searchPanel.hidden = !signedIn || route.page !== "search";
  els.accountPanel.hidden = !signedIn || route.page !== "profile";
  els.rbacPanel.hidden = !signedIn || route.page !== "rbac";
  els.settingsPanel.hidden = !signedIn || route.page !== "settings";
  els.engineWorkbench.hidden = !signedIn || route.page !== "automation";
  els.signedOut.hidden = signedIn;
  els.boardView.hidden = !signedIn || route.page !== "projects";
  els.issueView.hidden = !signedIn || route.page !== "issue";
  els.createPageView.hidden = !signedIn || route.page !== "create-page";
  els.ticketForm.hidden = !signedIn || route.page !== "projects" || !state.selectedProject;
  els.ticketFilterForm.hidden = !signedIn || route.page !== "projects" || !state.selectedProject;
  els.sessionState.textContent = signedIn ? state.user.username : "Signed out";

  renderNavigation(route);
  renderDashboard();
  renderProjects();
  renderPinnedProjectSavedViews();
  renderTickets();
  renderIssue();
  renderCreatePageView();
  renderNotifications();
  renderBacklog();
  renderWorkflowPanel();
  renderSprints();
  renderSprintReport();
  renderProjectLabels();
  renderComponents();
  renderVersions();
  renderVersionReport();
  renderCustomFields();
  renderRoadmap();
  renderRoadmapDependencies();
  renderTicketFormOptions();
  renderTicketCreateCustomFields();
  renderTicketFilters();
  renderCustomFieldSearchControls();
  renderSearchResults();
  renderSavedViews();
  renderTokens();
  renderRBAC();
  renderSettings();
  renderCronJobs();
  renderWebhooks();
  renderTicketHooks();
  renderCreatePages();
  renderEngineFields();
  renderEngineResult();
}

function isDocumentLink(pathname) {
  return pathname === "/docs" || pathname === "/api/docs" || pathname === "/api/docs/redoc";
}

function renderNavigation(route) {
  if (!els.appNav) {
    return;
  }
  els.appNav.querySelectorAll("a[href]").forEach((link) => {
    const target = link.getAttribute("href");
    const active =
      (route.page === "dashboard" && target === "/") ||
      (route.page === "projects" && target === "/projects") ||
      (route.page === "search" && target === "/search") ||
      (route.page === "automation" && target === "/automation") ||
      (route.page === "rbac" && target === "/rbac") ||
      (route.page === "settings" && target === "/settings") ||
      (route.page === "profile" && target === "/profile");
    if (active) {
      link.setAttribute("aria-current", "page");
    } else {
      link.removeAttribute("aria-current");
    }
  });
  renderNotificationBadge();
}

function renderNotificationBadge() {
  if (!els.navUnreadCount) {
    return;
  }
  const unreadCount = unreadNotificationCount();
  els.navUnreadCount.hidden = unreadCount === 0;
  els.navUnreadCount.textContent = unreadCount > 99 ? "99+" : String(unreadCount);
  els.navUnreadCount.setAttribute("aria-label", `${unreadCount} unread notifications`);
}

function renderDashboard() {
  if (!els.dashboardView) {
    return;
  }
  const totals = state.projectSummaries.reduce((acc, summary) => {
    acc.open += summary.open;
    acc.done += summary.done;
    return acc;
  }, { open: 0, done: 0 });
  els.metricProjects.textContent = String(state.projects.length);
  els.metricOpenTickets.textContent = String(totals.open);
  els.metricDoneTickets.textContent = String(totals.done);
  els.metricUnread.textContent = String(unreadNotificationCount());

  renderSummaryList(els.recentTickets, state.recentTickets, (ticket) => ticketSummaryNode(ticket));
  renderSummaryList(
    els.bigProjects,
    [...state.projectSummaries].sort((a, b) => b.count - a.count).slice(0, 3),
    (summary) => projectSummaryNode(summary)
  );
  renderSummaryList(els.dashboardSprints, state.activeSprints, (sprint) => sprintSummaryNode(sprint));
}

function renderSummaryList(container, items, nodeFactory) {
  if (!container) {
    return;
  }
  container.replaceChildren();
  if (!items.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "Nothing to show";
    container.append(empty);
    return;
  }
  for (const item of items) {
    container.append(nodeFactory(item));
  }
}

function ticketSummaryNode(ticket) {
  const row = document.createElement("a");
  row.className = "summary-item";
  row.href = `/issues/${ticket.id}`;
  const title = document.createElement("p");
  title.textContent = `${ticket.key} ${ticket.title}`;
  const meta = document.createElement("span");
  meta.textContent = [projectName(ticket.project_id), ticket.status, ticket.priority].filter(Boolean).join(" / ");
  row.append(title, meta);
  return row;
}

function projectSummaryNode(summary) {
  const row = document.createElement("a");
  row.className = "summary-item";
  row.href = `/projects/${summary.project.id}`;
  const title = document.createElement("p");
  title.textContent = `${summary.project.key} ${summary.project.name}`;
  const meta = document.createElement("span");
  meta.textContent = `${summary.count} tickets / ${summary.open} open / ${summary.done} done`;
  row.append(title, meta);
  return row;
}

function sprintSummaryNode(sprint) {
  const row = document.createElement("a");
  row.className = "summary-item";
  row.href = sprint.project ? `/projects/${sprint.project.id}` : "/projects";
  const title = document.createElement("p");
  title.textContent = sprint.name;
  const meta = document.createElement("span");
  meta.textContent = [sprint.project ? sprint.project.key : "", dateRange(sprint.start_date, sprint.end_date)].filter(Boolean).join(" / ");
  row.append(title, meta);
  return row;
}

function renderNotifications() {
  if (!els.notifications) {
    return;
  }
  const unreadCount = unreadNotificationCount();
  els.notificationCount.textContent = String(unreadCount);
  if (els.notificationUnreadOnly) {
    els.notificationUnreadOnly.checked = state.unreadNotificationsOnly;
  }
  els.notifications.replaceChildren();
  if (!state.notifications.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = state.unreadNotificationsOnly ? "No unread notifications" : "No notifications";
    els.notifications.append(empty);
    return;
  }
  for (const notification of state.notifications) {
    els.notifications.append(notificationNode(notification));
  }
}

function unreadNotificationCount() {
  return state.notifications.filter((notification) => !notification.read_at).length;
}

function notificationNode(notification) {
  const article = document.createElement("article");
  article.className = "notification-item";
  if (!notification.read_at) {
    article.classList.add("is-unread");
  }

  const body = document.createElement("p");
  body.textContent = notification.body || notification.type || "Notification";

  const meta = document.createElement("span");
  meta.textContent = [notificationTypeLabel(notification.type), notification.subject_type, notification.subject_id].filter(Boolean).join(" / ");

  const button = document.createElement("button");
  button.type = "button";
  button.dataset.notificationId = notification.id;
  button.dataset.notificationReadState = notification.read_at ? "unread" : "read";
  button.textContent = notification.read_at ? "Unread" : "Read";

  article.append(body, meta, button);
  return article;
}

function notificationTypeLabel(type) {
  const labels = {
    comment_added: "Comment",
    comment_mentioned: "Mention",
    release_changed: "Release",
    sprint_changed: "Sprint",
    ticket_assigned: "Assignment",
    ticket_status_changed: "Status"
  };
  return labels[type] || type || "";
}

function renderBacklog() {
  if (!els.backlog) {
    return;
  }
  els.backlog.replaceChildren();
  if (!state.selectedProject) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "Select a project to view backlog";
    els.backlog.append(empty);
    return;
  }
  if (!state.backlog.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No backlog tickets";
    els.backlog.append(empty);
    return;
  }
  els.backlog.append(backlogSummaryNode(state.backlog));
  state.backlog.forEach((ticket, index) => {
    els.backlog.append(backlogItemNode(ticket, index));
  });
}

function backlogSummaryNode(tickets) {
  const metrics = backlogSummaryMetrics(tickets);
  const section = document.createElement("section");
  section.className = "backlog-summary";
  section.append(
    backlogSummaryMetricNode("Tickets", metrics.total),
    backlogSummaryMetricNode("Sprint assigned", metrics.assigned),
    backlogSummaryMetricNode("Unassigned", metrics.unassigned),
    backlogSummaryMetricNode("Story points", metrics.story_points)
  );
  section.append(backlogEstimateCoverageNode(metrics.estimate_coverage));
  section.append(backlogStatusBreakdownNode(metrics.statuses));
  section.append(backlogPriorityBreakdownNode(metrics.priorities));
  section.append(backlogAssigneeBreakdownNode(metrics.assignees));
  section.append(backlogSprintWorkloadsNode(metrics.workloads));
  return section;
}

function backlogSummaryMetrics(tickets) {
  const list = Array.isArray(tickets) ? tickets : [];
  let storyPoints = 0;
  let hasStoryPoints = false;
  let estimated = 0;
  for (const ticket of list) {
    if (ticket.story_points === null || ticket.story_points === undefined || ticket.story_points === "") {
      continue;
    }
    const points = Number(ticket.story_points);
    if (Number.isFinite(points)) {
      storyPoints += points;
      hasStoryPoints = true;
      estimated += 1;
    }
  }
  const assigned = list.filter((ticket) => Boolean(ticket.sprint_id)).length;
  return {
    total: list.length,
    assigned,
    unassigned: list.length - assigned,
    story_points: hasStoryPoints ? formatStoryPoints(storyPoints) : "none",
    estimate_coverage: backlogEstimateCoverage(list.length, estimated),
    statuses: backlogStatusBreakdown(list),
    priorities: backlogPriorityBreakdown(list),
    assignees: backlogAssigneeBreakdown(list),
    workloads: backlogSprintWorkloads(list)
  };
}

function backlogEstimateCoverage(total, estimated) {
  const estimatedCount = Number.isFinite(estimated) ? estimated : 0;
  const totalCount = Number.isFinite(total) ? total : 0;
  return [
    { label: "Estimated", count: estimatedCount },
    { label: "Unestimated", count: Math.max(totalCount - estimatedCount, 0) }
  ];
}

function backlogEstimateCoverageNode(coverage) {
  const list = document.createElement("div");
  list.className = "backlog-estimate-coverage";
  if (!coverage.length) {
    return list;
  }
  for (const item of coverage) {
    const chip = document.createElement("span");
    chip.textContent = `${item.label}: ${item.count}`;
    list.append(chip);
  }
  return list;
}

function backlogStatusBreakdown(tickets) {
  const counts = new Map();
  for (const ticket of tickets) {
    const slug = ticket.status || "";
    counts.set(slug, (counts.get(slug) || 0) + 1);
  }
  const configured = (state.workflowStatuses.length ? state.workflowStatuses : defaultWorkflowStatuses())
    .slice()
    .sort((left, right) => Number(left.position || 0) - Number(right.position || 0));
  const known = new Set(configured.map((status) => status.slug));
  const ordered = configured
    .filter((status) => counts.has(status.slug))
    .map((status) => ({ label: status.name || status.slug, count: counts.get(status.slug) }));
  const unknown = Array.from(counts.entries())
    .filter(([slug]) => !known.has(slug))
    .map(([slug, count]) => ({ label: slug || "No status", count }))
    .sort((left, right) => {
      if (right.count !== left.count) {
        return right.count - left.count;
      }
      return left.label.localeCompare(right.label);
    });
  return ordered.concat(unknown);
}

function backlogStatusBreakdownNode(statuses) {
  const list = document.createElement("div");
  list.className = "backlog-status-breakdown";
  if (!statuses.length) {
    return list;
  }
  for (const status of statuses) {
    const item = document.createElement("span");
    item.textContent = `${status.label}: ${status.count}`;
    list.append(item);
  }
  return list;
}

function backlogPriorityBreakdown(tickets) {
  const priorities = new Map();
  for (const ticket of tickets) {
    const label = ticket.priority || "No priority";
    priorities.set(label, (priorities.get(label) || 0) + 1);
  }
  return Array.from(priorities.entries())
    .map(([label, count]) => ({ label, count }))
    .sort((left, right) => {
      if (right.count !== left.count) {
        return right.count - left.count;
      }
      return left.label.localeCompare(right.label);
    });
}

function backlogPriorityBreakdownNode(priorities) {
  const list = document.createElement("div");
  list.className = "backlog-priority-breakdown";
  if (!priorities.length) {
    return list;
  }
  for (const priority of priorities) {
    const item = document.createElement("span");
    item.textContent = `${priority.label}: ${priority.count}`;
    list.append(item);
  }
  return list;
}

function backlogAssigneeBreakdown(tickets) {
  const assignees = new Map();
  for (const ticket of tickets) {
    const key = ticket.assignee_id || "";
    assignees.set(key, (assignees.get(key) || 0) + 1);
  }
  return Array.from(assignees.entries())
    .map(([key, count]) => ({ key, label: key ? `assignee ${key}` : "Unassigned", count }))
    .sort((left, right) => {
      if (!left.key) {
        return 1;
      }
      if (!right.key) {
        return -1;
      }
      if (right.count !== left.count) {
        return right.count - left.count;
      }
      return left.label.localeCompare(right.label);
    });
}

function backlogAssigneeBreakdownNode(assignees) {
  const list = document.createElement("div");
  list.className = "backlog-assignee-breakdown";
  if (!assignees.length) {
    return list;
  }
  for (const assignee of assignees) {
    const item = document.createElement("span");
    item.textContent = `${assignee.label}: ${assignee.count}`;
    list.append(item);
  }
  return list;
}

function backlogSprintWorkloads(tickets) {
  const workloads = new Map();
  const ensureWorkload = (key, label, stateLabel = "") => {
    if (!workloads.has(key)) {
      workloads.set(key, {
        key,
        label,
        state: stateLabel,
        tickets: 0,
        story_points: 0,
        has_story_points: false
      });
    }
    return workloads.get(key);
  };
  for (const sprint of state.sprints) {
    ensureWorkload(sprint.id, sprint.name || sprint.id, sprint.state || "");
  }
  ensureWorkload("", "No sprint", "unassigned");
  for (const ticket of tickets) {
    const key = ticket.sprint_id || "";
    const sprint = state.sprints.find((item) => item.id === key);
    const workload = ensureWorkload(key, sprint ? sprint.name || sprint.id : sprintName(key), sprint ? sprint.state || "" : "unknown");
    workload.tickets += 1;
    if (ticket.story_points === null || ticket.story_points === undefined || ticket.story_points === "") {
      continue;
    }
    const points = Number(ticket.story_points);
    if (Number.isFinite(points)) {
      workload.story_points += points;
      workload.has_story_points = true;
    }
  }
  return Array.from(workloads.values())
    .filter((workload) => workload.tickets > 0)
    .sort((left, right) => {
      if (!left.key) {
        return 1;
      }
      if (!right.key) {
        return -1;
      }
      return 0;
    });
}

function backlogSprintWorkloadsNode(workloads) {
  const list = document.createElement("div");
  list.className = "backlog-sprint-workloads";
  if (!workloads.length) {
    return list;
  }
  for (const workload of workloads) {
    const item = document.createElement("span");
    const points = workload.has_story_points ? `${formatStoryPoints(workload.story_points)} pts` : "no estimates";
    item.textContent = [
      workload.label,
      workload.state,
      `${workload.tickets} ticket${workload.tickets === 1 ? "" : "s"}`,
      points
    ].filter(Boolean).join(" / ");
    list.append(item);
  }
  return list;
}

function backlogSummaryMetricNode(label, value) {
  const item = document.createElement("div");
  item.className = "backlog-summary-metric";
  const strong = document.createElement("strong");
  strong.textContent = String(value);
  const span = document.createElement("span");
  span.textContent = label;
  item.append(strong, span);
  return item;
}

function backlogItemNode(ticket, index) {
  const article = document.createElement("article");
  article.className = "backlog-item";
  article.draggable = true;
  article.dataset.backlogDragId = ticket.id;

  const body = document.createElement("div");
  body.className = "backlog-item-body";

  const title = document.createElement("a");
  title.href = `/issues/${encodeURIComponent(ticket.id)}`;
  title.textContent = `${ticket.key || ticket.id} ${ticket.title || "Untitled"}`;

  const meta = document.createElement("span");
  meta.textContent = [
    ticket.type || "task",
    ticket.status || "todo",
    ticket.priority || "",
    ticket.sprint_id ? sprintName(ticket.sprint_id) : "",
    ticket.assignee_id ? `assignee ${ticket.assignee_id}` : ""
  ].filter(Boolean).join(" / ");

  body.append(title, meta);

  const actions = document.createElement("div");
  actions.className = "backlog-actions";

  const up = document.createElement("button");
  up.type = "button";
  up.dataset.backlogMoveId = ticket.id;
  up.dataset.backlogMoveDirection = "up";
  up.disabled = index === 0;
  up.textContent = "Up";

  const down = document.createElement("button");
  down.type = "button";
  down.dataset.backlogMoveId = ticket.id;
  down.dataset.backlogMoveDirection = "down";
  down.disabled = index === state.backlog.length - 1;
  down.textContent = "Down";

  actions.append(up, down);
  article.append(body, backlogSprintControlNode(ticket), actions);
  return article;
}

function backlogSprintControlNode(ticket) {
  const section = document.createElement("section");
  section.className = "backlog-sprint";
  section.setAttribute("data-backlog-sprint-control", "true");
  section.setAttribute("aria-label", `${ticket.key || ticket.id} sprint planning`);

  const heading = document.createElement("p");
  heading.className = "backlog-sprint-heading";
  heading.textContent = ticket.sprint_id ? `Sprint: ${sprintName(ticket.sprint_id)}` : "No sprint";

  const controls = document.createElement("div");
  controls.className = "backlog-sprint-controls";

  const select = document.createElement("select");
  select.setAttribute("aria-label", "Backlog sprint");
  const empty = document.createElement("option");
  empty.value = "";
  empty.textContent = "Choose sprint";
  select.append(empty);
  for (const sprint of state.sprints) {
    if (sprint.state === "completed" && sprint.id !== ticket.sprint_id) {
      continue;
    }
    const option = document.createElement("option");
    option.value = sprint.id;
    option.textContent = `${sprint.name} (${sprint.state})`;
    option.selected = sprint.id === ticket.sprint_id;
    select.append(option);
  }

  const assign = document.createElement("button");
  assign.type = "button";
  assign.dataset.backlogAssignSprintId = ticket.id;
  assign.textContent = "Assign";
  assign.disabled = !state.sprints.some((sprint) => sprint.state !== "completed" || sprint.id === ticket.sprint_id);

  const remove = document.createElement("button");
  remove.type = "button";
  remove.dataset.backlogRemoveSprintId = ticket.id;
  remove.textContent = "Remove";
  remove.disabled = !ticket.sprint_id;

  controls.append(select, assign, remove);
  section.append(heading, controls);
  return section;
}

function selectedBoardSavedView() {
  if (!state.selectedBoardSavedViewID) {
    return null;
  }
  return state.boardSavedViews.find((view) => view.id === state.selectedBoardSavedViewID) || null;
}

function isApplicableBoardSavedView(view) {
  if (!view) {
    return false;
  }
  return !view.project_id || !state.selectedProject || view.project_id === state.selectedProject.id;
}

function compareBoardSavedViews(left, right) {
  const leftBoard = left.display_mode === "board" ? 0 : 1;
  const rightBoard = right.display_mode === "board" ? 0 : 1;
  if (leftBoard !== rightBoard) {
    return leftBoard - rightBoard;
  }
  const leftPinned = left.pinned ? 0 : 1;
  const rightPinned = right.pinned ? 0 : 1;
  if (leftPinned !== rightPinned) {
    return leftPinned - rightPinned;
  }
  return (left.name || "").localeCompare(right.name || "") || (left.id || "").localeCompare(right.id || "");
}

function boardTicketsFromSavedViewMatches(boardTickets, matches) {
  const ticketsByStatus = new Map();
  for (const ticket of matches) {
    const list = ticketsByStatus.get(ticket.status) || [];
    list.push(ticket);
    ticketsByStatus.set(ticket.status, list);
  }
  return {
    ...boardTickets,
    filtered_by_saved_view: true,
    columns: (boardTickets.columns || []).map((column) => {
      const tickets = ticketsByStatus.get(column.slug) || [];
      return {
        ...column,
        tickets,
        ticket_count: tickets.length,
        over_wip_limit: Number.isFinite(column.wip_limit) && column.wip_limit > 0 && tickets.length > column.wip_limit
      };
    })
  };
}

function renderBoardSavedViewFilter() {
  if (!els.boardSavedViewFilter || !els.boardSavedViewStatus) {
    return;
  }
  els.boardSavedViewFilter.replaceChildren();
  const empty = document.createElement("option");
  empty.value = "";
  empty.textContent = "No saved-view filter";
  els.boardSavedViewFilter.append(empty);
  for (const view of state.boardSavedViews) {
    const option = document.createElement("option");
    option.value = view.id;
    option.textContent = [
      view.name || "Saved view",
      view.display_mode === "board" ? "board" : view.display_mode || "list",
      view.scope_type || ""
    ].filter(Boolean).join(" / ");
    option.selected = view.id === state.selectedBoardSavedViewID;
    els.boardSavedViewFilter.append(option);
  }
  els.boardSavedViewFilter.value = state.selectedBoardSavedViewID || "";
  els.boardSavedViewFilter.disabled = !state.selectedProject || !state.boards.length || !state.boardSavedViews.length;
  if (state.boardSavedViewsError) {
    els.boardSavedViewStatus.textContent = state.boardSavedViewsError;
  } else if (!state.selectedProject) {
    els.boardSavedViewStatus.textContent = "Select a project to use board filters.";
  } else if (!state.boardSavedViews.length) {
    els.boardSavedViewStatus.textContent = "No saved views available for this project.";
  } else if (selectedBoardSavedView()) {
    const view = selectedBoardSavedView();
    els.boardSavedViewStatus.textContent = `Filtering board with ${view.name || "saved view"}.`;
  } else {
    els.boardSavedViewStatus.textContent = "Use a board saved view to filter cards.";
  }
}

function renderWorkflowPanel() {
  if (!els.workflowStatuses || !els.boards) {
    return;
  }
  renderStatusForm();
  renderWorkflowStatuses();
  renderBoardSavedViewFilter();
  renderBoards();
}

function renderStatusForm() {
  if (!els.statusForm || document.activeElement === els.statusForm.elements.statuses) {
    return;
  }
  const statuses = state.workflowStatuses.length ? state.workflowStatuses : defaultWorkflowStatuses();
  els.statusForm.elements.statuses.value = JSON.stringify(statuses.map((status) => ({
    slug: status.slug,
    name: status.name
  })), null, 2);
}

function renderWorkflowStatuses() {
  els.workflowStatuses.replaceChildren();
  if (!state.selectedProject) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "Select a project to manage workflow statuses";
    els.workflowStatuses.append(empty);
    return;
  }
  if (!state.workflowStatuses.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No workflow statuses";
    els.workflowStatuses.append(empty);
    return;
  }
  for (const status of state.workflowStatuses) {
    const item = document.createElement("article");
    item.className = "status-item";

    const title = document.createElement("p");
    title.textContent = status.name || status.slug;

    const meta = document.createElement("span");
    meta.textContent = [status.slug, `position ${status.position || 0}`].filter(Boolean).join(" / ");

    item.append(title, meta);
    els.workflowStatuses.append(item);
  }
}

function renderBoards() {
  els.boards.replaceChildren();
  if (!state.selectedProject) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "Select a project to manage boards";
    els.boards.append(empty);
    return;
  }
  if (!state.boards.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No boards";
    els.boards.append(empty);
    return;
  }
  for (const board of state.boards) {
    els.boards.append(boardNode(board));
  }
}

function boardNode(board) {
  const article = document.createElement("article");
  article.className = "board-item";
  if (board.id === state.selectedBoardID) {
    article.classList.add("is-active");
  }

  const body = document.createElement("div");
  body.className = "board-item-body";

  const title = document.createElement("p");
  title.textContent = board.name || "Board";

  body.append(title, boardMetadataNode(board));
  if (board.id === state.selectedBoardID) {
    body.append(boardColumnSettingsOverviewNode(board), boardStatusCoverageOverviewNode(board));
  }

  const edit = document.createElement("form");
  edit.className = "board-edit-form";
  edit.dataset.boardEditForm = board.id;
  edit.append(
    inputNode("name", board.name || "", "name"),
    inputNode("description", board.description || "", "description"),
    inputNode("status_slugs", board.status_slugs.join(", "), "status slugs"),
    inputNode("wip_limits", formatBoardWIPLimits(board.wip_limits), "wip limits")
  );
  const save = document.createElement("button");
  save.type = "submit";
  save.textContent = "Save";
  edit.append(save);

  const actions = document.createElement("div");
  actions.className = "board-actions";

  const select = document.createElement("button");
  select.type = "button";
  select.dataset.selectBoardId = board.id;
  select.disabled = board.id === state.selectedBoardID;
  select.textContent = board.id === state.selectedBoardID ? "Selected" : "Select";

  const remove = document.createElement("button");
  remove.type = "button";
  remove.dataset.deleteBoardId = board.id;
  remove.textContent = "Delete";

  actions.append(select, remove);
  article.append(body, edit, actions);
  return article;
}

function boardMetadataNode(board) {
  const metadata = document.createElement("div");
  metadata.className = "board-metadata";

  for (const item of boardMetadataItems(board)) {
    const chip = document.createElement("span");
    chip.textContent = item;
    metadata.append(chip);
  }

  return metadata;
}

function boardMetadataItems(board) {
  const statuses = Array.isArray(board.status_slugs) ? board.status_slugs : [];
  const wipLimits = board.wip_limits && typeof board.wip_limits === "object" ? board.wip_limits : {};
  const wipLimitCount = Object.values(wipLimits).filter((value) => Number(value) > 0).length;
  return [
    statuses.length ? `${statuses.length} column${statuses.length === 1 ? "" : "s"}` : "no columns",
    wipLimitCount ? `${wipLimitCount} WIP limit${wipLimitCount === 1 ? "" : "s"}` : "no WIP limits",
    board.description ? "has description" : "no description"
  ];
}

function boardColumnSettingsOverviewNode(board) {
  const overview = document.createElement("section");
  overview.className = "board-column-settings-overview";

  const title = document.createElement("strong");
  title.textContent = "Column settings";

  const chips = document.createElement("div");
  chips.className = "board-column-settings-chips";
  for (const item of boardColumnSettingsOverviewItems(board)) {
    const chip = document.createElement("span");
    chip.textContent = item;
    chips.append(chip);
  }

  overview.append(title, chips);
  return overview;
}

function boardColumnSettingsOverviewItems(board) {
  const statuses = Array.isArray(board.status_slugs) ? board.status_slugs : [];
  const columns = Array.isArray(board.columns) && board.columns.length
    ? board.columns
    : statuses.map((statusSlug) => ({ status_slug: statusSlug }));
  const coveredStatuses = new Set();
  for (const column of columns) {
    const slug = column && (column.status_slug || column.slug);
    if (slug) {
      coveredStatuses.add(slug);
    }
  }
  const wipLimits = board.wip_limits && typeof board.wip_limits === "object" ? board.wip_limits : {};
  const wipLimitedColumns = columns.filter((column) => {
    const columnLimit = column && Number(column.wip_limit);
    const slug = column && (column.status_slug || column.slug);
    const boardLimit = slug ? Number(wipLimits[slug]) : 0;
    return columnLimit > 0 || boardLimit > 0;
  }).length;

  return [
    `${columns.length} configured column${columns.length === 1 ? "" : "s"}`,
    `${coveredStatuses.size} covered status${coveredStatuses.size === 1 ? "" : "es"}`,
    wipLimitedColumns
      ? `${wipLimitedColumns} WIP-limited column${wipLimitedColumns === 1 ? "" : "s"}`
      : "no WIP-limited columns"
  ];
}

function boardStatusCoverageOverviewNode(board) {
  const overview = document.createElement("section");
  overview.className = "board-status-coverage-overview";

  const title = document.createElement("strong");
  title.textContent = "Status coverage";

  const chips = document.createElement("div");
  chips.className = "board-status-coverage-chips";
  for (const item of boardStatusCoverageOverviewItems(board)) {
    const chip = document.createElement("span");
    chip.textContent = item;
    chips.append(chip);
  }

  overview.append(title, chips);
  return overview;
}

function boardStatusCoverageOverviewItems(board) {
  const boardStatuses = Array.isArray(board.status_slugs) ? board.status_slugs.filter(Boolean) : [];
  const workflowStatuses = (state.workflowStatuses.length ? state.workflowStatuses : defaultWorkflowStatuses())
    .filter((status) => status && status.slug);
  const workflowSlugs = new Set(workflowStatuses.map((status) => status.slug));
  const boardSlugs = new Set(boardStatuses);
  const covered = workflowStatuses.filter((status) => boardSlugs.has(status.slug));
  const missing = workflowStatuses.filter((status) => !boardSlugs.has(status.slug));
  const extra = boardStatuses.filter((slug) => !workflowSlugs.has(slug));

  const items = [
    `${covered.length} covered workflow status${covered.length === 1 ? "" : "es"}`,
    missing.length
      ? `${missing.length} not on board: ${missing.map((status) => status.name || status.slug).join(", ")}`
      : "no workflow statuses off board",
    extra.length
      ? `${extra.length} extra board status${extra.length === 1 ? "" : "es"}: ${extra.map((slug) => statusName(slug)).join(", ")}`
      : "no extra board statuses"
  ];
  return items;
}

function renderSprints() {
  if (!els.sprints) {
    return;
  }
  els.sprints.replaceChildren();
  renderSprintFilter();
  if (!state.selectedProject) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "Select a project to manage sprints";
    els.sprints.append(empty);
    return;
  }
  if (!state.sprints.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No sprints";
    els.sprints.append(empty);
    return;
  }
  for (const sprint of state.sprints) {
    els.sprints.append(sprintNode(sprint));
  }
}

function renderSprintFilter() {
  if (!els.sprintForm || !els.sprintForm.elements.state_filter) {
    return;
  }
  els.sprintForm.elements.state_filter.value = state.sprintFilterState || "";
}

function renderSprintReport() {
  if (!els.sprintReport) {
    return;
  }
  els.sprintReport.replaceChildren();
  if (!state.selectedProject) {
    return;
  }

  const report = state.sprintReport;
  const sprint = report ? report.sprint : state.sprints.find((item) => item.id === state.selectedSprintReportID);
  if (!sprint) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "Select a sprint report";
    els.sprintReport.append(empty);
    return;
  }

  const header = document.createElement("div");
  header.className = "sprint-report-header";
  const title = document.createElement("h3");
  title.textContent = `${sprint.name || "Sprint"} report`;
  const scope = document.createElement("span");
  scope.textContent = sprintReportScopeText(report);
  header.append(title, scope);

  const progress = report ? report.progress : { total: 0, done: 0, by_status: {} };
  const hasPointMetrics = Number(progress.story_points_total || 0) > 0 || Number(progress.story_points_unestimated || 0) > 0;
  const metrics = document.createElement("div");
  metrics.className = "sprint-report-metrics";
  metrics.append(
    sprintMetricNode("Total", progress.total || 0),
    sprintMetricNode("Done", progress.done || 0),
    sprintMetricNode("Open", Math.max((progress.total || 0) - (progress.done || 0), 0))
  );
  if (hasPointMetrics) {
    metrics.append(
      sprintMetricNode("Points", formatStoryPoints(progress.story_points_total || 0)),
      sprintMetricNode("Done points", formatStoryPoints(progress.story_points_done || 0)),
      sprintMetricNode("Remaining points", formatStoryPoints(progress.story_points_remaining || 0)),
      sprintMetricNode("Unestimated", progress.story_points_unestimated || 0)
    );
  }

  const statuses = sprintReportStatusBreakdownNode(progress);
  const startDates = sprintReportStartDateBreakdownNode(report && report.tickets ? report.tickets : []);
  const dueDates = sprintReportDueDateBreakdownNode(report && report.tickets ? report.tickets : []);
  const ages = sprintReportAgeBreakdownNode(report && report.tickets ? report.tickets : []);
  const updates = sprintReportUpdateFreshnessNode(report && report.tickets ? report.tickets : []);
  const readiness = sprintReportReadinessSummaryNode(report && report.tickets ? report.tickets : []);
  const risks = sprintReportRiskSummaryNode(report && report.tickets ? report.tickets : []);
  const priorities = sprintReportPriorityBreakdownNode(report && report.tickets ? report.tickets : []);
  const types = sprintReportTypeBreakdownNode(report && report.tickets ? report.tickets : []);
  const labels = sprintReportLabelBreakdownNode(report && report.tickets ? report.tickets : []);
  const estimateCoverage = sprintReportEstimateCoverageNode(progress);
  const components = sprintReportComponentNode(report && report.tickets ? report.tickets : []);
  const versions = sprintReportVersionNode(report && report.tickets ? report.tickets : []);
  const epics = sprintReportEpicBreakdownNode(report && report.tickets ? report.tickets : []);
  const analytics = sprintReportAnalyticsNode(report ? report.analytics : null);
  const scopeChanges = sprintReportScopeChangesNode(report ? report.scope_changes : null);
  const reporters = sprintReportReporterBreakdownNode(report && report.tickets ? report.tickets : []);
  const assignees = sprintReportAssigneeWorkloadsNode(report && report.tickets ? report.tickets : []);

  const tickets = document.createElement("div");
  tickets.className = "sprint-report-tickets";
  if (report && report.tickets && report.tickets.length) {
    for (const ticket of report.tickets) {
      const row = sprintReportTicketNode(ticket);
      row.classList.add("version-report-ticket");
      tickets.append(row);
    }
  } else {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = report ? "No tickets in this report scope" : "Report tickets will appear here";
    tickets.append(empty);
  }

  els.sprintReport.append(header, sprintReportHealthNode(sprint), metrics, statuses, startDates, dueDates, ages, updates, readiness, risks, priorities, types, labels, estimateCoverage, components, versions, epics, analytics, scopeChanges, reporters, assignees, tickets);
}

function sprintReportHealthNode(sprint) {
  const health = sprintReportHealth(sprint);
  const section = document.createElement("section");
  section.className = `sprint-report-health is-${health.state}`;

  const title = document.createElement("strong");
  title.textContent = health.label;

  const detail = document.createElement("span");
  detail.textContent = health.detail;

  const dates = document.createElement("div");
  dates.className = "sprint-report-health-dates";
  for (const item of sprintReportHealthDates(sprint)) {
    const chip = document.createElement("span");
    chip.textContent = item;
    dates.append(chip);
  }

  section.append(title, detail, dates);
  return section;
}

function sprintReportHealth(sprint, todayValue = todayLocalISODate()) {
  const state = sprint.state || "planned";
  if (state === "completed") {
    return {
      state: "completed",
      label: "Completed",
      detail: sprint.completed_at ? `Completed ${formatDateTime(sprint.completed_at)}` : "Completed sprint"
    };
  }

  const start = dateToUTC(sprint.start_date);
  const end = dateToUTC(sprint.end_date);
  if (!start && !end) {
    return {
      state: "unscheduled",
      label: "No dates",
      detail: "Add sprint dates to track schedule health"
    };
  }

  const today = dateToUTC(todayValue);
  if (state === "active" && end) {
    const daysToEnd = daysBetween(today, end);
    if (daysToEnd < 0) {
      return {
        state: "overdue",
        label: "Overdue",
        detail: `${Math.abs(daysToEnd)} day${Math.abs(daysToEnd) === 1 ? "" : "s"} past end date`
      };
    }
    if (daysToEnd <= 3) {
      return {
        state: "due-soon",
        label: "Due soon",
        detail: daysToEnd === 0 ? "Sprint ends today" : `${daysToEnd} day${daysToEnd === 1 ? "" : "s"} remaining`
      };
    }
    return {
      state: "active",
      label: "Active",
      detail: `${daysToEnd} day${daysToEnd === 1 ? "" : "s"} remaining`
    };
  }

  if (start) {
    const daysToStart = daysBetween(today, start);
    if (daysToStart > 0) {
      return {
        state: "scheduled",
        label: "Scheduled",
        detail: `${daysToStart} day${daysToStart === 1 ? "" : "s"} to start`
      };
    }
  }

  return {
    state,
    label: titleize(state),
    detail: dateRange(sprint.start_date, sprint.end_date) || "Sprint dates are partially set"
  };
}

function sprintReportHealthDates(sprint) {
  return [
    dateRange(sprint.start_date, sprint.end_date) || "no date range",
    `state ${sprint.state || "planned"}`
  ];
}

function sprintReportScopeText(report) {
  if (!report) {
    return "Report not loaded";
  }
  if (report.scope === "completed_snapshot") {
    return report.snapshot_at ? `Completed snapshot ${formatDateTime(report.snapshot_at)}` : "Completed snapshot";
  }
  return "Live current assignment";
}

function sprintMetricNode(label, value) {
  const metric = document.createElement("div");
  metric.className = "sprint-report-metric";
  const number = document.createElement("strong");
  number.textContent = String(value);
  const caption = document.createElement("span");
  caption.textContent = label;
  metric.append(number, caption);
  return metric;
}

function sprintReportAnalyticsNode(analytics) {
  const section = document.createElement("div");
  section.className = "sprint-report-analytics";
  const velocity = analytics && analytics.velocity ? analytics.velocity : { completed: 0, unit: "tickets" };
  const latestBurndown = latestPoint(analytics && analytics.burndown);
  const latestBurnup = latestPoint(analytics && analytics.burnup);
  section.append(
    sprintMetricNode("Velocity", `${velocity.completed || 0} ${velocity.unit || "tickets"}`),
    sprintMetricNode("Remaining", latestBurndown ? latestBurndown.remaining || 0 : 0),
    sprintMetricNode("Burnup", latestBurnup ? `${latestBurnup.done || 0}/${latestBurnup.total || 0}` : "0/0")
  );

  const chart = document.createElement("div");
  chart.className = "sprint-report-chart";
  const burnup = analytics && Array.isArray(analytics.burnup) ? analytics.burnup : [];
  for (const point of burnup.slice(-14)) {
    const total = Math.max(Number(point.total) || 0, 1);
    const done = Math.max(Number(point.done) || 0, 0);
    const bar = document.createElement("span");
    bar.style.height = `${Math.max(8, Math.round((done / total) * 100))}%`;
    bar.title = `${point.date}: ${done}/${total} done`;
    chart.append(bar);
  }
  if (!chart.children.length) {
    const empty = document.createElement("small");
    empty.textContent = "Analytics will appear when the report is loaded";
    chart.append(empty);
  }
  section.append(chart);
  return section;
}

function latestPoint(points) {
  return Array.isArray(points) && points.length ? points[points.length - 1] : null;
}

function sprintReportScopeChangesNode(scopeChanges) {
  const section = document.createElement("section");
  section.className = "sprint-report-scope-changes";

  const heading = document.createElement("h4");
  heading.textContent = "Scope changes";

  const chips = document.createElement("div");
  chips.className = "sprint-report-scope-change-list";
  for (const item of sprintReportScopeChangeItems(scopeChanges)) {
    const chip = document.createElement("span");
    chip.textContent = item;
    chips.append(chip);
  }

  section.append(heading, chips);
  return section;
}

function sprintReportScopeChangeItems(scopeChanges) {
  const changes = scopeChanges || {};
  return [
    `current ${Number(changes.current) || 0}`,
    `snapshot ${Number(changes.snapshot) || 0}`,
    `added ${Number(changes.added) || 0}`,
    `removed ${Number(changes.removed) || 0}`,
    `unchanged ${Number(changes.unchanged) || 0}`
  ];
}

function sprintReportStatusBreakdownNode(progress) {
  const section = document.createElement("section");
  section.className = "sprint-report-statuses";

  const heading = document.createElement("h4");
  heading.textContent = "Status breakdown";
  section.append(heading);

  const list = document.createElement("div");
  list.className = "sprint-report-status-list";
  const statuses = sprintReportStatusBreakdown(progress);
  if (!statuses.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No status data";
    list.append(empty);
  } else {
    for (const status of statuses) {
      const item = document.createElement("span");
      item.textContent = `${status.label}: ${status.count}`;
      list.append(item);
    }
  }
  section.append(list);
  return section;
}

function sprintReportStatusBreakdown(progress) {
  const byStatus = progress && progress.by_status ? progress.by_status : {};
  const counts = new Map();
  for (const [slug, count] of Object.entries(byStatus)) {
    const value = Number(count) || 0;
    if (value > 0) {
      counts.set(slug || "", value);
    }
  }
  const configured = (state.workflowStatuses.length ? state.workflowStatuses : defaultWorkflowStatuses())
    .slice()
    .sort((left, right) => Number(left.position || 0) - Number(right.position || 0));
  const known = new Set(configured.map((status) => status.slug));
  const ordered = configured
    .filter((status) => counts.has(status.slug))
    .map((status) => ({ label: status.name || status.slug, count: counts.get(status.slug) }));
  const unknown = Array.from(counts.entries())
    .filter(([slug]) => !known.has(slug))
    .map(([slug, count]) => ({ label: slug || "No status", count }))
    .sort((left, right) => {
      if (right.count !== left.count) {
        return right.count - left.count;
      }
      return left.label.localeCompare(right.label);
    });
  return ordered.concat(unknown);
}

function sprintReportDueDateBreakdownNode(tickets) {
  const section = document.createElement("section");
  section.className = "sprint-report-due-dates";

  const heading = document.createElement("h4");
  heading.textContent = "Due date breakdown";
  section.append(heading);

  const list = document.createElement("div");
  list.className = "sprint-report-due-date-list";
  const buckets = sprintReportDueDateBreakdown(tickets);
  if (!buckets.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No due date data";
    list.append(empty);
  } else {
    for (const bucket of buckets) {
      const item = document.createElement("span");
      item.textContent = `${bucket.label}: ${bucket.count}`;
      list.append(item);
    }
  }
  section.append(list);
  return section;
}

function sprintReportDueDateBreakdown(tickets, todayValue = todayLocalISODate()) {
  const today = dateToUTC(todayValue) || dateToUTC(todayLocalISODate());
  const buckets = [
    { key: "overdue", label: "Overdue", count: 0 },
    { key: "today", label: "Due today", count: 0 },
    { key: "soon", label: "Due soon", count: 0 },
    { key: "later", label: "Later", count: 0 },
    { key: "none", label: "No due date", count: 0 }
  ];
  const byKey = new Map(buckets.map((bucket) => [bucket.key, bucket]));
  for (const ticket of Array.isArray(tickets) ? tickets : []) {
    const due = dateToUTC(ticket.due_date);
    if (!due) {
      byKey.get("none").count += 1;
      continue;
    }
    const days = daysBetween(today, due);
    if (days < 0) {
      byKey.get("overdue").count += 1;
    } else if (days === 0) {
      byKey.get("today").count += 1;
    } else if (days <= 3) {
      byKey.get("soon").count += 1;
    } else {
      byKey.get("later").count += 1;
    }
  }
  return buckets.filter((bucket) => bucket.count > 0);
}

function sprintReportAgeBreakdownNode(tickets) {
  const section = document.createElement("section");
  section.className = "sprint-report-ages";

  const heading = document.createElement("h4");
  heading.textContent = "Ticket age breakdown";
  section.append(heading);

  const list = document.createElement("div");
  list.className = "sprint-report-age-list";
  const buckets = sprintReportAgeBreakdown(tickets);
  if (!buckets.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No age data";
    list.append(empty);
  } else {
    for (const bucket of buckets) {
      const item = document.createElement("span");
      item.textContent = `${bucket.label}: ${bucket.count}`;
      list.append(item);
    }
  }
  section.append(list);
  return section;
}

function sprintReportAgeBreakdown(tickets, todayValue = todayLocalISODate()) {
  const today = dateToUTC(todayValue) || dateToUTC(todayLocalISODate());
  const buckets = [
    { key: "new", label: "New (0-7 days)", count: 0 },
    { key: "recent", label: "Recent (8-30 days)", count: 0 },
    { key: "older", label: "Older (31+ days)", count: 0 },
    { key: "none", label: "No created date", count: 0 }
  ];
  const byKey = new Map(buckets.map((bucket) => [bucket.key, bucket]));
  for (const ticket of Array.isArray(tickets) ? tickets : []) {
    const created = sprintReportCreatedDate(ticket.created_at);
    if (!created) {
      byKey.get("none").count += 1;
      continue;
    }
    const ageDays = Math.max(daysBetween(created, today), 0);
    if (ageDays <= 7) {
      byKey.get("new").count += 1;
    } else if (ageDays <= 30) {
      byKey.get("recent").count += 1;
    } else {
      byKey.get("older").count += 1;
    }
  }
  return buckets.filter((bucket) => bucket.count > 0);
}

function sprintReportCreatedDate(value) {
  if (!value) {
    return null;
  }
  return dateToUTC(String(value).slice(0, 10));
}

function sprintReportUpdateFreshnessNode(tickets) {
  const section = document.createElement("section");
  section.className = "sprint-report-updates";

  const heading = document.createElement("h4");
  heading.textContent = "Update freshness";
  section.append(heading);

  const list = document.createElement("div");
  list.className = "sprint-report-update-list";
  const buckets = sprintReportUpdateFreshness(tickets);
  if (!buckets.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No update data";
    list.append(empty);
  } else {
    for (const bucket of buckets) {
      const item = document.createElement("span");
      item.textContent = `${bucket.label}: ${bucket.count}`;
      list.append(item);
    }
  }
  section.append(list);
  return section;
}

function sprintReportUpdateFreshness(tickets, todayValue = todayLocalISODate()) {
  const today = dateToUTC(todayValue) || dateToUTC(todayLocalISODate());
  const buckets = [
    { key: "today", label: "Updated today", count: 0 },
    { key: "week", label: "Updated this week", count: 0 },
    { key: "stale", label: "Stale (8-30 days)", count: 0 },
    { key: "dormant", label: "Dormant (31+ days)", count: 0 },
    { key: "none", label: "No update date", count: 0 }
  ];
  const byKey = new Map(buckets.map((bucket) => [bucket.key, bucket]));
  for (const ticket of Array.isArray(tickets) ? tickets : []) {
    const updated = sprintReportUpdatedDate(ticket.updated_at);
    if (!updated) {
      byKey.get("none").count += 1;
      continue;
    }
    const ageDays = Math.max(daysBetween(updated, today), 0);
    if (ageDays === 0) {
      byKey.get("today").count += 1;
    } else if (ageDays <= 7) {
      byKey.get("week").count += 1;
    } else if (ageDays <= 30) {
      byKey.get("stale").count += 1;
    } else {
      byKey.get("dormant").count += 1;
    }
  }
  return buckets.filter((bucket) => bucket.count > 0);
}

function sprintReportUpdatedDate(value) {
  if (!value) {
    return null;
  }
  return dateToUTC(String(value).slice(0, 10));
}

function sprintReportReadinessSummaryNode(tickets) {
  const section = document.createElement("section");
  section.className = "sprint-report-readiness";

  const heading = document.createElement("h4");
  heading.textContent = "Readiness summary";
  section.append(heading);

  const list = document.createElement("div");
  list.className = "sprint-report-readiness-list";
  const buckets = sprintReportReadinessSummary(tickets);
  if (!buckets.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No readiness data";
    list.append(empty);
  } else {
    for (const bucket of buckets) {
      const item = document.createElement("span");
      item.textContent = `${bucket.label}: ${bucket.count}`;
      list.append(item);
    }
  }
  section.append(list);
  return section;
}

function sprintReportReadinessSummary(tickets) {
  const buckets = [
    { key: "ready", label: "Ready", count: 0 },
    { key: "missing_assignee", label: "Missing assignee", count: 0 },
    { key: "missing_estimate", label: "Missing estimate", count: 0 },
    { key: "missing_start", label: "Missing start date", count: 0 },
    { key: "missing_due", label: "Missing due date", count: 0 }
  ];
  const byKey = new Map(buckets.map((bucket) => [bucket.key, bucket]));
  for (const ticket of Array.isArray(tickets) ? tickets : []) {
    let gaps = 0;
    if (!String(ticket.assignee_id || "").trim()) {
      byKey.get("missing_assignee").count += 1;
      gaps += 1;
    }
    if (!sprintReportHasEstimate(ticket.story_points)) {
      byKey.get("missing_estimate").count += 1;
      gaps += 1;
    }
    if (!dateToUTC(ticket.start_date)) {
      byKey.get("missing_start").count += 1;
      gaps += 1;
    }
    if (!dateToUTC(ticket.due_date)) {
      byKey.get("missing_due").count += 1;
      gaps += 1;
    }
    if (gaps === 0) {
      byKey.get("ready").count += 1;
    }
  }
  return buckets.filter((bucket) => bucket.count > 0);
}

function sprintReportHasEstimate(value) {
  if (value === null || value === undefined || value === "") {
    return false;
  }
  if (typeof value === "string" && !value.trim()) {
    return false;
  }
  return Number.isFinite(Number(value));
}

function sprintReportRiskSummaryNode(tickets) {
  const section = document.createElement("section");
  section.className = "sprint-report-risks";

  const heading = document.createElement("h4");
  heading.textContent = "Risk summary";
  section.append(heading);

  const list = document.createElement("div");
  list.className = "sprint-report-risk-list";
  const buckets = sprintReportRiskSummary(tickets);
  if (!buckets.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No risk data";
    list.append(empty);
  } else {
    for (const bucket of buckets) {
      const item = document.createElement("span");
      item.textContent = `${bucket.label}: ${bucket.count}`;
      list.append(item);
    }
  }
  section.append(list);
  return section;
}

function sprintReportRiskSummary(tickets, todayValue = todayLocalISODate()) {
  const today = dateToUTC(todayValue) || dateToUTC(todayLocalISODate());
  const buckets = [
    { key: "overdue_open", label: "Open overdue", count: 0 },
    { key: "stale_open", label: "Stale open", count: 0 },
    { key: "unassigned_open", label: "Unassigned open", count: 0 },
    { key: "unscheduled_open", label: "Unscheduled open", count: 0 }
  ];
  const byKey = new Map(buckets.map((bucket) => [bucket.key, bucket]));
  for (const ticket of Array.isArray(tickets) ? tickets : []) {
    if (sprintReportTicketDone(ticket)) {
      continue;
    }
    const due = dateToUTC(ticket.due_date);
    if (due && daysBetween(today, due) < 0) {
      byKey.get("overdue_open").count += 1;
    }
    const updated = sprintReportUpdatedDate(ticket.updated_at);
    if (updated && daysBetween(updated, today) > 7) {
      byKey.get("stale_open").count += 1;
    }
    if (!String(ticket.assignee_id || "").trim()) {
      byKey.get("unassigned_open").count += 1;
    }
    if (!dateToUTC(ticket.start_date) || !due) {
      byKey.get("unscheduled_open").count += 1;
    }
  }
  return buckets.filter((bucket) => bucket.count > 0);
}

function sprintReportTicketDone(ticket) {
  return String(ticket && ticket.status ? ticket.status : "").toLowerCase() === "done";
}

function sprintReportStartDateBreakdownNode(tickets) {
  const section = document.createElement("section");
  section.className = "sprint-report-start-dates";

  const heading = document.createElement("h4");
  heading.textContent = "Start date breakdown";
  section.append(heading);

  const list = document.createElement("div");
  list.className = "sprint-report-start-date-list";
  const buckets = sprintReportStartDateBreakdown(tickets);
  if (!buckets.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No start date data";
    list.append(empty);
  } else {
    for (const bucket of buckets) {
      const item = document.createElement("span");
      item.textContent = `${bucket.label}: ${bucket.count}`;
      list.append(item);
    }
  }
  section.append(list);
  return section;
}

function sprintReportStartDateBreakdown(tickets, todayValue = todayLocalISODate()) {
  const today = dateToUTC(todayValue) || dateToUTC(todayLocalISODate());
  const buckets = [
    { key: "started", label: "Started", count: 0 },
    { key: "today", label: "Starts today", count: 0 },
    { key: "soon", label: "Starts soon", count: 0 },
    { key: "future", label: "Future start", count: 0 },
    { key: "none", label: "No start date", count: 0 }
  ];
  const byKey = new Map(buckets.map((bucket) => [bucket.key, bucket]));
  for (const ticket of Array.isArray(tickets) ? tickets : []) {
    const start = dateToUTC(ticket.start_date);
    if (!start) {
      byKey.get("none").count += 1;
      continue;
    }
    const days = daysBetween(today, start);
    if (days < 0) {
      byKey.get("started").count += 1;
    } else if (days === 0) {
      byKey.get("today").count += 1;
    } else if (days <= 3) {
      byKey.get("soon").count += 1;
    } else {
      byKey.get("future").count += 1;
    }
  }
  return buckets.filter((bucket) => bucket.count > 0);
}

function sprintReportPriorityBreakdownNode(tickets) {
  const section = document.createElement("section");
  section.className = "sprint-report-priorities";

  const heading = document.createElement("h4");
  heading.textContent = "Priority breakdown";
  section.append(heading);

  const list = document.createElement("div");
  list.className = "sprint-report-priority-list";
  const priorities = sprintReportPriorityBreakdown(tickets);
  if (!priorities.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No priority data";
    list.append(empty);
  } else {
    for (const priority of priorities) {
      const item = document.createElement("span");
      item.textContent = `${priority.label}: ${priority.count}`;
      list.append(item);
    }
  }
  section.append(list);
  return section;
}

function sprintReportPriorityBreakdown(tickets) {
  const priorities = new Map();
  for (const ticket of Array.isArray(tickets) ? tickets : []) {
    const label = ticket.priority || "No priority";
    priorities.set(label, (priorities.get(label) || 0) + 1);
  }
  return Array.from(priorities.entries())
    .map(([label, count]) => ({ label, count }))
    .sort((left, right) => {
      if (right.count !== left.count) {
        return right.count - left.count;
      }
      return left.label.localeCompare(right.label);
    });
}

function sprintReportTypeBreakdownNode(tickets) {
  const section = document.createElement("section");
  section.className = "sprint-report-types";

  const heading = document.createElement("h4");
  heading.textContent = "Issue type breakdown";
  section.append(heading);

  const list = document.createElement("div");
  list.className = "sprint-report-type-list";
  const types = sprintReportTypeBreakdown(tickets);
  if (!types.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No issue type data";
    list.append(empty);
  } else {
    for (const type of types) {
      const item = document.createElement("span");
      item.textContent = `${type.label}: ${type.count}`;
      list.append(item);
    }
  }
  section.append(list);
  return section;
}

function sprintReportTypeBreakdown(tickets) {
  const types = new Map();
  for (const ticket of Array.isArray(tickets) ? tickets : []) {
    const label = ticket.type || "No issue type";
    types.set(label, (types.get(label) || 0) + 1);
  }
  return Array.from(types.entries())
    .map(([label, count]) => ({ label, count }))
    .sort((left, right) => {
      if (right.count !== left.count) {
        return right.count - left.count;
      }
      return left.label.localeCompare(right.label);
    });
}

function sprintReportLabelBreakdownNode(tickets) {
  const section = document.createElement("section");
  section.className = "sprint-report-labels";

  const heading = document.createElement("h4");
  heading.textContent = "Label breakdown";
  section.append(heading);

  const list = document.createElement("div");
  list.className = "sprint-report-label-list";
  const labels = sprintReportLabelBreakdown(tickets);
  if (!labels.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No label data";
    list.append(empty);
  } else {
    for (const label of labels) {
      const item = document.createElement("span");
      item.textContent = `${label.label}: ${label.count}`;
      list.append(item);
    }
  }
  section.append(list);
  return section;
}

function sprintReportLabelBreakdown(tickets) {
  const labels = new Map();
  for (const ticket of Array.isArray(tickets) ? tickets : []) {
    const ticketLabels = Array.isArray(ticket.labels) ? ticket.labels.filter(Boolean) : [];
    if (!ticketLabels.length) {
      labels.set("No labels", (labels.get("No labels") || 0) + 1);
      continue;
    }
    for (const label of ticketLabels) {
      labels.set(label, (labels.get(label) || 0) + 1);
    }
  }
  return Array.from(labels.entries())
    .map(([label, count]) => ({ label, count }))
    .sort((left, right) => {
      if (right.count !== left.count) {
        return right.count - left.count;
      }
      if (left.label === "No labels") {
        return 1;
      }
      if (right.label === "No labels") {
        return -1;
      }
      return left.label.localeCompare(right.label);
    });
}

function sprintReportEstimateCoverageNode(progress) {
  const section = document.createElement("section");
  section.className = "sprint-report-estimate-coverage";

  const heading = document.createElement("h4");
  heading.textContent = "Estimate coverage";

  const list = document.createElement("div");
  list.className = "sprint-report-estimate-coverage-list";
  const coverage = sprintReportEstimateCoverage(progress);
  if (coverage.total === 0) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No sprint tickets";
    list.append(empty);
  } else {
    for (const item of coverage.items) {
      const chip = document.createElement("span");
      chip.textContent = `${item.label}: ${item.count}`;
      list.append(chip);
    }
  }

  section.append(heading, list);
  return section;
}

function sprintReportEstimateCoverage(progress) {
  const data = progress || {};
  const totalValue = Number(data.total || 0);
  const unestimatedValue = Number(data.story_points_unestimated || 0);
  const total = Math.max(Number.isFinite(totalValue) ? totalValue : 0, 0);
  const unestimated = Math.min(Math.max(Number.isFinite(unestimatedValue) ? unestimatedValue : 0, 0), total);
  return {
    total,
    items: [
      { label: "Estimated", count: Math.max(total - unestimated, 0) },
      { label: "Unestimated", count: unestimated }
    ]
  };
}

function sprintReportComponentNode(tickets) {
  const section = document.createElement("section");
  section.className = "sprint-report-components";

  const heading = document.createElement("h4");
  heading.textContent = "Component breakdown";

  const list = document.createElement("div");
  list.className = "sprint-report-component-list";
  const components = sprintReportComponents(tickets);
  if (!components.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No component assignments";
    list.append(empty);
  } else {
    for (const component of components) {
      list.append(sprintReportComponentItemNode(component));
    }
  }

  section.append(heading, list);
  return section;
}

function sprintReportComponents(tickets) {
  const groups = new Map();
  for (const ticket of Array.isArray(tickets) ? tickets : []) {
    const id = ticket.component_id || "";
    const key = id || "__unassigned__";
    if (!groups.has(key)) {
      groups.set(key, {
        id,
        name: id ? componentName(id) : "No component",
        total: 0,
        done: 0,
        story_points_total: 0,
        story_points_done: 0,
        unestimated: 0
      });
    }
    const group = groups.get(key);
    group.total += 1;
    if (ticket.status === "done") {
      group.done += 1;
    }
    if (ticket.story_points === null || ticket.story_points === undefined || ticket.story_points === "") {
      group.unestimated += 1;
    } else {
      const points = Number(ticket.story_points || 0);
      group.story_points_total += points;
      if (ticket.status === "done") {
        group.story_points_done += points;
      }
    }
  }
  return Array.from(groups.values()).sort((left, right) => {
    if (!left.id && right.id) {
      return 1;
    }
    if (left.id && !right.id) {
      return -1;
    }
    return left.name.localeCompare(right.name);
  });
}

function sprintReportComponentItemNode(component) {
  const item = document.createElement("article");
  item.className = component.id ? "sprint-report-component" : "sprint-report-component is-unassigned";
  const title = document.createElement("strong");
  title.textContent = component.name;
  const meta = document.createElement("small");
  const pointText = component.story_points_total > 0
    ? `${formatStoryPoints(component.story_points_done)}/${formatStoryPoints(component.story_points_total)} pts`
    : `${component.unestimated} unestimated`;
  meta.textContent = `${component.done}/${component.total} done / ${pointText}`;
  item.append(title, meta);
  return item;
}

function sprintReportVersionNode(tickets) {
  const section = document.createElement("section");
  section.className = "sprint-report-versions";

  const heading = document.createElement("h4");
  heading.textContent = "Version breakdown";

  const list = document.createElement("div");
  list.className = "sprint-report-version-list";
  const versions = sprintReportVersions(tickets);
  if (!versions.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No version assignments";
    list.append(empty);
  } else {
    for (const version of versions) {
      list.append(sprintReportVersionItemNode(version));
    }
  }

  section.append(heading, list);
  return section;
}

function sprintReportVersions(tickets) {
  const groups = new Map();
  for (const ticket of Array.isArray(tickets) ? tickets : []) {
    const id = ticket.version_id || "";
    const key = id || "__unassigned__";
    if (!groups.has(key)) {
      groups.set(key, {
        id,
        name: id ? versionName(id) : "No version",
        total: 0,
        done: 0,
        story_points_total: 0,
        story_points_done: 0,
        unestimated: 0
      });
    }
    const group = groups.get(key);
    group.total += 1;
    if (ticket.status === "done") {
      group.done += 1;
    }
    if (ticket.story_points === null || ticket.story_points === undefined || ticket.story_points === "") {
      group.unestimated += 1;
    } else {
      const points = Number(ticket.story_points || 0);
      group.story_points_total += points;
      if (ticket.status === "done") {
        group.story_points_done += points;
      }
    }
  }
  return Array.from(groups.values()).sort((left, right) => {
    if (!left.id && right.id) {
      return 1;
    }
    if (left.id && !right.id) {
      return -1;
    }
    return left.name.localeCompare(right.name);
  });
}

function sprintReportVersionItemNode(version) {
  const item = document.createElement("article");
  item.className = version.id ? "sprint-report-version" : "sprint-report-version is-unassigned";
  const title = document.createElement("strong");
  title.textContent = version.name;
  const meta = document.createElement("small");
  const pointText = version.story_points_total > 0
    ? `${formatStoryPoints(version.story_points_done)}/${formatStoryPoints(version.story_points_total)} pts`
    : `${version.unestimated} unestimated`;
  meta.textContent = `${version.done}/${version.total} done / ${pointText}`;
  item.append(title, meta);
  return item;
}

function sprintReportEpicBreakdownNode(tickets) {
  const section = document.createElement("section");
  section.className = "sprint-report-epics";

  const heading = document.createElement("h4");
  heading.textContent = "Epic breakdown";

  const list = document.createElement("div");
  list.className = "sprint-report-epic-list";
  const epics = sprintReportEpics(tickets);
  if (!epics.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No epic assignments";
    list.append(empty);
  } else {
    for (const epic of epics) {
      list.append(sprintReportEpicItemNode(epic));
    }
  }

  section.append(heading, list);
  return section;
}

function sprintReportEpics(tickets) {
  const groups = new Map();
  for (const ticket of Array.isArray(tickets) ? tickets : []) {
    const id = ticket.parent_ticket_id || "";
    const key = id || "__unassigned__";
    if (!groups.has(key)) {
      groups.set(key, {
        id,
        name: id ? roadmapEpicName(id) : "No epic",
        total: 0,
        done: 0,
        story_points_total: 0,
        story_points_done: 0,
        unestimated: 0
      });
    }
    const group = groups.get(key);
    group.total += 1;
    if (ticket.status === "done") {
      group.done += 1;
    }
    if (ticket.story_points === null || ticket.story_points === undefined || ticket.story_points === "") {
      group.unestimated += 1;
    } else {
      const points = Number(ticket.story_points || 0);
      group.story_points_total += points;
      if (ticket.status === "done") {
        group.story_points_done += points;
      }
    }
  }
  return Array.from(groups.values()).sort((left, right) => {
    if (!left.id && right.id) {
      return 1;
    }
    if (left.id && !right.id) {
      return -1;
    }
    return left.name.localeCompare(right.name);
  });
}

function sprintReportEpicItemNode(epic) {
  const item = document.createElement("article");
  item.className = epic.id ? "sprint-report-epic" : "sprint-report-epic is-unassigned";
  const title = document.createElement("strong");
  title.textContent = epic.name;
  const meta = document.createElement("small");
  const pointText = epic.story_points_total > 0
    ? `${formatStoryPoints(epic.story_points_done)}/${formatStoryPoints(epic.story_points_total)} pts`
    : `${epic.unestimated} unestimated`;
  meta.textContent = `${epic.done}/${epic.total} done / ${pointText}`;
  item.append(title, meta);
  return item;
}

function sprintReportReporterBreakdown(tickets) {
  const reporters = new Map();
  const ensureReporter = (reporterID) => {
    const key = reporterID || "";
    if (!reporters.has(key)) {
      reporters.set(key, {
        key,
        label: key ? `reporter ${key}` : "No reporter",
        tickets: 0,
        story_points: 0,
        has_story_points: false
      });
    }
    return reporters.get(key);
  };
  for (const ticket of Array.isArray(tickets) ? tickets : []) {
    const reporter = ensureReporter(String(ticket.reporter_id || "").trim());
    reporter.tickets += 1;
    if (ticket.story_points === null || ticket.story_points === undefined || ticket.story_points === "") {
      continue;
    }
    const points = Number(ticket.story_points);
    if (Number.isFinite(points)) {
      reporter.story_points += points;
      reporter.has_story_points = true;
    }
  }
  return Array.from(reporters.values()).sort((left, right) => {
    if (!left.key) {
      return 1;
    }
    if (!right.key) {
      return -1;
    }
    if (right.tickets !== left.tickets) {
      return right.tickets - left.tickets;
    }
    return left.label.localeCompare(right.label);
  });
}

function sprintReportReporterBreakdownNode(tickets) {
  const section = document.createElement("section");
  section.className = "sprint-report-reporters";

  const heading = document.createElement("h4");
  heading.textContent = "Reporter breakdown";
  section.append(heading);

  const list = document.createElement("div");
  list.className = "sprint-report-reporter-list";
  const reporters = sprintReportReporterBreakdown(tickets);
  if (!reporters.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No reporter data";
    list.append(empty);
  } else {
    for (const reporter of reporters) {
      const item = document.createElement("span");
      const points = reporter.has_story_points ? `${formatStoryPoints(reporter.story_points)} pts` : "no estimates";
      item.textContent = `${reporter.label}: ${reporter.tickets} ticket${reporter.tickets === 1 ? "" : "s"} / ${points}`;
      list.append(item);
    }
  }
  section.append(list);
  return section;
}

function sprintReportAssigneeWorkloads(tickets) {
  const workloads = new Map();
  const ensureWorkload = (assigneeID) => {
    const key = assigneeID || "";
    if (!workloads.has(key)) {
      workloads.set(key, {
        key,
        label: key ? `assignee ${key}` : "Unassigned",
        tickets: 0,
        story_points: 0,
        has_story_points: false
      });
    }
    return workloads.get(key);
  };
  for (const ticket of Array.isArray(tickets) ? tickets : []) {
    const workload = ensureWorkload(ticket.assignee_id || "");
    workload.tickets += 1;
    if (ticket.story_points === null || ticket.story_points === undefined || ticket.story_points === "") {
      continue;
    }
    const points = Number(ticket.story_points);
    if (Number.isFinite(points)) {
      workload.story_points += points;
      workload.has_story_points = true;
    }
  }
  return Array.from(workloads.values()).sort((left, right) => {
    if (!left.key) {
      return 1;
    }
    if (!right.key) {
      return -1;
    }
    if (right.tickets !== left.tickets) {
      return right.tickets - left.tickets;
    }
    return left.label.localeCompare(right.label);
  });
}

function sprintReportAssigneeWorkloadsNode(tickets) {
  const section = document.createElement("section");
  section.className = "sprint-report-assignees";

  const heading = document.createElement("h4");
  heading.textContent = "Assignee workload";
  section.append(heading);

  const list = document.createElement("div");
  list.className = "sprint-report-assignee-list";
  const workloads = sprintReportAssigneeWorkloads(tickets);
  if (!workloads.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "Assignee workload appears when report tickets are loaded";
    list.append(empty);
  } else {
    for (const workload of workloads) {
      const item = document.createElement("span");
      const points = workload.has_story_points ? `${formatStoryPoints(workload.story_points)} pts` : "no estimates";
      item.textContent = `${workload.label}: ${workload.tickets} ticket${workload.tickets === 1 ? "" : "s"} / ${points}`;
      list.append(item);
    }
  }
  section.append(list);
  return section;
}

function sprintReportTicketNode(ticket) {
  const row = document.createElement("a");
  row.className = "sprint-report-ticket";
  row.href = `/issues/${encodeURIComponent(ticket.id)}`;
  const title = document.createElement("span");
  title.textContent = `${ticket.key || ticket.id} ${ticket.title || "Untitled"}`;
  const meta = document.createElement("small");
  meta.textContent = [ticket.status || "todo", ticket.priority || "", storyPointLabel(ticket.story_points)].filter(Boolean).join(" / ");
  row.append(title, meta);
  return row;
}

function sprintNode(sprint) {
  const article = document.createElement("article");
  article.className = "sprint-item";
  article.dataset.sprintState = sprint.state || "planned";

  const body = document.createElement("div");
  body.className = "sprint-item-body";

  const name = document.createElement("p");
  name.textContent = sprint.name || "Sprint";

  const meta = document.createElement("span");
  meta.textContent = [
    sprint.state || "planned",
    dateRange(sprint.start_date, sprint.end_date),
    sprint.goal
  ].filter(Boolean).join(" / ");

  body.append(name, meta);

  let edit = null;
  if (sprint.state !== "completed") {
    edit = document.createElement("form");
    edit.className = "sprint-edit-form";
    edit.dataset.sprintEditForm = sprint.id;
    edit.append(
      inputNode("name", sprint.name || "", "name"),
      inputNode("goal", sprint.goal || "", "goal"),
      inputNode("start_date", sprint.start_date || "", "Start date", "date"),
      inputNode("end_date", sprint.end_date || "", "End date", "date")
    );
    const save = document.createElement("button");
    save.type = "submit";
    save.textContent = "Save";
    edit.append(save);
  }

  const actions = document.createElement("div");
  actions.className = "sprint-actions";

  const report = document.createElement("button");
  report.type = "button";
  report.dataset.sprintReportId = sprint.id;
  report.disabled = sprint.id === state.selectedSprintReportID && Boolean(state.sprintReport);
  report.textContent = sprint.id === state.selectedSprintReportID ? "Report" : "View report";
  actions.append(report);

  if (sprint.state === "planned") {
    const start = document.createElement("button");
    start.type = "button";
    start.dataset.startSprintId = sprint.id;
    start.textContent = "Start";
    actions.append(start);
  }

  if (sprint.state === "active") {
    const complete = document.createElement("button");
    complete.type = "button";
    complete.dataset.completeSprintId = sprint.id;
    complete.textContent = "Complete";
    actions.append(complete);
  }

  const remove = document.createElement("button");
  remove.type = "button";
  remove.dataset.deleteSprintId = sprint.id;
  remove.disabled = sprint.state === "active";
  remove.textContent = "Delete";
  actions.append(remove);

  article.append(body);
  if (edit) {
    article.append(edit);
  }
  article.append(actions);
  return article;
}

function renderComponents() {
  if (!els.components) {
    return;
  }
  els.components.replaceChildren();
  if (!state.selectedProject) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "Select a project to manage components";
    els.components.append(empty);
    return;
  }
  if (!state.components.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No components";
    els.components.append(empty);
    return;
  }
  for (const component of state.components) {
    els.components.append(componentNode(component));
  }
}

function componentNode(component) {
  const article = document.createElement("article");
  article.className = "component-item";

  const body = document.createElement("div");
  body.className = "component-item-body";

  const name = document.createElement("p");
  name.textContent = component.name || "Component";

  const meta = document.createElement("span");
  meta.textContent = component.description || component.id;

  body.append(name, meta);

  const edit = document.createElement("form");
  edit.className = "component-edit-form";
  edit.dataset.componentEditForm = component.id;
  edit.append(
    inputNode("name", component.name || "", "name"),
    inputNode("description", component.description || "", "description"),
    inputNode("owner_user_id", component.owner_user_id || "", "owner user id"),
    inputNode("default_assignee_id", component.default_assignee_id || "", "default assignee id")
  );

  const save = document.createElement("button");
  save.type = "submit";
  save.textContent = "Save";
  edit.append(save);

  const remove = document.createElement("button");
  remove.type = "button";
  remove.dataset.deleteComponentId = component.id;
  remove.textContent = "Delete";

  const actions = document.createElement("div");
  actions.className = "component-actions";
  actions.append(remove);

  article.append(body, edit, actions);
  return article;
}

function renderVersions() {
  if (!els.versions) {
    return;
  }
  els.versions.replaceChildren();
  if (!state.selectedProject) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "Select a project to manage versions";
    els.versions.append(empty);
    return;
  }
  if (!state.versions.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No versions";
    els.versions.append(empty);
    return;
  }
  for (const version of state.versions) {
    els.versions.append(versionNode(version));
  }
}

function renderVersionReport() {
  if (!els.versionReport) {
    return;
  }
  els.versionReport.replaceChildren();
  if (!state.selectedProject) {
    return;
  }

  const report = state.versionReport;
  const version = report ? report.version : state.versions.find((item) => item.id === state.selectedVersionReportID);
  if (!version) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "Select a version report";
    els.versionReport.append(empty);
    return;
  }

  const header = document.createElement("div");
  header.className = "version-report-header";
  const title = document.createElement("h3");
  title.textContent = `${version.name || "Version"} report`;
  const scope = document.createElement("span");
  scope.textContent = versionReportScopeText(report);
  header.append(title, scope);

  const progress = report ? report.progress : { total: 0, done: 0, open: 0, unassigned_component: 0, by_status: {} };
  const tickets = report && Array.isArray(report.tickets) ? report.tickets : [];
  els.versionReport.append(
    header,
    versionReportHealthNode(version),
    versionReportTimelineNode(version),
    versionReportSummaryNode(progress),
    versionReportEstimateCoverageNode(progress),
    versionReportScopeChangesNode(report ? report.scope_changes : null),
    versionReportAssigneeWorkloadsNode(tickets),
    versionReportBreakdownNode(progress, tickets),
    versionReportTicketListNode(report, tickets)
  );
}

function versionReportHealthNode(version) {
  const health = versionReleaseHealth(version);

  const section = document.createElement("section");
  section.className = `version-report-health is-${health.state}`;

  const title = document.createElement("strong");
  title.textContent = health.label;

  const detail = document.createElement("span");
  detail.textContent = health.detail;

  const dates = document.createElement("div");
  dates.className = "version-report-health-dates";
  for (const item of versionReportHealthDates(version)) {
    const chip = document.createElement("span");
    chip.textContent = item;
    dates.append(chip);
  }

  section.append(title, detail, dates);
  return section;
}

function versionReleaseHealth(version) {
  const state = version.state || "planned";
  if (state === "released") {
    return {
      state: "released",
      label: "Released",
      detail: version.release_date ? `Released on ${version.release_date}` : "Released without a release date"
    };
  }
  if (state === "archived") {
    return {
      state: "archived",
      label: "Archived",
      detail: version.release_date ? `Archived after release on ${version.release_date}` : "Archived without a release date"
    };
  }

  const target = dateToUTC(version.target_date);
  if (!target) {
    return {
      state: "unscheduled",
      label: "No target date",
      detail: "Add a target date to track release timing"
    };
  }

  const today = dateToUTC(todayISODate());
  const days = daysBetween(today, target);
  if (days < 0) {
    return {
      state: "overdue",
      label: "Overdue",
      detail: `${Math.abs(days)} day${Math.abs(days) === 1 ? "" : "s"} past target`
    };
  }
  if (days <= 14) {
    return {
      state: "due-soon",
      label: "Due soon",
      detail: days === 0 ? "Target date is today" : `${days} day${days === 1 ? "" : "s"} to target`
    };
  }
  return {
    state: "scheduled",
    label: "Scheduled",
    detail: `${days} day${days === 1 ? "" : "s"} to target`
  };
}

function versionReportHealthDates(version) {
  return [
    version.target_date ? `target ${version.target_date}` : "no target date",
    version.release_date ? `released ${version.release_date}` : "",
    `state ${version.state || "planned"}`
  ].filter(Boolean);
}

function versionReportTimelineNode(version) {
  const section = document.createElement("section");
  section.className = "version-report-timeline";

  const heading = document.createElement("h4");
  heading.textContent = "Release timeline";

  const chips = document.createElement("div");
  chips.className = "version-report-timeline-chips";
  for (const item of versionReportTimelineItems(version)) {
    const chip = document.createElement("span");
    chip.textContent = item;
    chips.append(chip);
  }

  section.append(heading, chips);
  return section;
}

function versionReportTimelineItems(version) {
  const target = dateToUTC(version.target_date);
  const released = dateToUTC(version.release_date);
  const releasedState = version.state === "released" || version.state === "archived";
  const items = [
    version.target_date ? `target ${version.target_date}` : "no target date",
    version.release_date ? `release ${version.release_date}` : releasedState ? "release date missing" : "not released"
  ];
  if (target && released) {
    const delta = daysBetween(target, released);
    if (delta < 0) {
      items.push(`${Math.abs(delta)} day${Math.abs(delta) === 1 ? "" : "s"} early`);
    } else if (delta > 0) {
      items.push(`${delta} day${delta === 1 ? "" : "s"} late`);
    } else {
      items.push("released on target");
    }
  } else if (target && !releasedState) {
    const days = daysBetween(dateToUTC(todayISODate()), target);
    if (days < 0) {
      items.push(`${Math.abs(days)} day${Math.abs(days) === 1 ? "" : "s"} past target`);
    } else if (days === 0) {
      items.push("target today");
    } else {
      items.push(`${days} day${days === 1 ? "" : "s"} to target`);
    }
  }
  items.push(`state ${version.state || "planned"}`);
  return items;
}

function versionReportSummaryNode(progress) {
  const total = Number(progress.total || 0);
  const done = Number(progress.done || 0);
  const open = progress.open !== undefined ? Number(progress.open || 0) : Math.max(total - done, 0);
  const percent = total > 0 ? Math.round((done / total) * 100) : 0;
  const hasPointMetrics = Number(progress.story_points_total || 0) > 0 || Number(progress.story_points_unestimated || 0) > 0;

  const section = document.createElement("section");
  section.className = "version-report-summary";

  const progressBox = document.createElement("div");
  progressBox.className = "version-report-progress";
  const label = document.createElement("span");
  label.textContent = "Progress";
  const value = document.createElement("strong");
  value.textContent = `${percent}%`;
  const bar = document.createElement("div");
  bar.className = "version-report-progress-bar";
  const fill = document.createElement("span");
  fill.style.width = `${percent}%`;
  bar.append(fill);
  const detail = document.createElement("small");
  detail.textContent = `${done} done / ${open} open / ${total} total`;
  progressBox.append(label, value, bar, detail);

  const metrics = document.createElement("div");
  metrics.className = "version-report-metrics";
  metrics.append(
    sprintMetricNode("Total", total),
    sprintMetricNode("Done", done),
    sprintMetricNode("Open", open),
    sprintMetricNode("No component", progress.unassigned_component || 0)
  );
  if (hasPointMetrics) {
    metrics.append(
      sprintMetricNode("Points", formatStoryPoints(progress.story_points_total || 0)),
      sprintMetricNode("Done points", formatStoryPoints(progress.story_points_done || 0)),
      sprintMetricNode("Remaining points", formatStoryPoints(progress.story_points_remaining || 0)),
      sprintMetricNode("Unestimated", progress.story_points_unestimated || 0)
    );
  }
  section.append(progressBox, metrics);
  return section;
}

function versionReportEstimateCoverageNode(progress) {
  const section = document.createElement("section");
  section.className = "version-report-estimate-coverage";

  const heading = document.createElement("h4");
  heading.textContent = "Estimate coverage";

  const list = document.createElement("div");
  list.className = "version-report-estimate-coverage-list";
  const coverage = versionReportEstimateCoverage(progress);
  if (coverage.total === 0) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No release tickets";
    list.append(empty);
  } else {
    for (const item of coverage.items) {
      const chip = document.createElement("span");
      chip.textContent = `${item.label}: ${item.count}`;
      list.append(chip);
    }
  }

  section.append(heading, list);
  return section;
}

function versionReportEstimateCoverage(progress) {
  const data = progress || {};
  const totalValue = Number(data.total || 0);
  const unestimatedValue = Number(data.story_points_unestimated || 0);
  const total = Math.max(Number.isFinite(totalValue) ? totalValue : 0, 0);
  const unestimated = Math.min(Math.max(Number.isFinite(unestimatedValue) ? unestimatedValue : 0, 0), total);
  return {
    total,
    items: [
      { label: "Estimated", count: Math.max(total - unestimated, 0) },
      { label: "Unestimated", count: unestimated }
    ]
  };
}

function versionReportScopeChangesNode(scopeChanges) {
  const section = document.createElement("section");
  section.className = "version-report-scope-changes";

  const heading = document.createElement("h4");
  heading.textContent = "Scope changes";

  const chips = document.createElement("div");
  chips.className = "version-report-scope-change-list";
  for (const item of versionReportScopeChangeItems(scopeChanges)) {
    const chip = document.createElement("span");
    chip.textContent = item;
    chips.append(chip);
  }

  section.append(heading, chips);
  return section;
}

function versionReportScopeChangeItems(scopeChanges) {
  const changes = scopeChanges || {};
  return [
    `current ${Number(changes.current) || 0}`,
    `snapshot ${Number(changes.snapshot) || 0}`,
    `added ${Number(changes.added) || 0}`,
    `removed ${Number(changes.removed) || 0}`,
    `unchanged ${Number(changes.unchanged) || 0}`
  ];
}

function versionReportBreakdownNode(progress, tickets) {
  const section = document.createElement("section");
  section.className = "version-report-breakdown";
  section.append(
    versionReportStatusNode(progress.by_status || {}),
    versionReportPriorityBreakdownNode(tickets),
    versionReportTypeBreakdownNode(tickets),
    versionReportComponentNode(tickets)
  );
  return section;
}

function versionReportStatusNode(statuses) {
  const section = document.createElement("div");
  section.className = "version-report-section";
  const title = document.createElement("h4");
  title.textContent = "Status breakdown";
  const list = document.createElement("div");
  list.className = "version-report-statuses";
  for (const [status, count] of Object.entries(statuses)) {
    const item = document.createElement("span");
    item.textContent = `${status}: ${count}`;
    list.append(item);
  }
  if (!list.childElementCount) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No status data";
    list.append(empty);
  }
  section.append(title, list);
  return section;
}

function versionReportComponentNode(tickets) {
  const section = document.createElement("div");
  section.className = "version-report-section";
  const title = document.createElement("h4");
  title.textContent = "Component breakdown";
  const list = document.createElement("div");
  list.className = "version-report-components";
  const components = versionReportComponents(tickets);
  for (const component of components) {
    list.append(versionReportComponentItemNode(component));
  }
  if (!components.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No component assignments";
    list.append(empty);
  }
  section.append(title, list);
  return section;
}

function versionReportComponents(tickets) {
  const groups = new Map();
  for (const ticket of tickets) {
    const id = ticket.component_id || "";
    const key = id || "__unassigned__";
    if (!groups.has(key)) {
      groups.set(key, {
        id,
        name: id ? componentName(id) : "No component",
        total: 0,
        done: 0,
        story_points_total: 0,
        story_points_done: 0,
        unestimated: 0
      });
    }
    const group = groups.get(key);
    group.total += 1;
    if (ticket.status === "done") {
      group.done += 1;
    }
    if (ticket.story_points === null || ticket.story_points === undefined || ticket.story_points === "") {
      group.unestimated += 1;
    } else {
      const points = Number(ticket.story_points || 0);
      group.story_points_total += points;
      if (ticket.status === "done") {
        group.story_points_done += points;
      }
    }
  }
  return Array.from(groups.values()).sort((a, b) => {
    if (!a.id && b.id) {
      return 1;
    }
    if (a.id && !b.id) {
      return -1;
    }
    return a.name.localeCompare(b.name);
  });
}

function versionReportComponentItemNode(component) {
  const item = document.createElement("article");
  item.className = component.id ? "version-report-component" : "version-report-component is-unassigned";
  const title = document.createElement("strong");
  title.textContent = component.name;
  const meta = document.createElement("small");
  const pointText = component.story_points_total > 0
    ? `${formatStoryPoints(component.story_points_done)}/${formatStoryPoints(component.story_points_total)} pts`
    : `${component.unestimated} unestimated`;
  meta.textContent = `${component.done}/${component.total} done / ${pointText}`;
  item.append(title, meta);
  return item;
}

function versionReportAssigneeWorkloadsNode(tickets) {
  const section = document.createElement("section");
  section.className = "version-report-assignees";
  const title = document.createElement("h4");
  title.textContent = "Assignee workload";
  const list = document.createElement("div");
  list.className = "version-report-assignee-list";
  const assignees = versionReportAssigneeWorkloads(tickets);
  for (const assignee of assignees) {
    list.append(versionReportAssigneeItemNode(assignee));
  }
  if (!assignees.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No assignee workload";
    list.append(empty);
  }
  section.append(title, list);
  return section;
}

function versionReportAssigneeWorkloads(tickets) {
  const groups = new Map();
  const ensureGroup = (assigneeID) => {
    const key = assigneeID || "";
    if (!groups.has(key)) {
      groups.set(key, {
        key,
        label: key ? `assignee ${key}` : "Unassigned",
        total: 0,
        done: 0,
        story_points_total: 0,
        story_points_done: 0,
        unestimated: 0
      });
    }
    return groups.get(key);
  };
  for (const ticket of tickets) {
    const group = ensureGroup(ticket.assignee_id || "");
    group.total += 1;
    if (ticket.status === "done") {
      group.done += 1;
    }
    if (ticket.story_points === null || ticket.story_points === undefined || ticket.story_points === "") {
      group.unestimated += 1;
    } else {
      const points = Number(ticket.story_points || 0);
      if (Number.isFinite(points)) {
        group.story_points_total += points;
        if (ticket.status === "done") {
          group.story_points_done += points;
        }
      } else {
        group.unestimated += 1;
      }
    }
  }
  return Array.from(groups.values()).sort((left, right) => {
    if (!left.key) {
      return 1;
    }
    if (!right.key) {
      return -1;
    }
    if (right.total !== left.total) {
      return right.total - left.total;
    }
    return left.label.localeCompare(right.label);
  });
}

function versionReportAssigneeItemNode(assignee) {
  const item = document.createElement("span");
  const pointText = assignee.story_points_total > 0
    ? `${formatStoryPoints(assignee.story_points_done)}/${formatStoryPoints(assignee.story_points_total)} pts`
    : `${assignee.unestimated} unestimated`;
  item.textContent = `${assignee.label}: ${assignee.done}/${assignee.total} done / ${pointText}`;
  return item;
}

function versionReportPriorityBreakdownNode(tickets) {
  const section = document.createElement("div");
  section.className = "version-report-section";
  const title = document.createElement("h4");
  title.textContent = "Priority breakdown";
  const list = document.createElement("div");
  list.className = "version-report-priorities";
  const priorities = versionReportPriorityBreakdown(tickets);
  for (const priority of priorities) {
    const item = document.createElement("span");
    item.textContent = `${priority.label}: ${priority.count}`;
    list.append(item);
  }
  if (!priorities.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No priority data";
    list.append(empty);
  }
  section.append(title, list);
  return section;
}

function versionReportPriorityBreakdown(tickets) {
  const priorities = new Map();
  for (const ticket of tickets) {
    const label = ticket.priority || "No priority";
    priorities.set(label, (priorities.get(label) || 0) + 1);
  }
  return Array.from(priorities.entries())
    .map(([label, count]) => ({ label, count }))
    .sort((left, right) => {
      if (right.count !== left.count) {
        return right.count - left.count;
      }
      return left.label.localeCompare(right.label);
    });
}

function versionReportTypeBreakdownNode(tickets) {
  const section = document.createElement("div");
  section.className = "version-report-section";
  const title = document.createElement("h4");
  title.textContent = "Issue type breakdown";
  const list = document.createElement("div");
  list.className = "version-report-types";
  const types = versionReportTypeBreakdown(tickets);
  for (const type of types) {
    const item = document.createElement("span");
    item.textContent = `${type.label}: ${type.count}`;
    list.append(item);
  }
  if (!types.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No issue type data";
    list.append(empty);
  }
  section.append(title, list);
  return section;
}

function versionReportTypeBreakdown(tickets) {
  const types = new Map();
  for (const ticket of tickets) {
    const label = ticket.type || "No issue type";
    types.set(label, (types.get(label) || 0) + 1);
  }
  return Array.from(types.entries())
    .map(([label, count]) => ({ label, count }))
    .sort((left, right) => {
      if (right.count !== left.count) {
        return right.count - left.count;
      }
      return left.label.localeCompare(right.label);
    });
}

function versionReportTicketListNode(report, tickets) {
  const section = document.createElement("section");
  section.className = "version-report-section";
  const title = document.createElement("h4");
  title.textContent = "Report tickets";
  const list = document.createElement("div");
  list.className = "version-report-tickets";
  if (tickets.length) {
    for (const ticket of tickets) {
      list.append(versionReportTicketNode(ticket));
    }
  } else {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = report ? "No tickets assigned to this version" : "Report tickets will appear here";
    list.append(empty);
  }
  section.append(title, list);
  return section;
}

function versionReportTicketNode(ticket) {
  const row = document.createElement("a");
  row.className = ticket.component_id ? "version-report-ticket" : "version-report-ticket is-unassigned";
  row.href = `/issues/${encodeURIComponent(ticket.id)}`;
  const title = document.createElement("span");
  title.textContent = `${ticket.key || ticket.id} ${ticket.title || "Untitled"}`;
  const meta = document.createElement("small");
  meta.textContent = [
    ticket.status || "todo",
    ticket.component_id ? componentName(ticket.component_id) : "No component",
    storyPointLabel(ticket.story_points)
  ].filter(Boolean).join(" / ");
  row.append(title, meta);
  return row;
}

function renderProjectLabels() {
  if (!els.projectLabels) {
    return;
  }
  els.projectLabels.replaceChildren();
  if (!state.selectedProject) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "Select a project to manage labels";
    els.projectLabels.append(empty);
    return;
  }
  if (!state.projectLabels.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No labels";
    els.projectLabels.append(empty);
    return;
  }
  for (const label of state.projectLabels) {
    els.projectLabels.append(projectLabelNode(label));
  }
}

function projectLabelNode(label) {
  const article = document.createElement("article");
  article.className = "project-label-item";

  const body = document.createElement("div");
  body.className = "project-label-body";

  const heading = document.createElement("p");
  const swatch = document.createElement("span");
  swatch.className = "label-color-swatch";
  swatch.style.backgroundColor = label.color || "transparent";
  swatch.setAttribute("aria-hidden", "true");
  const name = document.createElement("span");
  name.textContent = label.label || "label";
  heading.append(swatch, name);

  const meta = document.createElement("span");
  meta.textContent = [
    `${label.ticket_count || 0} ticket${label.ticket_count === 1 ? "" : "s"}`,
    label.description || ""
  ].filter(Boolean).join(" / ");
  body.append(heading, meta);

  const edit = document.createElement("form");
  edit.className = "project-label-edit-form";
  edit.dataset.projectLabelEditForm = label.label || "";
  edit.dataset.projectLabelCatalog = label.created_at ? "true" : "false";
  edit.append(
    inputNode("description", label.description || "", "description"),
    inputNode("color", label.color || "#5b7cfa", "color", "color")
  );

  const save = document.createElement("button");
  save.type = "submit";
  save.textContent = "Save";
  edit.append(save);

  const actions = document.createElement("div");
  actions.className = "project-label-actions";
  if (label.created_at) {
    const remove = document.createElement("button");
    remove.type = "button";
    remove.dataset.deleteProjectLabel = label.label || "";
    remove.textContent = "Delete";
    actions.append(remove);
  }

  article.append(body, edit, actions);
  return article;
}

function versionReportScopeText(report) {
  if (!report) {
    return "Report not loaded";
  }
  if (report.scope === "released_snapshot") {
    return report.snapshot_at ? `Released snapshot ${formatDateTime(report.snapshot_at)}` : "Released snapshot";
  }
  return "Live current assignment";
}

function versionNode(version) {
  const article = document.createElement("article");
  article.className = "version-item";
  article.dataset.versionState = version.state || "planned";

  const body = document.createElement("div");
  body.className = "version-item-body";

  const name = document.createElement("p");
  name.textContent = version.name || "Version";

  const meta = document.createElement("span");
  meta.textContent = [
    version.state || "planned",
    version.target_date ? `target ${version.target_date}` : "",
    version.release_date ? `released ${version.release_date}` : "",
    version.description
  ].filter(Boolean).join(" / ");

  body.append(name, meta);

  const edit = document.createElement("form");
  edit.className = "version-edit-form";
  edit.dataset.versionEditForm = version.id;
  edit.append(
    inputNode("name", version.name || "", "name"),
    inputNode("description", version.description || "", "description"),
    versionStatusSelect(version.state || "planned"),
    inputNode("target_date", version.target_date || "", "target date", "date"),
    inputNode("release_date", version.release_date || "", "release date", "date")
  );

  const save = document.createElement("button");
  save.type = "submit";
  save.textContent = "Save";
  edit.append(save);

  const actions = document.createElement("div");
  actions.className = "version-actions";

  const report = document.createElement("button");
  report.type = "button";
  report.dataset.versionReportId = version.id;
  report.disabled = version.id === state.selectedVersionReportID && Boolean(state.versionReport);
  report.textContent = version.id === state.selectedVersionReportID ? "Report" : "View report";
  actions.append(report);

  if (version.state !== "released") {
    const release = document.createElement("button");
    release.type = "button";
    release.dataset.versionId = version.id;
    release.dataset.versionStatus = "released";
    release.textContent = "Release";
    actions.append(release);
  }

  if (version.state !== "archived") {
    const archive = document.createElement("button");
    archive.type = "button";
    archive.dataset.versionId = version.id;
    archive.dataset.versionStatus = "archived";
    archive.textContent = "Archive";
    actions.append(archive);
  }

  const remove = document.createElement("button");
  remove.type = "button";
  remove.dataset.deleteVersionId = version.id;
  remove.textContent = "Delete";
  actions.append(remove);

  article.append(body, edit, actions);
  return article;
}

function versionStatusSelect(value) {
  const select = document.createElement("select");
  select.name = "status";
  for (const [statusValue, label] of [
    ["planned", "Planned"],
    ["released", "Released"],
    ["archived", "Archived"]
  ]) {
    const option = document.createElement("option");
    option.value = statusValue;
    option.textContent = label;
    option.selected = statusValue === value;
    select.append(option);
  }
  return select;
}

function renderCustomFields() {
  if (!els.customFields) {
    return;
  }
  els.customFields.replaceChildren();
  if (!state.selectedProject) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "Select a project to manage custom fields";
    els.customFields.append(empty);
    return;
  }
  if (!state.customFields.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No custom fields";
    els.customFields.append(empty);
    return;
  }
  els.customFields.append(customFieldLayoutOverviewNode(state.customFields));
  for (const field of state.customFields) {
    els.customFields.append(customFieldNode(field));
  }
}

function customFieldLayoutOverviewNode(fields) {
  const summary = customFieldLayoutSummary(fields);
  const section = document.createElement("section");
  section.className = "field-layout-overview";

  const heading = document.createElement("h3");
  heading.textContent = "Field layout";

  const chips = document.createElement("div");
  chips.className = "field-layout-chips";
  const items = [
    `${summary.total} field${summary.total === 1 ? "" : "s"}`,
    `${summary.required} required`,
    `${summary.optional} optional`
  ];
  for (const item of items) {
    const chip = document.createElement("span");
    chip.textContent = item;
    chips.append(chip);
  }

  section.append(heading, chips, customFieldTypeBreakdownNode(summary.types));
  return section;
}

function customFieldTypeBreakdownNode(types) {
  const group = document.createElement("div");
  group.className = "field-type-breakdown";

  const label = document.createElement("strong");
  label.textContent = "Field types";

  const chips = document.createElement("div");
  chips.className = "field-type-chips";
  const items = Array.isArray(types) && types.length ? types : [{ type: "text", count: 0 }];
  for (const item of items) {
    const chip = document.createElement("span");
    chip.textContent = `${item.type}: ${item.count}`;
    chips.append(chip);
  }

  group.append(label, chips);
  return group;
}

function customFieldLayoutSummary(fields) {
  const list = Array.isArray(fields) ? fields : [];
  const types = new Map();
  let required = 0;
  for (const field of list) {
    if (field.required) {
      required += 1;
    }
    const type = field.field_type || "text";
    types.set(type, (types.get(type) || 0) + 1);
  }
  return {
    total: list.length,
    required,
    optional: Math.max(list.length - required, 0),
    types: Array.from(types.entries())
      .map(([type, count]) => ({ type, count }))
      .sort((left, right) => left.type.localeCompare(right.type))
  };
}

function customFieldNode(field) {
  const optionsList = Array.isArray(field.options) ? field.options : [];
  const article = document.createElement("article");
  article.className = "field-item";
  article.dataset.fieldType = field.field_type || "text";

  const body = document.createElement("div");
  body.className = "field-item-body";

  const name = document.createElement("p");
  name.textContent = field.name || field.key;

  body.append(name, customFieldMetadataNode(field, optionsList));

  const edit = document.createElement("form");
  edit.className = "field-edit-form";
  edit.dataset.customFieldEditForm = field.id;

  edit.append(
    inputNode("key", field.key || "", "key"),
    inputNode("name", field.name || "", "name"),
    customFieldTypeSelect(field.field_type || "text")
  );

  const required = document.createElement("label");
  required.className = "inline-toggle";
  const requiredInput = document.createElement("input");
  requiredInput.type = "checkbox";
  requiredInput.name = "required";
  requiredInput.checked = Boolean(field.required);
  required.append(requiredInput, " Required");

  const options = inputNode("options", optionsList.join(", "), "options");
  const save = document.createElement("button");
  save.type = "submit";
  save.textContent = "Save";
  edit.append(required, options, save);

  const remove = document.createElement("button");
  remove.type = "button";
  remove.dataset.deleteFieldId = field.id;
  remove.textContent = "Delete";

  const actions = document.createElement("div");
  actions.className = "field-actions";
  actions.append(remove);

  article.append(body, edit, actions);
  return article;
}

function customFieldMetadataNode(field, optionsList = []) {
  const metadata = document.createElement("div");
  metadata.className = "field-metadata";

  for (const item of customFieldMetadataItems(field, optionsList)) {
    const chip = document.createElement("span");
    chip.textContent = item;
    metadata.append(chip);
  }

  return metadata;
}

function customFieldMetadataItems(field, optionsList = []) {
  const options = Array.isArray(optionsList) ? optionsList : [];
  return [
    field.key ? `key ${field.key}` : "",
    field.field_type ? `type ${field.field_type}` : "",
    field.required ? "required" : "optional",
    customFieldOptionsSummary(field, options)
  ].filter(Boolean);
}

function customFieldOptionsSummary(field, optionsList = []) {
  if (!["single_select", "multi_select"].includes(field.field_type || "")) {
    return "";
  }
  const count = optionsList.length;
  if (!count) {
    return "no options";
  }
  return `${count} option${count === 1 ? "" : "s"}`;
}

function customFieldTypeSelect(value) {
  const select = document.createElement("select");
  select.name = "field_type";
  for (const [fieldValue, label] of [
    ["text", "Text"],
    ["number", "Number"],
    ["boolean", "Boolean"],
    ["date", "Date"],
    ["single_select", "Single select"],
    ["multi_select", "Multi select"],
    ["user", "User"]
  ]) {
    const option = document.createElement("option");
    option.value = fieldValue;
    option.textContent = label;
    option.selected = fieldValue === value;
    select.append(option);
  }
  return select;
}

function renderTicketFormOptions() {
  replaceSelectOptions(els.ticketParentID, "Parent epic", roadmapEpics(), (epic) => `${epic.key} ${epic.title}`);
  replaceSelectOptions(els.ticketComponentID, "Component", state.components, (component) => component.name);
  replaceSelectOptions(els.ticketVersionID, "Version", state.versions, (version) => `${version.name} (${version.state})`);
}

function replaceSelectOptions(select, emptyLabel, items, label, selectedID) {
  if (!select) {
    return;
  }
  const current = selectedID === undefined ? select.value : selectedID;
  select.replaceChildren();
  appendSelectOptions(select, emptyLabel, items, label, current);
}

function appendSelectOptions(select, emptyLabel, items, label, selectedID = "") {
  const empty = document.createElement("option");
  empty.value = "";
  empty.textContent = emptyLabel;
  empty.selected = !selectedID;
  select.append(empty);
  for (const item of items) {
    const option = document.createElement("option");
    option.value = item.id;
    option.textContent = label(item);
    option.selected = item.id === selectedID;
    select.append(option);
  }
}

function renderRoadmap() {
  if (!els.roadmap) {
    return;
  }
  els.roadmap.replaceChildren();
  if (!state.selectedProject) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "Select a project to view the roadmap";
    els.roadmap.append(empty);
    return;
  }
  if (!state.roadmap.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No epics";
    els.roadmap.append(empty);
    return;
  }

  const scheduled = state.roadmap.filter((item) => item.epic && (item.epic.start_date || item.epic.due_date));
  const unscheduled = state.roadmap.filter((item) => !item.epic || (!item.epic.start_date && !item.epic.due_date));
  els.roadmap.append(roadmapCapacityNode(state.roadmap));
  if (scheduled.length) {
    els.roadmap.append(roadmapTimelineNode(scheduled));
  }
  if (unscheduled.length) {
    const group = document.createElement("section");
    group.className = "roadmap-unscheduled";
    const heading = document.createElement("h3");
    heading.textContent = "Unscheduled";
    group.append(heading);
    for (const item of unscheduled) {
      group.append(roadmapNode(item));
    }
    els.roadmap.append(group);
  }
}

function renderRoadmapDependencies() {
  if (!els.roadmapDependencies) {
    return;
  }
  els.roadmapDependencies.replaceChildren();
  if (!state.selectedProject) {
    return;
  }
  const heading = document.createElement("h3");
  heading.textContent = "Dependencies";
  els.roadmapDependencies.append(heading);
  els.roadmapDependencies.append(roadmapDependencyFormNode());
  els.roadmapDependencies.append(roadmapDependencyOverviewNode(state.roadmapDependencies));
  if (!state.roadmapDependencies.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No roadmap dependencies";
    els.roadmapDependencies.append(empty);
    return;
  }
  const list = document.createElement("div");
  list.className = "roadmap-dependency-list";
  for (const dependency of state.roadmapDependencies) {
    list.append(roadmapDependencyNode(dependency));
  }
  els.roadmapDependencies.append(list);
}

function roadmapDependencyOverviewNode(dependencies) {
  const summary = roadmapDependencyOverviewSummary(dependencies);
  const section = document.createElement("section");
  section.className = "roadmap-dependency-overview";

  const heading = document.createElement("h4");
  heading.textContent = "Dependency overview";

  const chips = document.createElement("div");
  chips.className = "roadmap-dependency-overview-chips";
  for (const item of [
    `${summary.total} total`,
    `${summary.cross_epic} cross-epic`,
    `${summary.same_epic} same-epic`,
    `${summary.incomplete} incomplete`
  ]) {
    const chip = document.createElement("span");
    chip.textContent = item;
    chips.append(chip);
  }

  section.append(heading, chips);
  return section;
}

function roadmapDependencyOverviewSummary(dependencies) {
  const list = Array.isArray(dependencies) ? dependencies : [];
  const summary = {
    total: list.length,
    cross_epic: 0,
    same_epic: 0,
    incomplete: 0
  };
  for (const dependency of list) {
    const sourceEpicID = dependency.source_epic_id || "";
    const targetEpicID = dependency.target_epic_id || "";
    if (sourceEpicID && targetEpicID && sourceEpicID !== targetEpicID) {
      summary.cross_epic += 1;
    } else {
      summary.same_epic += 1;
    }
    const link = dependency.link || {};
    if (!link.id || !link.source || !link.source.id || !link.target || !link.target.id) {
      summary.incomplete += 1;
    }
  }
  return summary;
}

function roadmapDependencyNode(dependency) {
  const link = dependency.link || {};
  const source = link.source || {};
  const target = link.target || {};
  const article = document.createElement("article");
  article.className = "roadmap-dependency";

  const title = document.createElement("p");
  title.textContent = `${ticketKeyTitle(source)} ${ticketLinkTypeLabel(link.link_type)} ${ticketKeyTitle(target)}`;

  const meta = document.createElement("span");
  meta.textContent = [
    roadmapEpicName(dependency.source_epic_id),
    dependency.target_epic_id && dependency.target_epic_id !== dependency.source_epic_id ? `to ${roadmapEpicName(dependency.target_epic_id)}` : "",
    source.status ? `source ${source.status}` : "",
    target.status ? `target ${target.status}` : ""
  ].filter(Boolean).join(" / ");

  const remove = document.createElement("button");
  remove.type = "button";
  remove.dataset.deleteRoadmapDependencyId = link.id || dependency.id || "";
  remove.dataset.sourceTicketId = source.id || "";
  remove.dataset.targetTicketId = target.id || "";
  remove.textContent = "Remove";
  remove.disabled = !link.id || !source.id;

  article.append(title, meta, remove);
  return article;
}

function roadmapDependencyFormNode() {
  const form = document.createElement("form");
  form.className = "roadmap-dependency-form";
  form.dataset.roadmapDependencyForm = "true";

  const source = document.createElement("select");
  source.name = "source_ticket_id";
  source.required = true;
  source.setAttribute("aria-label", "Source roadmap issue");
  appendRoadmapDependencyOptions(source, "Source");

  const type = document.createElement("select");
  type.name = "link_type";
  type.setAttribute("aria-label", "Dependency type");
  for (const [value, label] of [["blocks", "Blocks"], ["is_blocked_by", "Is blocked by"], ["relates_to", "Relates to"]]) {
    const option = document.createElement("option");
    option.value = value;
    option.textContent = label;
    type.append(option);
  }

  const target = document.createElement("select");
  target.name = "target_ticket_id";
  target.required = true;
  target.setAttribute("aria-label", "Target roadmap issue");
  appendRoadmapDependencyOptions(target, "Target");

  const submit = document.createElement("button");
  submit.type = "submit";
  submit.textContent = "Add dependency";
  submit.disabled = roadmapDependencyCandidates().length < 2;

  form.append(source, type, target, submit);
  syncRoadmapDependencyTargetOptions(form);
  return form;
}

function syncRoadmapDependencyTargetOptions(form) {
  const source = form.querySelector("select[name='source_ticket_id']");
  const target = form.querySelector("select[name='target_ticket_id']");
  if (!source || !target) {
    return;
  }
  for (const option of target.options) {
    option.disabled = Boolean(option.value && option.value === source.value);
  }
  if (target.value && target.value === source.value) {
    target.value = "";
  }
}

function appendRoadmapDependencyOptions(select, emptyLabel) {
  const empty = document.createElement("option");
  empty.value = "";
  empty.textContent = emptyLabel;
  select.append(empty);
  for (const ticket of roadmapDependencyCandidates()) {
    const option = document.createElement("option");
    option.value = ticket.id;
    option.textContent = `${ticket.key || ticket.id} ${ticket.title || ""}`.trim();
    select.append(option);
  }
}

function roadmapDependencyCandidates() {
  const ids = new Set();
  const candidates = [];
  for (const item of state.roadmap) {
    const epic = item.epic;
    if (epic && !ids.has(epic.id)) {
      ids.add(epic.id);
      candidates.push(epic);
    }
  }
  for (const ticket of state.tickets) {
    if (ticket.parent_ticket_id && ids.has(ticket.parent_ticket_id) && !ids.has(ticket.id)) {
      ids.add(ticket.id);
      candidates.push(ticket);
    }
  }
  return candidates;
}

function ticketKeyTitle(ticket) {
  if (!ticket) {
    return "issue";
  }
  return [ticket.key || ticket.id || "issue", ticket.title || ""].filter(Boolean).join(" ");
}

function roadmapEpicName(epicID) {
  const item = state.roadmap.find((entry) => entry.epic && entry.epic.id === epicID);
  if (!item || !item.epic) {
    return epicID || "";
  }
  return `${item.epic.key || item.epic.id} ${item.epic.title || ""}`.trim();
}

function roadmapCapacityNode(items) {
  const summary = roadmapCapacitySummary(items);
  const section = document.createElement("section");
  section.className = "roadmap-capacity";

  const header = document.createElement("div");
  header.className = "roadmap-capacity-header";
  const title = document.createElement("h3");
  title.textContent = "Capacity summary";
  const meta = document.createElement("span");
  meta.textContent = `${summary.scheduled} scheduled / ${summary.unscheduled} unscheduled epics`;
  header.append(title, meta);

  const metrics = document.createElement("div");
  metrics.className = "roadmap-capacity-metrics";
  metrics.append(
    sprintMetricNode("Epics", summary.total),
    sprintMetricNode("Scheduled", summary.scheduled),
    sprintMetricNode("Open children", summary.openChildren),
    sprintMetricNode("Done children", `${summary.doneChildren}/${summary.childTickets}`),
    sprintMetricNode("Points", formatStoryPoints(summary.storyPointsTotal)),
    sprintMetricNode("Remaining pts", formatStoryPoints(summary.storyPointsRemaining)),
    sprintMetricNode("Unestimated", summary.unestimatedChildren)
  );

  const insights = document.createElement("div");
  insights.className = "roadmap-capacity-insights";
  for (const insight of roadmapCapacityInsightItems(summary)) {
    const item = document.createElement("span");
    item.textContent = insight;
    insights.append(item);
  }

  const buckets = document.createElement("div");
  buckets.className = "roadmap-capacity-buckets";
  if (summary.buckets.length) {
    for (const bucket of summary.buckets) {
      buckets.append(roadmapCapacityBucketNode(bucket));
    }
  } else {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "Schedule epics to see monthly capacity.";
    buckets.append(empty);
  }

  section.append(header, metrics, insights, buckets);
  return section;
}

function roadmapCapacitySummary(items) {
  const summary = {
    total: 0,
    scheduled: 0,
    unscheduled: 0,
    childTickets: 0,
    doneChildren: 0,
    openChildren: 0,
    storyPointsTotal: 0,
    storyPointsDone: 0,
    storyPointsRemaining: 0,
    unestimatedChildren: 0,
    buckets: []
  };
  const buckets = new Map();
  for (const item of items) {
    if (!item || !item.epic) {
      summary.unscheduled += 1;
      continue;
    }
    summary.total += 1;
    const work = roadmapCapacityItemWork(item);
    summary.childTickets += work.childTickets;
    summary.doneChildren += work.doneChildren;
    summary.openChildren += work.openChildren;
    summary.storyPointsTotal += work.storyPointsTotal;
    summary.storyPointsDone += work.storyPointsDone;
    summary.storyPointsRemaining += work.storyPointsRemaining;
    summary.unestimatedChildren += work.unestimatedChildren;
    const epic = item.epic;
    const bucketKey = roadmapCapacityBucketKey(epic);
    if (!bucketKey) {
      summary.unscheduled += 1;
      continue;
    }
    summary.scheduled += 1;
    if (!buckets.has(bucketKey)) {
      buckets.set(bucketKey, {
        key: bucketKey,
        label: roadmapCapacityBucketLabel(bucketKey),
        epics: 0,
        childTickets: 0,
        doneChildren: 0,
        openChildren: 0,
        storyPointsTotal: 0,
        storyPointsDone: 0,
        storyPointsRemaining: 0,
        unestimatedChildren: 0
      });
    }
    const bucket = buckets.get(bucketKey);
    bucket.epics += 1;
    bucket.childTickets += work.childTickets;
    bucket.doneChildren += work.doneChildren;
    bucket.openChildren += work.openChildren;
    bucket.storyPointsTotal += work.storyPointsTotal;
    bucket.storyPointsDone += work.storyPointsDone;
    bucket.storyPointsRemaining += work.storyPointsRemaining;
    bucket.unestimatedChildren += work.unestimatedChildren;
  }
  summary.buckets = Array.from(buckets.values()).sort((a, b) => a.key.localeCompare(b.key));
  summary.atRiskBuckets = summary.buckets.filter(roadmapCapacityBucketAtRisk).length;
  summary.busiestBucket = summary.buckets.reduce((busiest, bucket) => {
    if (!busiest || bucket.openChildren > busiest.openChildren) {
      return bucket;
    }
    if (busiest.openChildren === bucket.openChildren && bucket.storyPointsRemaining > busiest.storyPointsRemaining) {
      return bucket;
    }
    return busiest;
  }, null);
  return summary;
}

function roadmapCapacityInsightItems(summary) {
  const items = [
    `${summary.buckets.length} visible months`,
    `${summary.atRiskBuckets || 0} at-risk months`
  ];
  if (summary.busiestBucket) {
    const points = summary.busiestBucket.storyPointsTotal > 0
      ? ` / ${formatStoryPoints(summary.busiestBucket.storyPointsRemaining)} pts`
      : "";
    items.push(`busiest ${summary.busiestBucket.label}: ${summary.busiestBucket.openChildren} open${points}`);
  }
  return items;
}

function roadmapCapacityItemWork(item) {
  const epic = item.epic || {};
  const children = roadmapCapacityChildTickets(epic.id);
  if (children.length) {
    const work = {
      childTickets: children.length,
      doneChildren: 0,
      openChildren: 0,
      storyPointsTotal: 0,
      storyPointsDone: 0,
      storyPointsRemaining: 0,
      unestimatedChildren: 0
    };
    for (const ticket of children) {
      const done = ticket.status === "done";
      if (done) {
        work.doneChildren += 1;
      } else {
        work.openChildren += 1;
      }
      if (ticket.story_points === null || ticket.story_points === undefined || ticket.story_points === "") {
        work.unestimatedChildren += 1;
      } else {
        const points = Number(ticket.story_points || 0);
        work.storyPointsTotal += points;
        if (done) {
          work.storyPointsDone += points;
        } else {
          work.storyPointsRemaining += points;
        }
      }
    }
    return work;
  }
  const progress = item.progress || {};
  const childTickets = Number(progress.total || 0);
  const doneChildren = Number(progress.done || 0);
  return {
    childTickets,
    doneChildren,
    openChildren: Math.max(childTickets - doneChildren, 0),
    storyPointsTotal: 0,
    storyPointsDone: 0,
    storyPointsRemaining: 0,
    unestimatedChildren: childTickets
  };
}

function roadmapCapacityChildTickets(epicID) {
  if (!epicID || !Array.isArray(state.roadmapCapacityTickets)) {
    return [];
  }
  return state.roadmapCapacityTickets.filter((ticket) => ticket.parent_ticket_id === epicID);
}

function roadmapCapacityBucketKey(epic) {
  const date = dateToUTC(epic.due_date || epic.start_date);
  if (!date) {
    return "";
  }
  const year = date.getUTCFullYear();
  const month = String(date.getUTCMonth() + 1).padStart(2, "0");
  return `${year}-${month}`;
}

function roadmapCapacityBucketLabel(key) {
  const date = dateToUTC(`${key}-01`);
  if (!date) {
    return key;
  }
  return date.toLocaleString(undefined, { month: "short", year: "numeric", timeZone: "UTC" });
}

function roadmapCapacityBucketNode(bucket) {
  const ratio = roadmapCapacityBucketCompletionRatio(bucket);
  const atRisk = roadmapCapacityBucketAtRisk(bucket);
  const article = document.createElement("article");
  article.className = atRisk ? "roadmap-capacity-bucket is-at-risk" : "roadmap-capacity-bucket";

  const title = document.createElement("strong");
  title.textContent = bucket.label;
  const meta = document.createElement("span");
  meta.textContent = `${bucket.epics} epics / ${bucket.openChildren} open children`;
  const progress = document.createElement("div");
  progress.className = "roadmap-progress";
  progress.setAttribute("aria-label", `${Math.round(ratio * 100)}% complete`);
  const fill = document.createElement("span");
  fill.style.width = `${Math.round(ratio * 100)}%`;
  progress.append(fill);
  const detail = document.createElement("small");
  const pointText = bucket.storyPointsTotal > 0
    ? `${formatStoryPoints(bucket.storyPointsRemaining)} pts remaining`
    : `${bucket.unestimatedChildren} unestimated`;
  detail.textContent = `${bucket.doneChildren}/${bucket.childTickets} child tickets done / ${pointText}${atRisk ? " / capacity risk" : ""}`;
  article.append(title, meta, progress, detail);
  return article;
}

function roadmapCapacityBucketCompletionRatio(bucket) {
  return bucket.childTickets > 0 ? bucket.doneChildren / bucket.childTickets : 1;
}

function roadmapCapacityBucketAtRisk(bucket) {
  const ratio = roadmapCapacityBucketCompletionRatio(bucket);
  return bucket.openChildren >= 8 || (bucket.childTickets > 0 && ratio < 0.35);
}

function roadmapTimelineNode(items) {
  const group = document.createElement("section");
  group.className = "roadmap-timeline";
  const heading = document.createElement("h3");
  heading.textContent = "Timeline";
  const track = document.createElement("div");
  track.className = "roadmap-timeline-track";
  track.dataset.roadmapTimelineTrack = "true";
  const bounds = roadmapTimelineBounds(items);
  for (const item of items) {
    const row = roadmapNode(item);
    row.classList.add("roadmap-timeline-item");
    const epic = item.epic || {};
    const start = dateToUTC(epic.start_date || epic.due_date);
    const due = dateToUTC(epic.due_date || epic.start_date);
    const offsetDays = Math.max(daysBetween(bounds.start, start), 0);
    const durationDays = Math.max(daysBetween(start, due) + 1, 1);
    const span = Math.max(daysBetween(bounds.start, bounds.end) + 1, 1);
    row.style.setProperty("--roadmap-offset", `${Math.round((offsetDays / span) * 100)}%`);
    row.style.setProperty("--roadmap-width", `${Math.max(12, Math.round((durationDays / span) * 100))}%`);
    row.draggable = true;
    row.dataset.roadmapDragId = epic.id || "";
    row.dataset.roadmapStartDate = formatISODate(start);
    row.dataset.roadmapDueDate = formatISODate(due);
    row.dataset.roadmapBoundsStart = formatISODate(bounds.start);
    row.dataset.roadmapBoundsEnd = formatISODate(bounds.end);
    track.append(row);
  }
  group.append(heading, track);
  return group;
}

function roadmapNode(item) {
  const article = document.createElement("article");
  article.className = "roadmap-item";

  const epic = item.epic;
  const progress = item.progress || { total: 0, done: 0, by_status: {} };
  const percent = progress.total > 0 ? Math.round((progress.done / progress.total) * 100) : 0;

  const title = document.createElement("p");
  title.textContent = `${epic.key} ${epic.title}`;

  const meta = document.createElement("span");
  meta.textContent = [
    epic.start_date || epic.due_date ? dateRange(epic.start_date, epic.due_date) : "unscheduled",
    `${progress.done}/${progress.total} done`,
    epic.priority,
    epic.version_id ? versionName(epic.version_id) : ""
  ].filter(Boolean).join(" / ");

  const bar = document.createElement("div");
  bar.className = "roadmap-progress";
  bar.setAttribute("aria-label", `${percent}% complete`);
  const fill = document.createElement("span");
  fill.style.width = `${percent}%`;
  bar.append(fill);

  const counts = document.createElement("small");
  counts.textContent = Object.entries(progress.by_status || {})
    .map(([status, count]) => `${status}: ${count}`)
    .join(" / ") || "No child tickets";

  article.append(title, meta, bar, counts);
  if (!epic.start_date && !epic.due_date) {
    article.append(roadmapQuickScheduleNode(epic));
  }
  article.append(roadmapScheduleFormNode(epic));
  return article;
}

function roadmapQuickScheduleNode(epic) {
  const actions = document.createElement("div");
  actions.className = "roadmap-quick-actions";
  const schedule = document.createElement("button");
  schedule.type = "button";
  schedule.dataset.roadmapQuickScheduleId = epic.id;
  schedule.textContent = "Schedule this week";
  actions.append(schedule);
  return actions;
}

function roadmapScheduleFormNode(epic) {
  const form = document.createElement("form");
  form.className = "roadmap-schedule-form";
  form.dataset.roadmapScheduleForm = epic.id;
  form.append(
    inputNode("start_date", epic.start_date || "", "start date", "date"),
    inputNode("due_date", epic.due_date || "", "due date", "date")
  );
  const save = document.createElement("button");
  save.type = "submit";
  save.textContent = "Schedule";
  form.append(save);
  return form;
}

function roadmapTimelineBounds(items) {
  const dates = [];
  for (const item of items) {
    const epic = item.epic || {};
    if (epic.start_date) {
      dates.push(dateToUTC(epic.start_date));
    }
    if (epic.due_date) {
      dates.push(dateToUTC(epic.due_date));
    }
  }
  const valid = dates.filter(Boolean);
  if (!valid.length) {
    const today = dateToUTC(new Date().toISOString().slice(0, 10));
    return { start: today, end: today };
  }
  return {
    start: new Date(Math.min(...valid.map((date) => date.getTime()))),
    end: new Date(Math.max(...valid.map((date) => date.getTime())))
  };
}

function dateToUTC(value) {
  if (!value) {
    return null;
  }
  const date = new Date(`${value}T00:00:00Z`);
  return Number.isNaN(date.getTime()) ? null : date;
}

function todayISODate() {
  return new Date().toISOString().slice(0, 10);
}

function todayLocalISODate() {
  const today = new Date();
  const year = today.getFullYear();
  const month = String(today.getMonth() + 1).padStart(2, "0");
  const day = String(today.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function addDays(date, days) {
  const next = new Date(date.getTime());
  next.setUTCDate(next.getUTCDate() + days);
  return next;
}

function addDaysISO(value, days) {
  const date = dateToUTC(value);
  return date ? formatISODate(addDays(date, days)) : "";
}

function formatISODate(date) {
  return date.toISOString().slice(0, 10);
}

function daysBetween(start, end) {
  if (!start || !end) {
    return 0;
  }
  return Math.round((end.getTime() - start.getTime()) / 86400000);
}

function renderSearchResults() {
  if (!els.searchResults) {
    return;
  }
  els.searchResultCount.textContent = String(state.searchResults.length);
  els.searchResults.replaceChildren();
  renderSearchPagination();
  if (!state.searchResults.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No search results";
    els.searchResults.append(empty);
    return;
  }
  for (const ticket of state.searchResults) {
    const row = document.createElement("article");
    row.className = "search-result-item";

    const title = document.createElement("p");
    title.textContent = `${ticket.key} ${ticket.title}`;

    const meta = document.createElement("span");
    meta.textContent = [ticket.status, ticket.type, ticket.priority, storyPointLabel(ticket.story_points)].filter(Boolean).join(" / ") || "Ticket";

    row.append(title, meta);
    els.searchResults.append(row);
  }
}

function renderSearchPagination() {
  if (!els.searchPagination) {
    return;
  }
  els.searchPagination.replaceChildren();

  const summary = document.createElement("span");
  summary.textContent = `Page ${state.searchCursorIndex + 1}`;

  const previous = document.createElement("button");
  previous.type = "button";
  previous.dataset.searchPrevious = "true";
  previous.disabled = state.searchCursorIndex <= 0;
  previous.textContent = "Previous";

  const next = document.createElement("button");
  next.type = "button";
  next.dataset.searchNext = "true";
  next.disabled = !state.searchNextCursor;
  next.textContent = "Next";

  els.searchPagination.append(summary, previous, next);
}

function renderCustomFieldSearchControls() {
  if (!els.customFieldSearchControls) {
    return;
  }
  els.customFieldSearchControls.replaceChildren();
  if (!state.selectedProject) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "Select a project to build custom-field filters";
    els.customFieldSearchControls.append(empty);
    return;
  }
  if (!state.customFields.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No custom fields for search filters";
    els.customFieldSearchControls.append(empty);
    return;
  }

  const form = document.createElement("form");
  form.className = "custom-field-search-form";
  form.dataset.customFieldSearchForm = "true";

  const field = document.createElement("select");
  field.name = "field_key";
  field.setAttribute("aria-label", "Custom field");
  for (const item of state.customFields) {
    const option = document.createElement("option");
    option.value = item.key;
    option.textContent = `${item.name || item.key} (${item.field_type || "text"})`;
    field.append(option);
  }

  const operator = document.createElement("select");
  operator.name = "operator";
  operator.setAttribute("aria-label", "Custom field operator");

  const value = document.createElement("span");
  value.className = "custom-field-search-value";
  value.dataset.customFieldSearchValue = "true";

  const add = document.createElement("button");
  add.type = "submit";
  add.textContent = "Add filter";

  form.append(field, operator, value, add);
  els.customFieldSearchControls.append(form);
  renderCustomFieldSearchValueControl(form);
}

function renderCustomFieldSearchValueControl(form) {
  const field = selectedCustomSearchField(form);
  const operator = form.querySelector("select[name='operator']");
  const value = form.querySelector("[data-custom-field-search-value]");
  if (!field || !operator || !value) {
    return;
  }
  const currentOperator = operator.value;
  operator.replaceChildren();
  for (const item of customFieldSearchOperators(field)) {
    const option = document.createElement("option");
    option.value = item.value;
    option.textContent = item.label;
    option.selected = item.value === currentOperator;
    operator.append(option);
  }
  if (![...operator.options].some((item) => item.selected)) {
    operator.selectedIndex = 0;
  }
  value.replaceChildren(customFieldSearchValueControl(field, operator.value));
}

function customFieldSearchOperators(field) {
  switch (field.field_type) {
    case "number":
    case "date":
      return [
        { value: "==", label: "=" },
        { value: "!=", label: "!=" },
        { value: ">=", label: ">=" },
        { value: "<=", label: "<=" },
        { value: ">", label: ">" },
        { value: "<", label: "<" }
      ];
    case "boolean":
      return [
        { value: "==", label: "is" },
        { value: "!=", label: "is not" }
      ];
    case "multi_select":
      return [
        { value: "contains", label: "contains" },
        { value: "not_contains", label: "does not contain" }
      ];
    default:
      return [
        { value: "==", label: "=" },
        { value: "!=", label: "!=" }
      ];
  }
}

function customFieldSearchValueControl(field) {
  if (field.field_type === "boolean") {
    const select = document.createElement("select");
    select.name = "value";
    appendCustomFieldOption(select, "true", "true");
    appendCustomFieldOption(select, "false", "false");
    return select;
  }
  if (field.field_type === "single_select" || field.field_type === "multi_select") {
    const select = document.createElement("select");
    select.name = "value";
    for (const option of field.options || []) {
      appendCustomFieldOption(select, option, option);
    }
    return select;
  }
  const input = document.createElement("input");
  input.name = "value";
  input.placeholder = field.name || field.key;
  if (field.field_type === "number") {
    input.type = "number";
    input.step = "any";
  } else if (field.field_type === "date") {
    input.type = "date";
  } else {
    input.type = "text";
  }
  return input;
}

function customFieldSearchExpression(form) {
  const field = selectedCustomSearchField(form);
  const operator = formData(form).operator || "==";
  const rawValue = formData(form).value;
  if (!field || rawValue === undefined || rawValue === "") {
    return "";
  }
  const literal = customFieldSearchLiteral(rawValue, field.field_type);
  if (!literal) {
    return "";
  }
  const key = `custom.${field.key}`;
  if (operator === "contains" || operator === "not_contains") {
    const expression = `${customFieldSearchLiteral(rawValue, "text")} in ${key}`;
    return operator === "not_contains" ? `!(${expression})` : expression;
  }
  return `${key} ${operator} ${literal}`;
}

function selectedCustomSearchField(form) {
  const key = formData(form).field_key || "";
  return state.customFields.find((field) => field.key === key) || state.customFields[0] || null;
}

function customFieldSearchLiteral(value, fieldType) {
  if (fieldType === "number") {
    const number = Number(value);
    return Number.isFinite(number) ? String(number) : "";
  }
  if (fieldType === "boolean") {
    return value === true || value === "true" ? "true" : "false";
  }
  return `"${String(value).replace(/\\/g, "\\\\").replace(/"/g, "\\\"")}"`;
}

function appendSearchFilter(current, expression) {
  const filter = String(current || "").trim();
  if (!filter) {
    return expression;
  }
  return `(${filter}) && (${expression})`;
}

function renderSavedViews() {
  if (!els.savedViews) {
    return;
  }
  els.savedViews.replaceChildren();
  renderSavedViewPagination();
  if (!state.savedViews.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No saved views";
    els.savedViews.append(empty);
    return;
  }
  els.savedViews.append(savedViewOverviewNode(state.savedViews));
  for (const view of state.savedViews) {
    els.savedViews.append(savedViewNode(view));
  }
}

function savedViewOverviewNode(views) {
  const summary = savedViewOverviewSummary(views);
  const section = document.createElement("section");
  section.className = "saved-view-overview";

  const heading = document.createElement("h3");
  heading.textContent = "Saved-view overview";

  const chips = document.createElement("div");
  chips.className = "saved-view-overview-chips";
  const items = [
    `${summary.total} visible`,
    `${summary.pinned} pinned`,
    ...summary.modes.map((item) => `mode ${item.key}: ${item.count}`)
  ];
  for (const item of items) {
    const chip = document.createElement("span");
    chip.textContent = item;
    chips.append(chip);
  }

  section.append(heading, chips, savedViewScopeBreakdownNode(summary.scopes));
  return section;
}

function savedViewScopeBreakdownNode(scopes) {
  const group = document.createElement("div");
  group.className = "saved-view-scope-breakdown";

  const label = document.createElement("strong");
  label.textContent = "Scopes";

  const chips = document.createElement("div");
  chips.className = "saved-view-scope-chips";
  const items = Array.isArray(scopes) && scopes.length ? scopes : [{ key: "user", count: 0 }];
  for (const item of items) {
    const chip = document.createElement("span");
    chip.textContent = `${item.key}: ${item.count}`;
    chips.append(chip);
  }

  group.append(label, chips);
  return group;
}

function savedViewOverviewSummary(views) {
  const list = Array.isArray(views) ? views : [];
  const scopes = new Map();
  const modes = new Map();
  let pinned = 0;
  for (const view of list) {
    if (view.pinned) {
      pinned += 1;
    }
    const scope = view.scope_type || "user";
    const mode = view.display_mode || "list";
    scopes.set(scope, (scopes.get(scope) || 0) + 1);
    modes.set(mode, (modes.get(mode) || 0) + 1);
  }
  const toItems = (map) => Array.from(map.entries())
    .map(([key, count]) => ({ key, count }))
    .sort((left, right) => left.key.localeCompare(right.key));
  return {
    total: list.length,
    pinned,
    scopes: toItems(scopes),
    modes: toItems(modes)
  };
}

function renderSavedViewPagination() {
  if (!els.savedViewPagination) {
    return;
  }
  els.savedViewPagination.replaceChildren();

  const summary = document.createElement("span");
  const first = state.savedViewOffset + 1;
  const last = state.savedViewOffset + state.savedViews.length;
  summary.textContent = state.savedViews.length ? `${first}-${last}` : "0";

  const previous = document.createElement("button");
  previous.type = "button";
  previous.dataset.savedViewPrevious = "true";
  previous.disabled = state.savedViewOffset <= 0;
  previous.textContent = "Previous";

  const next = document.createElement("button");
  next.type = "button";
  next.dataset.savedViewNext = "true";
  next.disabled = !state.savedViewHasMore;
  next.textContent = "Next";

  els.savedViewPagination.append(summary, previous, next);
}

function savedViewNode(view) {
  const article = document.createElement("article");
  article.className = "saved-view-item";
  if (view.pinned) {
    article.classList.add("is-pinned");
  }

  const name = document.createElement("p");
  name.textContent = view.name;

  const actions = document.createElement("div");
  actions.className = "saved-view-actions";

  const apply = document.createElement("button");
  apply.type = "button";
  apply.dataset.applySavedViewId = view.id;
  apply.textContent = "Apply";

  const edit = document.createElement("button");
  edit.type = "button";
  edit.dataset.editSavedViewId = view.id;
  edit.textContent = "Edit";

  const remove = document.createElement("button");
  remove.type = "button";
  remove.dataset.deleteSavedViewId = view.id;
  remove.textContent = "Delete";

  actions.append(apply, edit, remove);
  article.append(name, savedViewMetadataNode(view), actions);
  return article;
}

function savedViewMetadataNode(view, tagName = "div") {
  const metadata = document.createElement(tagName);
  metadata.className = "saved-view-metadata";

  for (const item of savedViewMetadataItems(view)) {
    const chip = document.createElement("span");
    chip.textContent = item;
    metadata.append(chip);
  }

  return metadata;
}

function savedViewMetadataItems(view) {
  const query = view.query || {};
  const sortCount = Array.isArray(view.sort) ? view.sort.length : 0;
  const columnCount = Array.isArray(view.columns) ? view.columns.length : 0;

  return [
    query.text ? `text: ${query.text}` : "",
    query.filter ? `filter: ${query.filter}` : "",
    sortCount ? `sort ${sortCount}` : "",
    columnCount ? `columns ${columnCount}` : "",
    view.display_mode ? `mode ${view.display_mode}` : "",
    view.group_by ? `group ${view.group_by}` : "",
    view.scope_type ? `scope ${view.scope_type}` : "",
    view.project_id ? `project ${view.project_id}` : "",
    view.pinned ? "pinned" : ""
  ].filter(Boolean);
}

function renderTokens() {
  if (!els.accountUser || !els.apiTokens || !els.createdToken) {
    return;
  }
  if (!state.user) {
    els.accountUser.textContent = "Signed out";
    els.createdToken.hidden = true;
    els.createdToken.replaceChildren();
    els.apiTokens.replaceChildren();
    return;
  }

  const displayName = state.user.display_name || state.user.username || state.user.id;
  els.accountUser.textContent = state.user.id ? `${displayName} (${state.user.id})` : displayName;

  els.createdToken.replaceChildren();
  if (state.createdToken && state.createdToken.token) {
    const label = document.createElement("p");
    label.textContent = "Shown once. Store this token before leaving the page.";
    const secret = document.createElement("pre");
    secret.textContent = state.createdToken.token;
    els.createdToken.append(label, secret);
    els.createdToken.hidden = false;
  } else {
    els.createdToken.hidden = true;
  }

  els.apiTokens.replaceChildren();
  if (!state.tokens.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No API tokens";
    els.apiTokens.append(empty);
    return;
  }
  for (const token of state.tokens) {
    els.apiTokens.append(tokenNode(token));
  }
}

function tokenNode(token) {
  const article = document.createElement("article");
  article.className = "token-item";
  if (token.revoked_at) {
    article.classList.add("is-revoked");
  }

  const body = document.createElement("div");
  body.className = "token-item-body";

  const name = document.createElement("p");
  name.textContent = token.name || "API token";

  const meta = document.createElement("span");
  meta.textContent = [
    token.created_at ? `created ${formatDateTime(token.created_at)}` : "",
    token.last_used_at ? `last used ${formatDateTime(token.last_used_at)}` : "never used",
    token.revoked_at ? `revoked ${formatDateTime(token.revoked_at)}` : ""
  ].filter(Boolean).join(" / ");

  body.append(name, meta);

  const revoke = document.createElement("button");
  revoke.type = "button";
  revoke.dataset.revokeTokenId = token.id;
  revoke.disabled = Boolean(token.revoked_at);
  revoke.textContent = token.revoked_at ? "Revoked" : "Revoke";

  article.append(body, revoke);
  return article;
}

function renderRBAC() {
  if (!els.rbacUsers || !els.rbacPermissionForm || !els.rbacPermissions) {
    return;
  }
  renderRBACFormOptions();
  renderAdminList(els.rbacUsers, state.rbac.users, (user) => {
    const meta = [user.display_name, user.disabled ? "disabled" : "enabled"].filter(Boolean).join(" / ");
    const row = adminItemNode(user.username || user.id, meta);
    const actions = document.createElement("div");
    actions.className = "admin-actions";

    const toggle = document.createElement("button");
    toggle.type = "button";
    toggle.dataset.rbacUserId = user.id;
    toggle.dataset.rbacUserDisabled = user.disabled ? "false" : "true";
    toggle.textContent = user.disabled ? "Enable" : "Disable";

    const remove = document.createElement("button");
    remove.type = "button";
    remove.dataset.deleteRbacUserId = user.id;
    remove.textContent = "Delete";

    actions.append(toggle, remove);
    row.append(actions);
    return row;
  });
  renderAdminList(els.rbacGroups, state.rbac.groups, (group) => {
    const members = state.rbac.members[group.id] || [];
    const row = adminItemNode(group.display_name || group.name || group.id, `${group.name || group.id} / ${members.length} members`);
    if (members.length) {
      const list = document.createElement("div");
      list.className = "member-list";
      for (const member of members) {
        const item = document.createElement("span");
        item.textContent = member.username || member.id;
        const remove = document.createElement("button");
        remove.type = "button";
        remove.dataset.removeGroupMember = "true";
        remove.dataset.groupId = group.id;
        remove.dataset.userId = member.id;
        remove.setAttribute("aria-label", `Remove ${member.username || member.id}`);
        remove.textContent = "x";
        item.append(remove);
        list.append(item);
      }
      row.append(list);
    }
    return row;
  });
  renderAdminList(els.rbacRoles, state.rbac.roles, (role) => {
    return adminItemNode(role.name, `${role.permissions.length} permissions`);
  });
  renderAdminList(els.rbacBindings, state.rbac.bindings, (binding) => {
    const subject = `${binding.subject_type}:${binding.subject_id}`;
    const scope = binding.scope === "project" ? `project ${binding.project_id}` : "global";
    const row = adminItemNode(binding.role_name, `${subject} / ${scope}`);
    const actions = document.createElement("div");
    actions.className = "admin-actions";
    const remove = document.createElement("button");
    remove.type = "button";
    remove.dataset.deleteBindingId = binding.id;
    remove.textContent = "Delete";
    actions.append(remove);
    row.append(actions);
    return row;
  });
  renderRBACPermissions();
}

function renderRBACFormOptions() {
  replaceSelectOptions(els.rbacMemberForm.elements.group_id, "Group", state.rbac.groups, (group) => group.display_name || group.name || group.id);
  replaceSelectOptions(els.rbacMemberForm.elements.user_id, "User", state.rbac.users, (user) => user.username || user.id);
  replaceSelectOptions(els.rbacBindingForm.elements.role_name, "Role", state.rbac.roles.map((role) => ({ id: role.name, name: role.name })), (role) => role.name);
  renderBindingSubjectOptions();
  replaceSelectOptions(els.rbacBindingForm.elements.project_id, "Global scope", state.projects, (project) => `${project.key} ${project.name}`);
  replaceSelectOptions(els.rbacPermissionForm.elements.user_id, "User", state.rbac.users, (user) => user.username || user.id);
  replaceSelectOptions(els.rbacPermissionForm.elements.project_id, "Global scope", state.projects, (project) => `${project.key} ${project.name}`);
  const projectScoped = els.rbacPermissionForm.elements.scope.value === "project";
  els.rbacPermissionForm.elements.project_id.disabled = !projectScoped;
}

function renderBindingSubjectOptions() {
  const type = els.rbacBindingForm.elements.subject_type.value || "user";
  const items = type === "group" ? state.rbac.groups : state.rbac.users;
  replaceSelectOptions(
    els.rbacBindingForm.elements.subject_id,
    type === "group" ? "Group" : "User",
    items,
    (item) => type === "group" ? (item.display_name || item.name || item.id) : (item.username || item.id)
  );
}

function renderRBACPermissions() {
  if (!els.rbacPermissions) {
    return;
  }
  els.rbacPermissions.replaceChildren();
  if (state.rbac.effectivePermissionsError) {
    const error = document.createElement("p");
    error.className = "muted";
    error.textContent = state.rbac.effectivePermissionsError;
    els.rbacPermissions.append(error);
    return;
  }
  const effective = state.rbac.effectivePermissions;
  if (!effective) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "Choose a user and scope to inspect permissions";
    els.rbacPermissions.append(empty);
    return;
  }
  const summary = document.createElement("p");
  summary.className = "permission-summary";
  summary.textContent = [
    userLabel(effective.user_id),
    effective.scope === "project" ? `project ${projectLabel(effective.project_id)}` : "global"
  ].filter(Boolean).join(" / ");
  els.rbacPermissions.append(summary);
  if (!effective.permissions.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No permissions granted";
    els.rbacPermissions.append(empty);
    return;
  }
  const chips = document.createElement("div");
  chips.className = "permission-chip-list";
  for (const permission of effective.permissions) {
    const chip = document.createElement("span");
    chip.className = "permission-chip";
    chip.textContent = permission;
    chips.append(chip);
  }
  els.rbacPermissions.append(chips);
}

function renderSettings() {
  if (!els.settingsForm || !els.preferenceForm || !els.projectPreferenceForm || !els.auditForm || !els.openRouterProviderForm || !els.notificationDestinationForm || !els.notificationPolicyForm || !els.notificationHookForm) {
    return;
  }

  if (state.settings) {
    setFormValue(els.settingsForm, "attachment_max_size_bytes", String(state.settings.attachment_max_size_bytes || 0));
    setFormValue(els.settingsForm, "attachment_allowed_content_types", state.settings.attachment_allowed_content_types.join(", "));
    setFormValue(els.settingsForm, "webhook_allowed_base_urls", state.settings.webhook_allowed_base_urls.join(", "));
    setFormValue(els.settingsForm, "system_health_note", state.settings.system_health_note || "");
    setFormChecked(els.settingsForm, "demo_warning_enabled", state.settings.demo_warning_enabled);
    setFormChecked(els.settingsForm, "backup_enabled", state.settings.backup_enabled);
    els.settingsForm.hidden = false;
    els.settingsStatus.textContent = [
      state.settings.attachment_policy_active ? "attachment policy active" : "",
      state.settings.webhook_allowlist_active ? "webhook allowlist active" : "",
      state.settings.demo_warning_visible ? "demo warning visible" : "",
      state.settings.backup_available ? "backup available" : ""
    ].filter(Boolean).join(" / ") || "No active global policy flags";
  } else {
    els.settingsForm.hidden = true;
    els.settingsStatus.textContent = state.settingsError || "Global settings require settings management permission";
  }

  const prefs = state.notificationPreferences;
  if (prefs) {
    for (const key of preferenceKeys()) {
      setFormChecked(els.preferenceForm, key, Boolean(prefs[key]));
    }
    els.preferenceStatus.textContent = prefs.customized ? "Customized preferences" : "Using defaults until saved";
  } else {
    els.preferenceStatus.textContent = "Notification preferences are not loaded";
  }

  renderProjectNotificationPreferences();
  renderNotificationDeliveries();
  renderAuditLog();
  renderOpenRouterProviders();
  renderNotificationDestinations();
  renderNotificationPolicies();
  renderNotificationHooks();
}

function renderProjectNotificationPreferences() {
  if (!els.projectPreferenceForm || !els.projectPreferenceProject) {
    return;
  }
  replaceSelectOptions(els.projectPreferenceProject, "Project", state.projects, (project) => `${project.key} ${project.name}`);
  if (state.selectedProject && state.projects.some((project) => project.id === state.selectedProject.id)) {
    els.projectPreferenceProject.value = state.selectedProject.id;
  }
  const prefs = state.projectNotificationPreferences;
  if (prefs) {
    for (const key of preferenceKeys()) {
      setFormChecked(els.projectPreferenceForm, key, Boolean(prefs[key]));
    }
    els.projectPreferenceStatus.textContent = prefs.customized ? "Customized project defaults" : "Using default project notification policy";
  } else {
    for (const key of preferenceKeys()) {
      setFormChecked(els.projectPreferenceForm, key, false);
    }
    els.projectPreferenceStatus.textContent = state.projectNotificationPreferencesError || "Project defaults require notification management permission";
  }
}

function renderNotificationDeliveries() {
  if (!els.notificationDeliveryForm || !els.notificationDeliveries) {
    return;
  }
  replaceSelectOptions(els.notificationDeliveryProject, "Project", state.projects, (project) => `${project.key} ${project.name}`);
  if (state.selectedProject && state.projects.some((project) => project.id === state.selectedProject.id)) {
    els.notificationDeliveryProject.value = state.selectedProject.id;
  }
  renderNotificationDeliverySummary();
  els.notificationDeliveries.replaceChildren();
  if (state.notificationDeliveriesError) {
    const error = document.createElement("p");
    error.className = "muted";
    error.textContent = state.notificationDeliveriesError;
    els.notificationDeliveries.append(error);
    return;
  }
  if (!state.notificationDeliveries.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No notification deliveries";
    els.notificationDeliveries.append(empty);
    return;
  }
  for (const delivery of state.notificationDeliveries) {
    els.notificationDeliveries.append(notificationDeliveryNode(delivery));
  }
}

function renderNotificationDeliverySummary() {
  if (!els.notificationDeliverySummary) {
    return;
  }
  els.notificationDeliverySummary.replaceChildren();
  if (state.notificationDeliveriesError) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "Delivery health summary unavailable";
    els.notificationDeliverySummary.append(empty);
    return;
  }
  if (!state.notificationDeliveries.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No loaded deliveries to summarize";
    els.notificationDeliverySummary.append(empty);
    return;
  }

  const summary = notificationDeliverySummary();
  const metrics = document.createElement("div");
  metrics.className = "notification-delivery-metrics";
  for (const metric of [
    ["Loaded", summary.total],
    ["Queued", summary.counts.queued],
    ["Sending", summary.counts.sending],
    ["Delivered", summary.counts.delivered],
    ["Failed", summary.counts.failed],
    ["Canceled", summary.counts.canceled],
    ["Retryable", summary.retryable],
    ["Delivered %", `${summary.deliveredPercent}%`],
    ["Oldest", summary.oldestDelivery ? notificationDeliverySummaryTime(summary.oldestDelivery) : "n/a"],
    ["Newest", summary.newestDelivery ? notificationDeliverySummaryTime(summary.newestDelivery) : "n/a"]
  ]) {
    const item = document.createElement("span");
    item.textContent = `${metric[0]}: ${metric[1]}`;
    metrics.append(item);
  }
  els.notificationDeliverySummary.append(metrics);

  if (summary.latestFailure) {
    const failure = document.createElement("p");
    failure.className = "notification-delivery-latest-failure";
    failure.textContent = [
      "Latest failure",
      summary.latestFailure.updated_at ? formatDateTime(summary.latestFailure.updated_at) : "",
      summary.latestFailure.destination_name || summary.latestFailure.destination_id || "",
      summary.latestFailure.last_error || summary.latestFailure.state
    ].filter(Boolean).join(" / ");
    els.notificationDeliverySummary.append(failure);
  }
}

function notificationDeliverySummary() {
  const counts = { queued: 0, sending: 0, delivered: 0, failed: 0, canceled: 0 };
  let latestFailure = null;
  let oldestDelivery = null;
  let newestDelivery = null;
  for (const delivery of state.notificationDeliveries) {
    if (Object.prototype.hasOwnProperty.call(counts, delivery.state)) {
      counts[delivery.state] += 1;
    }
    if ((delivery.state === "failed" || delivery.state === "canceled") && (!latestFailure || deliveryUpdatedAt(delivery) > deliveryUpdatedAt(latestFailure))) {
      latestFailure = delivery;
    }
    if (!oldestDelivery || deliveryUpdatedAt(delivery) < deliveryUpdatedAt(oldestDelivery)) {
      oldestDelivery = delivery;
    }
    if (!newestDelivery || deliveryUpdatedAt(delivery) > deliveryUpdatedAt(newestDelivery)) {
      newestDelivery = delivery;
    }
  }
  const total = state.notificationDeliveries.length;
  return {
    total,
    counts,
    retryable: counts.failed + counts.canceled,
    deliveredPercent: total ? Math.round((counts.delivered / total) * 100) : 0,
    latestFailure,
    oldestDelivery,
    newestDelivery
  };
}

function deliveryUpdatedAt(delivery) {
  const value = Date.parse(delivery.updated_at || delivery.last_attempt_at || delivery.created_at || "");
  return Number.isFinite(value) ? value : 0;
}

function notificationDeliverySummaryTime(delivery) {
  return formatDateTime(delivery.updated_at || delivery.last_attempt_at || delivery.created_at || "");
}

function notificationDeliveryNode(delivery) {
  const article = document.createElement("article");
  article.className = "notification-delivery-item";

  const header = document.createElement("div");
  header.className = "notification-delivery-header";
  const title = document.createElement("p");
  title.textContent = delivery.message || delivery.event_type || delivery.id;
  const stateLabel = document.createElement("span");
  stateLabel.className = delivery.state === "delivered" ? "delivery-state is-delivered" : "delivery-state";
  stateLabel.textContent = delivery.state || "unknown";
  header.append(title, stateLabel);

  const meta = document.createElement("span");
  meta.textContent = [
    delivery.scope_type === "project" ? `project ${projectLabel(delivery.project_id)}` : "global",
    delivery.policy_name || delivery.policy_id || "",
    delivery.destination_name || delivery.destination_id || "",
    delivery.destination_service || "",
    `attempts ${delivery.attempt_count}/${delivery.max_attempts || "?"}`,
    delivery.updated_at ? `updated ${formatDateTime(delivery.updated_at)}` : ""
  ].filter(Boolean).join(" / ");

  const actions = document.createElement("div");
  actions.className = "notification-delivery-actions";
  const retry = document.createElement("button");
  retry.type = "button";
  retry.dataset.retryNotificationDeliveryId = delivery.id;
  retry.disabled = !(delivery.state === "failed" || delivery.state === "canceled");
  retry.textContent = "Retry";
  actions.append(retry);

  article.append(header, meta, actions);
  if (delivery.last_error) {
    const error = document.createElement("pre");
    error.className = "destination-error";
    error.textContent = delivery.last_error;
    article.append(error);
  }
  const payload = document.createElement("pre");
  payload.className = "notification-delivery-payload";
  payload.textContent = JSON.stringify(delivery.payload || {}, null, 2);
  article.append(payload);
  return article;
}

function renderOpenRouterProviders() {
  if (!els.openRouterProviderStatus || !els.openRouterProviders) {
    return;
  }
  els.openRouterProviders.replaceChildren();
  if (state.openRouterProvidersError) {
    els.openRouterProviderForm.hidden = true;
    els.openRouterProviderStatus.textContent = state.openRouterProvidersError;
    return;
  }
  els.openRouterProviderForm.hidden = false;
  els.openRouterProviderStatus.textContent = state.openRouterProviders.length
    ? `${state.openRouterProviders.length} OpenRouter providers`
    : "No OpenRouter providers";
  if (!state.openRouterProviders.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "Create a provider to use AI engines";
    els.openRouterProviders.append(empty);
    return;
  }
  for (const provider of state.openRouterProviders) {
    els.openRouterProviders.append(openRouterProviderNode(provider));
  }
}

function openRouterProviderNode(provider) {
  const article = document.createElement("article");
  article.className = "openrouter-provider-item";

  const header = document.createElement("div");
  header.className = "openrouter-provider-header";
  const title = document.createElement("p");
  title.textContent = provider.name || provider.id;
  const stateLabel = document.createElement("span");
  stateLabel.className = provider.enabled ? "provider-state" : "provider-state is-disabled";
  stateLabel.textContent = provider.enabled ? "enabled" : "disabled";
  header.append(title, stateLabel);

  const meta = document.createElement("span");
  meta.textContent = [
    provider.api_key_set ? "key set" : "missing key",
    provider.allowed_models.length ? `${provider.allowed_models.length} allowed models` : "all models allowed",
    provider.updated_at ? `updated ${formatDateTime(provider.updated_at)}` : ""
  ].filter(Boolean).join(" / ");

  const form = document.createElement("form");
  form.className = "openrouter-provider-edit-form";
  form.dataset.openrouterProviderForm = provider.id;

  const name = inputNode("name", provider.name, "name");
  const model = inputNode("default_model", provider.default_model, "default model");
  const allowed = document.createElement("textarea");
  allowed.name = "allowed_models";
  allowed.rows = 2;
  allowed.placeholder = "allowed models";
  allowed.value = provider.allowed_models.join(", ");
  const timeout = inputNode("default_timeout_seconds", String(provider.default_timeout_seconds || 30), "timeout seconds", "number");
  timeout.min = "1";
  timeout.max = "300";
  timeout.step = "1";
  const tokens = inputNode("max_output_tokens", String(provider.max_output_tokens || 2048), "max output tokens", "number");
  tokens.min = "1";
  tokens.max = "32000";
  tokens.step = "1";
  const key = inputNode("api_key", "", "rotate api key", "password");
  key.autocomplete = "off";

  const enabled = document.createElement("label");
  enabled.className = "inline-toggle";
  const enabledInput = document.createElement("input");
  enabledInput.name = "enabled";
  enabledInput.type = "checkbox";
  enabledInput.checked = provider.enabled;
  enabled.append(enabledInput, " Enabled");

  const save = document.createElement("button");
  save.type = "submit";
  save.textContent = "Save";

  const remove = document.createElement("button");
  remove.type = "button";
  remove.dataset.deleteOpenrouterProviderId = provider.id;
  remove.textContent = "Delete";

  form.append(name, model, allowed, timeout, tokens, key, enabled, save, remove);
  article.append(header, meta, form);
  return article;
}

function renderNotificationDestinations() {
  if (!els.notificationDestinationStatus || !els.notificationDestinations) {
    return;
  }
  renderNotificationDestinationProjectOptions();
  els.notificationDestinations.replaceChildren();
  if (state.notificationDestinationsError) {
    els.notificationDestinationForm.hidden = true;
    els.notificationDestinationStatus.textContent = state.notificationDestinationsError;
    return;
  }
  els.notificationDestinationForm.hidden = false;
  els.notificationDestinationStatus.textContent = state.notificationDestinations.length
    ? `${state.notificationDestinations.length} notification destinations`
    : "No notification destinations";
  if (!state.notificationDestinations.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "Create a Shoutrrr destination for notification hooks and policies";
    els.notificationDestinations.append(empty);
    return;
  }
  for (const destination of state.notificationDestinations) {
    els.notificationDestinations.append(notificationDestinationNode(destination));
  }
}

function renderNotificationDestinationProjectOptions() {
  if (!els.notificationDestinationProject) {
    return;
  }
  replaceSelectOptions(
    els.notificationDestinationProject,
    "Project",
    state.projects,
    (project) => `${project.key} ${project.name}`
  );
  if (state.selectedProject && state.projects.some((project) => project.id === state.selectedProject.id)) {
    els.notificationDestinationProject.value = state.selectedProject.id;
  }
  const scopeType = els.notificationDestinationScope ? els.notificationDestinationScope.value : "global";
  els.notificationDestinationProject.disabled = scopeType !== "project";
}

function notificationDestinationNode(destination) {
  const article = document.createElement("article");
  article.className = "notification-destination-item";

  const header = document.createElement("div");
  header.className = "notification-destination-header";
  const title = document.createElement("p");
  title.textContent = destination.name || destination.id;
  const stateLabel = document.createElement("span");
  stateLabel.className = destination.enabled ? "destination-state" : "destination-state is-disabled";
  stateLabel.textContent = destination.enabled ? "enabled" : "disabled";
  header.append(title, stateLabel);

  const meta = document.createElement("span");
  meta.textContent = [
    destination.scope_type === "project" ? `project ${projectLabel(destination.project_id)}` : "global",
    destination.type || "unknown service",
    destination.url_set ? "url set" : "missing url",
    destination.last_delivery_status ? `last ${destination.last_delivery_status}` : "",
    destination.last_delivery_at ? formatDateTime(destination.last_delivery_at) : ""
  ].filter(Boolean).join(" / ");

  const form = document.createElement("form");
  form.className = "notification-destination-edit-form";
  form.dataset.notificationDestinationForm = destination.id;
  form.dataset.notificationDestinationScope = destination.scope_type;

  const name = inputNode("name", destination.name, "name");
  const url = inputNode("shoutrrr_url", "", "rotate Shoutrrr URL", "password");
  url.autocomplete = "off";
  const testMessage = inputNode("test_message", "", "test message");
  const enabled = document.createElement("label");
  enabled.className = "inline-toggle";
  const enabledInput = document.createElement("input");
  enabledInput.name = "enabled";
  enabledInput.type = "checkbox";
  enabledInput.checked = destination.enabled;
  enabled.append(enabledInput, " Enabled");

  const save = document.createElement("button");
  save.type = "submit";
  save.textContent = "Save";

  const test = document.createElement("button");
  test.type = "button";
  test.dataset.testNotificationDestinationId = destination.id;
  test.textContent = "Test";

  const remove = document.createElement("button");
  remove.type = "button";
  remove.dataset.deleteNotificationDestinationId = destination.id;
  remove.textContent = "Delete";

  form.append(name, url, testMessage, enabled, save, test, remove);
  article.append(header, meta, form);
  if (destination.last_error) {
    const error = document.createElement("pre");
    error.className = "destination-error";
    error.textContent = destination.last_error;
    article.append(error);
  }
  return article;
}

function renderNotificationPolicies() {
  if (!els.notificationPolicyStatus || !els.notificationPolicies) {
    return;
  }
  renderNotificationPolicyProjectOptions();
  renderNotificationPolicyDestinationOptions();
  renderNotificationHookPreviewPolicyOptions();
  els.notificationPolicies.replaceChildren();
  if (state.notificationPoliciesError) {
    els.notificationPolicyForm.hidden = true;
    els.notificationPolicyStatus.textContent = state.notificationPoliciesError;
    return;
  }
  els.notificationPolicyForm.hidden = false;
  els.notificationPolicyStatus.textContent = state.notificationPolicies.length
    ? `${state.notificationPolicies.length} notification policies`
    : "No notification policies";
  if (!state.notificationPolicies.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "Create a policy to route notification events to destinations";
    els.notificationPolicies.append(empty);
    return;
  }
  for (const policy of state.notificationPolicies) {
    els.notificationPolicies.append(notificationPolicyNode(policy));
  }
}

function renderNotificationPolicyProjectOptions() {
  if (!els.notificationPolicyProject) {
    return;
  }
  replaceSelectOptions(
    els.notificationPolicyProject,
    "Project",
    state.projects,
    (project) => `${project.key} ${project.name}`
  );
  if (state.selectedProject && state.projects.some((project) => project.id === state.selectedProject.id)) {
    els.notificationPolicyProject.value = state.selectedProject.id;
  }
  const scopeType = els.notificationPolicyScope ? els.notificationPolicyScope.value : "global";
  els.notificationPolicyProject.disabled = scopeType !== "project";
}

function renderNotificationPolicyDestinationOptions() {
  if (!els.notificationPolicyDestinations || !els.notificationPolicyForm) {
    return;
  }
  const data = formData(els.notificationPolicyForm);
  const scopeType = data.scope_type || "global";
  const projectID = data.project_id || selectedNotificationPolicyProjectID();
  renderDestinationMultiSelect(els.notificationPolicyDestinations, [], scopeType, projectID);
}

function notificationPolicyNode(policy) {
  const article = document.createElement("article");
  article.className = "notification-policy-item";

  const header = document.createElement("div");
  header.className = "notification-policy-header";
  const title = document.createElement("p");
  title.textContent = policy.name || policy.id;
  const stateLabel = document.createElement("span");
  stateLabel.className = policy.enabled ? "policy-state" : "policy-state is-disabled";
  stateLabel.textContent = policy.enabled ? "enabled" : "disabled";
  header.append(title, stateLabel);

  const meta = document.createElement("span");
  meta.textContent = [
    policy.scope_type === "project" ? `project ${projectLabel(policy.project_id)}` : "global",
    policy.event_types.length ? policy.event_types.join(", ") : "no events",
    policy.destination_ids.length ? policy.destination_ids.map(destinationLabel).join(", ") : "no destinations"
  ].filter(Boolean).join(" / ");

  const form = document.createElement("form");
  form.className = "notification-policy-edit-form";
  form.dataset.notificationPolicyForm = policy.id;
  form.dataset.notificationPolicyScope = policy.scope_type;

  const name = inputNode("name", policy.name, "name");
  const events = document.createElement("textarea");
  events.name = "event_types";
  events.rows = 2;
  events.placeholder = "event types";
  events.value = policy.event_types.join(", ");
  const destinations = document.createElement("select");
  destinations.name = "destination_ids";
  destinations.multiple = true;
  destinations.size = 3;
  renderDestinationMultiSelect(destinations, policy.destination_ids, policy.scope_type, policy.project_id);
  const enabled = document.createElement("label");
  enabled.className = "inline-toggle";
  const enabledInput = document.createElement("input");
  enabledInput.name = "enabled";
  enabledInput.type = "checkbox";
  enabledInput.checked = policy.enabled;
  enabled.append(enabledInput, " Enabled");

  const save = document.createElement("button");
  save.type = "submit";
  save.textContent = "Save";

  const remove = document.createElement("button");
  remove.type = "button";
  remove.dataset.deleteNotificationPolicyId = policy.id;
  remove.textContent = "Delete";

  form.append(name, events, destinations, enabled, save, remove);
  article.append(header, meta, form);
  return article;
}

function renderNotificationHooks() {
  if (!els.notificationHookStatus || !els.notificationHooks) {
    return;
  }
  renderNotificationHookProjectOptions();
  renderNotificationHookEngineFields();
  renderNotificationHookPreviewPolicyOptions();
  renderNotificationHookPreviewDestinationOptions();
  renderNotificationHookPreview();
  els.notificationHooks.replaceChildren();
  if (state.notificationHooksError) {
    els.notificationHookForm.hidden = true;
    els.notificationHookPreviewForm.hidden = true;
    els.notificationHookStatus.textContent = state.notificationHooksError;
    return;
  }
  els.notificationHookForm.hidden = false;
  els.notificationHookPreviewForm.hidden = false;
  els.notificationHookStatus.textContent = state.notificationHooks.length
    ? `${state.notificationHooks.length} notification hooks`
    : "No notification hooks";
  if (!state.notificationHooks.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "Create a hook to transform or suppress notification delivery plans";
    els.notificationHooks.append(empty);
    return;
  }
  for (const hook of state.notificationHooks) {
    els.notificationHooks.append(notificationHookNode(hook));
  }
}

function renderNotificationHookProjectOptions() {
  if (!els.notificationHookProject) {
    return;
  }
  replaceSelectOptions(
    els.notificationHookProject,
    "Project",
    state.projects,
    (project) => `${project.key} ${project.name}`
  );
  if (state.selectedProject && state.projects.some((project) => project.id === state.selectedProject.id)) {
    els.notificationHookProject.value = state.selectedProject.id;
  }
  const scopeType = els.notificationHookScope ? els.notificationHookScope.value : "global";
  els.notificationHookProject.disabled = scopeType !== "project";
}

function renderNotificationHookEngineFields() {
  const type = els.notificationHookEngineType ? els.notificationHookEngineType.value : "lua";
  document.querySelectorAll("[data-notification-hook-engine-field]").forEach((field) => {
    field.hidden = field.dataset.notificationHookEngineField !== type;
  });
}

function renderNotificationHookPreview() {
  if (!els.notificationHookPreviewOutput) {
    return;
  }
  els.notificationHookPreviewOutput.textContent = JSON.stringify(notificationHookPreviewDisplay(state.notificationHookPreview), null, 2);
}

function renderNotificationHookPreviewPolicyOptions() {
  if (!els.notificationHookPreviewPolicy) {
    return;
  }
  const current = els.notificationHookPreviewPolicy.value;
  els.notificationHookPreviewPolicy.replaceChildren();
  const empty = document.createElement("option");
  empty.value = "";
  empty.textContent = "Preview policy";
  empty.selected = !current;
  els.notificationHookPreviewPolicy.append(empty);
  for (const policy of state.notificationPolicies) {
    const option = document.createElement("option");
    option.value = policy.id;
    option.textContent = [
      policy.name || policy.id,
      policy.scope_type === "project" ? projectLabel(policy.project_id) : "global"
    ].filter(Boolean).join(" / ");
    option.selected = policy.id === current;
    els.notificationHookPreviewPolicy.append(option);
  }
}

function renderNotificationHookPreviewDestinationOptions() {
  if (!els.notificationHookPreviewDestinations || !els.notificationHookPreviewForm) {
    return;
  }
  const data = formData(els.notificationHookPreviewForm);
  const policy = state.notificationPolicies.find((item) => item.id === data.policy_id);
  const projectID = policy ? policy.project_id : selectedNotificationHookProjectID();
  const scopeType = policy ? policy.scope_type : (projectID ? "project" : "global");
  renderDestinationMultiSelect(
    els.notificationHookPreviewDestinations,
    selectedFormValues(els.notificationHookPreviewForm, "destination_ids"),
    scopeType,
    projectID
  );
}

function applyNotificationHookPreviewPolicy(policyID) {
  const policy = state.notificationPolicies.find((item) => item.id === policyID);
  if (!policy || !els.notificationHookPreviewForm) {
    renderNotificationHookPreviewDestinationOptions();
    return;
  }
  setFormValue(els.notificationHookPreviewForm, "event_type", policy.event_types[0] || "");
  renderDestinationMultiSelect(els.notificationHookPreviewDestinations, policy.destination_ids, policy.scope_type, policy.project_id);
}

function notificationHookPreviewDisplay(preview) {
  if (!preview || !preview.status) {
    return {};
  }
  const plan = preview.status.plan || {};
  return {
    state: preview.status.state || "",
    suppressed: Boolean(plan.suppressed),
    plan: {
      event_type: plan.event_type || "",
      project: plan.project_id ? projectLabel(plan.project_id) : "",
      subject_type: plan.subject_type || "",
      subject_id: plan.subject_id || "",
      message: plan.message || "",
      destination_ids: plan.destination_ids || [],
      destinations: (plan.destination_ids || []).map(destinationLabel),
      payload: plan.payload || {}
    },
    output: preview.status.output || {},
    error: preview.status.error || "",
    run_id: preview.metadata ? preview.metadata.run_id || "" : ""
  };
}

function notificationHookNode(hook) {
  const article = document.createElement("article");
  article.className = "notification-hook-item";

  const header = document.createElement("div");
  header.className = "notification-hook-header";
  const title = document.createElement("p");
  title.textContent = hook.name || hook.id;
  const stateLabel = document.createElement("span");
  stateLabel.className = hook.enabled ? "hook-state" : "hook-state is-disabled";
  stateLabel.textContent = hook.enabled ? "enabled" : "disabled";
  header.append(title, stateLabel);

  const meta = document.createElement("span");
  meta.textContent = [
    hook.scope_type === "project" ? `project ${projectLabel(hook.project_id)}` : "global",
    hook.event_types.length ? hook.event_types.join(", ") : "all events",
    hook.engine.type || "engine",
    hook.actor_user_id ? `actor ${hook.actor_user_id}` : "",
    hook.last_error ? `error: ${hook.last_error}` : ""
  ].filter(Boolean).join(" / ");

  const actions = document.createElement("div");
  actions.className = "ticket-hook-actions";

  const preview = document.createElement("button");
  preview.type = "button";
  preview.dataset.previewNotificationHookId = hook.id;
  preview.textContent = "Preview";

  const runs = document.createElement("button");
  runs.type = "button";
  runs.dataset.loadNotificationHookRunsId = hook.id;
  runs.textContent = "Runs";

  const toggle = document.createElement("button");
  toggle.type = "button";
  toggle.dataset.toggleNotificationHookId = hook.id;
  toggle.dataset.notificationHookEnabled = hook.enabled ? "false" : "true";
  toggle.textContent = hook.enabled ? "Disable" : "Enable";

  const remove = document.createElement("button");
  remove.type = "button";
  remove.dataset.deleteNotificationHookId = hook.id;
  remove.textContent = "Delete";

  actions.append(preview, runs, toggle, remove);
  article.append(header, meta, actions);

  const runsList = state.notificationHookRuns[hook.id] || [];
  if (runsList.length) {
    article.append(notificationHookRunListNode(runsList));
  }
  return article;
}

function notificationHookRunListNode(runs) {
  const list = document.createElement("div");
  list.className = "notification-hook-run-list";
  for (const run of runs) {
    const item = document.createElement("article");
    item.className = "cron-run-item";
    const summary = document.createElement("span");
    summary.textContent = [
      run.state || "queued",
      run.trigger_type,
      run.created_at ? formatDateTime(run.created_at) : "",
      run.error ? `error: ${run.error}` : ""
    ].filter(Boolean).join(" / ");
    const output = document.createElement("pre");
    output.textContent = JSON.stringify(run.output || {}, null, 2);
    item.append(summary, output);
    list.append(item);
  }
  return list;
}

function projectLabel(projectID) {
  const project = state.projects.find((item) => item.id === projectID);
  return project ? `${project.key} ${project.name}` : projectID;
}

function userLabel(userID) {
  const user = state.rbac.users.find((item) => item.id === userID);
  return user ? (user.username || user.display_name || user.id) : userID;
}

function destinationLabel(destinationID) {
  const destination = state.notificationDestinations.find((item) => item.id === destinationID);
  return destination ? `${destination.name || destination.id} (${destination.id})` : destinationID;
}

function availableNotificationDestinations(scopeType, projectID) {
  return state.notificationDestinations.filter((destination) => {
    if (destination.deleted) {
      return false;
    }
    if (scopeType !== "project") {
      return destination.scope_type === "global";
    }
    return destination.scope_type === "global" || destination.project_id === projectID;
  });
}

function renderDestinationMultiSelect(select, selectedIDs, scopeType, projectID) {
  if (!select) {
    return;
  }
  const selected = new Set(selectedIDs || []);
  const destinations = availableNotificationDestinations(scopeType, projectID);
  select.replaceChildren();
  for (const destination of destinations) {
    const option = document.createElement("option");
    option.value = destination.id;
    option.textContent = destinationLabel(destination.id);
    option.selected = selected.has(destination.id);
    select.append(option);
  }
  for (const id of selected) {
    if (destinations.some((destination) => destination.id === id)) {
      continue;
    }
    const option = document.createElement("option");
    option.value = id;
    option.textContent = destinationLabel(id);
    option.selected = true;
    select.append(option);
  }
}

function inputNode(name, value, placeholder, type = "text") {
  const input = document.createElement("input");
  input.name = name;
  input.type = type;
  input.placeholder = placeholder;
  input.value = value || "";
  return input;
}

function renderCronJobs() {
  if (!els.cronJobs || !els.cronJobProject) {
    return;
  }
  replaceSelectOptions(els.cronJobProject, "Project", state.projects, (project) => `${project.key} ${project.name}`);
  if (state.selectedProject && state.projects.some((project) => project.id === state.selectedProject.id)) {
    els.cronJobProject.value = state.selectedProject.id;
  }
  renderCronJobEngineFields();

  els.cronJobs.replaceChildren();
  if (state.cronJobsError) {
    els.cronJobStatus.textContent = state.cronJobsError;
    return;
  }
  const projectID = selectedCronJobProjectID();
  els.cronJobStatus.textContent = projectID
    ? `${state.cronJobs.length} cron jobs`
    : "Choose a project to manage cron jobs";
  if (!projectID || !state.cronJobs.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = projectID ? "No cron jobs for this project" : "Select a project first";
    els.cronJobs.append(empty);
    return;
  }
  for (const job of state.cronJobs) {
    els.cronJobs.append(cronJobNode(job));
  }
}

function renderCronJobEngineFields() {
  const type = els.cronJobEngineType ? els.cronJobEngineType.value : "lua";
  document.querySelectorAll("[data-cron-job-engine-field]").forEach((field) => {
    field.hidden = field.dataset.cronJobEngineField !== type;
  });
}

function renderCronJobEditEngineFields(form) {
  const type = form && form.elements.engine_type ? form.elements.engine_type.value : "lua";
  form.querySelectorAll("[data-cron-job-edit-engine-field]").forEach((field) => {
    field.hidden = field.dataset.cronJobEditEngineField !== type;
  });
}

function cronJobNode(job) {
  const article = document.createElement("article");
  article.className = "cron-job-item";

  const header = document.createElement("div");
  header.className = "ticket-hook-item-header";
  const title = document.createElement("p");
  title.textContent = job.name || job.id;
  const stateLabel = document.createElement("span");
  stateLabel.className = job.enabled ? "hook-state" : "hook-state is-disabled";
  stateLabel.textContent = job.enabled ? "enabled" : "disabled";
  header.append(title, stateLabel);

  const meta = document.createElement("span");
  meta.textContent = [
    job.schedule,
    job.timezone,
    job.engine.type,
    job.owner_user_id ? `owner ${job.owner_user_id}` : "",
    job.next_run_at ? `next ${formatDateTime(job.next_run_at)}` : "",
    job.last_run_status ? `last ${job.last_run_status}` : "",
    job.last_error ? `error: ${job.last_error}` : ""
  ].filter(Boolean).join(" / ");

  const actions = document.createElement("div");
  actions.className = "ticket-hook-actions";

  const runs = document.createElement("button");
  runs.type = "button";
  runs.dataset.loadCronRunsId = job.id;
  runs.textContent = "Runs";

  const run = document.createElement("button");
  run.type = "button";
  run.dataset.runCronJobId = job.id;
  run.textContent = "Run now";

  const toggle = document.createElement("button");
  toggle.type = "button";
  toggle.dataset.toggleCronJobId = job.id;
  toggle.dataset.cronJobEnabled = job.enabled ? "false" : "true";
  toggle.textContent = job.enabled ? "Disable" : "Enable";

  const remove = document.createElement("button");
  remove.type = "button";
  remove.dataset.deleteCronJobId = job.id;
  remove.textContent = "Delete";

  actions.append(runs, run, toggle, remove);
  article.append(header, meta, actions);
  article.append(cronJobEditForm(job));

  const jobRuns = state.cronRuns[job.id] || [];
  if (jobRuns.length) {
    article.append(cronRunListNode(job.id, jobRuns));
  }
  return article;
}

function cronJobEditForm(job) {
  const form = document.createElement("form");
  form.className = "cron-job-edit-form";
  form.dataset.cronJobForm = job.id;

  form.append(
    inputNode("name", job.name || "", "name"),
    inputNode("schedule", job.schedule || "", "cron schedule"),
    inputNode("timezone", job.timezone || "UTC", "timezone"),
    inputNode("owner_user_id", job.owner_user_id || "", "owner user id"),
    automationEngineSelect("engine_type", [
      ["lua", "Lua"],
      ["ai", "OpenRouter AI"]
    ], job.engine.type || "lua")
  );

  const enabled = document.createElement("label");
  enabled.className = "inline-toggle";
  const enabledInput = document.createElement("input");
  enabledInput.name = "enabled";
  enabledInput.type = "checkbox";
  enabledInput.checked = job.enabled;
  enabled.append(enabledInput, " Enabled");

  const lua = document.createElement("label");
  lua.dataset.cronJobEditEngineField = "lua";
  lua.append("Lua Script", automationTextarea("script", job.engine.script || "", 4));

  const ai = document.createElement("div");
  ai.dataset.cronJobEditEngineField = "ai";
  const prompt = document.createElement("label");
  prompt.append("AI Prompt", automationTextarea("prompt", job.engine.prompt || "", 4));
  const provider = document.createElement("label");
  provider.append("Provider ID", inputNode("provider_id", job.engine.provider_id || "", "openrouter_provider_..."));
  ai.append(prompt, provider);

  const save = document.createElement("button");
  save.type = "submit";
  save.textContent = "Save cron job";

  form.append(enabled, lua, ai, save);
  renderCronJobEditEngineFields(form);
  return form;
}

function automationEngineSelect(name, options, selectedValue) {
  const select = document.createElement("select");
  select.name = name;
  for (const [value, label] of options) {
    const option = document.createElement("option");
    option.value = value;
    option.textContent = label;
    option.selected = value === selectedValue;
    select.append(option);
  }
  return select;
}

function automationTextarea(name, value, rows) {
  const textarea = document.createElement("textarea");
  textarea.name = name;
  textarea.rows = rows;
  textarea.spellcheck = false;
  textarea.value = value || "";
  return textarea;
}

function cronRunListNode(jobID, runs) {
  const list = document.createElement("div");
  list.className = "cron-run-list";
  const filterKey = automationRunFilterKey("cron", jobID);
  const visibleRuns = filterAutomationRuns(runs, state.automationRunFilters[filterKey] || "all");
  list.append(automationRunSummaryNode(runs, visibleRuns, filterKey));
  if (!visibleRuns.length) {
    list.append(automationRunEmptyNode());
    return list;
  }
  for (const run of visibleRuns) {
    const item = document.createElement("article");
    item.className = "cron-run-item";
    const summary = document.createElement("span");
    summary.textContent = [
      run.state || "queued",
      run.trigger_type,
      run.created_at ? formatDateTime(run.created_at) : "",
      automationRunDurationLabel(run),
      run.error ? `error: ${run.error}` : ""
    ].filter(Boolean).join(" / ");
    const output = document.createElement("pre");
    output.textContent = JSON.stringify(run.output || {}, null, 2);
    item.append(summary, output);
    list.append(item);
  }
  return list;
}

function automationRunSummaryNode(runs, visibleRuns, filterKey) {
  const summary = summarizeAutomationRuns(runs);
  const section = document.createElement("div");
  section.className = "automation-run-summary";
  section.append(automationRunFilterNode(filterKey, visibleRuns.length, summary.total));

  for (const metric of [
    ["total", summary.total],
    ["completed", summary.completed],
    ["failed", summary.failed],
    ["active", summary.active],
    ["completion rate", summary.completionRateLabel],
    ["failure rate", summary.failureRateLabel],
    ["avg duration", summary.averageDurationLabel],
    ["max duration", summary.maxDurationLabel],
    ["oldest", summary.oldestRunLabel],
    ["newest", summary.newestRunLabel]
  ]) {
    if (metric[1] === "") {
      continue;
    }
    const item = document.createElement("span");
    item.textContent = `${metric[0]} ${metric[1]}`;
    section.append(item);
  }

  if (summary.latestFailure) {
    const error = document.createElement("span");
    error.className = "automation-run-summary-error";
    error.textContent = `latest failure: ${summary.latestFailure}`;
    section.append(error);
  }

  for (const [trigger, count] of Object.entries(summary.triggerCounts)) {
    const item = document.createElement("span");
    item.textContent = `trigger ${trigger} ${count}`;
    section.append(item);
  }

  return section;
}

function automationRunFilterNode(filterKey, visibleCount, totalCount) {
  const label = document.createElement("label");
  label.className = "automation-run-filter";
  const select = document.createElement("select");
  select.dataset.automationRunFilter = filterKey;
  const selectedValue = state.automationRunFilters[filterKey] || "all";
  for (const [value, text] of [
    ["all", "All"],
    ["active", "Active"],
    ["completed", "Completed"],
    ["failed", "Failed"]
  ]) {
    const option = document.createElement("option");
    option.value = value;
    option.textContent = text;
    option.selected = value === selectedValue;
    select.append(option);
  }
  const count = document.createElement("span");
  count.textContent = `showing ${visibleCount} of ${totalCount}`;
  label.append("Show", select, count);
  return label;
}

function automationRunEmptyNode() {
  const empty = document.createElement("p");
  empty.className = "muted";
  empty.textContent = "No runs match this filter";
  return empty;
}

function automationRunFilterKey(kind, id) {
  return `${kind}:${id || ""}`;
}

function handleAutomationRunFilterChange(event, renderFn) {
  const select = event.target.closest("[data-automation-run-filter]");
  if (!select) {
    return false;
  }
  state.automationRunFilters[select.dataset.automationRunFilter] = select.value || "all";
  renderFn();
  return true;
}

function filterAutomationRuns(runs, filter) {
  if (!filter || filter === "all") {
    return runs || [];
  }
  return (runs || []).filter((run) => automationRunStateGroup(run) === filter);
}

function automationRunStateGroup(run) {
  const stateValue = String(run && run.state ? run.state : "queued").toLowerCase();
  if ((run && run.error) || ["failed", "error", "canceled", "cancelled"].includes(stateValue)) {
    return "failed";
  }
  if (["completed", "complete", "success", "succeeded", "done"].includes(stateValue)) {
    return "completed";
  }
  return "active";
}

function summarizeAutomationRuns(runs) {
  const summary = {
    total: 0,
    completed: 0,
    failed: 0,
    active: 0,
    latestFailure: "",
    completionRateLabel: "",
    failureRateLabel: "",
    averageDurationLabel: "",
    maxDurationLabel: "",
    oldestRunLabel: "",
    newestRunLabel: "",
    triggerCounts: {}
  };
  let durationCount = 0;
  let totalDurationMs = 0;
  let maxDurationMs = 0;
  let oldestRunMs = 0;
  let newestRunMs = 0;
  for (const run of runs || []) {
    if (!run) {
      continue;
    }
    summary.total += 1;
    const stateGroup = automationRunStateGroup(run);
    const triggerType = run.trigger_type || "unknown";
    summary.triggerCounts[triggerType] = (summary.triggerCounts[triggerType] || 0) + 1;
    if (stateGroup === "failed") {
      summary.failed += 1;
      if (!summary.latestFailure) {
        summary.latestFailure = run.error || run.state || "failed";
      }
    } else if (stateGroup === "completed") {
      summary.completed += 1;
    } else {
      summary.active += 1;
    }
    const durationMs = automationRunDurationMs(run);
    if (durationMs > 0) {
      durationCount += 1;
      totalDurationMs += durationMs;
      maxDurationMs = Math.max(maxDurationMs, durationMs);
    }
    const createdMs = Date.parse(run.created_at || "");
    if (Number.isFinite(createdMs)) {
      oldestRunMs = oldestRunMs === 0 ? createdMs : Math.min(oldestRunMs, createdMs);
      newestRunMs = Math.max(newestRunMs, createdMs);
    }
  }
  if (durationCount > 0) {
    summary.averageDurationLabel = formatDuration(Math.round(totalDurationMs / durationCount));
    summary.maxDurationLabel = formatDuration(maxDurationMs);
  }
  if (summary.total > 0) {
    summary.completionRateLabel = formatRunRate(summary.completed, summary.total);
    summary.failureRateLabel = formatRunRate(summary.failed, summary.total);
  }
  if (oldestRunMs > 0) {
    summary.oldestRunLabel = formatDateTime(new Date(oldestRunMs).toISOString());
  }
  if (newestRunMs > 0) {
    summary.newestRunLabel = formatDateTime(new Date(newestRunMs).toISOString());
  }
  return summary;
}

function formatRunRate(count, total) {
  return `${Math.round((count / total) * 100)}%`;
}

function automationRunDurationLabel(run) {
  const durationMs = automationRunDurationMs(run);
  return durationMs > 0 ? `duration ${formatDuration(durationMs)}` : "";
}

function automationRunDurationMs(run) {
  if (!run) {
    return 0;
  }
  const started = Date.parse(run.started_at || run.created_at || "");
  const finished = Date.parse(run.finished_at || "");
  if (!Number.isFinite(started) || !Number.isFinite(finished) || finished <= started) {
    return 0;
  }
  return finished - started;
}

function formatDuration(durationMs) {
  const totalSeconds = Math.max(1, Math.round(durationMs / 1000));
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;
  if (hours > 0) {
    return `${hours}h ${minutes}m`;
  }
  if (minutes > 0) {
    return `${minutes}m ${seconds}s`;
  }
  return `${seconds}s`;
}

function renderWebhooks() {
  if (!els.webhooks || !els.webhookProject) {
    return;
  }
  replaceSelectOptions(els.webhookProject, "Project", state.projects, (project) => `${project.key} ${project.name}`);
  if (state.selectedProject && state.projects.some((project) => project.id === state.selectedProject.id)) {
    els.webhookProject.value = state.selectedProject.id;
  }
  renderWebhookEngineFields();

  els.webhooks.replaceChildren();
  if (state.webhooksError) {
    els.webhookStatus.textContent = state.webhooksError;
    return;
  }
  const projectID = selectedWebhookProjectID();
  els.webhookStatus.textContent = projectID
    ? `${state.webhooks.length} webhooks`
    : "Choose a project to manage webhooks";
  if (!projectID || !state.webhooks.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = projectID ? "No webhooks for this project" : "Select a project first";
    els.webhooks.append(empty);
    return;
  }
  for (const webhook of state.webhooks) {
    els.webhooks.append(webhookNode(webhook));
  }
}

function renderWebhookEngineFields() {
  const type = els.webhookEngineType ? els.webhookEngineType.value : "lua";
  document.querySelectorAll("[data-webhook-engine-field]").forEach((field) => {
    field.hidden = field.dataset.webhookEngineField !== type;
  });
}

function renderWebhookEditEngineFields(form) {
  const type = form && form.elements.engine_type ? form.elements.engine_type.value : "lua";
  form.querySelectorAll("[data-webhook-edit-engine-field]").forEach((field) => {
    field.hidden = field.dataset.webhookEditEngineField !== type;
  });
}

function webhookNode(webhook) {
  const article = document.createElement("article");
  article.className = "webhook-item";

  const header = document.createElement("div");
  header.className = "ticket-hook-item-header";
  const title = document.createElement("p");
  title.textContent = webhook.name || webhook.id;
  const stateLabel = document.createElement("span");
  stateLabel.className = webhook.enabled ? "hook-state" : "hook-state is-disabled";
  stateLabel.textContent = webhook.enabled ? "enabled" : "disabled";
  header.append(title, stateLabel);

  const meta = document.createElement("span");
  meta.textContent = [
    webhook.direction,
    webhook.engine.type,
    webhook.actor_user_id ? `actor ${webhook.actor_user_id}` : "",
    webhook.event_types.length ? webhook.event_types.join(", ") : "",
    webhook.token_set ? "token set" : "",
    webhook.last_error ? `error: ${webhook.last_error}` : ""
  ].filter(Boolean).join(" / ");

  const token = state.webhookTokens[webhook.id];
  if (token) {
    const tokenBlock = document.createElement("pre");
    tokenBlock.className = "webhook-token";
    tokenBlock.textContent = `Webhook token: ${token}`;
    article.append(header, meta, tokenBlock);
  } else {
    article.append(header, meta);
  }

  const actions = document.createElement("div");
  actions.className = "ticket-hook-actions";

  const runs = document.createElement("button");
  runs.type = "button";
  runs.dataset.loadWebhookRunsId = webhook.id;
  runs.textContent = "Runs";
  actions.append(runs);

  if (webhook.direction === "outgoing") {
    const deliveries = document.createElement("button");
    deliveries.type = "button";
    deliveries.dataset.loadWebhookDeliveriesId = webhook.id;
    deliveries.textContent = "Deliveries";
    actions.append(deliveries);
  }

  if (webhook.direction === "incoming") {
    const rotate = document.createElement("button");
    rotate.type = "button";
    rotate.dataset.rotateWebhookTokenId = webhook.id;
    rotate.textContent = "Rotate token";
    actions.append(rotate);
  }

  const toggle = document.createElement("button");
  toggle.type = "button";
  toggle.dataset.toggleWebhookId = webhook.id;
  toggle.dataset.webhookEnabled = webhook.enabled ? "false" : "true";
  toggle.textContent = webhook.enabled ? "Disable" : "Enable";

  const remove = document.createElement("button");
  remove.type = "button";
  remove.dataset.deleteWebhookId = webhook.id;
  remove.textContent = "Delete";

  actions.append(toggle, remove);
  article.append(actions);

  article.append(webhookEditForm(webhook));

  const runsList = state.webhookRuns[webhook.id] || [];
  if (runsList.length) {
    article.append(webhookRunListNode(webhook.id, runsList));
  }
  const deliveriesList = state.webhookDeliveries[webhook.id] || [];
  if (deliveriesList.length) {
    article.append(webhookDeliveryListNode(webhook.id, deliveriesList));
  }
  return article;
}

function webhookEditForm(webhook) {
  const form = document.createElement("form");
  form.className = "webhook-edit-form";
  form.dataset.webhookForm = webhook.id;

  form.append(
    inputNode("name", webhook.name || "", "name"),
    webhookSelect("direction", [
      ["incoming", "Incoming"],
      ["outgoing", "Outgoing"]
    ], webhook.direction || "incoming"),
    inputNode("actor_user_id", webhook.actor_user_id || "", "actor user id"),
    inputNode("event_types", (webhook.event_types || []).join(", "), "event types"),
    webhookSelect("engine_type", [
      ["lua", "Lua"],
      ["ai", "OpenRouter AI"]
    ], webhook.engine.type || "lua")
  );

  const enabled = document.createElement("label");
  enabled.className = "inline-toggle";
  const enabledInput = document.createElement("input");
  enabledInput.name = "enabled";
  enabledInput.type = "checkbox";
  enabledInput.checked = webhook.enabled;
  enabled.append(enabledInput, " Enabled");

  const lua = document.createElement("label");
  lua.dataset.webhookEditEngineField = "lua";
  lua.append("Lua Script", webhookTextarea("script", webhook.engine.script || "", 4));

  const ai = document.createElement("div");
  ai.dataset.webhookEditEngineField = "ai";
  const prompt = document.createElement("label");
  prompt.append("AI Prompt", webhookTextarea("prompt", webhook.engine.prompt || "", 4));
  const provider = document.createElement("label");
  provider.append("Provider ID", inputNode("provider_id", webhook.engine.provider_id || "", "openrouter_provider_..."));
  ai.append(prompt, provider);

  const save = document.createElement("button");
  save.type = "submit";
  save.textContent = "Save webhook";

  form.append(enabled, lua, ai, save);
  renderWebhookEditEngineFields(form);
  return form;
}

function webhookSelect(name, options, selectedValue) {
  const select = document.createElement("select");
  select.name = name;
  for (const [value, label] of options) {
    const option = document.createElement("option");
    option.value = value;
    option.textContent = label;
    option.selected = value === selectedValue;
    select.append(option);
  }
  return select;
}

function webhookTextarea(name, value, rows) {
  const textarea = document.createElement("textarea");
  textarea.name = name;
  textarea.rows = rows;
  textarea.spellcheck = false;
  textarea.value = value || "";
  return textarea;
}

function webhookRunListNode(webhookID, runs) {
  const list = document.createElement("div");
  list.className = "webhook-run-list";
  const filterKey = automationRunFilterKey("webhook", webhookID);
  const visibleRuns = filterAutomationRuns(runs, state.automationRunFilters[filterKey] || "all");
  list.append(automationRunSummaryNode(runs, visibleRuns, filterKey));
  if (!visibleRuns.length) {
    list.append(automationRunEmptyNode());
    return list;
  }
  for (const run of visibleRuns) {
    const item = document.createElement("article");
    item.className = "cron-run-item";
    const summary = document.createElement("span");
    summary.textContent = [
      run.state || "queued",
      run.trigger_type,
      run.created_at ? formatDateTime(run.created_at) : "",
      automationRunDurationLabel(run),
      run.error ? `error: ${run.error}` : ""
    ].filter(Boolean).join(" / ");
    const output = document.createElement("pre");
    output.textContent = JSON.stringify(run.output || {}, null, 2);
    item.append(summary, output);
    list.append(item);
  }
  return list;
}

function webhookDeliveryListNode(webhookID, deliveries) {
  const list = document.createElement("div");
  list.className = "webhook-delivery-list";
  for (const delivery of deliveries) {
    const item = document.createElement("article");
    item.className = "cron-run-item";
    const summary = document.createElement("span");
    summary.textContent = [
      delivery.state,
      delivery.event_type,
      delivery.subject_id,
      delivery.attempt_count ? `${delivery.attempt_count} attempts` : "",
      delivery.last_error ? `error: ${delivery.last_error}` : ""
    ].filter(Boolean).join(" / ");
    item.append(summary);
    if (delivery.state === "failed" || delivery.state === "canceled") {
      const retry = document.createElement("button");
      retry.type = "button";
      retry.dataset.retryWebhookDeliveryId = delivery.id;
      retry.dataset.webhookId = webhookID;
      retry.textContent = "Retry";
      item.append(retry);
    }
    list.append(item);
  }
  return list;
}

function renderAuditLog() {
  if (!els.auditStatus || !els.auditLog) {
    return;
  }
  els.auditLog.replaceChildren();
  if (state.auditLogError) {
    els.auditStatus.textContent = state.auditLogError;
    return;
  }
  els.auditStatus.textContent = state.auditLog.length ? `${state.auditLog.length} audit entries` : "No audit entries";
  if (!state.auditLog.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No audit events match the current filters";
    els.auditLog.append(empty);
    return;
  }
  for (const entry of state.auditLog) {
    els.auditLog.append(auditEntryNode(entry));
  }
}

function auditEntryNode(entry) {
  const article = document.createElement("article");
  article.className = "audit-entry";

  const header = document.createElement("div");
  header.className = "audit-entry-header";
  const title = document.createElement("p");
  title.textContent = entry.event_type || "audit event";
  const outcome = document.createElement("span");
  outcome.className = entry.outcome === "failure" ? "audit-outcome is-failure" : "audit-outcome";
  outcome.textContent = entry.outcome || "success";
  header.append(title, outcome);

  const meta = document.createElement("span");
  meta.textContent = [
    entry.occurred_at ? formatDateTime(entry.occurred_at) : "",
    entry.actor_user_id ? `actor ${entry.actor_user_id}` : "",
    entry.auth_kind || "",
    entry.subject_type ? `${entry.subject_type}${entry.subject_id ? ` ${entry.subject_id}` : ""}` : ""
  ].filter(Boolean).join(" / ");

  const payload = document.createElement("pre");
  payload.textContent = JSON.stringify(entry.payload || {}, null, 2);

  article.append(header, meta, payload);
  return article;
}

function renderAdminList(container, items, nodeFactory) {
  if (!container) {
    return;
  }
  container.replaceChildren();
  if (!items.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No accessible records";
    container.append(empty);
    return;
  }
  for (const item of items) {
    container.append(nodeFactory(item));
  }
}

function adminItemNode(titleText, metaText) {
  const row = document.createElement("article");
  row.className = "admin-item";
  const title = document.createElement("p");
  title.textContent = titleText || "Untitled";
  const meta = document.createElement("span");
  meta.textContent = metaText || "";
  row.append(title, meta);
  return row;
}

function renderProjects() {
  els.projects.replaceChildren();
  if (!state.projects.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No projects";
    els.projects.append(empty);
    return;
  }
  for (const project of state.projects) {
    const button = document.createElement("button");
    button.type = "button";
    button.className = "project-item";
    if (state.selectedProject && state.selectedProject.id === project.id) {
      button.classList.add("is-active");
    }
    button.addEventListener("click", async () => {
      state.selectedProject = project;
      if (els.engineProjectID && !els.engineProjectID.value) {
        els.engineProjectID.value = project.id;
      }
      await navigate(`/projects/${project.id}`);
    });

    const key = document.createElement("span");
    key.className = "project-key";
    key.textContent = project.key;
    const name = document.createElement("span");
    name.textContent = project.name;
    button.append(key, name);
    els.projects.append(button);
  }
}

function renderPinnedProjectSavedViews() {
  if (!els.pinnedProjectViews) {
    return;
  }
  els.pinnedProjectViews.replaceChildren();
  if (!state.selectedProject) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "Select a project";
    els.pinnedProjectViews.append(empty);
    return;
  }
  if (state.pinnedProjectSavedViewsLoading) {
    const loading = document.createElement("p");
    loading.className = "muted";
    loading.textContent = "Loading views";
    els.pinnedProjectViews.append(loading);
    return;
  }
  if (state.pinnedProjectSavedViewsError) {
    const error = document.createElement("p");
    error.className = "muted";
    error.textContent = state.pinnedProjectSavedViewsError;
    els.pinnedProjectViews.append(error);
    return;
  }
  if (!state.pinnedProjectSavedViews.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No pinned views";
    els.pinnedProjectViews.append(empty);
    return;
  }
  for (const view of state.pinnedProjectSavedViews) {
    els.pinnedProjectViews.append(pinnedProjectViewNode(view));
  }
}

function pinnedProjectViewNode(view) {
  const button = document.createElement("button");
  button.type = "button";
  button.className = "pinned-project-view";
  button.dataset.applyPinnedProjectViewId = view.id;

  const title = document.createElement("span");
  title.className = "pinned-project-view-title";
  title.textContent = view.name || "Saved view";

  const metadata = savedViewMetadataNode(view, "span");
  metadata.classList.add("pinned-project-view-metadata");

  button.append(title, metadata);
  return button;
}

async function applySavedView(view, options = {}) {
  const query = view.query || {};
  if (options.navigateToSearch) {
    await navigate("/search");
  }
  setSearchForm(query);
  await runSearch({
    text: query.text || "",
    filter: query.filter || "",
    project_id: view.project_id || (state.selectedProject ? state.selectedProject.id : ""),
    sort: view.sort && view.sort.length ? view.sort : [{ field: "updated_at", direction: "desc" }]
  }, { reset: true });
}

function renderTicketHooks() {
  if (!els.ticketHooks || !els.ticketHookProject) {
    return;
  }
  replaceSelectOptions(els.ticketHookProject, "Project", state.projects, (project) => `${project.key} ${project.name}`);
  if (state.selectedProject && state.projects.some((project) => project.id === state.selectedProject.id)) {
    els.ticketHookProject.value = state.selectedProject.id;
  }
  renderTicketHookEngineFields();
  renderTicketHookPreview();

  els.ticketHooks.replaceChildren();
  if (state.ticketHooksError) {
    els.ticketHookStatus.textContent = state.ticketHooksError;
    return;
  }
  const projectID = selectedTicketHookProjectID();
  els.ticketHookStatus.textContent = projectID
    ? `${state.ticketHooks.length} ticket hooks`
    : "Choose a project to manage ticket hooks";
  if (!projectID || !state.ticketHooks.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = projectID ? "No ticket hooks for this project" : "Select a project first";
    els.ticketHooks.append(empty);
    return;
  }
  for (const hook of state.ticketHooks) {
    els.ticketHooks.append(ticketHookNode(hook));
  }
}

function renderTicketHookEngineFields() {
  const type = els.ticketHookEngineType ? els.ticketHookEngineType.value : "lua";
  document.querySelectorAll("[data-ticket-hook-engine-field]").forEach((field) => {
    field.hidden = field.dataset.ticketHookEngineField !== type;
  });
}

function renderTicketHookEditEngineFields(form) {
  const type = form && form.elements.engine_type ? form.elements.engine_type.value : "lua";
  form.querySelectorAll("[data-ticket-hook-edit-engine-field]").forEach((field) => {
    field.hidden = field.dataset.ticketHookEditEngineField !== type;
  });
}

function renderTicketHookPreview() {
  if (!els.ticketHookPreviewOutput) {
    return;
  }
  els.ticketHookPreviewOutput.textContent = JSON.stringify(state.ticketHookPreview || {}, null, 2);
}

function ticketHookNode(hook) {
  const article = document.createElement("article");
  article.className = "ticket-hook-item";

  const header = document.createElement("div");
  header.className = "ticket-hook-item-header";
  const title = document.createElement("p");
  title.textContent = hook.name || hook.id;
  const stateLabel = document.createElement("span");
  stateLabel.className = hook.enabled ? "hook-state" : "hook-state is-disabled";
  stateLabel.textContent = hook.enabled ? "enabled" : "disabled";
  header.append(title, stateLabel);

  const meta = document.createElement("span");
  meta.textContent = [
    hook.event,
    hook.phase,
    `position ${hook.position}`,
    hook.engine.type,
    hook.last_error ? `last error: ${hook.last_error}` : ""
  ].filter(Boolean).join(" / ");

  const actions = document.createElement("div");
  actions.className = "ticket-hook-actions";

  const preview = document.createElement("button");
  preview.type = "button";
  preview.dataset.previewTicketHookId = hook.id;
  preview.textContent = "Preview";

  const runs = document.createElement("button");
  runs.type = "button";
  runs.dataset.loadTicketHookRunsId = hook.id;
  runs.textContent = "Runs";

  const toggle = document.createElement("button");
  toggle.type = "button";
  toggle.dataset.toggleTicketHookId = hook.id;
  toggle.dataset.ticketHookEnabled = hook.enabled ? "false" : "true";
  toggle.textContent = hook.enabled ? "Disable" : "Enable";

  const remove = document.createElement("button");
  remove.type = "button";
  remove.dataset.deleteTicketHookId = hook.id;
  remove.textContent = "Delete";

  actions.append(preview, runs, toggle, remove);
  article.append(header, meta, actions, ticketHookEditForm(hook));
  const runsList = state.ticketHookRuns[hook.id] || [];
  if (runsList.length) {
    article.append(ticketHookRunListNode(hook.id, runsList));
  }
  return article;
}

function ticketHookRunListNode(hookID, runs) {
  const list = document.createElement("div");
  list.className = "ticket-hook-run-list";
  const filterKey = automationRunFilterKey("ticket-hook", hookID);
  const visibleRuns = filterAutomationRuns(runs, state.automationRunFilters[filterKey] || "all");
  list.append(automationRunSummaryNode(runs, visibleRuns, filterKey));
  if (!visibleRuns.length) {
    list.append(automationRunEmptyNode());
    return list;
  }
  for (const run of visibleRuns) {
    const item = document.createElement("article");
    item.className = "cron-run-item";
    const summary = document.createElement("span");
    summary.textContent = [
      run.state || "queued",
      run.trigger_type,
      run.ticket_id ? `ticket ${run.ticket_id}` : "",
      run.created_at ? formatDateTime(run.created_at) : "",
      automationRunDurationLabel(run),
      run.error ? `error: ${run.error}` : ""
    ].filter(Boolean).join(" / ");
    const output = document.createElement("pre");
    output.textContent = JSON.stringify(run.output || {}, null, 2);
    item.append(summary, output);
    list.append(item);
  }
  return list;
}

function ticketHookEditForm(hook) {
  const form = document.createElement("form");
  form.className = "ticket-hook-edit-form";
  form.dataset.ticketHookForm = hook.id;

  form.append(
    inputNode("name", hook.name || "", "name"),
    ticketHookSelect("event", [
      ["ticket_create", "Ticket create"],
      ["ticket_update", "Ticket update"]
    ], hook.event || "ticket_create"),
    ticketHookSelect("phase", [
      ["before", "Before"],
      ["after", "After"]
    ], hook.phase || "before"),
    inputNode("position", String(hook.position || 0), "position", "number"),
    ticketHookSelect("engine_type", [
      ["lua", "Lua"],
      ["ai", "OpenRouter AI"]
    ], hook.engine.type || "lua")
  );

  const enabled = document.createElement("label");
  enabled.className = "inline-toggle";
  const enabledInput = document.createElement("input");
  enabledInput.name = "enabled";
  enabledInput.type = "checkbox";
  enabledInput.checked = hook.enabled;
  enabled.append(enabledInput, " Enabled");

  const lua = document.createElement("label");
  lua.dataset.ticketHookEditEngineField = "lua";
  lua.append("Lua Script", ticketHookTextarea("script", hook.engine.script || "", 4));

  const ai = document.createElement("div");
  ai.dataset.ticketHookEditEngineField = "ai";
  const prompt = document.createElement("label");
  prompt.append("AI Prompt", ticketHookTextarea("prompt", hook.engine.prompt || "", 4));
  const provider = document.createElement("label");
  provider.append("Provider ID", inputNode("provider_id", hook.engine.provider_id || "", "openrouter_provider_..."));
  ai.append(prompt, provider);

  const save = document.createElement("button");
  save.type = "submit";
  save.textContent = "Save hook";

  form.append(enabled, lua, ai, save);
  renderTicketHookEditEngineFields(form);
  return form;
}

function ticketHookSelect(name, options, selectedValue) {
  const select = document.createElement("select");
  select.name = name;
  for (const [value, label] of options) {
    const option = document.createElement("option");
    option.value = value;
    option.textContent = label;
    option.selected = value === selectedValue;
    select.append(option);
  }
  return select;
}

function ticketHookTextarea(name, value, rows) {
  const textarea = document.createElement("textarea");
  textarea.name = name;
  textarea.rows = rows;
  textarea.spellcheck = false;
  textarea.value = value || "";
  return textarea;
}

function renderCreatePages() {
  if (!els.createPages || !els.createPageProject) {
    return;
  }
  replaceSelectOptions(els.createPageProject, "Project", state.projects, (project) => `${project.key} ${project.name}`);
  if (state.selectedProject && state.projects.some((project) => project.id === state.selectedProject.id)) {
    els.createPageProject.value = state.selectedProject.id;
  }
  renderCreatePageLogicFields();

  els.createPages.replaceChildren();
  if (state.createPagesError) {
    els.createPageStatus.textContent = state.createPagesError;
    return;
  }
  const projectID = selectedCreatePageProjectID();
  els.createPageStatus.textContent = projectID
    ? `${state.createPages.length} create pages`
    : "Choose a project to manage create pages";
  if (!projectID || !state.createPages.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = projectID ? "No create pages for this project" : "Select a project first";
    els.createPages.append(empty);
    return;
  }
  for (const page of state.createPages) {
    els.createPages.append(createPageNode(page));
  }
}

function renderCreatePageLogicFields() {
  const type = els.createPageLogicType ? els.createPageLogicType.value : "none";
  document.querySelectorAll("[data-create-page-logic-field]").forEach((field) => {
    field.hidden = field.dataset.createPageLogicField !== type;
  });
}

function renderCreatePageEditLogicFields(form) {
  const type = form && form.elements.logic_type ? form.elements.logic_type.value : "none";
  form.querySelectorAll("[data-create-page-edit-logic-field]").forEach((field) => {
    field.hidden = field.dataset.createPageEditLogicField !== type;
  });
}

function createPageNode(page) {
  const article = document.createElement("article");
  article.className = "create-page-item";

  const header = document.createElement("div");
  header.className = "ticket-hook-item-header";
  const title = document.createElement("p");
  title.textContent = page.name || page.slug || page.id;
  const stateLabel = document.createElement("span");
  stateLabel.className = page.enabled ? "hook-state" : "hook-state is-disabled";
  stateLabel.textContent = page.enabled ? "enabled" : "disabled";
  header.append(title, stateLabel);

  const meta = document.createElement("span");
  meta.textContent = [
    page.slug ? `/${page.slug}` : "",
    page.target_type ? `type ${page.target_type}` : "",
    page.target_status ? `status ${page.target_status}` : "",
    page.owner_user_id ? `owner ${page.owner_user_id}` : "",
    page.has_lua ? "Lua form logic" : "",
    page.has_ai ? "AI form logic" : ""
  ].filter(Boolean).join(" / ");

  const actions = document.createElement("div");
  actions.className = "ticket-hook-actions";

  const schema = document.createElement("button");
  schema.type = "button";
  schema.dataset.loadCreatePageSchemaId = page.id;
  schema.dataset.createPageProjectId = page.project_id;
  schema.dataset.createPageSlug = page.slug;
  schema.textContent = "Schema";

  const runs = document.createElement("button");
  runs.type = "button";
  runs.dataset.loadCreatePageRunsId = page.id;
  runs.textContent = "Runs";

  const toggle = document.createElement("button");
  toggle.type = "button";
  toggle.dataset.toggleCreatePageId = page.id;
  toggle.dataset.createPageEnabled = page.enabled ? "false" : "true";
  toggle.textContent = page.enabled ? "Disable" : "Enable";

  const remove = document.createElement("button");
  remove.type = "button";
  remove.dataset.deleteCreatePageId = page.id;
  remove.textContent = "Delete";

  const open = document.createElement("a");
  open.className = "text-link";
  open.href = `/projects/${encodeURIComponent(page.project_id)}/create/${encodeURIComponent(page.slug)}`;
  open.textContent = "Open";

  actions.append(open, schema, runs, toggle, remove);
  article.append(header, meta, actions);

  if (page.description) {
    const description = document.createElement("p");
    description.className = "muted";
    description.textContent = page.description;
    article.append(description);
  }

  const config = document.createElement("pre");
  config.className = "create-page-config";
  config.textContent = JSON.stringify({
    field_layout: page.field_layout,
    defaults: page.defaults
  }, null, 2);
  article.append(config);

  article.append(createPageEditForm(page));

  if (page.schema) {
    const schemaOutput = document.createElement("pre");
    schemaOutput.className = "create-page-config";
    schemaOutput.textContent = JSON.stringify(page.schema, null, 2);
    article.append(schemaOutput);
  }
  const runsList = state.createPageRuns[page.id] || [];
  if (runsList.length) {
    article.append(createPageRunListNode(page.id, runsList));
  }
  return article;
}

function createPageEditForm(page) {
  const form = document.createElement("form");
  form.className = "create-page-edit-form";
  form.dataset.createPageForm = page.id;
  const logicType = page.form_lua_script ? "lua" : (page.form_ai_prompt || page.form_ai_provider_id ? "ai" : "none");

  form.append(
    inputNode("name", page.name || "", "name"),
    inputNode("slug", page.slug || "", "slug"),
    inputNode("target_type", page.target_type || "", "target type"),
    inputNode("target_status", page.target_status || "", "target status"),
    inputNode("owner_user_id", page.owner_user_id || "", "owner user id"),
    createPageSelect("logic_type", [
      ["none", "Static"],
      ["lua", "Lua"],
      ["ai", "OpenRouter AI"]
    ], logicType)
  );

  const enabled = document.createElement("label");
  enabled.className = "inline-toggle";
  const enabledInput = document.createElement("input");
  enabledInput.name = "enabled";
  enabledInput.type = "checkbox";
  enabledInput.checked = page.enabled;
  enabled.append(enabledInput, " Enabled");

  const description = document.createElement("label");
  description.className = "create-page-edit-wide";
  description.append("Description", createPageTextarea("description", page.description || "", 2));

  const fieldLayout = document.createElement("label");
  fieldLayout.className = "create-page-edit-wide";
  fieldLayout.append("Field Layout JSON", createPageTextarea("field_layout", JSON.stringify(page.field_layout || [], null, 2), 5));

  const defaults = document.createElement("label");
  defaults.className = "create-page-edit-wide";
  defaults.append("Defaults JSON", createPageTextarea("defaults", JSON.stringify(page.defaults || {}, null, 2), 4));

  const lua = document.createElement("label");
  lua.dataset.createPageEditLogicField = "lua";
  lua.append("Lua Script", createPageTextarea("form_lua_script", page.form_lua_script || "", 5));

  const ai = document.createElement("div");
  ai.dataset.createPageEditLogicField = "ai";
  const prompt = document.createElement("label");
  prompt.append("AI Prompt", createPageTextarea("form_ai_prompt", page.form_ai_prompt || "", 5));
  const provider = document.createElement("label");
  provider.append("Provider ID", inputNode("form_ai_provider_id", page.form_ai_provider_id || "", "openrouter_provider_..."));
  ai.append(prompt, provider);

  const save = document.createElement("button");
  save.type = "submit";
  save.textContent = "Save page";

  form.append(enabled, description, fieldLayout, defaults, lua, ai, save);
  renderCreatePageEditLogicFields(form);
  return form;
}

function createPageSelect(name, options, selectedValue) {
  const select = document.createElement("select");
  select.name = name;
  for (const [value, label] of options) {
    const option = document.createElement("option");
    option.value = value;
    option.textContent = label;
    option.selected = value === selectedValue;
    select.append(option);
  }
  return select;
}

function createPageTextarea(name, value, rows) {
  const textarea = document.createElement("textarea");
  textarea.name = name;
  textarea.rows = rows;
  textarea.spellcheck = false;
  textarea.value = value || "";
  return textarea;
}

function createPageRunListNode(pageID, runs) {
  const list = document.createElement("div");
  list.className = "create-page-run-list";
  const filterKey = automationRunFilterKey("create-page", pageID);
  const visibleRuns = filterAutomationRuns(runs, state.automationRunFilters[filterKey] || "all");
  list.append(automationRunSummaryNode(runs, visibleRuns, filterKey));
  if (!visibleRuns.length) {
    list.append(automationRunEmptyNode());
    return list;
  }
  for (const run of visibleRuns) {
    const item = document.createElement("article");
    item.className = "cron-run-item";
    const summary = document.createElement("span");
    summary.textContent = [
      run.state || "queued",
      run.trigger_type,
      run.trigger_ref ? `ref ${run.trigger_ref}` : "",
      run.project_id ? `project ${run.project_id}` : "",
      run.ticket_id ? `ticket ${run.ticket_id}` : "",
      run.created_at ? `created ${formatDateTime(run.created_at)}` : "",
      run.started_at ? `started ${formatDateTime(run.started_at)}` : "",
      run.finished_at ? `finished ${formatDateTime(run.finished_at)}` : "",
      automationRunDurationLabel(run),
      run.error ? `error: ${run.error}` : ""
    ].filter(Boolean).join(" / ");
    const input = document.createElement("pre");
    input.textContent = JSON.stringify({ input: run.input || {} }, null, 2);
    const output = document.createElement("pre");
    output.textContent = JSON.stringify({ output: run.output || {} }, null, 2);
    item.append(summary, input, output);
    list.append(item);
  }
  return list;
}

function renderEngineFields() {
  const type = els.engineType ? els.engineType.value : "lua";
  document.querySelectorAll("[data-engine-field]").forEach((field) => {
    field.hidden = field.dataset.engineField !== type;
  });
  if (els.engineProjectID && state.selectedProject && !els.engineProjectID.value) {
    els.engineProjectID.value = state.selectedProject.id;
  }
}

function renderEngineResult() {
  renderEngineResultSummary();
  if (!els.engineOutput) {
    return;
  }
  els.engineOutput.textContent = JSON.stringify(state.engineResult || {}, null, 2);
}

function renderEngineResultSummary() {
  if (!els.engineResultSummary) {
    return;
  }
  els.engineResultSummary.replaceChildren();
  const result = state.engineResult;
  if (!result || !result.status) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "Run an engine test to inspect output, logs, errors, and action previews.";
    els.engineResultSummary.append(empty);
    return;
  }

  const status = result.status || {};
  const badges = document.createElement("div");
  badges.className = "engine-result-badges";
  badges.append(
    engineBadge("state", status.state || "unknown"),
    engineBadge("mode", status.mode || "executed")
  );
  if (status.engine && status.engine.type) {
    badges.append(engineBadge("engine", status.engine.type));
  }
  els.engineResultSummary.append(badges);

  if (status.error) {
    const error = document.createElement("pre");
    error.className = "engine-result-error";
    error.textContent = status.error;
    els.engineResultSummary.append(error);
  }

  const previews = Array.isArray(status.action_previews) ? status.action_previews : [];
  if (previews.length > 0) {
    const section = document.createElement("section");
    section.className = "engine-action-previews";
    const title = document.createElement("h4");
    title.textContent = `Action previews (${previews.length})`;
    const list = document.createElement("div");
    list.className = "engine-action-preview-list";
    previews.forEach((preview, index) => {
      const item = document.createElement("article");
      item.className = "engine-action-preview";
      const heading = document.createElement("strong");
      heading.textContent = preview.action || preview.type || `Action ${index + 1}`;
      const payload = document.createElement("pre");
      payload.textContent = JSON.stringify(preview, null, 2);
      item.append(heading, payload);
      list.append(item);
    });
    section.append(title, list);
    els.engineResultSummary.append(section);
  }
}

function engineBadge(label, value) {
  const badge = document.createElement("span");
  badge.className = "engine-result-badge";
  badge.textContent = `${label}: ${value}`;
  return badge;
}

function engineTestSpec(form) {
  const data = formData(form);
  const type = data.engine_type || "lua";
  const engine = { type };
  if (type === "lua") {
    engine.script = data.script || "";
  } else if (type === "ai") {
    engine.prompt = data.prompt || "";
    engine.provider_id = data.provider_id || "";
  } else if (type === "wasm") {
    engine.module_base64 = data.module_base64 || "";
  }
  return {
    surface: data.surface || "scratch",
    project_id: data.project_id || "",
    engine,
    context: parseJSONField(data.context, "Context JSON"),
    input: parseJSONField(data.input, "Input JSON"),
    dry_run: true
  };
}

function selectedTicketHookProjectID() {
  if (els.ticketHookProject && els.ticketHookProject.value) {
    return els.ticketHookProject.value;
  }
  return state.selectedProject ? state.selectedProject.id : "";
}

function selectedCronJobProjectID() {
  if (els.cronJobProject && els.cronJobProject.value) {
    return els.cronJobProject.value;
  }
  return state.selectedProject ? state.selectedProject.id : "";
}

function selectedWebhookProjectID() {
  if (els.webhookProject && els.webhookProject.value) {
    return els.webhookProject.value;
  }
  return state.selectedProject ? state.selectedProject.id : "";
}

function selectedCreatePageProjectID() {
  if (els.createPageProject && els.createPageProject.value) {
    return els.createPageProject.value;
  }
  return state.selectedProject ? state.selectedProject.id : "";
}

function selectedNotificationDestinationProjectID() {
  if (els.notificationDestinationProject && els.notificationDestinationProject.value) {
    return els.notificationDestinationProject.value;
  }
  return state.selectedProject ? state.selectedProject.id : "";
}

function selectedNotificationPolicyProjectID() {
  if (els.notificationPolicyProject && els.notificationPolicyProject.value) {
    return els.notificationPolicyProject.value;
  }
  return state.selectedProject ? state.selectedProject.id : "";
}

function selectedNotificationHookProjectID() {
  if (els.notificationHookProject && els.notificationHookProject.value) {
    return els.notificationHookProject.value;
  }
  return state.selectedProject ? state.selectedProject.id : "";
}

function selectedProjectPreferenceProjectID() {
  if (els.projectPreferenceProject && els.projectPreferenceProject.value) {
    return els.projectPreferenceProject.value;
  }
  return state.selectedProject ? state.selectedProject.id : "";
}

function selectedNotificationDeliveryProjectID() {
  if (els.notificationDeliveryProject && els.notificationDeliveryProject.value) {
    return els.notificationDeliveryProject.value;
  }
  return state.selectedProject ? state.selectedProject.id : "";
}

function cronJobSpec(form) {
  const data = formData(form);
  const projectID = selectedCronJobProjectID();
  const spec = {
    project_id: projectID,
    owner_user_id: data.owner_user_id || "",
    name: data.name || "",
    schedule: data.schedule || "",
    timezone: data.timezone || "UTC",
    enabled: Boolean(data.enabled),
    engine: cronJobEngineSpec(data)
  };
  if (!spec.owner_user_id) {
    delete spec.owner_user_id;
  }
  return spec;
}

function cronJobEngineSpec(data) {
  const type = data.engine_type || "lua";
  if (type === "ai") {
    return {
      type,
      prompt: data.prompt || "",
      provider_id: data.provider_id || ""
    };
  }
  return {
    type: "lua",
    script: data.script || ""
  };
}

function webhookSpec(form) {
  const data = formData(form);
  return {
    name: data.name || "",
    direction: data.direction || "incoming",
    enabled: Boolean(data.enabled),
    actor_user_id: data.actor_user_id || "",
    event_types: parseCommaList(data.event_types),
    engine: webhookEngineSpec(data)
  };
}

function webhookEngineSpec(data) {
  const type = data.engine_type || "lua";
  if (type === "ai") {
    return {
      type,
      prompt: data.prompt || "",
      provider_id: data.provider_id || ""
    };
  }
  return {
    type: "lua",
    script: data.script || ""
  };
}

function notificationDestinationCollectionPath(scopeType, projectID) {
  if (scopeType === "project") {
    return `/api/projects/${projectID}/notification-destinations`;
  }
  return "/api/notification-destinations";
}

function notificationPolicyCollectionPath(scopeType, projectID) {
  if (scopeType === "project") {
    return `/api/projects/${projectID}/notification-policies`;
  }
  return "/api/notification-policies";
}

function notificationHookCollectionPath(scopeType, projectID) {
  if (scopeType === "project") {
    return `/api/projects/${projectID}/notification-hooks`;
  }
  return "/api/notification-hooks";
}

function notificationDeliveryQuery() {
  const params = new URLSearchParams();
  if (els.notificationDeliveryForm) {
    const data = formData(els.notificationDeliveryForm);
    if (data.status) {
      params.set("status", data.status);
    }
    if (data.policy_id) {
      params.set("policy_id", data.policy_id);
    }
    if (data.destination_id) {
      params.set("destination_id", data.destination_id);
    }
  }
  params.set("limit", "20");
  const query = params.toString();
  return query ? `?${query}` : "";
}

function notificationPolicySpec(form) {
  const data = formData(form);
  return {
    name: data.name || "",
    event_types: parseCommaList(data.event_types),
    destination_ids: selectedFormValues(form, "destination_ids"),
    enabled: Boolean(data.enabled)
  };
}

function notificationPolicyUpdateSpec(form) {
  return notificationPolicySpec(form);
}

function notificationHookSpec(form) {
  const data = formData(form);
  const spec = {
    name: data.name || "",
    actor_user_id: data.actor_user_id || "",
    event_types: parseCommaList(data.event_types),
    enabled: Boolean(data.enabled),
    engine: notificationHookEngineSpec(data)
  };
  if (!spec.actor_user_id) {
    delete spec.actor_user_id;
  }
  return spec;
}

function notificationHookEngineSpec(data) {
  const type = data.engine_type || "lua";
  if (type === "ai") {
    return {
      type,
      prompt: data.prompt || "",
      provider_id: data.provider_id || ""
    };
  }
  return {
    type: "lua",
    script: data.script || ""
  };
}

function notificationHookPreviewSpec() {
  if (!els.notificationHookPreviewForm) {
    return {};
  }
  const data = formData(els.notificationHookPreviewForm);
  const spec = {
    policy_id: data.policy_id || "",
    event_type: data.event_type || "",
    message: data.message || "",
    destination_ids: selectedFormValues(els.notificationHookPreviewForm, "destination_ids"),
    payload: parseJSONField(data.payload, "Notification hook payload JSON")
  };
  const projectID = selectedNotificationHookProjectID();
  if (projectID) {
    spec.project_id = projectID;
  }
  return spec;
}

function createPageSpec(form, options = {}) {
  const data = formData(form);
  const logicType = data.logic_type || "none";
  const spec = {
    name: data.name || "",
    slug: data.slug || "",
    description: data.description || "",
    enabled: Boolean(data.enabled),
    target_type: data.target_type || "",
    target_status: data.target_status || "",
    owner_user_id: data.owner_user_id || "",
    field_layout: parseJSONArrayField(data.field_layout, "Field layout JSON"),
    defaults: parseJSONField(data.defaults, "Defaults JSON")
  };
  if (logicType === "lua") {
    spec.form_lua_script = data.form_lua_script || "";
    if (options.clearUnselectedLogic) {
      spec.form_ai_prompt = "";
      spec.form_ai_provider_id = "";
    }
  } else if (logicType === "ai") {
    if (options.clearUnselectedLogic) {
      spec.form_lua_script = "";
    }
    spec.form_ai_prompt = data.form_ai_prompt || "";
    spec.form_ai_provider_id = data.form_ai_provider_id || "";
  } else if (options.clearUnselectedLogic) {
    spec.form_lua_script = "";
    spec.form_ai_prompt = "";
    spec.form_ai_provider_id = "";
  }
  for (const key of ["target_type", "target_status", "owner_user_id", "form_lua_script", "form_ai_prompt", "form_ai_provider_id"]) {
    if (!spec[key]) {
      if (!options.includeEmptyOptionals) {
        delete spec[key];
      }
    }
  }
  return spec;
}

function ticketHookSpec(form) {
  const data = formData(form);
  return {
    name: data.name || "",
    event: data.event || "ticket_create",
    phase: data.phase || "before",
    enabled: Boolean(data.enabled),
    position: Number(data.position || 0),
    engine: ticketHookEngineSpec(data)
  };
}

function ticketHookEngineSpec(data) {
  const type = data.engine_type || "lua";
  if (type === "ai") {
    return {
      type,
      prompt: data.prompt || "",
      provider_id: data.provider_id || ""
    };
  }
  return {
    type: "lua",
    script: data.script || ""
  };
}

function ticketHookPreviewSpec() {
  const data = formData(els.ticketHookPreviewForm);
  const spec = {
    ticket: parseJSONField(data.ticket, "Preview ticket JSON")
  };
  const currentText = (data.current || "").trim();
  if (currentText) {
    spec.current = parseJSONField(currentText, "Current ticket JSON");
  }
  return spec;
}

function parseJSONField(value, label) {
  const text = (value || "").trim();
  if (!text) {
    return {};
  }
  try {
    const parsed = JSON.parse(text);
    if (!parsed || Array.isArray(parsed) || typeof parsed !== "object") {
      throw new Error(`${label} must be a JSON object`);
    }
    return parsed;
  } catch (error) {
    if (error.message && error.message.includes("must be a JSON object")) {
      throw error;
    }
    throw new Error(`${label} is not valid JSON`);
  }
}

function parseJSONArrayField(value, label) {
  const text = (value || "").trim();
  if (!text) {
    return [];
  }
  try {
    const parsed = JSON.parse(text);
    if (!Array.isArray(parsed)) {
      throw new Error(`${label} must be a JSON array`);
    }
    return parsed;
  } catch (error) {
    if (error.message && error.message.includes("must be a JSON array")) {
      throw error;
    }
    throw new Error(`${label} is not valid JSON`);
  }
}

function renderTickets() {
  els.selectedProject.textContent = state.selectedProject ? `${state.selectedProject.key} ${state.selectedProject.name}` : "No project selected";
  els.ticketColumns.replaceChildren();
  const columns = ticketColumns();
  if (state.boardTickets) {
    els.ticketColumns.append(boardSummaryNode(state.boardTickets, columns));
  }
  for (const column of columns) {
    els.ticketColumns.append(ticketColumnNode(column));
  }
  const filteredTickets = filteredProjectTickets();
  const byID = new Map(state.tickets.map((ticket) => [ticket.id, ticket]));
  const rendered = new Set();
  if (state.boardTickets && Array.isArray(state.boardTickets.columns)) {
    for (const column of state.boardTickets.columns) {
      const list = els.ticketColumns.querySelector(`[data-status="${cssEscape(column.slug)}"] .ticket-list`);
      if (!list) {
        continue;
      }
      for (const ticket of column.tickets) {
        const fullTicket = byID.get(ticket.id) || ticket;
        if (!ticketMatchesFilters(fullTicket)) {
          continue;
        }
        list.append(ticketNode(fullTicket));
        rendered.add(fullTicket.id);
      }
    }
    if (state.boardTickets.filtered_by_saved_view) {
      renderTicketFilterSummary(rendered.size, state.tickets.length);
      return;
    }
  }
  for (const ticket of filteredTickets) {
    if (rendered.has(ticket.id)) {
      continue;
    }
    const list = els.ticketColumns.querySelector(`[data-status="${cssEscape(ticket.status)}"] .ticket-list`) ||
      els.ticketColumns.querySelector(".ticket-list");
    if (list) {
      list.append(ticketNode(ticket));
    }
  }
  renderTicketFilterSummary(filteredTickets.length, state.tickets.length);
}

function filteredProjectTickets() {
  return state.tickets.filter(ticketMatchesFilters);
}

function ticketMatchesFilters(ticket) {
  if (state.ticketFilters.label && (!Array.isArray(ticket.labels) || !ticket.labels.includes(state.ticketFilters.label))) {
    return false;
  }
  if (state.ticketFilters.component_id && ticket.component_id !== state.ticketFilters.component_id) {
    return false;
  }
  if (state.ticketFilters.version_id && ticket.version_id !== state.ticketFilters.version_id) {
    return false;
  }
  return true;
}

function pruneTicketFilters() {
  if (state.ticketFilters.label && !state.projectLabels.some((label) => label.label === state.ticketFilters.label)) {
    state.ticketFilters.label = "";
  }
  if (state.ticketFilters.component_id && !state.components.some((component) => component.id === state.ticketFilters.component_id)) {
    state.ticketFilters.component_id = "";
  }
  if (state.ticketFilters.version_id && !state.versions.some((version) => version.id === state.ticketFilters.version_id)) {
    state.ticketFilters.version_id = "";
  }
}

function renderTicketFilters() {
  replaceSelectOptions(els.ticketFilterLabel, "All labels", state.projectLabels, (label) => `${label.label} (${label.ticket_count})`, state.ticketFilters.label);
  replaceSelectOptions(els.ticketFilterComponent, "All components", state.components, (component) => component.name || component.id, state.ticketFilters.component_id);
  replaceSelectOptions(els.ticketFilterVersion, "All versions", state.versions, (version) => version.name || version.id, state.ticketFilters.version_id);
  renderTicketFilterSummary(filteredProjectTickets().length, state.tickets.length);
}

function renderTicketFilterSummary(shown, total) {
  if (!els.ticketFilterSummary) {
    return;
  }
  const active = hasTicketFilters();
  if (!total) {
    els.ticketFilterSummary.textContent = active ? "No tickets match these filters" : "No tickets";
    return;
  }
  if (!active) {
    els.ticketFilterSummary.textContent = `Showing all ${total} ticket${total === 1 ? "" : "s"}`;
    return;
  }
  els.ticketFilterSummary.textContent = `Showing ${shown} filtered ticket${shown === 1 ? "" : "s"}`;
}

function emptyTicketFilters() {
  return { label: "", component_id: "", version_id: "" };
}

function ticketFiltersFromForm() {
  return {
    label: els.ticketFilterLabel ? els.ticketFilterLabel.value : "",
    component_id: els.ticketFilterComponent ? els.ticketFilterComponent.value : "",
    version_id: els.ticketFilterVersion ? els.ticketFilterVersion.value : ""
  };
}

function ticketFilterParams() {
  pruneTicketFilters();
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(state.ticketFilters)) {
    if (value) {
      params.set(key, value);
    }
  }
  return params;
}

function hasTicketFilters() {
  return Boolean(state.ticketFilters.label || state.ticketFilters.component_id || state.ticketFilters.version_id);
}

function ticketColumns() {
  if (state.boardTickets && Array.isArray(state.boardTickets.columns) && state.boardTickets.columns.length) {
    return state.boardTickets.columns.map((column) => ({
      slug: column.slug,
      name: column.name || statusName(column.slug),
      ticket_count: column.ticket_count,
      wip_limit: column.wip_limit,
      over_wip_limit: column.over_wip_limit
    }));
  }
  const statuses = state.workflowStatuses.length ? state.workflowStatuses : defaultWorkflowStatuses();
  return statuses.map((status) => ({ slug: status.slug, name: status.name }));
}

function boardSummaryNode(boardTickets, columns) {
  const metrics = boardSummaryMetrics(boardTickets, columns);
  const section = document.createElement("section");
  section.className = "board-summary";
  section.append(
    boardSummaryMetricNode("Tickets", metrics.total_tickets),
    boardSummaryMetricNode("Columns", metrics.column_count),
    boardSummaryMetricNode("WIP warnings", metrics.wip_warnings),
    boardSummaryMetricNode("Saved view", metrics.saved_view_filter)
  );
  return section;
}

function boardSummaryMetrics(boardTickets, columns = []) {
  const boardColumns = Array.isArray(boardTickets && boardTickets.columns) ? boardTickets.columns : columns;
  const totalTickets = boardColumns.reduce((sum, column) => {
    if (Number.isFinite(column.ticket_count)) {
      return sum + column.ticket_count;
    }
    return sum + (Array.isArray(column.tickets) ? column.tickets.length : 0);
  }, 0);
  return {
    total_tickets: totalTickets,
    column_count: boardColumns.length,
    wip_warnings: boardColumns.filter((column) => column.over_wip_limit).length,
    saved_view_filter: boardTickets && boardTickets.filtered_by_saved_view ? "filtered" : "all tickets"
  };
}

function boardSummaryMetricNode(label, value) {
  const item = document.createElement("div");
  item.className = "board-summary-metric";
  const strong = document.createElement("strong");
  strong.textContent = String(value);
  const span = document.createElement("span");
  span.textContent = label;
  item.append(strong, span);
  return item;
}

function ticketColumnNode(column) {
  const section = document.createElement("section");
  section.className = "column";
  section.dataset.status = column.slug;

  const heading = document.createElement("h3");
  heading.textContent = column.name || column.slug;

  const meta = document.createElement("span");
  meta.className = "column-capacity";
  const ticketCount = Number.isFinite(column.ticket_count) ? column.ticket_count : 0;
  if (Number.isFinite(column.wip_limit)) {
    meta.textContent = `${ticketCount}/${column.wip_limit}`;
    if (column.over_wip_limit) {
      section.classList.add("is-over-wip");
    }
  } else {
    meta.textContent = `${ticketCount}`;
  }

  const header = document.createElement("div");
  header.className = "column-header";
  header.append(heading, meta);

  const list = document.createElement("div");
  list.className = "ticket-list";
  list.dataset.boardDropStatus = column.slug;

  section.append(header, list);
  return section;
}

function renderIssue() {
  if (!els.issueDetail) {
    return;
  }
  const ticket = state.selectedIssue;
  els.issueDetail.replaceChildren();
  if (!ticket) {
    els.issueTitle.textContent = "No issue selected";
    els.issueProjectLink.href = "/projects";
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "Choose an issue from a project or dashboard.";
    els.issueDetail.append(empty);
    return;
  }

  els.issueTitle.textContent = `${ticket.key} ${ticket.title}`;
  els.issueProjectLink.href = `/projects/${ticket.project_id}`;
  els.issueProjectLink.textContent = projectName(ticket.project_id) || "Project";

  const overview = document.createElement("section");
  overview.className = "issue-overview";
  for (const [label, value] of [
    ["Status", ticket.status],
    ["Priority", ticket.priority || "None"],
    ["Type", ticket.type || "None"],
    ["Story points", storyPointLabel(ticket.story_points) || "Unestimated"],
    ["Project", projectName(ticket.project_id)],
    ["Sprint", ticket.sprint_id ? sprintName(ticket.sprint_id) : "None"],
    ["Component", ticket.component_id ? componentName(ticket.component_id) : "None"],
    ["Version", ticket.version_id ? versionName(ticket.version_id) : "None"],
    ["Watchers", String(ticket.watcher_count || 0)],
    ["Updated", ticket.updated_at ? formatDateTime(ticket.updated_at) : ""]
  ]) {
    const item = document.createElement("article");
    const key = document.createElement("span");
    key.textContent = label;
    const val = document.createElement("strong");
    val.textContent = value || "None";
    item.append(key, val);
    overview.append(item);
  }

  const description = document.createElement("section");
  description.className = "issue-section";
  const descriptionTitle = document.createElement("h3");
  descriptionTitle.textContent = "Description";
  const descriptionBody = document.createElement("p");
  descriptionBody.textContent = ticket.description || "No description";
  description.append(descriptionTitle, descriptionBody);

  els.issueDetail.append(
    overview,
    description,
    ticketDeleteNode(ticket),
    labelControlNode(ticket),
    storyPointControlNode(ticket),
    customFieldControlNode(ticket),
    planningControlNode(ticket),
    watcherNode(ticket),
    ticketLinksNode(ticket),
    sprintControlNode(ticket),
    commentNode(ticket),
    attachmentNode(ticket),
    activityNode(ticket)
  );
}

function renderCreatePageView() {
  if (!els.createPageSubmitForm) {
    return;
  }
  const route = currentRoute();
  const schema = state.selectedCreatePageSchema;
  els.createPageSubmitForm.replaceChildren();
  if (!schema || route.page !== "create-page") {
    els.createPageTitle.textContent = "Create Ticket";
    els.createPageProjectLink.href = "/projects";
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "Choose a create page.";
    els.createPageSubmitForm.append(empty);
    return;
  }

  els.createPageTitle.textContent = schema.name || "Create Ticket";
  els.createPageProjectLink.href = `/projects/${encodeURIComponent(schema.project_id)}`;
  els.createPageProjectLink.textContent = projectName(schema.project_id) || "Project";

  if (schema.description) {
    const description = document.createElement("p");
    description.className = "create-page-description";
    description.textContent = schema.description;
    els.createPageSubmitForm.append(description);
  }

  for (const item of createPageLayout(schema)) {
    const node = createPageLayoutNode(item, schema.defaults);
    if (node) {
      els.createPageSubmitForm.append(node);
    }
  }

  const actions = document.createElement("div");
  actions.className = "create-page-submit-actions";
  const submit = document.createElement("button");
  submit.type = "submit";
  submit.textContent = "Create ticket";
  actions.append(submit);
  els.createPageSubmitForm.append(actions);

  if (state.createPageSubmission) {
    const result = document.createElement("article");
    result.className = "create-page-result";
    const title = document.createElement("p");
    const link = document.createElement("a");
    link.href = `/issues/${encodeURIComponent(state.createPageSubmission.id)}`;
    link.textContent = `${state.createPageSubmission.key || state.createPageSubmission.id} ${state.createPageSubmission.title || ""}`;
    title.append(link);
    const meta = document.createElement("span");
    meta.textContent = [state.createPageSubmission.status, state.createPageSubmission.type, state.createPageSubmission.priority].filter(Boolean).join(" / ");
    result.append(title, meta);
    els.createPageSubmitForm.append(result);
  }
}

function createPageLayout(schema) {
  const layout = Array.isArray(schema.field_layout) && schema.field_layout.length
    ? schema.field_layout.map(normalizeCreatePageLayoutItem).filter(Boolean)
    : [];
  if (!createPageLayoutHasField(layout, "title")) {
    layout.unshift(normalizeCreatePageLayoutItem({ key: "title", label: "Title", type: "text", required: true }));
  }
  if (!createPageLayoutHasField(layout, "description")) {
    layout.push(normalizeCreatePageLayoutItem({ key: "description", label: "Description", type: "textarea", required: false }));
  }
  return layout;
}

function normalizeCreatePageLayoutItem(item) {
  if (!item || typeof item !== "object" || item.html !== undefined || item.hidden) {
    return null;
  }
  const key = String(item.key || "").trim();
  const kind = String(item.widget || item.kind || item.layout || item.type || (key ? "field" : "section")).trim().toLowerCase();
  const fields = Array.isArray(item.fields)
    ? item.fields.map(normalizeCreatePageLayoutItem).filter(Boolean)
    : [];
  const title = String(item.title || item.heading || item.label || item.name || "").trim();
  const text = String(item.text || item.help || item.description || item.body || "").trim();
  if (!key && !fields.length && !title && !text) {
    return null;
  }
  return {
    key,
    label: String(item.label || item.name || title || titleize(key)).trim(),
    title,
    text,
    kind,
    type: String(item.type || "text").trim().toLowerCase(),
    required: Boolean(item.required),
    options: Array.isArray(item.options) ? item.options.map(String) : [],
    fields
  };
}

function createPageLayoutHasField(items, key) {
  for (const item of items) {
    if (item.key === key || createPageLayoutHasField(item.fields || [], key)) {
      return true;
    }
  }
  return false;
}

function createPageLayoutNode(item, defaults) {
  if (!item) {
    return null;
  }
  if (item.key) {
    return createPageFieldNode(item, defaults);
  }
  if (createPageTextLayoutKinds().has(item.kind) && !item.fields.length) {
    return createPageTextNode(item);
  }
  return createPageGroupNode(item, defaults);
}

function createPageTextLayoutKinds() {
  return new Set(["heading", "header", "title", "help", "hint", "note", "paragraph", "text", "description", "copy"]);
}

function createPageTextNode(item) {
  const node = document.createElement(item.kind === "heading" || item.kind === "header" || item.kind === "title" ? "h3" : "p");
  node.className = item.kind === "heading" || item.kind === "header" || item.kind === "title"
    ? "create-page-layout-heading"
    : "create-page-layout-text";
  node.textContent = item.text || item.title;
  return node;
}

function createPageGroupNode(item, defaults) {
  const section = document.createElement("section");
  section.className = createPageColumnsLayoutKinds().has(item.kind)
    ? "create-page-layout-group create-page-layout-columns"
    : "create-page-layout-group";
  if (item.title) {
    const heading = document.createElement("h3");
    heading.className = "create-page-layout-heading";
    heading.textContent = item.title;
    section.append(heading);
  }
  if (item.text) {
    const body = document.createElement("p");
    body.className = "create-page-layout-text";
    body.textContent = item.text;
    section.append(body);
  }
  const fields = document.createElement("div");
  fields.className = "create-page-layout-fields";
  for (const child of item.fields) {
    const childNode = createPageLayoutNode(child, defaults);
    if (childNode) {
      fields.append(childNode);
    }
  }
  if (fields.childElementCount) {
    section.append(fields);
  }
  return section.childElementCount ? section : null;
}

function createPageColumnsLayoutKinds() {
  return new Set(["columns", "column", "row", "grid"]);
}

function createPageFieldNode(field, defaults) {
  const label = document.createElement("label");
  label.className = "create-page-field";
  const text = document.createElement("span");
  text.textContent = field.required ? `${field.label} *` : field.label;
  label.append(text, createPageFieldControl(field, defaults));
  return label;
}

function createPageFieldControl(field, defaults) {
  const defaultValue = createPageDefaultValue(defaults, field.key);
  let control;
  if (field.type === "textarea" || field.type === "markdown" || field.key === "description" || field.key === "custom_fields") {
    control = document.createElement("textarea");
    control.rows = field.key === "custom_fields" ? 4 : 3;
  } else if (field.type === "select" || field.type === "single-select" || field.type === "single_select") {
    control = document.createElement("select");
    const empty = document.createElement("option");
    empty.value = "";
    empty.textContent = "Choose";
    control.append(empty);
    for (const option of field.options) {
      const item = document.createElement("option");
      item.value = option;
      item.textContent = option;
      control.append(item);
    }
  } else {
    control = document.createElement("input");
    if (field.type === "date" || field.key === "start_date" || field.key === "due_date") {
      control.type = "date";
    } else if (field.type === "number") {
      control.type = "number";
    } else if (field.type === "checkbox" || field.type === "boolean") {
      control.type = "checkbox";
    } else {
      control.type = "text";
    }
  }
  control.name = field.key;
  control.dataset.createPageField = field.key;
  control.dataset.createPageFieldType = field.type;
  control.required = field.required && control.type !== "checkbox";
  if (control.type === "checkbox") {
    control.checked = Boolean(defaultValue);
  } else if (defaultValue !== undefined && defaultValue !== null) {
    control.value = createPageDisplayValue(defaultValue);
  }
  return control;
}

function createPageDefaultValue(defaults, key) {
  if (!defaults || typeof defaults !== "object") {
    return "";
  }
  if (Object.prototype.hasOwnProperty.call(defaults, key)) {
    return defaults[key];
  }
  if (defaults.custom_fields && typeof defaults.custom_fields === "object" && Object.prototype.hasOwnProperty.call(defaults.custom_fields, key)) {
    return defaults.custom_fields[key];
  }
  const customFieldKey = createPageCustomFieldKey(key);
  if (customFieldKey && defaults.custom_fields && typeof defaults.custom_fields === "object" && Object.prototype.hasOwnProperty.call(defaults.custom_fields, customFieldKey)) {
    return defaults.custom_fields[customFieldKey];
  }
  return "";
}

function createPageDisplayValue(value) {
  if (Array.isArray(value)) {
    return value.join(", ");
  }
  if (value && typeof value === "object") {
    return JSON.stringify(value, null, 2);
  }
  return String(value);
}

function createPageTicketSpec(form, schema) {
  const ticket = {};
  const customFields = {};
  if (schema && schema.defaults && schema.defaults.custom_fields && typeof schema.defaults.custom_fields === "object") {
    Object.assign(customFields, schema.defaults.custom_fields);
  }
  for (const control of form.querySelectorAll("[data-create-page-field]")) {
    const key = control.dataset.createPageField;
    const fieldType = control.dataset.createPageFieldType || "text";
    const value = createPageControlValue(control, fieldType);
    if (value === "" || value === undefined || value === null) {
      continue;
    }
    if (key === "labels") {
      ticket.labels = Array.isArray(value) ? value : parseLabels(value);
    } else if (key === "custom_fields") {
      Object.assign(customFields, parseJSONField(value, "Custom fields JSON"));
    } else if (key === "story_points") {
      applyStoryPointsSpec(ticket, value);
    } else if (createPageCustomFieldKey(key)) {
      customFields[createPageCustomFieldKey(key)] = value;
    } else if (ticketFieldKeys().has(key)) {
      ticket[key] = value;
    } else {
      customFields[key] = value;
    }
  }
  if (Object.keys(customFields).length) {
    ticket.custom_fields = customFields;
  }
  return ticket;
}

function createPageControlValue(control, fieldType) {
  if (control.type === "checkbox") {
    return control.checked;
  }
  const value = control.value.trim();
  if (!value) {
    return "";
  }
  if (fieldType === "number") {
    const parsed = Number(value);
    return Number.isNaN(parsed) ? value : parsed;
  }
  if (fieldType === "json" || fieldType === "object") {
    return parseJSONField(value, `${control.name} JSON`);
  }
  return value;
}

function createPageCustomFieldKey(key) {
  const prefix = "custom_fields.";
  if (typeof key === "string" && key.startsWith(prefix) && key.length > prefix.length) {
    return key.slice(prefix.length);
  }
  return "";
}

function ticketFieldKeys() {
  return new Set([
    "title",
    "description",
    "status",
    "priority",
    "type",
    "assignee_id",
    "parent_ticket_id",
    "sprint_id",
    "component_id",
    "version_id",
    "rank",
    "start_date",
    "due_date",
    "story_points"
  ]);
}

function applyStoryPointsSpec(spec, rawValue) {
  const value = rawValue === undefined || rawValue === null ? "" : String(rawValue).trim();
  if (value === "") {
    spec.story_points = null;
    return;
  }
  const parsed = Number(value);
  spec.story_points = Number.isFinite(parsed) ? parsed : value;
}

function storyPointLabel(value) {
  if (value === null || value === undefined || value === "") {
    return "";
  }
  return `${formatStoryPoints(value)} pt`;
}

function formatStoryPoints(value) {
  const number = Number(value);
  if (!Number.isFinite(number)) {
    return String(value);
  }
  return Number.isInteger(number) ? String(number) : String(Number(number.toFixed(2)));
}

function titleize(value) {
  return String(value || "")
    .replace(/[_-]+/g, " ")
    .replace(/\b\w/g, (match) => match.toUpperCase());
}

function ticketNode(ticket) {
  const article = document.createElement("article");
  article.className = "ticket";
  article.draggable = true;
  article.dataset.boardTicketId = ticket.id;

  const key = document.createElement("p");
  key.className = "ticket-key";
  key.textContent = ticket.key;

  const title = document.createElement("h4");
  const titleLink = document.createElement("a");
  titleLink.href = `/issues/${ticket.id}`;
  titleLink.textContent = ticket.title;
  title.append(titleLink);

  const meta = document.createElement("p");
  meta.className = "ticket-meta";
  meta.textContent = [ticket.type, ticket.priority, storyPointLabel(ticket.story_points)].filter(Boolean).join(" / ") || "Unclassified";

  const actions = document.createElement("div");
  actions.className = "ticket-actions";
  for (const action of statusActions(ticket.status)) {
    const button = document.createElement("button");
    button.type = "button";
    button.dataset.ticketId = ticket.id;
    button.dataset.ticketStatus = action.status;
    button.textContent = action.label;
    actions.append(button);
  }
  actions.append(ticketDeleteButton(ticket));

  article.append(key, title, meta, watcherNode(ticket), labelControlNode(ticket), storyPointControlNode(ticket), customFieldControlNode(ticket), planningControlNode(ticket), ticketLinksNode(ticket), sprintControlNode(ticket), commentNode(ticket), attachmentNode(ticket), actions);
  return article;
}

function ticketDeleteNode(ticket) {
  const section = document.createElement("section");
  section.className = "ticket-delete";
  section.setAttribute("aria-label", `${ticket.key} delete`);
  section.append(ticketDeleteButton(ticket));
  return section;
}

function ticketDeleteButton(ticket) {
  const button = document.createElement("button");
  button.type = "button";
  button.dataset.deleteTicketId = ticket.id;
  button.dataset.projectId = ticket.project_id;
  button.textContent = "Delete";
  return button;
}

function watcherNode(ticket) {
  const section = document.createElement("section");
  section.className = "ticket-watchers";
  section.setAttribute("aria-label", `${ticket.key} watchers`);

  const watchers = state.ticketWatchers[ticket.id] || [];
  const summary = document.createElement("p");
  summary.className = "watcher-heading";
  const count = Number(ticket.watcher_count || watchers.length || 0);
  summary.textContent = `${count} watcher${count === 1 ? "" : "s"}`;

  const names = document.createElement("div");
  names.className = "watcher-list";
  if (!watchers.length) {
    const empty = document.createElement("span");
    empty.className = "muted";
    empty.textContent = "No watchers";
    names.append(empty);
  } else {
    for (const watcher of watchers) {
      const item = document.createElement("span");
      item.textContent = watcher.display_name || watcher.username || watcher.user_id;
      names.append(item);
    }
  }

  const button = document.createElement("button");
  button.type = "button";
  if (ticket.watching) {
    button.dataset.unwatchTicketId = ticket.id;
    button.textContent = "Unwatch";
  } else {
    button.dataset.watchTicketId = ticket.id;
    button.textContent = "Watch";
  }

  section.append(summary, names, button);
  return section;
}

function labelControlNode(ticket) {
  const section = document.createElement("section");
  section.className = "ticket-labels";
  section.setAttribute("data-ticket-label-control", "true");
  section.setAttribute("aria-label", `${ticket.key} labels`);

  const chips = document.createElement("div");
  chips.className = "label-chips";
  if (ticket.labels.length) {
    for (const label of ticket.labels) {
      const chip = document.createElement("span");
      chip.textContent = label;
      chips.append(chip);
    }
  } else {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No labels";
    chips.append(empty);
  }

  const controls = document.createElement("div");
  controls.className = "ticket-label-controls";

  const input = document.createElement("input");
  input.name = "labels";
  input.value = ticket.labels.join(", ");
  input.placeholder = "backend, auth";
  input.setAttribute("aria-label", "Labels");

  const update = document.createElement("button");
  update.type = "button";
  update.dataset.updateLabelsId = ticket.id;
  update.textContent = "Labels";

  controls.append(input, update);
  section.append(chips, controls);
  return section;
}

function storyPointControlNode(ticket) {
  const section = document.createElement("section");
  section.className = "ticket-story-points";
  section.setAttribute("data-ticket-story-points-control", "true");
  section.setAttribute("aria-label", `${ticket.key} story points`);

  const heading = document.createElement("p");
  heading.className = "story-points-heading";
  heading.textContent = storyPointLabel(ticket.story_points) || "Unestimated";

  const controls = document.createElement("div");
  controls.className = "ticket-story-points-controls";

  const input = document.createElement("input");
  input.name = "story_points";
  input.type = "number";
  input.min = "0";
  input.step = "0.5";
  input.value = ticket.story_points === null || ticket.story_points === undefined ? "" : String(ticket.story_points);
  input.placeholder = "Story points";
  input.setAttribute("aria-label", "Story points");

  const update = document.createElement("button");
  update.type = "button";
  update.dataset.updateStoryPointsId = ticket.id;
  update.textContent = "Points";

  controls.append(input, update);
  section.append(heading, controls);
  return section;
}

function planningControlNode(ticket) {
  const section = document.createElement("section");
  section.className = "ticket-planning";
  section.setAttribute("data-ticket-planning-control", "true");
  section.setAttribute("aria-label", `${ticket.key} component and version`);

  const heading = document.createElement("p");
  heading.className = "planning-heading";
  const component = ticket.component_id ? componentName(ticket.component_id) : "No component";
  const version = ticket.version_id ? versionName(ticket.version_id) : "No version";
  heading.textContent = `${component} / ${version}`;

  const controls = document.createElement("div");
  controls.className = "ticket-planning-controls";

  const componentSelect = document.createElement("select");
  componentSelect.setAttribute("aria-label", "Component");
  componentSelect.setAttribute("data-ticket-component-select", "true");
  appendSelectOptions(componentSelect, "Component", state.components, (item) => item.name, ticket.component_id);

  const versionSelect = document.createElement("select");
  versionSelect.setAttribute("aria-label", "Version");
  versionSelect.setAttribute("data-ticket-version-select", "true");
  appendSelectOptions(versionSelect, "Version", state.versions, (item) => `${item.name} (${item.state})`, ticket.version_id);

  const assign = document.createElement("button");
  assign.type = "button";
  assign.dataset.assignPlanningId = ticket.id;
  assign.textContent = "Set";

  controls.append(componentSelect, versionSelect, assign);
  section.append(heading, controls);
  return section;
}

function customFieldControlNode(ticket) {
  const section = document.createElement("section");
  section.className = "ticket-custom-fields";
  section.setAttribute("data-ticket-custom-field-control", "true");
  section.setAttribute("aria-label", `${ticket.key} custom fields`);

  const heading = document.createElement("p");
  heading.className = "custom-field-heading";
  heading.textContent = state.customFields.length ? `Custom fields (${state.customFields.length})` : "Custom fields";

  const summary = document.createElement("div");
  summary.className = "custom-field-summary";
  const entries = Object.entries(ticket.custom_fields || {});
  if (entries.length) {
    for (const [key, value] of entries) {
      const row = document.createElement("span");
      row.textContent = `${key}: ${customFieldValueLabel(value)}`;
      summary.append(row);
    }
  } else {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No custom values";
    summary.append(empty);
  }

  const controls = document.createElement("div");
  controls.className = "ticket-custom-field-controls";
  renderCustomFieldInputs(controls, ticket.custom_fields || {});

  const update = document.createElement("button");
  update.type = "button";
  update.dataset.updateCustomFieldsId = ticket.id;
  update.textContent = "Fields";

  controls.append(update);
  section.append(heading, summary, controls);
  return section;
}

function renderTicketCreateCustomFields() {
  if (!els.ticketCustomFields) {
    return;
  }
  let current = {};
  try {
    current = customFieldsFromControls(els.ticketCustomFields);
  } catch (_) {
    current = {};
  }
  renderCustomFieldInputs(els.ticketCustomFields, current);
}

function renderCustomFieldInputs(container, values = {}) {
  if (!container) {
    return;
  }
  container.replaceChildren();
  if (!state.customFields.length) {
    const input = document.createElement("textarea");
    input.name = "custom_fields";
    input.rows = 3;
    input.value = formatCustomFields(values);
    input.placeholder = customFieldPlaceholder();
    input.setAttribute("aria-label", "Custom fields JSON");
    container.append(input);
    return;
  }
  for (const field of state.customFields) {
    const hasValue = values && Object.prototype.hasOwnProperty.call(values, field.key);
    container.append(customFieldInputNode(field, hasValue ? values[field.key] : undefined, hasValue));
  }
}

function customFieldInputNode(field, value, hasValue = false) {
  const label = document.createElement("label");
  label.className = "custom-field-input";
  label.dataset.customFieldKey = field.key;
  label.dataset.customFieldType = field.field_type || "text";
  label.dataset.customFieldRequired = field.required ? "true" : "false";
  label.dataset.customFieldValueSet = hasValue ? "true" : "false";

  const text = document.createElement("span");
  text.textContent = field.required ? `${field.name || field.key} *` : field.name || field.key;
  label.append(text);

  let control;
  switch (field.field_type) {
    case "number":
      control = document.createElement("input");
      control.type = "number";
      control.step = "any";
      control.value = value === undefined || value === null ? "" : String(value);
      break;
    case "boolean":
      control = document.createElement("input");
      control.type = "checkbox";
      control.checked = Boolean(value);
      break;
    case "date":
      control = document.createElement("input");
      control.type = "date";
      control.value = value ? String(value) : "";
      break;
    case "single_select":
      control = document.createElement("select");
      appendCustomFieldOption(control, "", "Choose");
      for (const option of field.options || []) {
        appendCustomFieldOption(control, option, option, value === option);
      }
      break;
    case "multi_select":
      control = document.createElement("select");
      control.multiple = true;
      control.size = Math.min(Math.max((field.options || []).length, 2), 5);
      for (const option of field.options || []) {
        appendCustomFieldOption(control, option, option, Array.isArray(value) && value.includes(option));
      }
      break;
    case "user":
      control = document.createElement("input");
      control.type = "text";
      control.placeholder = "User ID";
      control.value = value ? String(value) : "";
      break;
    default:
      control = document.createElement("input");
      control.type = "text";
      control.value = value === undefined || value === null ? "" : String(value);
  }
  control.dataset.customFieldInput = "true";
  control.setAttribute("aria-label", field.name || field.key);
  label.append(control);
  return label;
}

function appendCustomFieldOption(select, value, label, selected = false) {
  const option = document.createElement("option");
  option.value = value;
  option.textContent = label;
  option.selected = Boolean(selected);
  select.append(option);
}

function sprintControlNode(ticket) {
  const section = document.createElement("section");
  section.className = "ticket-sprint";
  section.setAttribute("data-ticket-sprint-control", "true");
  section.setAttribute("aria-label", `${ticket.key} sprint`);

  const heading = document.createElement("p");
  heading.className = "sprint-heading";
  heading.textContent = ticket.sprint_id ? `Sprint: ${sprintName(ticket.sprint_id)}` : "Sprint";

  const controls = document.createElement("div");
  controls.className = "ticket-sprint-controls";

  const select = document.createElement("select");
  select.setAttribute("aria-label", "Sprint");
  const empty = document.createElement("option");
  empty.value = "";
  empty.textContent = "Choose sprint";
  select.append(empty);
  for (const sprint of state.sprints) {
    if (sprint.state === "completed" && sprint.id !== ticket.sprint_id) {
      continue;
    }
    const option = document.createElement("option");
    option.value = sprint.id;
    option.textContent = `${sprint.name} (${sprint.state})`;
    option.selected = sprint.id === ticket.sprint_id;
    select.append(option);
  }

  const assign = document.createElement("button");
  assign.type = "button";
  assign.dataset.assignSprintId = ticket.id;
  assign.textContent = "Assign";
  assign.disabled = !state.sprints.length;

  const remove = document.createElement("button");
  remove.type = "button";
  remove.dataset.removeSprintId = ticket.id;
  remove.textContent = "Remove";
  remove.disabled = !ticket.sprint_id;

  controls.append(select, assign, remove);
  section.append(heading, controls);
  return section;
}

function ticketLinksNode(ticket) {
  const section = document.createElement("section");
  section.className = "ticket-links";
  section.setAttribute("aria-label", `${ticket.key} linked issues`);

  const links = state.ticketLinks[ticket.id] || [];
  const heading = document.createElement("p");
  heading.className = "ticket-link-heading";
  heading.textContent = `Links (${links.length})`;
  section.append(heading);

  section.append(ticketLinkDependencyOverviewNode(links));

  const list = document.createElement("div");
  list.className = "ticket-link-list";
  if (!links.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No linked issues";
    list.append(empty);
  } else {
    for (const link of links) {
      list.append(ticketLinkItemNode(link, ticket.id));
    }
  }
  section.append(list);

  const form = document.createElement("form");
  form.className = "ticket-link-form";
  form.dataset.ticketLinkForm = "true";
  form.dataset.ticketId = ticket.id;

  const type = document.createElement("select");
  type.name = "link_type";
  type.setAttribute("aria-label", "Link type");
  for (const [value, label] of [["blocks", "Blocks"], ["is_blocked_by", "Is blocked by"], ["relates_to", "Relates to"]]) {
    const option = document.createElement("option");
    option.value = value;
    option.textContent = label;
    type.append(option);
  }

  const target = document.createElement("select");
  target.name = "target_ticket_id";
  target.required = true;
  target.setAttribute("aria-label", "Linked issue");
  const empty = document.createElement("option");
  empty.value = "";
  empty.textContent = "Issue";
  target.append(empty);
  const candidates = linkableTickets(ticket.id);
  for (const candidate of candidates) {
    const option = document.createElement("option");
    option.value = candidate.id;
    option.textContent = `${candidate.key} ${candidate.title}`;
    target.append(option);
  }

  const submit = document.createElement("button");
  submit.type = "submit";
  submit.textContent = "Link";
  submit.disabled = !candidates.length;

  form.append(type, target, submit);
  section.append(form);
  return section;
}

function ticketLinkDependencyOverviewNode(links) {
  const summary = ticketLinkDependencySummary(links);
  const overview = document.createElement("div");
  overview.className = "ticket-link-dependency-overview";

  if (!summary.total) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No dependencies";
    overview.append(empty);
    return overview;
  }

  for (const item of summary.items) {
    const chip = document.createElement("span");
    chip.textContent = `${item.label}: ${item.count}`;
    overview.append(chip);
  }
  return overview;
}

function ticketLinkDependencySummary(links) {
  const counts = {
    blocks: 0,
    is_blocked_by: 0,
    relates_to: 0
  };
  for (const link of Array.isArray(links) ? links : []) {
    if (Object.prototype.hasOwnProperty.call(counts, link.link_type)) {
      counts[link.link_type] += 1;
    }
  }
  return {
    total: Object.values(counts).reduce((sum, count) => sum + count, 0),
    items: [
      { label: "Blocks", count: counts.blocks },
      { label: "Blocked by", count: counts.is_blocked_by },
      { label: "Related", count: counts.relates_to }
    ]
  };
}

function ticketLinkItemNode(link, ticketID) {
  const row = document.createElement("article");
  row.className = "ticket-link-item";

  const label = document.createElement("span");
  label.textContent = ticketLinkTypeLabel(link.link_type);

  const target = document.createElement("a");
  target.href = `/issues/${encodeURIComponent(link.target.id)}`;
  target.textContent = `${link.target.key} ${link.target.title}`;

  const meta = document.createElement("span");
  meta.textContent = [link.target.status, link.target.type].filter(Boolean).join(" / ") || "issue";

  const remove = document.createElement("button");
  remove.type = "button";
  remove.dataset.deleteTicketLinkId = link.id;
  remove.dataset.ticketId = ticketID;
  remove.setAttribute("aria-label", `Remove link to ${link.target.key}`);
  remove.textContent = "Remove";

  row.append(label, target, meta, remove);
  return row;
}

function linkableTickets(ticketID) {
  return state.tickets.filter((ticket) => ticket.id !== ticketID);
}

function ticketLinkTypeLabel(value) {
  const labels = {
    blocks: "blocks",
    is_blocked_by: "is blocked by",
    relates_to: "relates to"
  };
  return labels[value] || value || "links";
}

function commentNode(ticket) {
  const section = document.createElement("section");
  section.className = "ticket-comments";
  section.setAttribute("aria-label", `${ticket.key} comments`);

  const comments = state.comments[ticket.id] || [];
  const heading = document.createElement("p");
  heading.className = "comment-heading";
  heading.textContent = `Comments (${comments.length})`;
  section.append(heading);

  const list = document.createElement("div");
  list.className = "comment-list";
  if (!comments.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No comments";
    list.append(empty);
  } else {
    for (const comment of comments) {
      const row = document.createElement("article");
      row.className = "comment-item";

      const body = document.createElement("p");
      body.textContent = comment.body;

      const meta = document.createElement("span");
      meta.textContent = comment.author_id ? `by ${comment.author_id}` : "comment";

      const remove = document.createElement("button");
      remove.type = "button";
      remove.dataset.deleteCommentId = comment.id;
      remove.dataset.ticketId = ticket.id;
      remove.setAttribute("aria-label", "Delete comment");
      remove.textContent = "Delete";

      row.append(body, meta, remove);
      list.append(row);
    }
  }
  section.append(list);

  const form = document.createElement("form");
  form.className = "comment-form";
  form.dataset.commentForm = "true";
  form.dataset.ticketId = ticket.id;

  const textarea = document.createElement("textarea");
  textarea.name = "body";
  textarea.rows = 2;
  textarea.placeholder = "Add a comment. Mention people with @username";
  textarea.required = true;

  const submit = document.createElement("button");
  submit.type = "submit";
  submit.textContent = "Comment";

  form.append(textarea, submit);
  section.append(form);
  return section;
}

function attachmentNode(ticket) {
  const section = document.createElement("section");
  section.className = "ticket-attachments";
  section.setAttribute("aria-label", `${ticket.key} attachments`);

  const attachments = state.attachments[ticket.id] || [];
  const heading = document.createElement("p");
  heading.className = "attachment-heading";
  heading.textContent = `Attachments (${attachments.length})`;
  section.append(heading);

  const list = document.createElement("div");
  list.className = "attachment-list";
  if (!attachments.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No files";
    list.append(empty);
  } else {
    for (const attachment of attachments) {
      const row = document.createElement("div");
      row.className = "attachment-item";

      const link = document.createElement("a");
      link.href = `/api/attachments/${attachment.id}/download`;
      link.textContent = attachment.file_name;

      const size = document.createElement("span");
      size.textContent = formatBytes(attachment.size_bytes);

      const remove = document.createElement("button");
      remove.type = "button";
      remove.dataset.deleteAttachmentId = attachment.id;
      remove.dataset.ticketId = ticket.id;
      remove.setAttribute("aria-label", `Delete ${attachment.file_name}`);
      remove.textContent = "Delete";

      row.append(link, size, remove);
      list.append(row);
    }
  }
  section.append(list);

  const form = document.createElement("form");
  form.className = "attachment-form";
  form.dataset.attachmentForm = "true";
  form.dataset.ticketId = ticket.id;

  const input = document.createElement("input");
  input.type = "file";
  input.name = "file";
  input.required = true;

  const submit = document.createElement("button");
  submit.type = "submit";
  submit.textContent = "Attach";

  form.append(input, submit);
  section.append(form);
  return section;
}

function activityNode(ticket) {
  const section = document.createElement("section");
  section.className = "ticket-activity";
  section.setAttribute("aria-label", `${ticket.key} activity`);

  const activities = state.activities[ticket.id] || [];
  const heading = document.createElement("p");
  heading.className = "activity-heading";
  heading.textContent = `Activity (${activities.length})`;
  section.append(heading);

  const list = document.createElement("div");
  list.className = "activity-list";
  if (!activities.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No activity";
    list.append(empty);
  } else {
    for (const activity of activities) {
      list.append(activityItemNode(activity));
    }
  }
  section.append(list);
  return section;
}

function activityItemNode(activity) {
  const row = document.createElement("article");
  row.className = "activity-item";

  const title = document.createElement("p");
  title.textContent = activityLabel(activity);

  const meta = document.createElement("span");
  meta.textContent = [
    activity.actor_id ? `by ${activity.actor_id}` : "",
    activity.created_at ? formatDateTime(activity.created_at) : ""
  ].filter(Boolean).join(" / ");

  const details = document.createElement("code");
  details.textContent = activityDataLabel(activity.data);

  row.append(title, meta);
  if (details.textContent) {
    row.append(details);
  }
  return row;
}

function activityLabel(activity) {
  const labels = {
    "ticket.created": "Ticket created",
    "ticket.updated": "Ticket updated",
    "ticket.deleted": "Ticket deleted",
    "ticket.link_created": "Issue link added",
    "ticket.link_deleted": "Issue link removed",
    "ticket.watcher_added": "Watcher added",
    "ticket.watcher_removed": "Watcher removed",
    "comment.created": "Comment added",
    "comment.deleted": "Comment deleted",
    "attachment.uploaded": "Attachment uploaded",
    "attachment.deleted": "Attachment deleted"
  };
  return labels[activity.activity_type] || activity.activity_type || "Activity";
}

function activityDataLabel(data) {
  if (!data || typeof data !== "object" || Array.isArray(data) || !Object.keys(data).length) {
    return "";
  }
  const parts = [];
  if (data.key) {
    parts.push(`key: ${data.key}`);
  }
  if (data.changes && typeof data.changes === "object" && !Array.isArray(data.changes)) {
    parts.push(`changed: ${Object.keys(data.changes).join(", ")}`);
  }
  if (data.custom_fields) {
    parts.push(`custom fields: ${data.custom_fields}`);
  }
  if (data.labels && Array.isArray(data.labels)) {
    parts.push(`labels: ${data.labels.join(", ")}`);
  }
  if (data.body) {
    parts.push(`body: ${data.body}`);
  }
  if (data.comment_id) {
    parts.push(`comment: ${data.comment_id}`);
  }
  if (data.file_name) {
    parts.push(`file: ${data.file_name}`);
  }
  if (data.size_bytes) {
    parts.push(`size: ${formatBytes(data.size_bytes)}`);
  }
  if (parts.length) {
    return parts.join(" / ");
  }
  return JSON.stringify(data);
}

function statusActions(status) {
  if (state.workflowStatuses.length > 1) {
    return state.workflowStatuses
      .filter((item) => item.slug !== status)
      .map((item) => ({ label: item.name || item.slug, status: item.slug }));
  }
  switch (status) {
    case "todo":
      return [{ label: "Start", status: "in_progress" }, { label: "Done", status: "done" }];
    case "in_progress":
      return [{ label: "Todo", status: "todo" }, { label: "Done", status: "done" }];
    case "done":
      return [{ label: "Reopen", status: "todo" }];
    default:
      return [{ label: "Todo", status: "todo" }];
  }
}

function statusName(slug) {
  const status = state.workflowStatuses.find((item) => item.slug === slug);
  return status ? status.name : slug;
}

function defaultWorkflowStatuses() {
  return [
    { slug: "todo", name: "Todo", position: 0 },
    { slug: "in_progress", name: "In Progress", position: 1 },
    { slug: "done", name: "Done", position: 2 }
  ];
}

function cssEscape(value) {
  if (window.CSS && typeof window.CSS.escape === "function") {
    return window.CSS.escape(String(value || ""));
  }
  return String(value || "").replace(/"/g, '\\"');
}

function dataTransferHasType(event, type) {
  return Array.from(event.dataTransfer && event.dataTransfer.types ? event.dataTransfer.types : []).includes(type);
}

function formData(form) {
  return Object.fromEntries(new FormData(form).entries());
}

function selectedFormValues(form, name) {
  if (!form || !form.elements[name]) {
    return [];
  }
  const field = form.elements[name];
  if (field.selectedOptions) {
    return [...field.selectedOptions].map((option) => option.value).filter(Boolean);
  }
  return parseCommaList(field.value);
}

function setFormValue(form, name, value) {
  if (form && form.elements[name]) {
    form.elements[name].value = value;
  }
}

function setFormChecked(form, name, checked) {
  if (form && form.elements[name]) {
    form.elements[name].checked = Boolean(checked);
  }
}

function parseLabels(value) {
  return String(value || "")
    .split(",")
    .map((label) => label.trim())
    .filter(Boolean);
}

function parseCommaList(value) {
  return String(value || "")
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);
}

function parseBoardWIPLimits(value) {
  const limits = {};
  for (const item of parseCommaList(value)) {
    const [rawSlug, rawLimit, ...rest] = item.split("=");
    const slug = String(rawSlug || "").trim();
    const limitText = String(rawLimit || "").trim();
    if (!slug || rest.length || limitText === "") {
      throw new Error("Board WIP limits must use status=limit pairs");
    }
    const limit = Number(limitText);
    if (!Number.isInteger(limit) || limit < 0) {
      throw new Error("Board WIP limits must be whole numbers zero or greater");
    }
    limits[slug] = limit;
  }
  return limits;
}

function formatBoardWIPLimits(limits) {
  return Object.entries(limits || {})
    .map(([status, limit]) => `${status}=${limit}`)
    .join(", ");
}

function parseOptions(value) {
  return String(value || "")
    .split(",")
    .map((option) => option.trim())
    .filter(Boolean);
}

function parseCustomFields(value) {
  const text = String(value || "").trim();
  if (!text) {
    return {};
  }
  try {
    const parsed = JSON.parse(text);
    if (!parsed || Array.isArray(parsed) || typeof parsed !== "object") {
      throw new Error("Custom fields must be a JSON object");
    }
    return parsed;
  } catch (error) {
    if (error.message && error.message.includes("must be a JSON object")) {
      throw error;
    }
    throw new Error("Custom fields JSON is not valid");
  }
}

function customFieldsFromControls(container) {
  if (!container) {
    return {};
  }
  const fallback = container.querySelector("textarea[name='custom_fields']");
  if (fallback) {
    return parseCustomFields(fallback.value);
  }
  const fields = {};
  for (const wrapper of container.querySelectorAll("[data-custom-field-key]")) {
    const key = wrapper.dataset.customFieldKey;
    const type = wrapper.dataset.customFieldType || "text";
    const input = wrapper.querySelector("[data-custom-field-input]");
    if (!key || !input) {
      continue;
    }
    if (type === "boolean") {
      if (input.checked || wrapper.dataset.customFieldRequired === "true" || wrapper.dataset.customFieldValueSet === "true") {
        fields[key] = Boolean(input.checked);
      }
      continue;
    }
    if (type === "multi_select") {
      const selected = Array.from(input.selectedOptions).map((option) => option.value).filter(Boolean);
      if (selected.length) {
        fields[key] = selected;
      }
      continue;
    }
    const value = String(input.value || "").trim();
    if (!value) {
      continue;
    }
    if (type === "number") {
      const number = Number(value);
      if (!Number.isFinite(number)) {
        throw new Error(`${key} must be a number`);
      }
      fields[key] = number;
      continue;
    }
    fields[key] = value;
  }
  return fields;
}

function formatCustomFields(value) {
  const fields = value && typeof value === "object" && !Array.isArray(value) ? value : {};
  if (!Object.keys(fields).length) {
    return "";
  }
  return JSON.stringify(fields, null, 2);
}

function customFieldValueLabel(value) {
  if (Array.isArray(value)) {
    return value.join(", ");
  }
  if (value === null) {
    return "null";
  }
  if (typeof value === "object") {
    return JSON.stringify(value);
  }
  return String(value);
}

function customFieldPlaceholder() {
  if (!state.customFields.length) {
    return "{}";
  }
  const example = {};
  for (const field of state.customFields.slice(0, 3)) {
    example[field.key] = customFieldExampleValue(field);
  }
  return JSON.stringify(example);
}

function customFieldExampleValue(field) {
  switch (field.field_type) {
    case "number":
      return 1;
    case "boolean":
      return true;
    case "date":
      return "2026-06-17";
    case "multi_select":
      return field.options.length ? [field.options[0]] : [];
    case "single_select":
      return field.options[0] || "";
    default:
      return "";
  }
}

function customFieldUpdateSpec(form) {
  const data = formData(form);
  return {
    key: data.key || "",
    name: data.name || "",
    field_type: data.field_type || "text",
    required: Boolean(data.required),
    options: parseOptions(data.options)
  };
}

function componentUpdateSpec(form) {
  const data = formData(form);
  return {
    name: data.name || "",
    description: data.description || "",
    owner_user_id: data.owner_user_id || "",
    default_assignee_id: data.default_assignee_id || ""
  };
}

function versionUpdateSpec(form) {
  const data = formData(form);
  return {
    name: data.name || "",
    description: data.description || "",
    status: data.status || "planned",
    target_date: data.target_date || "",
    release_date: data.release_date || ""
  };
}

function preferenceKeys() {
  return [
    "in_app_enabled",
    "external_enabled",
    "assignment_enabled",
    "comment_enabled",
    "status_change_enabled",
    "sprint_change_enabled",
    "release_change_enabled",
    "automation_failure_enabled"
  ];
}

function notificationPreferenceSpec(form) {
  const data = formData(form);
  const spec = {};
  for (const key of preferenceKeys()) {
    spec[key] = Boolean(data[key]);
  }
  return spec;
}

function auditQuery() {
  const params = new URLSearchParams();
  if (!els.auditForm) {
    params.set("limit", "25");
    return `?${params.toString()}`;
  }
  const data = formData(els.auditForm);
  for (const key of ["event_type", "actor_user_id", "subject_type", "subject_id", "outcome"]) {
    if (data[key]) {
      params.set(key, data[key]);
    }
  }
  params.set("limit", String(Math.min(Math.max(Number(data.limit || 25), 1), 500)));
  return `?${params.toString()}`;
}

function openRouterProviderSpec(form) {
  const data = formData(form);
  return {
    name: data.name || "",
    default_model: data.default_model || "",
    api_key: data.api_key || "",
    allowed_models: parseCommaList(data.allowed_models),
    default_timeout_seconds: Number(data.default_timeout_seconds || 0),
    max_output_tokens: Number(data.max_output_tokens || 0),
    enabled: Boolean(data.enabled)
  };
}

function openRouterProviderUpdateSpec(form) {
  const spec = openRouterProviderSpec(form);
  if (!spec.api_key) {
    delete spec.api_key;
  }
  return spec;
}

function notificationDestinationSpec(form) {
  const data = formData(form);
  return {
    name: data.name || "",
    shoutrrr_url: data.shoutrrr_url || "",
    enabled: Boolean(data.enabled)
  };
}

function notificationDestinationUpdateSpec(form) {
  const spec = notificationDestinationSpec(form);
  if (!spec.shoutrrr_url) {
    delete spec.shoutrrr_url;
  }
  return spec;
}

function searchSpecFromForm(form) {
  const data = formData(form);
  return {
    text: data.text || "",
    filter: data.filter || "",
    project_id: state.selectedProject ? state.selectedProject.id : ""
  };
}

function savedViewSpecFromForm(form) {
  const data = formData(form);
  const scopeType = data.scope_type || "user";
  const currentSearch = searchSpecFromForm(els.searchForm);
  return {
    name: data.name || "Saved view",
    scope_type: scopeType,
    project_id: scopeType === "project" && state.selectedProject ? state.selectedProject.id : "",
    query: {
      text: currentSearch.text || "",
      filter: currentSearch.filter || ""
    },
    sort: parseSavedViewSort(data.sort),
    columns: parseCommaList(data.columns || "key, title, status, priority"),
    display_mode: data.display_mode || "list",
    group_by: data.group_by || "",
    pinned: Boolean(data.pinned)
  };
}

function parseSavedViewSort(value) {
  const text = String(value || "").trim();
  if (!text) {
    return [{ field: "updated_at", direction: "desc" }];
  }
  return parseJSONArrayField(text, "Saved view sort JSON");
}

function savedViewUpdateSpec(spec) {
  return {
    name: spec.name,
    query: spec.query,
    sort: spec.sort,
    columns: spec.columns,
    display_mode: spec.display_mode,
    group_by: spec.group_by,
    pinned: spec.pinned
  };
}

function resetSearchPagination() {
  state.searchNextCursor = "";
  state.searchCursorStack = [""];
  state.searchCursorIndex = 0;
}

function resetSavedViewPagination() {
  state.savedViewOffset = 0;
  state.savedViewHasMore = false;
}

function setSearchForm(query) {
  if (!els.searchForm) {
    return;
  }
  els.searchForm.elements.text.value = query.text || "";
  els.searchForm.elements.filter.value = query.filter || "";
}

function editSavedView(view) {
  if (!els.savedViewForm) {
    return;
  }
  els.savedViewForm.dataset.savedViewEditId = view.id;
  setFormValue(els.savedViewForm, "name", view.name || "");
  setFormValue(els.savedViewForm, "scope_type", view.scope_type || "user");
  setFormValue(els.savedViewForm, "display_mode", view.display_mode || "list");
  setFormValue(els.savedViewForm, "group_by", view.group_by || "");
  setFormValue(els.savedViewForm, "columns", (view.columns || []).join(", "));
  setFormValue(els.savedViewForm, "sort", JSON.stringify(view.sort && view.sort.length ? view.sort : [{ field: "updated_at", direction: "desc" }]));
  setFormChecked(els.savedViewForm, "pinned", Boolean(view.pinned));
  setSearchForm(view.query || {});
  if (els.savedViewCancelEdit) {
    els.savedViewCancelEdit.hidden = false;
  }
}

function resetSavedViewForm() {
  if (!els.savedViewForm) {
    return;
  }
  delete els.savedViewForm.dataset.savedViewEditId;
  els.savedViewForm.reset();
  setFormValue(els.savedViewForm, "display_mode", "list");
  setFormValue(els.savedViewForm, "sort", `[{"field":"updated_at","direction":"desc"}]`);
  setFormValue(els.savedViewForm, "columns", "key, title, status, priority");
  if (els.savedViewCancelEdit) {
    els.savedViewCancelEdit.hidden = true;
  }
}

function listItems(data) {
  if (!data) {
    return [];
  }
  if (Array.isArray(data.items)) {
    return data.items;
  }
  if (data.status && Array.isArray(data.status.items)) {
    return data.status.items;
  }
  return [];
}

function normalizeProject(project) {
  if (!project) {
    return null;
  }
  if (project.metadata && project.spec) {
    return {
      id: project.metadata.id,
      key: project.spec.key,
      name: project.spec.name,
      description: project.spec.description || "",
      lead_user_id: project.spec.lead_user_id || ""
    };
  }
  return project;
}

function normalizeProjectLabel(label) {
  if (!label) {
    return null;
  }
  if (label.metadata && label.spec && label.status) {
    return {
      id: label.metadata.id || label.spec.label || "",
      project_id: label.metadata.project_id || "",
      created_at: label.metadata.created_at || "",
      updated_at: label.metadata.updated_at || "",
      label: label.spec.label || label.metadata.id || "",
      description: label.spec.description || "",
      color: label.spec.color || "",
      ticket_count: Number(label.status.ticket_count) || 0
    };
  }
  return label;
}

function normalizeTicket(ticket) {
  if (!ticket) {
    return null;
  }
  if (ticket.metadata && ticket.spec && ticket.status) {
    return {
      id: ticket.metadata.id,
      project_id: ticket.metadata.project_id,
      created_at: ticket.metadata.created_at,
      updated_at: ticket.metadata.updated_at,
      key: ticket.status.key,
      title: ticket.spec.title,
      description: ticket.spec.description || "",
      status: ticket.spec.status,
      priority: ticket.spec.priority || "",
      type: ticket.spec.type || "",
      reporter_id: ticket.status.reporter_id || "",
      assignee_id: ticket.spec.assignee_id || "",
      parent_ticket_id: ticket.spec.parent_ticket_id || "",
      sprint_id: ticket.spec.sprint_id || "",
      component_id: ticket.spec.component_id || "",
      version_id: ticket.spec.version_id || "",
      rank: ticket.spec.rank || "",
      start_date: ticket.spec.start_date || "",
      due_date: ticket.spec.due_date || "",
      story_points: ticket.spec.story_points === undefined ? null : ticket.spec.story_points,
      labels: ticket.spec.labels || [],
      custom_fields: ticket.spec.custom_fields || {},
      watcher_count: Number(ticket.status.watcher_count || 0),
      watching: Boolean(ticket.status.watching)
    };
  }
  return ticket;
}

function normalizeTicketWatcher(watcher) {
  if (!watcher) {
    return null;
  }
  if (watcher.metadata && watcher.spec) {
    return {
      ticket_id: watcher.metadata.ticket_id || "",
      user_id: watcher.metadata.user_id || "",
      created_at: watcher.metadata.created_at || "",
      username: watcher.spec.username || "",
      display_name: watcher.spec.display_name || ""
    };
  }
  return watcher;
}

function normalizeTicketLink(link) {
  if (!link) {
    return null;
  }
  if (link.metadata && link.spec && link.status) {
    return {
      id: link.metadata.id,
      project_id: link.metadata.project_id,
      created_at: link.metadata.created_at,
      link_type: link.spec.link_type || "",
      source: normalizeTicket(link.spec.source),
      target: normalizeTicket(link.spec.target),
      created_by: link.status.created_by || ""
    };
  }
  return link;
}

function normalizeRoadmapDependency(dependency) {
  if (!dependency) {
    return null;
  }
  if (dependency.metadata && dependency.spec) {
    return {
      id: dependency.metadata.id || "",
      project_id: dependency.metadata.project_id || "",
      source_epic_id: dependency.spec.source_epic_id || "",
      target_epic_id: dependency.spec.target_epic_id || "",
      link: normalizeTicketLink(dependency.spec.link)
    };
  }
  return dependency;
}

function normalizeWorkflowStatus(status) {
  if (!status) {
    return null;
  }
  if (status.metadata && status.spec) {
    return {
      id: status.metadata.id,
      project_id: status.metadata.project_id,
      slug: status.spec.slug || "",
      name: status.spec.name || "",
      position: Number(status.spec.position || 0)
    };
  }
  return status;
}

function normalizeBoard(board) {
  if (!board) {
    return null;
  }
  if (board.metadata && board.spec && board.status) {
    return {
      id: board.metadata.id,
      project_id: board.metadata.project_id,
      name: board.spec.name || "",
      description: board.spec.description || "",
      status_slugs: board.spec.status_slugs || [],
      wip_limits: board.spec.wip_limits || {},
      columns: board.status.columns || []
    };
  }
  return board;
}

function normalizeBoardTickets(data) {
  if (!data || !data.metadata || !data.status) {
    return null;
  }
  return {
    id: data.metadata.id,
    project_id: data.metadata.project_id,
    board: data.spec && data.spec.board ? normalizeBoard(data.spec.board) : null,
    columns: (data.status.columns || []).map((column) => ({
      slug: column.column ? column.column.status_slug : "",
      name: column.column ? column.column.name : "",
      wip_limit: column.column && Number.isFinite(column.column.wip_limit) ? column.column.wip_limit : null,
      ticket_count: Number.isFinite(column.ticket_count) ? column.ticket_count : (column.tickets || []).length,
      over_wip_limit: Boolean(column.over_wip_limit),
      tickets: (column.tickets || []).map(normalizeTicket).filter(Boolean)
    }))
  };
}

function normalizeAttachment(attachment) {
  if (!attachment) {
    return null;
  }
  if (attachment.metadata && attachment.spec && attachment.status) {
    return {
      id: attachment.metadata.id,
      ticket_id: attachment.metadata.ticket_id,
      created_at: attachment.metadata.created_at,
      file_name: attachment.spec.file_name,
      content_type: attachment.spec.content_type,
      size_bytes: attachment.status.size_bytes || 0,
      uploader_id: attachment.status.uploader_id || ""
    };
  }
  return attachment;
}

function normalizeNotification(notification) {
  if (!notification) {
    return null;
  }
  if (notification.metadata && notification.spec && notification.status) {
    return {
      id: notification.metadata.id,
      user_id: notification.metadata.user_id,
      created_at: notification.metadata.created_at,
      type: notification.spec.type || "",
      subject_type: notification.spec.subject_type || "",
      subject_id: notification.spec.subject_id || "",
      body: notification.spec.body || "",
      data: notification.spec.data || {},
      read_at: notification.status.read_at || null
    };
  }
  return notification;
}

function normalizeSavedView(view) {
  if (!view) {
    return null;
  }
  if (view.metadata && view.spec) {
    return {
      id: view.metadata.id,
      created_at: view.metadata.created_at,
      updated_at: view.metadata.updated_at,
      owner_user_id: view.spec.owner_user_id || "",
      project_id: view.spec.project_id || "",
      scope_type: view.spec.scope_type || "user",
      name: view.spec.name || "",
      query: view.spec.query || {},
      sort: view.spec.sort || [],
      columns: view.spec.columns || [],
      display_mode: view.spec.display_mode || "list",
      group_by: view.spec.group_by || "",
      pinned: Boolean(view.spec.pinned)
    };
  }
  return view;
}

function normalizeSprint(sprint) {
  if (!sprint) {
    return null;
  }
  if (sprint.metadata && sprint.spec && sprint.status) {
    return {
      id: sprint.metadata.id,
      project_id: sprint.metadata.project_id,
      created_at: sprint.metadata.created_at,
      updated_at: sprint.metadata.updated_at,
      name: sprint.spec.name || "",
      goal: sprint.spec.goal || "",
      start_date: sprint.spec.start_date || "",
      end_date: sprint.spec.end_date || "",
      state: sprint.status.state || "planned",
      started_at: sprint.status.started_at || "",
      completed_at: sprint.status.completed_at || ""
    };
  }
  return sprint;
}

function normalizeSprintReport(report) {
  if (!report) {
    return null;
  }
  if (report.metadata && report.spec && report.status) {
    return {
      sprint: normalizeSprint(report.spec.sprint) || {
        id: report.metadata.id,
        project_id: report.metadata.project_id
      },
      scope: report.status.scope || "current",
      snapshot_at: report.status.snapshot_at || "",
      progress: {
        total: Number(report.status.progress && report.status.progress.total) || 0,
        done: Number(report.status.progress && report.status.progress.done) || 0,
        story_points_total: Number(report.status.progress && report.status.progress.story_points_total) || 0,
        story_points_done: Number(report.status.progress && report.status.progress.story_points_done) || 0,
        story_points_remaining: Number(report.status.progress && report.status.progress.story_points_remaining) || 0,
        story_points_unestimated: Number(report.status.progress && report.status.progress.story_points_unestimated) || 0,
        by_status: (report.status.progress && report.status.progress.by_status) || {}
      },
      analytics: normalizeSprintAnalytics(report.status.analytics),
      scope_changes: normalizeSprintReportScopeChanges(report.status.scope_changes),
      tickets: listItems({ items: report.status.tickets || [] }).map(normalizeTicket).filter(Boolean)
    };
  }
  return report;
}

function normalizeSprintReportScopeChanges(scopeChanges) {
  scopeChanges = scopeChanges || {};
  return {
    current: Number(scopeChanges.current) || 0,
    snapshot: Number(scopeChanges.snapshot) || 0,
    added: Number(scopeChanges.added) || 0,
    removed: Number(scopeChanges.removed) || 0,
    unchanged: Number(scopeChanges.unchanged) || 0
  };
}

function normalizeSprintAnalytics(analytics) {
  analytics = analytics || {};
  return {
    burndown: Array.isArray(analytics.burndown) ? analytics.burndown.map(normalizeBurndownPoint) : [],
    burnup: Array.isArray(analytics.burnup) ? analytics.burnup.map(normalizeBurnupPoint) : [],
    velocity: {
      completed: Number(analytics.velocity && analytics.velocity.completed) || 0,
      unit: (analytics.velocity && analytics.velocity.unit) || "tickets"
    }
  };
}

function normalizeBurndownPoint(point) {
  return {
    date: point && point.date ? point.date : "",
    remaining: Number(point && point.remaining) || 0
  };
}

function normalizeBurnupPoint(point) {
  return {
    date: point && point.date ? point.date : "",
    total: Number(point && point.total) || 0,
    done: Number(point && point.done) || 0
  };
}

function normalizeComponent(component) {
  if (!component) {
    return null;
  }
  if (component.metadata && component.spec) {
    return {
      id: component.metadata.id,
      project_id: component.metadata.project_id,
      created_at: component.metadata.created_at,
      updated_at: component.metadata.updated_at,
      name: component.spec.name || "",
      description: component.spec.description || "",
      owner_user_id: component.spec.owner_user_id || "",
      default_assignee_id: component.spec.default_assignee_id || ""
    };
  }
  return component;
}

function normalizeVersion(version) {
  if (!version) {
    return null;
  }
  if (version.metadata && version.spec && version.status) {
    return {
      id: version.metadata.id,
      project_id: version.metadata.project_id,
      created_at: version.metadata.created_at,
      updated_at: version.metadata.updated_at,
      name: version.spec.name || "",
      description: version.spec.description || "",
      target_date: version.spec.target_date || "",
      release_date: version.spec.release_date || "",
      state: version.status.state || "planned"
    };
  }
  return version;
}

function normalizeVersionReport(report) {
  if (!report) {
    return null;
  }
  if (report.metadata && report.spec && report.status) {
    return {
      version: normalizeVersion(report.spec.version) || {
        id: report.metadata.id,
        project_id: report.metadata.project_id
      },
      scope: report.status.scope || "current",
      snapshot_at: report.status.snapshot_at || "",
      progress: {
        total: Number(report.status.progress && report.status.progress.total) || 0,
        done: Number(report.status.progress && report.status.progress.done) || 0,
        open: Number(report.status.progress && report.status.progress.open) || 0,
        unassigned_component: Number(report.status.progress && report.status.progress.unassigned_component) || 0,
        story_points_total: Number(report.status.progress && report.status.progress.story_points_total) || 0,
        story_points_done: Number(report.status.progress && report.status.progress.story_points_done) || 0,
        story_points_remaining: Number(report.status.progress && report.status.progress.story_points_remaining) || 0,
        story_points_unestimated: Number(report.status.progress && report.status.progress.story_points_unestimated) || 0,
        by_status: (report.status.progress && report.status.progress.by_status) || {}
      },
      scope_changes: normalizeVersionReportScopeChanges(report.status.scope_changes),
      tickets: listItems({ items: report.status.tickets || [] }).map(normalizeTicket).filter(Boolean)
    };
  }
  return report;
}

function normalizeVersionReportScopeChanges(scopeChanges) {
  scopeChanges = scopeChanges || {};
  return {
    current: Number(scopeChanges.current) || 0,
    snapshot: Number(scopeChanges.snapshot) || 0,
    added: Number(scopeChanges.added) || 0,
    removed: Number(scopeChanges.removed) || 0,
    unchanged: Number(scopeChanges.unchanged) || 0
  };
}

function normalizeCustomField(field) {
  if (!field) {
    return null;
  }
  if (field.metadata && field.spec) {
    return {
      id: field.metadata.id,
      project_id: field.metadata.project_id,
      created_at: field.metadata.created_at,
      updated_at: field.metadata.updated_at,
      key: field.spec.key || "",
      name: field.spec.name || "",
      field_type: field.spec.field_type || "text",
      required: Boolean(field.spec.required),
      options: field.spec.options || []
    };
  }
  return field;
}

function normalizeRoadmapItem(item) {
  if (!item) {
    return null;
  }
  if (item.metadata && item.spec && item.status) {
    return {
      id: item.metadata.id,
      project_id: item.metadata.project_id,
      epic: normalizeTicket(item.spec.epic),
      progress: item.status.progress || { total: 0, done: 0, by_status: {} }
    };
  }
  return item;
}

function normalizeToken(token) {
  if (!token) {
    return null;
  }
  if (token.metadata && token.spec && token.status) {
    return {
      id: token.metadata.id,
      name: token.spec.name || "",
      created_at: token.status.created_at || "",
      last_used_at: token.status.last_used_at || "",
      expires_at: token.status.expires_at || "",
      revoked_at: token.status.revoked_at || "",
      token: token.status.token || ""
    };
  }
  return token;
}

function normalizeUser(user) {
  if (!user) {
    return null;
  }
  if (user.metadata && user.spec && user.status) {
    return {
      id: user.metadata.id,
      username: user.spec.username || "",
      display_name: user.spec.display_name || "",
      disabled: Boolean(user.status.disabled || user.spec.disabled),
      generated_password: user.status.password || ""
    };
  }
  return user;
}

function normalizeGroup(group) {
  if (!group) {
    return null;
  }
  if (group.metadata && group.spec) {
    return {
      id: group.metadata.id,
      name: group.spec.name || "",
      display_name: group.spec.display_name || ""
    };
  }
  return group;
}

function normalizeRole(role) {
  if (!role) {
    return null;
  }
  if (role.metadata && role.spec) {
    return {
      id: role.metadata.id,
      name: role.spec.name || "",
      description: role.spec.description || "",
      permissions: role.spec.permissions || []
    };
  }
  return role;
}

function normalizeRoleBinding(binding) {
  if (!binding) {
    return null;
  }
  if (binding.metadata && binding.spec) {
    return {
      id: binding.metadata.id,
      role_name: binding.spec.role_name || "",
      subject_type: binding.spec.subject_type || "",
      subject_id: binding.spec.subject_id || "",
      scope: binding.spec.scope || "global",
      project_id: binding.spec.project_id || ""
    };
  }
  return binding;
}

function normalizeEffectivePermissions(effective) {
  if (!effective) {
    return null;
  }
  if (effective.metadata && effective.spec && effective.status) {
    return {
      user_id: effective.metadata.user_id || "",
      scope: effective.spec.scope || "global",
      project_id: effective.spec.project_id || "",
      permissions: effective.status.permissions || []
    };
  }
  return effective;
}

function normalizeSettings(settings) {
  if (!settings) {
    return null;
  }
  if (settings.metadata && settings.spec && settings.status) {
    return {
      id: settings.metadata.id,
      updated_at: settings.metadata.updated_at,
      updated_by: settings.metadata.updated_by || "",
      attachment_max_size_bytes: settings.spec.attachment_max_size_bytes || 0,
      attachment_allowed_content_types: settings.spec.attachment_allowed_content_types || [],
      webhook_allowed_base_urls: settings.spec.webhook_allowed_base_urls || [],
      demo_warning_enabled: Boolean(settings.spec.demo_warning_enabled),
      backup_enabled: Boolean(settings.spec.backup_enabled),
      system_health_note: settings.spec.system_health_note || "",
      attachment_policy_active: Boolean(settings.status.attachment_policy_active),
      webhook_allowlist_active: Boolean(settings.status.webhook_allowlist_active),
      demo_warning_visible: Boolean(settings.status.demo_warning_visible),
      backup_available: Boolean(settings.status.backup_available)
    };
  }
  return settings;
}

function normalizePreferences(preferences) {
  if (!preferences) {
    return null;
  }
  if (preferences.metadata && preferences.spec && preferences.status) {
    const normalized = {
      id: preferences.metadata.id || "",
      scope_type: preferences.metadata.scope_type || "",
      customized: Boolean(preferences.status.customized)
    };
    for (const key of preferenceKeys()) {
      normalized[key] = Boolean(preferences.spec[key]);
    }
    return normalized;
  }
  return preferences;
}

function normalizeNotificationDelivery(delivery) {
  if (!delivery) {
    return null;
  }
  if (delivery.metadata && delivery.spec && delivery.status) {
    return {
      id: delivery.metadata.id || "",
      domain_event_id: delivery.metadata.domain_event_id || "",
      scope_type: delivery.metadata.scope_type || "",
      project_id: delivery.metadata.project_id || "",
      policy_id: delivery.metadata.policy_id || "",
      policy_name: delivery.metadata.policy_name || "",
      destination_id: delivery.metadata.destination_id || "",
      destination_name: delivery.metadata.destination_name || "",
      destination_service: delivery.metadata.destination_service || "",
      created_at: delivery.metadata.created_at || "",
      updated_at: delivery.metadata.updated_at || "",
      event_type: delivery.spec.event_type || "",
      subject_type: delivery.spec.subject_type || "",
      subject_id: delivery.spec.subject_id || "",
      message: delivery.spec.message || "",
      payload: delivery.spec.payload || {},
      max_attempts: Number(delivery.spec.max_attempts || 0),
      state: delivery.status.state || "",
      attempt_count: Number(delivery.status.attempt_count || 0),
      next_attempt_at: delivery.status.next_attempt_at || "",
      last_attempt_at: delivery.status.last_attempt_at || "",
      delivered_at: delivery.status.delivered_at || "",
      last_error: delivery.status.last_error || ""
    };
  }
  return delivery;
}

function normalizeAuditEntry(entry) {
  if (!entry) {
    return null;
  }
  if (entry.metadata && entry.spec && entry.status) {
    return {
      id: entry.metadata.id || "",
      occurred_at: entry.metadata.occurred_at || "",
      event_type: entry.spec.event_type || "",
      actor_user_id: entry.spec.actor_user_id || "",
      auth_kind: entry.spec.auth_kind || "",
      subject_type: entry.spec.subject_type || "",
      subject_id: entry.spec.subject_id || "",
      outcome: entry.spec.outcome || "",
      payload: entry.spec.payload || {},
      security_event: Boolean(entry.status.security_event)
    };
  }
  return entry;
}

function normalizeTicketHook(hook) {
  if (!hook) {
    return null;
  }
  if (hook.metadata && hook.spec && hook.status) {
    return {
      id: hook.metadata.id || "",
      project_id: hook.metadata.project_id || "",
      created_at: hook.metadata.created_at || "",
      updated_at: hook.metadata.updated_at || "",
      name: hook.spec.name || "",
      event: hook.spec.event || "",
      phase: hook.spec.phase || "",
      enabled: Boolean(hook.spec.enabled),
      position: Number(hook.spec.position || 0),
      engine: hook.spec.engine || { type: "" },
      last_error: hook.status.last_error || ""
    };
  }
  return hook;
}

function normalizeTicketHookRun(run) {
  if (!run) {
    return null;
  }
  if (run.metadata && run.spec && run.status) {
    return {
      id: run.metadata.id || "",
      created_at: run.metadata.created_at || "",
      trigger_type: run.spec.trigger_type || "",
      trigger_ref: run.spec.trigger_ref || "",
      project_id: run.spec.project_id || "",
      ticket_id: run.spec.ticket_id || "",
      input: run.spec.input || {},
      state: run.status.state || "",
      output: run.status.output || {},
      error: run.status.error || "",
      started_at: run.status.started_at || "",
      finished_at: run.status.finished_at || ""
    };
  }
  return run;
}

function normalizeCreatePage(page) {
  if (!page) {
    return null;
  }
  if (page.metadata && page.spec && page.status) {
    return {
      id: page.metadata.id || "",
      project_id: page.metadata.project_id || "",
      owner_user_id: page.metadata.owner_user_id || page.spec.owner_user_id || "",
      created_at: page.metadata.created_at || "",
      updated_at: page.metadata.updated_at || "",
      name: page.spec.name || "",
      slug: page.spec.slug || "",
      description: page.spec.description || "",
      enabled: Boolean(page.spec.enabled),
      target_type: page.spec.target_type || "",
      target_status: page.spec.target_status || "",
      field_layout: page.spec.field_layout || [],
      defaults: page.spec.defaults || {},
      has_lua: Boolean(page.spec.form_lua_script),
      has_ai: Boolean(page.spec.form_ai_prompt || page.spec.form_ai_provider_id),
      form_lua_script: page.spec.form_lua_script || "",
      form_ai_prompt: page.spec.form_ai_prompt || "",
      form_ai_provider_id: page.spec.form_ai_provider_id || "",
      deleted_at: page.status.deleted_at || "",
      schema: null
    };
  }
  return page;
}

function normalizeCreatePageRun(run) {
  if (!run) {
    return null;
  }
  if (run.metadata && run.spec && run.status) {
    return {
      id: run.metadata.id || "",
      created_at: run.metadata.created_at || "",
      trigger_type: run.spec.trigger_type || "",
      trigger_ref: run.spec.trigger_ref || "",
      project_id: run.spec.project_id || "",
      ticket_id: run.spec.ticket_id || "",
      input: run.spec.input || {},
      state: run.status.state || "",
      output: run.status.output || {},
      error: run.status.error || "",
      started_at: run.status.started_at || "",
      finished_at: run.status.finished_at || ""
    };
  }
  return run;
}

function normalizeCreatePageSchema(schema) {
  if (!schema) {
    return null;
  }
  if (schema.metadata && schema.spec && schema.status) {
    return {
      page_id: schema.metadata.page_id || "",
      project_id: schema.metadata.project_id || "",
      slug: schema.metadata.slug || schema.spec.slug || "",
      name: schema.spec.name || "",
      description: schema.spec.description || "",
      enabled: Boolean(schema.status.enabled || schema.spec.enabled),
      target_type: schema.spec.target_type || "",
      target_status: schema.spec.target_status || "",
      field_layout: schema.spec.field_layout || [],
      defaults: schema.spec.defaults || {},
      owner_user_id: schema.spec.owner_user_id || ""
    };
  }
  return schema;
}

function normalizeCronJob(job) {
  if (!job) {
    return null;
  }
  if (job.metadata && job.spec && job.status) {
    return {
      id: job.metadata.id || "",
      created_at: job.metadata.created_at || "",
      updated_at: job.metadata.updated_at || "",
      owner_user_id: job.spec.owner_user_id || "",
      project_id: job.spec.project_id || "",
      name: job.spec.name || "",
      schedule: job.spec.schedule || "",
      timezone: job.spec.timezone || "UTC",
      enabled: Boolean(job.spec.enabled),
      engine: job.spec.engine || { type: "" },
      last_run_status: job.status.last_run_status || "",
      last_run_at: job.status.last_run_at || "",
      next_run_at: job.status.next_run_at || "",
      last_error: job.status.last_error || ""
    };
  }
  return job;
}

function normalizeCronRun(run) {
  if (!run) {
    return null;
  }
  if (run.metadata && run.spec && run.status) {
    return {
      id: run.metadata.id || "",
      created_at: run.metadata.created_at || "",
      trigger_type: run.spec.trigger_type || "",
      trigger_ref: run.spec.trigger_ref || "",
      project_id: run.spec.project_id || "",
      ticket_id: run.spec.ticket_id || "",
      input: run.spec.input || {},
      state: run.status.state || "",
      output: run.status.output || {},
      error: run.status.error || "",
      started_at: run.status.started_at || "",
      finished_at: run.status.finished_at || ""
    };
  }
  return run;
}

function normalizeWebhook(webhook) {
  if (!webhook) {
    return null;
  }
  if (webhook.metadata && webhook.spec && webhook.status) {
    return {
      id: webhook.metadata.id || "",
      project_id: webhook.metadata.project_id || "",
      created_at: webhook.metadata.created_at || "",
      updated_at: webhook.metadata.updated_at || "",
      name: webhook.spec.name || "",
      direction: webhook.spec.direction || "",
      enabled: Boolean(webhook.spec.enabled),
      actor_user_id: webhook.spec.actor_user_id || "",
      event_types: webhook.spec.event_types || [],
      engine: webhook.spec.engine || { type: "" },
      token_set: Boolean(webhook.status.token_set),
      token_rotated_at: webhook.status.token_rotated_at || "",
      token: webhook.status.token || "",
      last_error: webhook.status.last_error || ""
    };
  }
  return webhook;
}

function normalizeWebhookRun(run) {
  if (!run) {
    return null;
  }
  if (run.metadata && run.spec && run.status) {
    return {
      id: run.metadata.id || "",
      created_at: run.metadata.created_at || "",
      trigger_type: run.spec.trigger_type || "",
      trigger_ref: run.spec.trigger_ref || "",
      project_id: run.spec.project_id || "",
      ticket_id: run.spec.ticket_id || "",
      input: run.spec.input || {},
      state: run.status.state || "",
      output: run.status.output || {},
      error: run.status.error || "",
      started_at: run.status.started_at || "",
      finished_at: run.status.finished_at || ""
    };
  }
  return run;
}

function normalizeWebhookDelivery(delivery) {
  if (!delivery) {
    return null;
  }
  if (delivery.metadata && delivery.spec && delivery.status) {
    return {
      id: delivery.metadata.id || "",
      webhook_id: delivery.metadata.webhook_id || "",
      webhook_name: delivery.metadata.webhook_name || "",
      domain_event_id: delivery.metadata.domain_event_id || "",
      project_id: delivery.metadata.project_id || "",
      created_at: delivery.metadata.created_at || "",
      updated_at: delivery.metadata.updated_at || "",
      event_type: delivery.spec.event_type || "",
      subject_type: delivery.spec.subject_type || "",
      subject_id: delivery.spec.subject_id || "",
      payload: delivery.spec.payload || {},
      max_attempts: Number(delivery.spec.max_attempts || 0),
      state: delivery.status.state || "",
      attempt_count: Number(delivery.status.attempt_count || 0),
      next_attempt_at: delivery.status.next_attempt_at || "",
      last_attempt_at: delivery.status.last_attempt_at || "",
      delivered_at: delivery.status.delivered_at || "",
      last_error: delivery.status.last_error || ""
    };
  }
  return delivery;
}

function normalizeOpenRouterProvider(provider) {
  if (!provider) {
    return null;
  }
  if (provider.metadata && provider.spec && provider.status) {
    return {
      id: provider.metadata.id || "",
      created_at: provider.metadata.created_at || "",
      updated_at: provider.metadata.updated_at || "",
      name: provider.spec.name || "",
      default_model: provider.spec.default_model || "",
      allowed_models: provider.spec.allowed_models || [],
      default_timeout_seconds: Number(provider.spec.default_timeout_seconds || 0),
      max_output_tokens: Number(provider.spec.max_output_tokens || 0),
      enabled: Boolean(provider.spec.enabled),
      api_key_set: Boolean(provider.status.api_key_set),
      deleted: Boolean(provider.status.deleted)
    };
  }
  return provider;
}

function normalizeNotificationDestination(destination) {
  if (!destination) {
    return null;
  }
  if (destination.metadata && destination.spec && destination.status) {
    return {
      id: destination.metadata.id || "",
      scope_type: destination.metadata.scope_type || "global",
      project_id: destination.metadata.project_id || "",
      dashboard_id: destination.metadata.dashboard_id || "",
      created_at: destination.metadata.created_at || "",
      updated_at: destination.metadata.updated_at || "",
      name: destination.spec.name || "",
      type: destination.spec.type || "",
      enabled: Boolean(destination.spec.enabled),
      url_set: Boolean(destination.status.url_set),
      last_delivery_status: destination.status.last_delivery_status || "",
      last_delivery_at: destination.status.last_delivery_at || "",
      last_error: destination.status.last_error || "",
      deleted: Boolean(destination.status.deleted)
    };
  }
  return destination;
}

function normalizeNotificationPolicy(policy) {
  if (!policy) {
    return null;
  }
  if (policy.metadata && policy.spec && policy.status) {
    return {
      id: policy.metadata.id || "",
      scope_type: policy.metadata.scope_type || "global",
      project_id: policy.metadata.project_id || "",
      created_at: policy.metadata.created_at || "",
      updated_at: policy.metadata.updated_at || "",
      name: policy.spec.name || "",
      event_types: policy.spec.event_types || [],
      destination_ids: policy.spec.destination_ids || [],
      enabled: Boolean(policy.spec.enabled),
      deleted: Boolean(policy.status.deleted)
    };
  }
  return policy;
}

function normalizeNotificationHook(hook) {
  if (!hook) {
    return null;
  }
  if (hook.metadata && hook.spec && hook.status) {
    return {
      id: hook.metadata.id || "",
      scope_type: hook.metadata.scope_type || "global",
      project_id: hook.metadata.project_id || "",
      created_at: hook.metadata.created_at || "",
      updated_at: hook.metadata.updated_at || "",
      name: hook.spec.name || "",
      actor_user_id: hook.spec.actor_user_id || "",
      event_types: hook.spec.event_types || [],
      enabled: Boolean(hook.spec.enabled),
      engine: hook.spec.engine || { type: "" },
      last_error: hook.status.last_error || ""
    };
  }
  return hook;
}

function normalizeNotificationHookRun(run) {
  if (!run) {
    return null;
  }
  if (run.metadata && run.spec && run.status) {
    return {
      id: run.metadata.id || "",
      created_at: run.metadata.created_at || "",
      trigger_type: run.spec.trigger_type || "",
      trigger_ref: run.spec.trigger_ref || "",
      project_id: run.spec.project_id || "",
      ticket_id: run.spec.ticket_id || "",
      input: run.spec.input || {},
      state: run.status.state || "",
      output: run.status.output || {},
      error: run.status.error || "",
      started_at: run.status.started_at || "",
      finished_at: run.status.finished_at || ""
    };
  }
  return run;
}

function normalizeComment(comment) {
  if (!comment) {
    return null;
  }
  if (comment.metadata && comment.spec && comment.status) {
    return {
      id: comment.metadata.id,
      ticket_id: comment.metadata.ticket_id,
      created_at: comment.metadata.created_at,
      updated_at: comment.metadata.updated_at,
      body: comment.spec.body,
      author_id: comment.status.author_id || ""
    };
  }
  return comment;
}

function normalizeActivity(activity) {
  if (!activity) {
    return null;
  }
  if (activity.metadata && activity.spec && activity.status) {
    return {
      id: activity.metadata.id,
      ticket_id: activity.metadata.ticket_id,
      created_at: activity.metadata.created_at,
      activity_type: activity.spec.activity_type || "",
      data: activity.spec.data || {},
      actor_id: activity.status.actor_id || ""
    };
  }
  return activity;
}

function formatBytes(bytes) {
  const value = Number(bytes) || 0;
  if (value < 1024) {
    return `${value} B`;
  }
  if (value < 1024 * 1024) {
    return `${(value / 1024).toFixed(1)} KB`;
  }
  return `${(value / (1024 * 1024)).toFixed(1)} MB`;
}

function formatDateTime(value) {
  if (!value) {
    return "";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString();
}

function sprintName(id) {
  const sprint = state.sprints.find((item) => item.id === id);
  return sprint ? sprint.name : id;
}

function projectName(id) {
  const project = state.projects.find((item) => item.id === id);
  return project ? `${project.key} ${project.name}` : id;
}

function componentName(id) {
  const component = state.components.find((item) => item.id === id);
  return component ? component.name : id;
}

function versionName(id) {
  const version = state.versions.find((item) => item.id === id);
  return version ? version.name : id;
}

function roadmapEpics() {
  return state.roadmap
    .map((item) => item.epic)
    .filter(Boolean);
}

function dateRange(start, end) {
  if (start && end) {
    return `${start} to ${end}`;
  }
  return start || end || "";
}

function cookieValue(name) {
  return document.cookie
    .split(";")
    .map((part) => part.trim())
    .find((part) => part.startsWith(`${name}=`))
    ?.slice(name.length + 1) || "";
}

function mutates(method) {
  return !["GET", "HEAD", "OPTIONS"].includes(method);
}

function setNotice(message) {
  els.notice.textContent = message;
}

if (typeof module !== "undefined" && module.exports) {
  module.exports = {
    dateRange,
    dateToUTC,
    daysBetween,
    sprintReportComponents,
    sprintReportVersions,
    sprintReportEpics,
    sprintReportReporterBreakdown,
    sprintReportLabelBreakdown,
    sprintReportStatusBreakdown,
    sprintReportAgeBreakdown,
    sprintReportUpdateFreshness,
    sprintReportReadinessSummary,
    sprintReportRiskSummary,
    sprintReportDueDateBreakdown,
    sprintReportStartDateBreakdown,
    sprintReportHealth,
    sprintReportHealthDates,
    sprintReportEstimateCoverage,
    sprintReportPriorityBreakdown,
    sprintReportTypeBreakdown,
    ticketLinkDependencySummary,
    todayLocalISODate,
    versionReportAssigneeWorkloads,
    versionReportEstimateCoverage,
    versionReportPriorityBreakdown,
    versionReportTypeBreakdown,
    versionReportTimelineItems
  };
}
