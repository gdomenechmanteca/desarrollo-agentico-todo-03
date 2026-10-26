package actions

import (
	"net/url"
	"strings"

	"todo/models"

	"github.com/gofrs/uuid"
)

func (as *ActionSuite) Test_TasksIndex_EmptyState() {
	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "No hay tareas todavía")
}

func (as *ActionSuite) Test_TasksIndex_ListsTasks() {
	as.LoadFixture("tres tareas")

	res := as.HTML("/").Get()
	as.Equal(200, res.Code)

	body := res.Body.String()
	as.Contains(body, "Comprar material de oficina")
	as.Contains(body, "Llamar al gestor")
	as.Contains(body, "Publicar la oferta de empleo")
	as.Contains(body, "Completada")
}

func (as *ActionSuite) Test_TasksCreate_Valid() {
	res := as.HTML("/tasks").Post(&models.Task{Title: "  Nueva tarea de prueba  "})
	as.Equal(303, res.Code)
	as.Equal("/", res.Header().Get("Location"))

	task := &models.Task{}
	as.NoError(as.DB.First(task))
	as.Equal("Nueva tarea de prueba", task.Title)
	as.False(task.Completed)
}

func (as *ActionSuite) Test_TasksCreate_EmptyTitle_Invalid() {
	res := as.HTML("/tasks").Post(&models.Task{Title: ""})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksCreate_OnlySpaces_Invalid() {
	res := as.HTML("/tasks").Post(&models.Task{Title: "     "})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksToggle_CompletesAndReopens() {
	task := &models.Task{Title: "Pendiente de completar"}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/tasks/%s/toggle", task.ID).Put(url.Values{})
	as.Equal(303, res.Code)

	as.NoError(as.DB.Reload(task))
	as.True(task.Completed)

	res = as.HTML("/tasks/%s/toggle", task.ID).Put(url.Values{})
	as.Equal(303, res.Code)

	as.NoError(as.DB.Reload(task))
	as.False(task.Completed)
}

func (as *ActionSuite) Test_TasksToggle_NotFound() {
	id := uuid.Must(uuid.NewV4())
	res := as.HTML("/tasks/%s/toggle", id).Put(url.Values{})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksToggle_InvalidID() {
	res := as.HTML("/tasks/no-es-un-uuid/toggle").Put(url.Values{})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksDestroy() {
	task := &models.Task{Title: "Tarea que se elimina"}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/tasks/%s", task.ID).Delete()
	as.Equal(303, res.Code)

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksDestroy_NotFound() {
	id := uuid.Must(uuid.NewV4())
	res := as.HTML("/tasks/%s", id).Delete()
	as.Equal(404, res.Code)
}

func (as *ActionSuite) taskForEditing(completed bool) *models.Task {
	task := &models.Task{Title: "Título original", Completed: completed}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())
	as.NoError(as.DB.Reload(task))
	return task
}

func (as *ActionSuite) Test_TasksIndex_EditLink() {
	task := as.taskForEditing(false)
	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "/tasks/"+task.ID.String()+"/edit")
	as.Contains(res.Body.String(), "Editar tarea: Título original")
}

func (as *ActionSuite) Test_TasksEdit_PrefillsTitle() {
	task := as.taskForEditing(true)
	res := as.HTML("/tasks/%s/edit", task.ID).Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), `value="Título original"`)
	as.Contains(res.Body.String(), `value="PUT"`)
	as.Contains(res.Body.String(), "Guardar cambios")
	as.Contains(res.Body.String(), "Cancelar")
	as.NoError(as.DB.Reload(task))
	as.True(task.Completed)
}

func (as *ActionSuite) Test_TasksEdit_NotFound() {
	res := as.HTML("/tasks/%s/edit", uuid.Must(uuid.NewV4())).Get()
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksEdit_InvalidID() {
	res := as.HTML("/tasks/no-es-un-uuid/edit").Get()
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksEdit_EscapesTitle() {
	task := &models.Task{Title: `<img src=x onerror="alert(1)">`}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())
	res := as.HTML("/tasks/%s/edit", task.ID).Get()
	as.Equal(200, res.Code)
	as.NotContains(res.Body.String(), `<img src=x`)
	as.Contains(res.Body.String(), "&lt;img")
}

func (as *ActionSuite) Test_TasksUpdate_PersistsTitleOnly() {
	task := as.taskForEditing(true)
	id, createdAt, updatedAt := task.ID, task.CreatedAt, task.UpdatedAt
	res := as.HTML("/tasks/%s", id).Put(url.Values{
		"title": {"  Título actualizado  "}, "completed": {"false"},
		"id": {uuid.Must(uuid.NewV4()).String()}, "created_at": {"2000-01-01T00:00:00Z"},
	})
	as.Equal(303, res.Code)
	as.Equal("/", res.Header().Get("Location"))
	as.NoError(as.DB.Reload(task))
	as.Equal("Título actualizado", task.Title)
	as.Equal(id, task.ID)
	as.True(task.Completed)
	as.True(task.CreatedAt.Equal(createdAt))
	as.False(task.UpdatedAt.Before(updatedAt))
	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(1, count)
}

func (as *ActionSuite) Test_TasksUpdate_PendingTaskRemainsPending() {
	task := as.taskForEditing(false)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {"Nuevo título"}, "completed": {"true"}})
	as.Equal(303, res.Code)
	as.NoError(as.DB.Reload(task))
	as.Equal("Nuevo título", task.Title)
	as.False(task.Completed)
}

func (as *ActionSuite) Test_TasksUpdate_EmptyTitle() {
	task := as.taskForEditing(false)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {""}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")
	as.Contains(res.Body.String(), `aria-describedby="title-error"`)
	as.NoError(as.DB.Reload(task))
	as.Equal("Título original", task.Title)
}

func (as *ActionSuite) Test_TasksUpdate_OnlySpaces() {
	task := as.taskForEditing(true)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {"   "}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")
	as.NoError(as.DB.Reload(task))
	as.Equal("Título original", task.Title)
	as.True(task.Completed)
}

func (as *ActionSuite) Test_TasksUpdate_TitleTooLong() {
	task := as.taskForEditing(false)
	title := strings.Repeat("á", models.TitleMaxLength+1)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {title}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título no puede superar los 200 caracteres.")
	as.Contains(res.Body.String(), title)
	as.NoError(as.DB.Reload(task))
	as.Equal("Título original", task.Title)
}

func (as *ActionSuite) Test_TasksUpdate_TitleAtLimit() {
	task := as.taskForEditing(false)
	title := strings.Repeat("á", models.TitleMaxLength)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {title}})
	as.Equal(303, res.Code)
	as.NoError(as.DB.Reload(task))
	as.Equal(title, task.Title)
}

func (as *ActionSuite) Test_TasksUpdate_NotFound() {
	res := as.HTML("/tasks/%s", uuid.Must(uuid.NewV4())).Put(url.Values{"title": {"Título nuevo"}})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksUpdate_InvalidID() {
	res := as.HTML("/tasks/no-es-un-uuid").Put(url.Values{"title": {"Título nuevo"}})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) createTaskAt(title string, position int) *models.Task {
	task := &models.Task{Title: title, Position: position}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())
	return task
}

func (as *ActionSuite) positionOf(task *models.Task) int {
	reloaded := &models.Task{}
	as.NoError(as.DB.Find(reloaded, task.ID))
	return reloaded.Position
}

func (as *ActionSuite) Test_TasksIndex_OrdersByPosition() {
	as.createTaskAt("Creada primero", 3)
	as.createTaskAt("Creada segundo", 2)
	as.createTaskAt("Creada tercero", 1)

	res := as.HTML("/").Get()
	as.Equal(200, res.Code)

	body := res.Body.String()
	first := strings.Index(body, "Creada tercero")
	second := strings.Index(body, "Creada segundo")
	third := strings.Index(body, "Creada primero")
	as.True(first >= 0 && first < second && second < third)
}

func (as *ActionSuite) Test_TasksIndex_RendersDragHandle() {
	task := as.createTaskAt("Tarea movible", 1)

	res := as.HTML("/").Get()
	as.Equal(200, res.Code)

	body := res.Body.String()
	as.Contains(body, `data-reorder-url="/tasks/reorder"`)
	as.Contains(body, `data-task-id="`+task.ID.String()+`"`)
	as.Contains(body, "Mover tarea: Tarea movible")
}

func (as *ActionSuite) Test_TasksCreate_AppendsAtEnd() {
	as.createTaskAt("Existente", 7)

	res := as.HTML("/tasks").Post(&models.Task{Title: "Nueva al final"})
	as.Equal(303, res.Code)

	task := &models.Task{}
	as.NoError(as.DB.Where("title = ?", "Nueva al final").First(task))
	as.Equal(8, task.Position)
}

func (as *ActionSuite) Test_TasksReorder_Persists() {
	a := as.createTaskAt("Tarea A", 1)
	b := as.createTaskAt("Tarea B", 2)
	c := as.createTaskAt("Tarea C", 3)

	res := as.JSON("/tasks/reorder").Post(map[string]interface{}{
		"ids": []string{c.ID.String(), a.ID.String(), b.ID.String()},
	})
	as.Equal(204, res.Code)

	as.Equal(1, as.positionOf(c))
	as.Equal(2, as.positionOf(a))
	as.Equal(3, as.positionOf(b))

	body := as.HTML("/").Get().Body.String()
	as.True(strings.Index(body, "Tarea C") < strings.Index(body, "Tarea A"))
	as.True(strings.Index(body, "Tarea A") < strings.Index(body, "Tarea B"))
}

func (as *ActionSuite) Test_TasksReorder_IncompleteList() {
	a := as.createTaskAt("Tarea A", 1)
	b := as.createTaskAt("Tarea B", 2)

	res := as.JSON("/tasks/reorder").Post(map[string]interface{}{
		"ids": []string{b.ID.String()},
	})
	as.Equal(422, res.Code)

	as.Equal(1, as.positionOf(a))
	as.Equal(2, as.positionOf(b))
}

func (as *ActionSuite) Test_TasksReorder_UnknownID() {
	a := as.createTaskAt("Tarea A", 1)

	res := as.JSON("/tasks/reorder").Post(map[string]interface{}{
		"ids": []string{uuid.Must(uuid.NewV4()).String()},
	})
	as.Equal(422, res.Code)
	as.Equal(1, as.positionOf(a))
}

func (as *ActionSuite) Test_TasksReorder_InvalidUUID() {
	as.createTaskAt("Tarea A", 1)

	res := as.JSON("/tasks/reorder").Post(map[string]interface{}{
		"ids": []string{"no-es-un-uuid"},
	})
	as.Equal(400, res.Code)
}

func (as *ActionSuite) taskWithDueDate(title, dueDate string) *models.Task {
	d, err := models.ParseDueDate(dueDate)
	as.NoError(err)
	task := &models.Task{Title: title, DueDate: d}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())
	return task
}

func (as *ActionSuite) Test_TasksCreate_WithDueDate() {
	res := as.HTML("/tasks").Post(url.Values{"title": {"Con fecha"}, "due_date": {"2026-11-03"}})
	as.Equal(303, res.Code)

	task := &models.Task{}
	as.NoError(as.DB.First(task))
	as.Equal("Con fecha", task.Title)
	as.Equal("2026-11-03", task.DueDateValue())
}

func (as *ActionSuite) Test_TasksCreate_WithoutDueDate() {
	res := as.HTML("/tasks").Post(url.Values{"title": {"Sin fecha"}, "due_date": {""}})
	as.Equal(303, res.Code)

	task := &models.Task{}
	as.NoError(as.DB.First(task))
	as.Nil(task.DueDate)
}

func (as *ActionSuite) Test_TasksCreate_InvalidDueDate() {
	res := as.HTML("/tasks").Post(url.Values{"title": {"Fecha rota"}, "due_date": {"2026-13-45"}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "La fecha límite no es válida.")

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksIndex_ShowsDueDate() {
	as.taskWithDueDate("Tarea con fecha", "2026-11-03")

	res := as.HTML("/").Get()
	as.Equal(200, res.Code)

	body := res.Body.String()
	as.Contains(body, `datetime="2026-11-03"`)
	as.Contains(body, "03/11/2026")
	as.Contains(body, `name="due_date"`)
}

func (as *ActionSuite) Test_TasksEdit_PrefillsDueDate() {
	task := as.taskWithDueDate("Tarea con fecha", "2026-11-03")

	res := as.HTML("/tasks/%s/edit", task.ID).Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), `value="2026-11-03"`)
}

func (as *ActionSuite) Test_TasksUpdate_SetsAndClearsDueDate() {
	task := as.createTaskAt("Tarea editable", 7)
	task.Completed = true
	as.NoError(as.DB.Update(task))

	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {"Tarea editable"}, "due_date": {"2026-12-24"}})
	as.Equal(303, res.Code)
	as.NoError(as.DB.Reload(task))
	as.Equal("2026-12-24", task.DueDateValue())
	as.True(task.Completed)
	as.Equal(7, task.Position)

	res = as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {"Tarea editable"}, "due_date": {""}})
	as.Equal(303, res.Code)
	as.NoError(as.DB.Reload(task))
	as.Nil(task.DueDate)
	as.True(task.Completed)
	as.Equal(7, task.Position)
}

func (as *ActionSuite) Test_TasksUpdate_InvalidDueDate() {
	task := as.taskWithDueDate("Título original", "2026-11-03")

	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {"Título nuevo"}, "due_date": {"no-es-fecha"}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "La fecha límite no es válida.")

	as.NoError(as.DB.Reload(task))
	as.Equal("Título original", task.Title)
	as.Equal("2026-11-03", task.DueDateValue())
}
