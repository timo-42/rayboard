const state = {
  user: null,
  projects: [],
  selectedProject: null,
  tickets: [],
  attachments: {},
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
      state.attachments = {};
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

  els.ticketColumns.addEventListener("click", async (event) => {
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
  } catch (error) {
    state.user = null;
    render();
  }
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
    await loadTickets();
  } else {
    state.tickets = [];
    state.attachments = {};
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
  await Promise.all(state.tickets.map((ticket) => loadAttachments(ticket.id, { renderAfter: false })));
  render();
}

async function loadAttachments(ticketID, options = {}) {
  const data = await api(`/api/tickets/${ticketID}/attachments`);
  state.attachments[ticketID] = listItems(data).map(normalizeAttachment);
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
  els.engineWorkbench.hidden = !signedIn;
  els.signedOut.hidden = signedIn;
  els.boardView.hidden = !signedIn;
  els.ticketForm.hidden = !signedIn || !state.selectedProject;
  els.sessionState.textContent = signedIn ? state.user.username : "Signed out";

  renderProjects();
  renderTickets();
  renderEngineFields();
  renderEngineResult();
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

  article.append(key, title, meta, attachmentNode(ticket), actions);
  return article;
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
