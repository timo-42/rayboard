const state = {
  user: null,
  projects: [],
  selectedProject: null,
  tickets: [],
  sprints: [],
  components: [],
  versions: [],
  attachments: {},
  comments: {},
  notifications: [],
  unreadNotificationsOnly: false,
  searchResults: [],
  savedViews: [],
  lastSearchSpec: { text: "", filter: "" },
  tokens: [],
  createdToken: null,
  engineResult: null
};

const els = {
  loginForm: document.querySelector("#login-form"),
  projectForm: document.querySelector("#project-form"),
  ticketForm: document.querySelector("#ticket-form"),
  engineForm: document.querySelector("#engine-form"),
  engineType: document.querySelector("#engine-type"),
  engineProjectID: document.querySelector("#engine-project-id"),
  engineWorkbench: document.querySelector("#engine-workbench"),
  engineOutput: document.querySelector("#engine-output"),
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
  els.loginForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const data = formData(event.currentTarget);
    await runAction(async () => {
      await api("/api/login", { method: "POST", body: { spec: data } });
      event.currentTarget.reset();
      await refreshSession();
    }, "Signed in");
  });

  els.logoutButton.addEventListener("click", async () => {
    await runAction(async () => {
      await api("/api/logout", { method: "POST" });
      state.user = null;
      state.projects = [];
      state.selectedProject = null;
      state.tickets = [];
      state.sprints = [];
      state.components = [];
      state.versions = [];
      state.attachments = {};
      state.comments = {};
      state.notifications = [];
      state.unreadNotificationsOnly = false;
      state.searchResults = [];
      state.savedViews = [];
      state.lastSearchSpec = { text: "", filter: "" };
      state.tokens = [];
      state.createdToken = null;
      render();
    }, "Signed out");
  });

  els.projectForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const data = formData(event.currentTarget);
    await runAction(async () => {
      const project = normalizeProject(await api("/api/projects", { method: "POST", body: { spec: data } }));
      event.currentTarget.reset();
      await loadProjects(project.id);
    }, "Project created");
  });

  els.ticketForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    if (!state.selectedProject) {
      return;
    }
    const data = formData(event.currentTarget);
    await runAction(async () => {
      await api(`/api/projects/${state.selectedProject.id}/tickets`, { method: "POST", body: { spec: data } });
      event.currentTarget.reset();
      await loadTickets();
    }, "Ticket created");
  });

  els.engineType.addEventListener("change", () => {
    renderEngineFields();
  });

  els.engineForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    await runAction(async () => {
      const spec = engineTestSpec(event.currentTarget);
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
    const data = formData(event.currentTarget);
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
      event.currentTarget.reset();
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
    const data = formData(event.currentTarget);
    await runAction(async () => {
      await api(`/api/projects/${state.selectedProject.id}/components`, {
        method: "POST",
        body: { spec: { name: data.name || "", description: data.description || "" } }
      });
      event.currentTarget.reset();
      await loadComponents();
    }, "Component created");
  });

  els.versionForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    if (!state.selectedProject) {
      setNotice("Select a project before creating a version");
      return;
    }
    const data = formData(event.currentTarget);
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
      event.currentTarget.reset();
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
    await runAction(async () => {
      await runSearch(searchSpecFromForm(event.currentTarget));
    }, "Search complete");
  });

  els.savedViewForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    if (!state.selectedProject && event.currentTarget.elements.scope_type.value === "project") {
      setNotice("Select a project before saving a project view");
      return;
    }
    await runAction(async () => {
      const data = formData(event.currentTarget);
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
      event.currentTarget.reset();
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
    const data = formData(event.currentTarget);
    await runAction(async () => {
      const created = await api("/api/tokens", { method: "POST", body: { spec: { name: data.name || "api-token" } } });
      state.createdToken = normalizeToken(created);
      event.currentTarget.reset();
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

  els.ticketColumns.addEventListener("click", async (event) => {
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
        await loadTickets();
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
        await loadTickets();
      }, "Ticket assigned to sprint");
      return;
    }

    const removeSprint = event.target.closest("[data-remove-sprint-id]");
    if (removeSprint) {
      await runAction(async () => {
        await api(`/api/tickets/${removeSprint.dataset.removeSprintId}/sprint`, { method: "DELETE" });
        await loadTickets();
      }, "Ticket removed from sprint");
      return;
    }

    const deleteComment = event.target.closest("[data-delete-comment-id]");
    if (deleteComment) {
      await runAction(async () => {
        await api(`/api/comments/${deleteComment.dataset.deleteCommentId}`, { method: "DELETE" });
        await loadComments(deleteComment.dataset.ticketId);
      }, "Comment deleted");
      return;
    }

    const deleteAttachment = event.target.closest("[data-delete-attachment-id]");
    if (deleteAttachment) {
      await runAction(async () => {
        await api(`/api/attachments/${deleteAttachment.dataset.deleteAttachmentId}`, { method: "DELETE" });
        await loadAttachments(deleteAttachment.dataset.ticketId);
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
      await loadTickets();
    }, "Ticket updated");
  });

  els.ticketColumns.addEventListener("submit", async (event) => {
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
    }, "Attachment uploaded");
  });
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
  if (selectedID) {
    state.selectedProject = state.projects.find((project) => project.id === selectedID) || null;
  } else if (!state.selectedProject && state.projects.length > 0) {
    state.selectedProject = state.projects[0];
  } else if (state.selectedProject) {
    state.selectedProject = state.projects.find((project) => project.id === state.selectedProject.id) || null;
  }
  if (state.selectedProject) {
    await loadSprints({ renderTickets: false });
    await loadComponents({ renderTickets: false });
    await loadVersions({ renderTickets: false });
    await loadTickets();
    await loadSavedViews();
  } else {
    state.tickets = [];
    state.sprints = [];
    state.components = [];
    state.versions = [];
    state.attachments = {};
    state.comments = {};
    state.searchResults = [];
    state.savedViews = [];
  }
  render();
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
  await Promise.all(state.tickets.flatMap((ticket) => [
    loadAttachments(ticket.id, { renderAfter: false }),
    loadComments(ticket.id, { renderAfter: false })
  ]));
  render();
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
  els.loginForm.hidden = signedIn;
  els.logoutButton.hidden = !signedIn;
  els.projectCreate.hidden = !signedIn;
  els.notificationInbox.hidden = !signedIn;
  els.sprintPanel.hidden = !signedIn || !state.selectedProject;
  els.releasePanel.hidden = !signedIn || !state.selectedProject;
  els.searchPanel.hidden = !signedIn;
  els.accountPanel.hidden = !signedIn;
  els.engineWorkbench.hidden = !signedIn;
  els.signedOut.hidden = signedIn;
  els.boardView.hidden = !signedIn;
  els.ticketForm.hidden = !signedIn || !state.selectedProject;
  els.sessionState.textContent = signedIn ? state.user.username : "Signed out";

  renderProjects();
  renderTickets();
  renderNotifications();
  renderSprints();
  renderComponents();
  renderVersions();
  renderTicketFormOptions();
  renderSearchResults();
  renderSavedViews();
  renderTokens();
  renderEngineFields();
  renderEngineResult();
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

function renderTicketFormOptions() {
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
      await loadSprints({ renderTickets: false });
      await loadComponents({ renderTickets: false });
      await loadVersions({ renderTickets: false });
      await loadTickets();
      await loadSavedViews();
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

function ticketNode(ticket) {
  const article = document.createElement("article");
  article.className = "ticket";

  const key = document.createElement("p");
  key.className = "ticket-key";
  key.textContent = ticket.key;

  const title = document.createElement("h4");
  title.textContent = ticket.title;

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

  article.append(key, title, meta, planningControlNode(ticket), sprintControlNode(ticket), commentNode(ticket), attachmentNode(ticket), actions);
  return article;
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

function componentName(id) {
  const component = state.components.find((item) => item.id === id);
  return component ? component.name : id;
}

function versionName(id) {
  const version = state.versions.find((item) => item.id === id);
  return version ? version.name : id;
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
