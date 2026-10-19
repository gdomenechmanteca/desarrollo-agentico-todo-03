package models

import (
	"errors"
	"strings"

	"github.com/gofrs/uuid"
)

func (ms *ModelSuite) Test_Task_Create_Valid() {
	task := &Task{Title: "Comprar café"}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	ms.NotZero(task.ID)
	ms.False(task.Completed)

	count, err := ms.DB.Count(&Task{})
	ms.NoError(err)
	ms.Equal(1, count)
}

func (ms *ModelSuite) Test_Task_Title_Required() {
	task := &Task{Title: ""}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.True(verrs.HasAny())
	ms.Contains(verrs.Get("title"), "El título es obligatorio.")

	count, err := ms.DB.Count(&Task{})
	ms.NoError(err)
	ms.Equal(0, count)
}

func (ms *ModelSuite) Test_Task_Title_OnlySpaces_Invalid() {
	task := &Task{Title: "   \t  "}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.True(verrs.HasAny())
	ms.Contains(verrs.Get("title"), "El título es obligatorio.")
}

func (ms *ModelSuite) Test_Task_Title_MaxLength() {
	task := &Task{Title: strings.Repeat("a", TitleMaxLength)}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.False(verrs.HasAny())
}

func (ms *ModelSuite) Test_Task_Title_TooLong_Invalid() {
	task := &Task{Title: strings.Repeat("a", TitleMaxLength+1)}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.True(verrs.HasAny())
	ms.NotEmpty(verrs.Get("title"))
}

func (ms *ModelSuite) Test_Task_Toggle_Persists() {
	task := &Task{Title: "Revisar el correo"}
	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	task.Completed = true
	ms.NoError(ms.DB.Update(task))

	reloaded := &Task{}
	ms.NoError(ms.DB.Find(reloaded, task.ID))
	ms.True(reloaded.Completed)
}

func (ms *ModelSuite) createTaskAt(title string, position int) *Task {
	task := &Task{Title: title, Position: position}
	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.False(verrs.HasAny())
	return task
}

func (ms *ModelSuite) positionOf(task *Task) int {
	reloaded := &Task{}
	ms.NoError(ms.DB.Find(reloaded, task.ID))
	return reloaded.Position
}

func (ms *ModelSuite) Test_NextTaskPosition_EmptyTable() {
	next, err := NextTaskPosition(ms.DB)
	ms.NoError(err)
	ms.Equal(1, next)
}

func (ms *ModelSuite) Test_NextTaskPosition_AfterMax() {
	ms.createTaskAt("Primera", 1)
	ms.createTaskAt("Quinta", 5)

	next, err := NextTaskPosition(ms.DB)
	ms.NoError(err)
	ms.Equal(6, next)
}

func (ms *ModelSuite) Test_ReorderTasks_PersistsPositions() {
	a := ms.createTaskAt("A", 1)
	b := ms.createTaskAt("B", 2)
	c := ms.createTaskAt("C", 3)

	ms.NoError(ReorderTasks(ms.DB, []uuid.UUID{c.ID, a.ID, b.ID}))

	ms.Equal(1, ms.positionOf(c))
	ms.Equal(2, ms.positionOf(a))
	ms.Equal(3, ms.positionOf(b))
}

func (ms *ModelSuite) assertInvalidOrderKeepsPositions(ids func(a, b, c *Task) []uuid.UUID) {
	a := ms.createTaskAt("A", 1)
	b := ms.createTaskAt("B", 2)
	c := ms.createTaskAt("C", 3)

	err := ReorderTasks(ms.DB, ids(a, b, c))
	ms.True(errors.Is(err, ErrInvalidTaskOrder))

	ms.Equal(1, ms.positionOf(a))
	ms.Equal(2, ms.positionOf(b))
	ms.Equal(3, ms.positionOf(c))
}

func (ms *ModelSuite) Test_ReorderTasks_MissingID_Invalid() {
	ms.assertInvalidOrderKeepsPositions(func(a, b, c *Task) []uuid.UUID {
		return []uuid.UUID{c.ID, a.ID}
	})
}

func (ms *ModelSuite) Test_ReorderTasks_UnknownID_Invalid() {
	ms.assertInvalidOrderKeepsPositions(func(a, b, c *Task) []uuid.UUID {
		return []uuid.UUID{c.ID, a.ID, uuid.Must(uuid.NewV4())}
	})
}

func (ms *ModelSuite) Test_ReorderTasks_DuplicateID_Invalid() {
	ms.assertInvalidOrderKeepsPositions(func(a, b, c *Task) []uuid.UUID {
		return []uuid.UUID{c.ID, a.ID, a.ID}
	})
}
