const state = {
  user: null,
  projects: [],
  selectedProject: null,
  selectedIssue: null,
  tickets: [],
  projectSummaries: [],
  recentTickets: [],
  activeSprints: [],
  sprints: [],
  components: [],
  versions: [],
  customFields: [],
  roadmap: [],
  attachments: {},
  comments: {},
  activities: {},
  notifications: [],
  unreadNotificationsOnly: false,
  searchResults: [],
  savedViews: [],
  lastSearchSpec: { text: "", filter: "" },
  tokens: [],
  createdToken: null,
  rbac: { users: [], groups: [], roles: [], bindings: [], members: {} },
  settings: null,
  notificationPreferences: null,
  auditLog: [],
  openRouterProviders: [],
  notificationDestinations: [],
  notificationPolicies: [],
  settingsError: "",
  auditLogError: "",
  openRouterProvidersError: "",
  notificationDestinationsError: "",
  notificationPoliciesError: "",
  cronJobs: [],
  cronRuns: {},
  cronJobsError: "",
  webhooks: [],
  webhookRuns: {},
  webhookDeliveries: {},
  webhookTokens: {},
  webhooksError: "",
  ticketHooks: [],
  ticketHooksError: "",
  ticketHookPreview: null,
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
  rbacPanel: document.querySelector("#rbac-panel"),
  rbacRefresh: document.querySelector("#rbac-refresh"),
  rbacUserForm: document.querySelector("#rbac-user-form"),
  rbacGroupForm: document.querySelector("#rbac-group-form"),
  rbacMemberForm: document.querySelector("#rbac-member-form"),
  rbacBindingForm: document.querySelector("#rbac-binding-form"),
  rbacUsers: document.querySelector("#rbac-users"),
  rbacGroups: document.querySelector("#rbac-groups"),
  rbacRoles: document.querySelector("#rbac-roles"),
  rbacBindings: document.querySelector("#rbac-bindings"),
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
  notificationPolicyStatus: document.querySelector("#notification-policy-status"),
  notificationPolicies: document.querySelector("#notification-policies"),
  preferenceForm: document.querySelector("#preference-form"),
  preferenceStatus: document.querySelector("#preference-status"),
  notificationInbox: document.querySelector("#notification-inbox"),
  notificationCount: document.querySelector("#notification-count"),
  notificationUnreadOnly: document.querySelector("#notifications-unread-only"),
  notificationRefresh: document.querySelector("#notifications-refresh"),
  notificationReadAll: document.querySelector("#notifications-read-all"),
  notifications: document.querySelector("#notifications"),
  sprintPanel: document.querySelector("#sprint-panel"),
  sprintForm: document.querySelector("#sprint-form"),
  sprints: document.querySelector("#sprints"),
  releasePanel: document.querySelector("#release-panel"),
  componentForm: document.querySelector("#component-form"),
  versionForm: document.querySelector("#version-form"),
  components: document.querySelector("#components"),
  versions: document.querySelector("#versions"),
  fieldPanel: document.querySelector("#field-panel"),
  fieldForm: document.querySelector("#field-form"),
  customFields: document.querySelector("#custom-fields"),
  roadmapPanel: document.querySelector("#roadmap-panel"),
  roadmap: document.querySelector("#roadmap"),
  ticketParentID: document.querySelector("#ticket-parent-id"),
  ticketComponentID: document.querySelector("#ticket-component-id"),
  ticketVersionID: document.querySelector("#ticket-version-id"),
  searchPanel: document.querySelector("#search-panel"),
  searchForm: document.querySelector("#search-form"),
  savedViewForm: document.querySelector("#saved-view-form"),
  searchResults: document.querySelector("#search-results"),
  searchResultCount: document.querySelector("#search-result-count"),
  savedViews: document.querySelector("#saved-views"),
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
  selectedProject: document.querySelector("#selected-project"),
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
      state.tickets = [];
      state.projectSummaries = [];
      state.recentTickets = [];
      state.activeSprints = [];
      state.sprints = [];
      state.components = [];
      state.versions = [];
      state.customFields = [];
      state.roadmap = [];
      state.attachments = {};
      state.comments = {};
      state.activities = {};
      state.notifications = [];
      state.unreadNotificationsOnly = false;
      state.searchResults = [];
      state.savedViews = [];
      state.lastSearchSpec = { text: "", filter: "" };
      state.tokens = [];
      state.createdToken = null;
      state.rbac = { users: [], groups: [], roles: [], bindings: [], members: {} };
      state.settings = null;
      state.notificationPreferences = null;
      state.auditLog = [];
      state.openRouterProviders = [];
      state.notificationDestinations = [];
      state.settingsError = "";
      state.auditLogError = "";
      state.openRouterProvidersError = "";
      state.notificationDestinationsError = "";
      state.cronJobs = [];
      state.cronRuns = {};
      state.cronJobsError = "";
      state.webhooks = [];
      state.webhookRuns = {};
      state.webhookDeliveries = {};
      state.webhookTokens = {};
      state.webhooksError = "";
      state.ticketHooks = [];
      state.ticketHooksError = "";
      state.ticketHookPreview = null;
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
    data.custom_fields = parseCustomFields(data.custom_fields);
    await runAction(async () => {
      await api(`/api/projects/${state.selectedProject.id}/tickets`, { method: "POST", body: { spec: data } });
      form.reset();
      await loadRoadmap({ renderTickets: false });
      await loadTickets();
    }, "Ticket created");
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
    await Promise.all([loadCronJobs(projectID), loadWebhooks(projectID), loadTicketHooks(projectID)]);
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
        await loadCronJobs();
      }, "Cron job deleted");
    }
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
    await Promise.all([loadCronJobs(projectID), loadWebhooks(projectID), loadTicketHooks(projectID)]);
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
        await loadWebhooks();
      }, "Webhook deleted");
    }
  });

  els.ticketHookProject.addEventListener("change", async () => {
    const projectID = els.ticketHookProject.value;
    state.selectedProject = state.projects.find((project) => project.id === projectID) || state.selectedProject;
    await Promise.all([loadCronJobs(projectID), loadWebhooks(projectID), loadTicketHooks(projectID)]);
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
        await loadTicketHooks();
      }, "Ticket hook deleted");
    }
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
      await loadSprints();
    }, "Sprint created");
  });

  els.sprints.addEventListener("click", async (event) => {
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

  els.versions.addEventListener("click", async (event) => {
    const status = event.target.closest("[data-version-status]");
    if (status) {
      await runAction(async () => {
        await api(`/api/versions/${status.dataset.versionId}`, {
          method: "PATCH",
          body: { spec: { status: status.dataset.versionStatus } }
        });
        await loadVersions();
      }, "Version updated");
      return;
    }

    const remove = event.target.closest("[data-delete-version-id]");
    if (!remove) {
      return;
    }
    await runAction(async () => {
      await api(`/api/versions/${remove.dataset.deleteVersionId}`, { method: "DELETE" });
      await loadVersions();
      await loadTickets();
    }, "Version deleted");
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
      await runSearch(searchSpecFromForm(form));
    }, "Search complete");
  });

  els.savedViewForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    if (!state.selectedProject && form.elements.scope_type.value === "project") {
      setNotice("Select a project before saving a project view");
      return;
    }
    await runAction(async () => {
      const data = formData(form);
      const scopeType = data.scope_type || "user";
      const currentSearch = searchSpecFromForm(els.searchForm);
      await api("/api/saved-views", {
        method: "POST",
        body: {
          spec: {
            name: data.name || "Saved view",
            scope_type: scopeType,
            project_id: scopeType === "project" && state.selectedProject ? state.selectedProject.id : "",
            query: {
              text: currentSearch.text || "",
              filter: currentSearch.filter || ""
            },
            sort: [{ field: "updated_at", direction: "desc" }],
            columns: ["key", "title", "status", "priority"],
            display_mode: "list",
            group_by: "",
            pinned: Boolean(data.pinned)
          }
        }
      });
      form.reset();
      await loadSavedViews();
    }, "Saved view created");
  });

  els.savedViews.addEventListener("click", async (event) => {
    const apply = event.target.closest("[data-apply-saved-view-id]");
    if (apply) {
      const view = state.savedViews.find((item) => item.id === apply.dataset.applySavedViewId);
      if (!view) {
        return;
      }
      setSearchForm(view.query);
      await runAction(async () => {
        await runSearch({
          text: view.query.text || "",
          filter: view.query.filter || "",
          project_id: view.project_id || (state.selectedProject ? state.selectedProject.id : "")
        });
      }, "Saved view applied");
      return;
    }

    const remove = event.target.closest("[data-delete-saved-view-id]");
    if (remove) {
      await runAction(async () => {
        await api(`/api/saved-views/${remove.dataset.deleteSavedViewId}`, { method: "DELETE" });
        await loadSavedViews();
      }, "Saved view deleted");
    }
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

  els.rbacBindingForm.elements.subject_type.addEventListener("change", () => {
    renderBindingSubjectOptions();
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
    const data = formData(event.currentTarget);
    await runAction(async () => {
      await api("/api/me/notification-preferences", {
        method: "PATCH",
        body: {
          spec: {
            in_app_enabled: Boolean(data.in_app_enabled),
            external_enabled: Boolean(data.external_enabled),
            assignment_enabled: Boolean(data.assignment_enabled),
            comment_enabled: Boolean(data.comment_enabled),
            status_change_enabled: Boolean(data.status_change_enabled),
            sprint_change_enabled: Boolean(data.sprint_change_enabled),
            release_change_enabled: Boolean(data.release_change_enabled),
            automation_failure_enabled: Boolean(data.automation_failure_enabled)
          }
        }
      });
      await loadNotificationPreferences();
    }, "Notification preferences saved");
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
      await loadNotificationPolicies();
    }, "Notification policies refreshed");
  });

  els.notificationPolicyProject.addEventListener("change", async () => {
    const project = state.projects.find((item) => item.id === els.notificationPolicyProject.value);
    if (project) {
      state.selectedProject = project;
    }
    await runAction(async () => {
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
        await refreshTicketViews(updateLabels.dataset.updateLabelsId);
      }, "Ticket labels updated");
      return;
    }

    const updateCustomFields = event.target.closest("[data-update-custom-fields-id]");
    if (updateCustomFields) {
      const control = updateCustomFields.closest("[data-ticket-custom-field-control]");
      const input = control ? control.querySelector("textarea[name='custom_fields']") : null;
      await runAction(async () => {
        await api(`/api/tickets/${updateCustomFields.dataset.updateCustomFieldsId}`, {
          method: "PATCH",
          body: { spec: { custom_fields: parseCustomFields(input ? input.value : "") } }
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
      }, "Ticket planning fields updated");
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
      }, "Ticket assigned to sprint");
      return;
    }

    const removeSprint = event.target.closest("[data-remove-sprint-id]");
    if (removeSprint) {
      await runAction(async () => {
        await api(`/api/tickets/${removeSprint.dataset.removeSprintId}/sprint`, { method: "DELETE" });
        await refreshTicketViews(removeSprint.dataset.removeSprintId, { roadmap: false });
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
    }, "Ticket updated");
  });

  document.addEventListener("submit", async (event) => {
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
  if (route.page === "automation") {
    if (!state.selectedProject && state.projects.length) {
      state.selectedProject = state.projects[0];
    }
    await Promise.all([loadCronJobs(), loadWebhooks(), loadTicketHooks()]);
  }
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
    renderNotifications();
    return;
  }
  const query = state.unreadNotificationsOnly ? "?unread=true&limit=20" : "?limit=20";
  const data = await api(`/api/notifications${query}`);
  state.notifications = listItems(data).map(normalizeNotification);
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
    renderSavedViews();
    return;
  }
  const projectPart = state.selectedProject ? `project_id=${encodeURIComponent(state.selectedProject.id)}&` : "";
  const data = await api(`/api/saved-views?${projectPart}limit=20`);
  state.savedViews = listItems(data).map(normalizeSavedView);
  renderSavedViews();
}

async function loadSprints(options = {}) {
  if (!state.user || !state.selectedProject) {
    state.sprints = [];
    renderSprints();
    if (options.renderTickets !== false) {
      renderTickets();
    }
    return;
  }
  const data = await api(`/api/projects/${state.selectedProject.id}/sprints`);
  state.sprints = listItems(data).map(normalizeSprint);
  renderSprints();
  if (options.renderTickets !== false) {
    renderTickets();
  }
}

async function loadComponents(options = {}) {
  if (!state.user || !state.selectedProject) {
    state.components = [];
    renderComponents();
    renderTicketFormOptions();
    if (options.renderTickets !== false) {
      renderTickets();
    }
    return;
  }
  const data = await api(`/api/projects/${state.selectedProject.id}/components`);
  state.components = listItems(data).map(normalizeComponent);
  renderComponents();
  renderTicketFormOptions();
  if (options.renderTickets !== false) {
    renderTickets();
  }
}

async function loadVersions(options = {}) {
  if (!state.user || !state.selectedProject) {
    state.versions = [];
    renderVersions();
    renderTicketFormOptions();
    if (options.renderTickets !== false) {
      renderTickets();
    }
    return;
  }
  const data = await api(`/api/projects/${state.selectedProject.id}/versions`);
  state.versions = listItems(data).map(normalizeVersion);
  renderVersions();
  renderTicketFormOptions();
  if (options.renderTickets !== false) {
    renderTickets();
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
  if (options.renderTickets !== false) {
    renderTickets();
  }
}

async function loadRoadmap(options = {}) {
  if (!state.user || !state.selectedProject) {
    state.roadmap = [];
    renderRoadmap();
    renderTicketFormOptions();
    if (options.renderTickets !== false) {
      renderTickets();
    }
    return;
  }
  const data = await api(`/api/projects/${state.selectedProject.id}/roadmap`);
  state.roadmap = listItems(data).map(normalizeRoadmapItem);
  renderRoadmap();
  renderTicketFormOptions();
  if (options.renderTickets !== false) {
    renderTickets();
  }
}

async function loadProjectDetails() {
  if (!state.selectedProject) {
    return;
  }
  await loadSprints({ renderTickets: false });
  await loadComponents({ renderTickets: false });
  await loadVersions({ renderTickets: false });
  await loadCustomFields({ renderTickets: false });
  await loadRoadmap({ renderTickets: false });
  await loadTickets();
  await loadSavedViews();
}

async function runSearch(spec) {
  const normalized = {
    project_id: spec.project_id || (state.selectedProject ? state.selectedProject.id : ""),
    text: spec.text || "",
    filter: spec.filter || "",
    sort: [{ field: "updated_at", direction: "desc" }],
    limit: 20
  };
  const data = await api("/api/search", { method: "POST", body: { spec: normalized } });
  state.lastSearchSpec = { text: normalized.text, filter: normalized.filter };
  state.searchResults = listItems(data).map(normalizeTicket);
  renderSearchResults();
}

async function loadProjects(selectedID = "") {
  const data = await api("/api/projects");
  state.projects = listItems(data).map(normalizeProject);
  const route = currentRoute();
  if (route.projectID) {
    state.selectedProject = state.projects.find((project) => project.id === route.projectID) || null;
  } else if (selectedID) {
    state.selectedProject = state.projects.find((project) => project.id === selectedID) || null;
  } else if (!state.selectedProject && state.projects.length > 0) {
    state.selectedProject = state.projects[0];
  } else if (state.selectedProject) {
    state.selectedProject = state.projects.find((project) => project.id === state.selectedProject.id) || null;
  }
  await loadDashboardSummaries();
  if (state.selectedProject && route.page === "projects") {
    await loadProjectDetails();
  } else {
    state.tickets = [];
    state.sprints = [];
    state.components = [];
    state.versions = [];
    state.customFields = [];
    state.roadmap = [];
    state.attachments = {};
    state.comments = {};
    state.activities = {};
    state.searchResults = [];
    state.savedViews = [];
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
  const data = await api(`/api/projects/${state.selectedProject.id}/tickets`);
  state.tickets = listItems(data).map(normalizeTicket);
  state.attachments = {};
  state.comments = {};
  state.activities = {};
  await Promise.all(state.tickets.flatMap((ticket) => [
    loadAttachments(ticket.id, { renderAfter: false }),
    loadComments(ticket.id, { renderAfter: false })
  ]));
  render();
}

async function loadSelectedIssue(ticketID) {
  const ticket = normalizeTicket(await api(`/api/tickets/${ticketID}`));
  state.selectedIssue = ticket;
  state.selectedProject = state.projects.find((project) => project.id === ticket.project_id) || state.selectedProject;
  if (state.selectedProject) {
    await Promise.all([
      loadSprints({ renderTickets: false }),
      loadComponents({ renderTickets: false }),
      loadVersions({ renderTickets: false }),
      loadCustomFields({ renderTickets: false })
    ]);
  }
  await Promise.all([
    loadAttachments(ticket.id, { renderAfter: false }),
    loadComments(ticket.id, { renderAfter: false }),
    loadActivity(ticket.id, { renderAfter: false })
  ]);
}

async function refreshTicketViews(ticketID, options = {}) {
  if (options.roadmap !== false) {
    await loadRoadmap({ renderTickets: false });
  }
  await loadTickets();
  if (state.selectedIssue && state.selectedIssue.id === ticketID) {
    await loadSelectedIssue(ticketID);
    renderIssue();
  }
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
    members: {}
  };
  const memberEntries = await Promise.all(state.rbac.groups.map(async (group) => {
    const members = await api(`/api/groups/${group.id}/members`).catch(() => null);
    return [group.id, listItems(members).map(normalizeUser)];
  }));
  state.rbac.members = Object.fromEntries(memberEntries);
  renderRBAC();
}

async function loadSettingsPage() {
  await Promise.all([
    loadGlobalSettings(),
    loadNotificationPreferences(),
    loadAuditLog(),
    loadOpenRouterProviders(),
    loadNotificationDestinations(),
    loadNotificationPolicies()
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
  state.notificationPreferences = normalizePreferences(await api("/api/me/notification-preferences"));
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
  els.ticketForm.hidden = !signedIn || route.page !== "projects" || !state.selectedProject;
  els.sessionState.textContent = signedIn ? state.user.username : "Signed out";

  renderNavigation(route);
  renderDashboard();
  renderProjects();
  renderTickets();
  renderIssue();
  renderNotifications();
  renderSprints();
  renderComponents();
  renderVersions();
  renderCustomFields();
  renderRoadmap();
  renderTicketFormOptions();
  renderSearchResults();
  renderSavedViews();
  renderTokens();
  renderRBAC();
  renderSettings();
  renderCronJobs();
  renderWebhooks();
  renderTicketHooks();
  renderEngineFields();
  renderEngineResult();
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
  els.metricUnread.textContent = String(state.notifications.filter((notification) => !notification.read_at).length);

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
  const unreadCount = state.notifications.filter((notification) => !notification.read_at).length;
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

function notificationNode(notification) {
  const article = document.createElement("article");
  article.className = "notification-item";
  if (!notification.read_at) {
    article.classList.add("is-unread");
  }

  const body = document.createElement("p");
  body.textContent = notification.body || notification.type || "Notification";

  const meta = document.createElement("span");
  meta.textContent = [notification.type, notification.subject_type, notification.subject_id].filter(Boolean).join(" / ");

  const button = document.createElement("button");
  button.type = "button";
  button.dataset.notificationId = notification.id;
  button.dataset.notificationReadState = notification.read_at ? "unread" : "read";
  button.textContent = notification.read_at ? "Unread" : "Read";

  article.append(body, meta, button);
  return article;
}

function renderSprints() {
  if (!els.sprints) {
    return;
  }
  els.sprints.replaceChildren();
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

  const actions = document.createElement("div");
  actions.className = "sprint-actions";

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

  article.append(body, actions);
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

  const remove = document.createElement("button");
  remove.type = "button";
  remove.dataset.deleteComponentId = component.id;
  remove.textContent = "Delete";

  article.append(body, remove);
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

  const actions = document.createElement("div");
  actions.className = "version-actions";

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

  article.append(body, actions);
  return article;
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
  for (const field of state.customFields) {
    els.customFields.append(customFieldNode(field));
  }
}

function customFieldNode(field) {
  const article = document.createElement("article");
  article.className = "field-item";
  article.dataset.fieldType = field.field_type || "text";

  const body = document.createElement("div");
  body.className = "field-item-body";

  const name = document.createElement("p");
  name.textContent = field.name || field.key;

  const meta = document.createElement("span");
  meta.textContent = [
    field.key,
    field.field_type,
    field.required ? "required" : "optional",
    field.options.length ? `options: ${field.options.join(", ")}` : ""
  ].filter(Boolean).join(" / ");

  body.append(name, meta);

  const remove = document.createElement("button");
  remove.type = "button";
  remove.dataset.deleteFieldId = field.id;
  remove.textContent = "Delete";

  article.append(body, remove);
  return article;
}

function renderTicketFormOptions() {
  replaceSelectOptions(els.ticketParentID, "Parent epic", roadmapEpics(), (epic) => `${epic.key} ${epic.title}`);
  replaceSelectOptions(els.ticketComponentID, "Component", state.components, (component) => component.name);
  replaceSelectOptions(els.ticketVersionID, "Version", state.versions, (version) => `${version.name} (${version.state})`);
}

function replaceSelectOptions(select, emptyLabel, items, label) {
  if (!select) {
    return;
  }
  const current = select.value;
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
  for (const item of state.roadmap) {
    els.roadmap.append(roadmapNode(item));
  }
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
  return article;
}

function renderSearchResults() {
  if (!els.searchResults) {
    return;
  }
  els.searchResultCount.textContent = String(state.searchResults.length);
  els.searchResults.replaceChildren();
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
    meta.textContent = [ticket.status, ticket.type, ticket.priority].filter(Boolean).join(" / ") || "Ticket";

    row.append(title, meta);
    els.searchResults.append(row);
  }
}

function renderSavedViews() {
  if (!els.savedViews) {
    return;
  }
  els.savedViews.replaceChildren();
  if (!state.savedViews.length) {
    const empty = document.createElement("p");
    empty.className = "muted";
    empty.textContent = "No saved views";
    els.savedViews.append(empty);
    return;
  }
  for (const view of state.savedViews) {
    els.savedViews.append(savedViewNode(view));
  }
}

function savedViewNode(view) {
  const article = document.createElement("article");
  article.className = "saved-view-item";
  if (view.pinned) {
    article.classList.add("is-pinned");
  }

  const name = document.createElement("p");
  name.textContent = view.name;

  const meta = document.createElement("span");
  meta.textContent = [view.scope_type, view.display_mode, view.pinned ? "pinned" : ""].filter(Boolean).join(" / ");

  const actions = document.createElement("div");
  actions.className = "saved-view-actions";

  const apply = document.createElement("button");
  apply.type = "button";
  apply.dataset.applySavedViewId = view.id;
  apply.textContent = "Apply";

  const remove = document.createElement("button");
  remove.type = "button";
  remove.dataset.deleteSavedViewId = view.id;
  remove.textContent = "Delete";

  actions.append(apply, remove);
  article.append(name, meta, actions);
  return article;
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
  if (!els.rbacUsers) {
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
}

function renderRBACFormOptions() {
  replaceSelectOptions(els.rbacMemberForm.elements.group_id, "Group", state.rbac.groups, (group) => group.display_name || group.name || group.id);
  replaceSelectOptions(els.rbacMemberForm.elements.user_id, "User", state.rbac.users, (user) => user.username || user.id);
  replaceSelectOptions(els.rbacBindingForm.elements.role_name, "Role", state.rbac.roles.map((role) => ({ id: role.name, name: role.name })), (role) => role.name);
  renderBindingSubjectOptions();
  replaceSelectOptions(els.rbacBindingForm.elements.project_id, "Global scope", state.projects, (project) => `${project.key} ${project.name}`);
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

function renderSettings() {
  if (!els.settingsForm || !els.preferenceForm || !els.auditForm || !els.openRouterProviderForm || !els.notificationDestinationForm || !els.notificationPolicyForm) {
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

  renderAuditLog();
  renderOpenRouterProviders();
  renderNotificationDestinations();
  renderNotificationPolicies();
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
  const destinations = document.createElement("textarea");
  destinations.name = "destination_ids";
  destinations.rows = 2;
  destinations.placeholder = "destination ids";
  destinations.value = policy.destination_ids.join(", ");
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

function projectLabel(projectID) {
  const project = state.projects.find((item) => item.id === projectID);
  return project ? `${project.key} ${project.name}` : projectID;
}

function destinationLabel(destinationID) {
  const destination = state.notificationDestinations.find((item) => item.id === destinationID);
  return destination ? `${destination.name || destination.id} (${destination.id})` : destinationID;
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

  const jobRuns = state.cronRuns[job.id] || [];
  if (jobRuns.length) {
    article.append(cronRunListNode(jobRuns));
  }
  return article;
}

function cronRunListNode(runs) {
  const list = document.createElement("div");
  list.className = "cron-run-list";
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

  const runsList = state.webhookRuns[webhook.id] || [];
  if (runsList.length) {
    article.append(webhookRunListNode(runsList));
  }
  const deliveriesList = state.webhookDeliveries[webhook.id] || [];
  if (deliveriesList.length) {
    article.append(webhookDeliveryListNode(webhook.id, deliveriesList));
  }
  return article;
}

function webhookRunListNode(runs) {
  const list = document.createElement("div");
  list.className = "webhook-run-list";
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

  const toggle = document.createElement("button");
  toggle.type = "button";
  toggle.dataset.toggleTicketHookId = hook.id;
  toggle.dataset.ticketHookEnabled = hook.enabled ? "false" : "true";
  toggle.textContent = hook.enabled ? "Disable" : "Enable";

  const remove = document.createElement("button");
  remove.type = "button";
  remove.dataset.deleteTicketHookId = hook.id;
  remove.textContent = "Delete";

  actions.append(preview, toggle, remove);
  article.append(header, meta, actions);
  return article;
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
  if (!els.engineOutput) {
    return;
  }
  els.engineOutput.textContent = JSON.stringify(state.engineResult || {}, null, 2);
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

function notificationPolicySpec(form) {
  const data = formData(form);
  return {
    name: data.name || "",
    event_types: parseCommaList(data.event_types),
    destination_ids: parseCommaList(data.destination_ids),
    enabled: Boolean(data.enabled)
  };
}

function notificationPolicyUpdateSpec(form) {
  return notificationPolicySpec(form);
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

function renderTickets() {
  els.selectedProject.textContent = state.selectedProject ? `${state.selectedProject.key} ${state.selectedProject.name}` : "No project selected";
  const lists = els.ticketColumns.querySelectorAll(".ticket-list");
  for (const list of lists) {
    list.replaceChildren();
  }
  for (const ticket of state.tickets) {
    const list = els.ticketColumns.querySelector(`[data-status="${ticket.status}"] .ticket-list`) ||
      els.ticketColumns.querySelector('[data-status="todo"] .ticket-list');
    list.append(ticketNode(ticket));
  }
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
    ["Project", projectName(ticket.project_id)],
    ["Sprint", ticket.sprint_id ? sprintName(ticket.sprint_id) : "None"],
    ["Component", ticket.component_id ? componentName(ticket.component_id) : "None"],
    ["Version", ticket.version_id ? versionName(ticket.version_id) : "None"],
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
    labelControlNode(ticket),
    customFieldControlNode(ticket),
    planningControlNode(ticket),
    sprintControlNode(ticket),
    commentNode(ticket),
    attachmentNode(ticket),
    activityNode(ticket)
  );
}

function ticketNode(ticket) {
  const article = document.createElement("article");
  article.className = "ticket";

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
  meta.textContent = [ticket.type, ticket.priority].filter(Boolean).join(" / ") || "Unclassified";

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

  article.append(key, title, meta, labelControlNode(ticket), customFieldControlNode(ticket), planningControlNode(ticket), sprintControlNode(ticket), commentNode(ticket), attachmentNode(ticket), actions);
  return article;
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

  const input = document.createElement("textarea");
  input.name = "custom_fields";
  input.rows = 3;
  input.value = formatCustomFields(ticket.custom_fields);
  input.placeholder = customFieldPlaceholder();
  input.setAttribute("aria-label", "Custom fields JSON");

  const update = document.createElement("button");
  update.type = "button";
  update.dataset.updateCustomFieldsId = ticket.id;
  update.textContent = "Fields";

  controls.append(input, update);
  section.append(heading, summary, controls);
  return section;
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
  textarea.placeholder = "Add a comment";
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

function formData(form) {
  return Object.fromEntries(new FormData(form).entries());
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

function setSearchForm(query) {
  if (!els.searchForm) {
    return;
  }
  els.searchForm.elements.text.value = query.text || "";
  els.searchForm.elements.filter.value = query.filter || "";
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
      labels: ticket.spec.labels || [],
      custom_fields: ticket.spec.custom_fields || {}
    };
  }
  return ticket;
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
