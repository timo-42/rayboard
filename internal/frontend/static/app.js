const state = {
  user: null,
  projects: [],
  selectedProject: null,
  tickets: []
};

const els = {
  loginForm: document.querySelector("#login-form"),
  projectForm: document.querySelector("#project-form"),
  ticketForm: document.querySelector("#ticket-form"),
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
      await api("/api/login", { method: "POST", body: data });
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
      await api(`/api/projects/${state.selectedProject.id}/tickets`, { method: "POST", body: data });
      event.currentTarget.reset();
      await loadTickets();
    }, "Ticket created");
  });

  els.ticketColumns.addEventListener("click", async (event) => {
    const button = event.target.closest("[data-ticket-status]");
    if (!button) {
      return;
    }
    await runAction(async () => {
      await api(`/api/tickets/${button.dataset.ticketId}`, {
        method: "PATCH",
        body: { status: button.dataset.ticketStatus }
      });
      await loadTickets();
    }, "Ticket updated");
  });
}

async function refreshSession() {
  try {
    const data = await api("/api/me");
    state.user = data.user;
    await loadProjects();
  } catch (error) {
    state.user = null;
    render();
  }
}

async function loadProjects(selectedID = "") {
  const data = await api("/api/projects");
  state.projects = (data.items || []).map(normalizeProject);
  if (selectedID) {
    state.selectedProject = state.projects.find((project) => project.id === selectedID) || null;
  } else if (!state.selectedProject && state.projects.length > 0) {
    state.selectedProject = state.projects[0];
  } else if (state.selectedProject) {
    state.selectedProject = state.projects.find((project) => project.id === state.selectedProject.id) || null;
  }
  if (state.selectedProject) {
    await loadTickets();
  } else {
    state.tickets = [];
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
  state.tickets = data.items || [];
  render();
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
    request.body = JSON.stringify(options.body);
    request.headers["Content-Type"] = "application/json";
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
  els.signedOut.hidden = signedIn;
  els.boardView.hidden = !signedIn;
  els.ticketForm.hidden = !signedIn || !state.selectedProject;
  els.sessionState.textContent = signedIn ? state.user.username : "Signed out";

  renderProjects();
  renderTickets();
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
      await loadTickets();
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

  article.append(key, title, meta, actions);
  return article;
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
