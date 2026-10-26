package models

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// TitleMaxLength es el número máximo de caracteres permitido en el título.
const TitleMaxLength = 200

// TaskOrder es el orden en que se muestra la lista de tareas.
const TaskOrder = "position asc, created_at asc, id asc"

// ErrInvalidTaskOrder indica que un orden no incluye cada tarea exactamente una vez.
var ErrInvalidTaskOrder = errors.New("el orden debe incluir cada tarea exactamente una vez")

// DueDateLayout es el formato de la fecha límite en formularios (input type="date").
const DueDateLayout = "2006-01-02"

// ErrInvalidDueDate indica que la fecha límite no tiene el formato AAAA-MM-DD.
var ErrInvalidDueDate = errors.New("La fecha límite no es válida.")

// Task representa una tarea de la lista.
type Task struct {
	ID        uuid.UUID  `json:"id" db:"id" form:"-"`
	Title     string     `json:"title" db:"title" form:"title"`
	Completed bool       `json:"completed" db:"completed" form:"-"`
	Position  int        `json:"position" db:"position" form:"-"`
	DueDate   *time.Time `json:"due_date" db:"due_date" form:"-"`
	CreatedAt time.Time  `json:"created_at" db:"created_at" form:"-"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at" form:"-"`
}

// Tasks es una colección de Task.
type Tasks []Task

// ParseDueDate interpreta una fecha límite AAAA-MM-DD; la cadena vacía significa sin fecha.
func ParseDueDate(s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	d, err := time.Parse(DueDateLayout, s)
	if err != nil {
		return nil, ErrInvalidDueDate
	}
	return &d, nil
}

// DueDateValue devuelve la fecha límite en formato AAAA-MM-DD, o "" si no tiene.
func (t Task) DueDateValue() string {
	if t.DueDate == nil {
		return ""
	}
	return t.DueDate.Format(DueDateLayout)
}

// DueDateLabel devuelve la fecha límite en formato dd/mm/aaaa, o "" si no tiene.
func (t Task) DueDateLabel() string {
	if t.DueDate == nil {
		return ""
	}
	return t.DueDate.Format("02/01/2006")
}

// Validate se ejecuta en cada llamada a los métodos "pop.Validate*".
func (t *Task) Validate(tx *pop.Connection) (*validate.Errors, error) {
	verrs := validate.NewErrors()

	title := strings.TrimSpace(t.Title)
	if title == "" {
		verrs.Add("title", "El título es obligatorio.")
	} else if len([]rune(title)) > TitleMaxLength {
		verrs.Add("title", fmt.Sprintf("El título no puede superar los %d caracteres.", TitleMaxLength))
	}

	return verrs, nil
}

// NextTaskPosition devuelve la posición que ocupa una tarea añadida al final.
func NextTaskPosition(tx *pop.Connection) (int, error) {
	next := struct {
		Position int `db:"position"`
	}{}
	if err := tx.RawQuery("SELECT COALESCE(MAX(position), 0) + 1 AS position FROM tasks").First(&next); err != nil {
		return 0, err
	}
	return next.Position, nil
}

// ReorderTasks asigna a las tareas el orden de ids, que debe incluirlas todas exactamente una vez.
func ReorderTasks(tx *pop.Connection, ids []uuid.UUID) error {
	tasks := Tasks{}
	if err := tx.Select("id").All(&tasks); err != nil {
		return err
	}
	if len(ids) != len(tasks) {
		return ErrInvalidTaskOrder
	}

	pending := make(map[uuid.UUID]bool, len(tasks))
	for _, t := range tasks {
		pending[t.ID] = true
	}
	for _, id := range ids {
		if !pending[id] {
			return ErrInvalidTaskOrder
		}
		delete(pending, id)
	}

	for i, id := range ids {
		if err := tx.RawQuery("UPDATE tasks SET position = ? WHERE id = ?", i+1, id).Exec(); err != nil {
			return err
		}
	}
	return nil
}
