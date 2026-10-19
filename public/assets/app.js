// Pide confirmación antes de enviar formularios marcados con data-confirm.
document.addEventListener("submit", function (event) {
  var form = event.target;
  if (form.dataset.confirm && !window.confirm(form.dataset.confirm)) {
    event.preventDefault();
  }

  if (!form.matches("[data-task-toggle]")) {
    return;
  }

  event.preventDefault();

  var task = form.closest(".task");
  var completed = task.classList.toggle("task-completed");
  var button = form.querySelector("button");
  var status = task.querySelector(".task-status");

  var label = completed ? "Reabrir" : "Completar";
  button.querySelector(".toggle-label").textContent = label;
  button.setAttribute("aria-pressed", String(completed));
  button.setAttribute("aria-label", label + " tarea: " + task.querySelector(".task-title").textContent);
  button.setAttribute("title", label + " tarea");

  if (completed && !status) {
    status = document.createElement("span");
    status.className = "task-status";
    status.textContent = "Completada";
    task.insertBefore(status, task.querySelector(".task-actions"));
  } else if (!completed && status) {
    status.remove();
  }

  // Envía el formulario tal cual (POST con _method=PUT y authenticity_token);
  // si falla, recarga para mostrar el estado real.
  fetch(form.action, {
    method: form.method.toUpperCase(),
    body: new URLSearchParams(new FormData(form)),
    credentials: "same-origin",
  })
    .then(function (response) {
      if (!response.ok) {
        throw new Error("toggle failed: " + response.status);
      }
    })
    .catch(function () {
      window.location.reload();
    });
});

// Reordena la lista arrastrando el asa de cada tarea o con las flechas del teclado.
(function () {
  var list = document.querySelector("[data-reorder-url]");
  if (!list) {
    return;
  }

  var dragged = null;
  var initialOrder = "";

  function tasks() {
    return Array.prototype.slice.call(list.querySelectorAll(":scope > .task"));
  }

  function currentOrder() {
    return tasks().map(function (task) {
      return task.dataset.taskId;
    });
  }

  function saveOrder() {
    fetch(list.dataset.reorderUrl, {
      method: "POST",
      headers: { "Content-Type": "application/json", "X-CSRF-Token": list.dataset.csrfToken },
      body: JSON.stringify({ ids: currentOrder() }),
      credentials: "same-origin"
    })
      .then(function (res) {
        if (!res.ok) {
          window.location.reload();
        }
      })
      .catch(function () {
        window.location.reload();
      });
  }

  list.addEventListener("pointerdown", function (event) {
    var handle = event.target.closest("[data-drag-handle]");
    if (!handle || event.button !== 0) {
      return;
    }

    event.preventDefault();
    dragged = handle.closest(".task");
    initialOrder = currentOrder().join(",");
    handle.setPointerCapture(event.pointerId);
    dragged.classList.add("task-dragging");
    list.classList.add("is-reordering");
  });

  list.addEventListener("pointermove", function (event) {
    if (!dragged) {
      return;
    }

    var target = null;
    var siblings = tasks();
    for (var i = 0; i < siblings.length; i++) {
      if (siblings[i] === dragged) {
        continue;
      }
      var rect = siblings[i].getBoundingClientRect();
      if (event.clientY < rect.top + rect.height / 2) {
        target = siblings[i];
        break;
      }
    }

    if (target) {
      if (dragged.nextElementSibling !== target) {
        list.insertBefore(dragged, target);
      }
    } else if (list.lastElementChild !== dragged) {
      list.appendChild(dragged);
    }
  });

  function endDrag() {
    if (!dragged) {
      return;
    }

    dragged.classList.remove("task-dragging");
    list.classList.remove("is-reordering");
    dragged = null;

    if (currentOrder().join(",") !== initialOrder) {
      saveOrder();
    }
  }

  list.addEventListener("pointerup", endDrag);
  list.addEventListener("pointercancel", endDrag);

  list.addEventListener("keydown", function (event) {
    var handle = event.target.closest("[data-drag-handle]");
    if (!handle || (event.key !== "ArrowUp" && event.key !== "ArrowDown")) {
      return;
    }

    var task = handle.closest(".task");
    if (event.key === "ArrowUp") {
      if (!task.previousElementSibling) {
        return;
      }
      list.insertBefore(task, task.previousElementSibling);
    } else {
      if (!task.nextElementSibling) {
        return;
      }
      list.insertBefore(task.nextElementSibling, task);
    }

    event.preventDefault();
    handle.focus();

    var all = tasks();
    var status = document.querySelector("[data-reorder-status]");
    if (status) {
      status.textContent = task.querySelector(".task-title").textContent + " movida a la posición " +
        (all.indexOf(task) + 1) + " de " + all.length;
    }

    saveOrder();
  });
})();
