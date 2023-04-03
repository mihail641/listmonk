package main

import (
	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

// handleGetProjects получает список всех проектов
func handleGetProjects(c echo.Context) error {
	var (
		app = c.Get("app").(*App)
	)
	out, err := app.core.GetProjects()
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{out})
}

//handleGetProject получает конкретный 1 проект, указанный в URL адресе.
func handleGetProject(c echo.Context) error {
	var (
		app = c.Get("app").(*App)
	)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, app.i18n.T("globals.messages.invalidID"))
		return err
	}
	out, err := app.core.GetProject(id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{out})
}

// handleCreatePtojects создает новый проект.
func handleCreateProjects(c echo.Context) error {
	var (
		app = c.Get("app").(*App)
		l   = models.Project{}
	)
	//Bind привязывает тело запроса к предоставленной структре. Связующее по умолчанию
	//// делает это на основе заголовка Content-Type.
	if err := c.Bind(&l); err != nil {
		return err
	}
	//strHasLen проверяет, имеет ли заданная строка длину в пределах min-max.
	if !strHasLen(l.Name, 1, stdInputMaxLen) {
		return echo.NewHTTPError(http.StatusBadRequest, app.i18n.T("projects.invalidName"))
	}
	out, err := app.core.CreateProject(l)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{out})
}

// handleUpdateProjects handles projects modification.
func handleUpdateProjects(c echo.Context) error {
	var (
		app = c.Get("app").(*App)
	)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, app.i18n.T("globals.messages.invalidID"))
	}
	var l models.Project
	if err := c.Bind(&l); err != nil {
		return err
	}
	// strHasLen проверяет, имеет ли заданная строка длину в пределах min-max.
	if !strHasLen(l.Name, 1, stdInputMaxLen) {
		return echo.NewHTTPError(http.StatusBadRequest, app.i18n.T("projects.invalidName"))
	}
	out, err := app.core.UpdateProject(id, l)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{out})
}

// handleDeleteProjects удаляет один проект
func handleDeleteProjects(c echo.Context) error {
	var (
		app = c.Get("app").(*App)
	)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, app.i18n.T("globals.messages.invalidID"))
	}
	var ids []int
	if id < 1 && len(ids) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, app.i18n.T("globals.messages.invalidID"))
	}
	if id > 0 {
		ids = append(ids, id)
	}
	if err := app.core.DeleteProjects(ids); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, okResp{true})
}
