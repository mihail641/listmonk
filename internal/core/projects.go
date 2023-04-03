package core

import (
	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"net/http"
)

// GetProjects  получает список всех проектов
func (c *Core) GetProjects() ([]models.Project, error) {
	out := []models.Project{}
	if err := c.q.GetProjects.Select(&out); err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "{globals.terms.projects}", "error", pqErrMsg(err)))
	}

	return out, nil
}

// GetProject получает проект по номеру id указанному в URL строке
func (c *Core) GetProject(id int) (models.Project, error) {
	var out []models.Project
	if err := c.q.GetProject.Select(&out, id); err != nil {
		return models.Project{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "{globals.terms.projects}", "error", pqErrMsg(err)))
	}
	if len(out) == 0 {
		return models.Project{}, echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "{globals.terms.project}"))
	}
	return out[0], nil
}

// CreateProject  создает новый проект.
func (c *Core) CreateProject(l models.Project) (models.Project, error) {
	var newID int
	if err := c.q.CreateProject.Get(&newID, l.Name, l.SenderEmail, l.SenderName, l.Description); err != nil {
		c.log.Printf("error creating project: %v", err)
		return models.Project{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "{globals.terms.project}", "error", pqErrMsg(err)))
	}
	return c.GetProject(newID)
}

// UpdateProject  изменяет проект id, которого указан в URL строке.
func (c *Core) UpdateProject(id int, l models.Project) (models.Project, error) {
	res, err := c.q.UpdateProject.Exec(id, l.Name, l.Description, l.SenderName, l.SenderEmail)
	if err != nil {
		c.log.Printf("error updating project: %v", err)
		return models.Project{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "{globals.terms.project}", "error", pqErrMsg(err)))
	}
	// RowsAffected возвращает количество строк, затронутых
	// обновлением, вставкой или удалением.
	if n, _ := res.RowsAffected(); n == 0 {
		return models.Project{}, echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "{globals.terms.project}"))
	}

	return c.GetProject(id)
}

// DeleteProjects  удаляет по Id из URL проекты.
func (c *Core) DeleteProjects(ids []int) error {
	if _, err := c.q.DeleteProjects.Exec(pq.Array(ids)); err != nil {
		c.log.Printf("error deleting projects: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorDeleting", "name", "{globals.terms.project}", "error", pqErrMsg(err)))
	}
	return nil
}
