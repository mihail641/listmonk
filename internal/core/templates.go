package core

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
	"net/http"
)

// GetTemplates retrieves all templates.
func (c *Core) GetTemplates(status string, noBody bool) ([]models.Template, error) {
	var err error
	out := []models.Template{}
	if err := c.q.GetTemplates.Select(&out, 0, noBody, status); err != nil {
		return nil, echo.NewHTTPError(
			http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "{globals.terms.templates}", "error", pqErrMsg(err)))
	}
	for i := range out {
		//определение fk_project_id в  Templates
		projectId := out[i].ProjectId
		//доступ к полю Projects экземпляра структуры Templates по уникальному ключу
		out[i].Projects, err = c.GetProject(projectId)
		if err != nil {
			return nil, echo.NewHTTPError(
				http.StatusInternalServerError,
				c.i18n.Ts("globals.messages.errorFetching", "name", "{globals.templates.GetProject}", "error", pqErrMsg(err)))
		}
		templateAttributesId := out[i].ID
		out[i].TemplateAttributes, err = c.GetTemplateAttributesByTemplate(templateAttributesId)
		if err != nil {
			return nil, echo.NewHTTPError(
				http.StatusInternalServerError,
				c.i18n.Ts("globals.messages.errorFetching", "name", "{globals.templates.GetTemplateAttributesByTemplate}", "error", pqErrMsg(err)))
		}
	}
	return out, nil
}

// GetTemplate retrieves a given template.
func (c *Core) GetTemplate(id int, noBody bool) (models.Template, error) {
	var err error
	var out []models.Template
	if err := c.q.GetTemplates.Select(&out, id, noBody, ""); err != nil {
		return models.Template{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "{globals.terms.templates}", "error", pqErrMsg(err)))
	}
	if len(out) == 0 {
		return models.Template{}, echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "{globals.terms.template}"))
	}
	for i := range out {
		//определение fk_project_id в  Templates
		projectId := out[i].ProjectId
		//доступ к полю Projects экземпляра структуры Templates c уникальным ключом и изменению его значению
		out[i].Projects, err = c.GetProject(projectId)
		if err != nil {
			return out[0], echo.NewHTTPError(
				http.StatusInternalServerError,
				c.i18n.Ts("globals.messages.errorFetching", "name", "{globals.templates.GetProject}", "error", pqErrMsg(err)))
		}
		templateAttributesId := out[i].ID
		out[i].TemplateAttributes, err = c.GetTemplateAttributesByTemplate(templateAttributesId)
		if err != nil {
			return out[0], echo.NewHTTPError(
				http.StatusInternalServerError,
				c.i18n.Ts("globals.messages.errorFetching", "name", "{globals.templates.GetTemplateAttributesByTemplate}", "error", pqErrMsg(err)))
		}
	}
	return out[0], nil
}

// CreateTemplate creates a new template.
func (c *Core) CreateTemplate(name, typ, subject string, body []byte, projectId int, templateAttributes []models.TemplateAttribute) (models.Template, error) {
	ctx := context.Background()
	tx, err := c.db.BeginTxx(ctx, nil)
	if err != nil {
		return models.Template{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "{globals.templates.transaction}", "error", pqErrMsg(err)))
	}
	defer tx.Rollback()
	var newID int
	createStringTemplate := fmt.Sprintf(c.q.CreateTemplates)
	err = tx.QueryRowContext(ctx, createStringTemplate, name, typ, subject, body, projectId).Scan(&newID)
	if err != nil {
		return models.Template{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "{globals.terms.template}", "error", pqErrMsg(err)))
	}
	for _, templateAttribute := range templateAttributes {
		createStringTemplateAttributes := fmt.Sprintf(c.q.CreateTemplateAttributes)

		_, err := tx.ExecContext(ctx, createStringTemplateAttributes, templateAttribute.Key, templateAttribute.Description, templateAttribute.Required, templateAttribute.DefaultValue, templateAttribute.Type, newID)
		if err != nil {
			return models.Template{}, echo.NewHTTPError(
				http.StatusInternalServerError,
				c.i18n.Ts("globals.messages.errorCreating", "name", "{globals.templates.CreateTemplateAttributes}", "error", pqErrMsg(err)))
		}
	}
	if err = tx.Commit(); err != nil {
		return models.Template{}, echo.NewHTTPError(
			http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "{globals.templates.Commit}", "error", pqErrMsg(err)))
	}
	return c.GetTemplate(newID, false)
}

// UpdateTemplate updates a given template.
func (c *Core) UpdateTemplate(templateId int, template models.Template) (models.Template, error) {
	ctx := context.Background()
	tx, err := c.db.BeginTxx(ctx, nil)
	if err != nil {
		return models.Template{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "{globals.templates.transaction}", "error", pqErrMsg(err)))
	}
	defer tx.Rollback()
	updateTemplateString := fmt.Sprintf(c.q.UpdateTemplate)
	_, err = tx.ExecContext(ctx, updateTemplateString, templateId, template.Name, template.Subject, template.Body, template.Projects.ID)
	if err != nil {
		return models.Template{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "{globals.terms.template}", "error", pqErrMsg(err)))
	}
	templateAttributesDB, err := c.GetTemplateAttributesByTemplate(templateId)
	if err != nil {
		return models.Template{}, echo.NewHTTPError(
			http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "{globals.templates.GetTemplateAttributesByTemplate}", "error", pqErrMsg(err)))
	}
	//слайс Id на удаление
	var sliceDeleteTemplateAttributes []models.TemplateAttribute
	//слайс Id на изменение
	var sliceUpdateTemplateAttributes []models.TemplateAttribute
	//слайс Id на добавление
	var sliceCreateTemplateAttributes []models.TemplateAttribute
	////мапа ключ-id, значение 1
	mapIdTemplateAttributes := make(map[int]int)
	//цикл для получения слайса Id пришедших из вне
	for templateAttrKey := range template.TemplateAttributes {
		if template.TemplateAttributes[templateAttrKey].ID != 0 {
			mapIdTemplateAttributes[template.TemplateAttributes[templateAttrKey].ID]++
		}
		if template.TemplateAttributes[templateAttrKey].ID == 0 {
			sliceCreateTemplateAttributes = append(sliceCreateTemplateAttributes, template.TemplateAttributes[templateAttrKey])
		}
	}
	//цикл для получения слайса Id пришедших из БД
	for templateAttributeDBKey := range templateAttributesDB {
		if _, ok := mapIdTemplateAttributes[templateAttributesDB[templateAttributeDBKey].ID]; ok {
			sliceUpdateTemplateAttributes = append(sliceUpdateTemplateAttributes, templateAttributesDB[templateAttributeDBKey])
		} else {
			sliceDeleteTemplateAttributes = append(sliceDeleteTemplateAttributes, templateAttributesDB[templateAttributeDBKey])
		}
	}
	err = c.DeleteTeplateAttributes(sliceDeleteTemplateAttributes, ctx, tx)
	if err != nil {
		return models.Template{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "{globals.templates.DeleteTemplateAttributes}", "error", pqErrMsg(err)))
	}
	err = c.updateTemplateAttribute(templateId, sliceCreateTemplateAttributes, sliceUpdateTemplateAttributes, ctx, tx)
	if err != nil {
		return models.Template{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "{globals.templates.UpdateTemplateAttributes}", "error", pqErrMsg(err)))
	}
	if err = tx.Commit(); err != nil {
		return models.Template{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "{globals.templates.Commit}", "error", pqErrMsg(err)))
	}
	return c.GetTemplate(templateId, false)
}

//updateTemplate-приватная функция которая сравнивает слайс структур полученный от клиента со слайсом структур в БД
func (c *Core) updateTemplateAttribute(templateId int, sliceCreateTemplateAttributes []models.TemplateAttribute, sliceUpdateTemplateAttributes []models.TemplateAttribute, ctx context.Context, tx *sqlx.Tx) error {
	var err error
	for _, sliceUpdateTemplateAttribute := range sliceUpdateTemplateAttributes {
		updateString := fmt.Sprintf(c.q.UpdateTemplateAttribute)
		_, err := tx.ExecContext(ctx, updateString, sliceUpdateTemplateAttribute.ID, sliceUpdateTemplateAttribute.Key, sliceUpdateTemplateAttribute.Description, sliceUpdateTemplateAttribute.DefaultValue, sliceUpdateTemplateAttribute.IDTemplate)
		if err != nil {
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				c.i18n.Ts("globals.messages.errorUpdating", "name", "{globals.templates.UpdateTemplateAttributes}", "error", pqErrMsg(err)))
		}

	}
	for _, sliceCreateTemplateAttribute := range sliceCreateTemplateAttributes {
		createString := fmt.Sprintf(c.q.CreateTemplateAttributes)
		_, err := tx.ExecContext(ctx, createString, sliceCreateTemplateAttribute.Key, sliceCreateTemplateAttribute.Description, sliceCreateTemplateAttribute.Required, sliceCreateTemplateAttribute.DefaultValue, sliceCreateTemplateAttribute.Type, templateId)
		if err != nil {
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				c.i18n.Ts("globals.messages.errorUpdating", "name", "{globals.templates.CreateTemplateAttributes}", "error", pqErrMsg(err)))
		}
	}

	return err
}

// SetDefaultTemplate sets a template as default.
func (c *Core) SetDefaultTemplate(id int) error {
	if _, err := c.q.SetDefaultTemplate.Exec(id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "{globals.terms.template}", "error", pqErrMsg(err)))
	}
	return nil
}

// DeleteTemplate deletes a given template.
func (c *Core) DeleteTemplate(id int) error {
	var delID int
	if err := c.q.DeleteTemplate.Get(&delID, id); err != nil && err != sql.ErrNoRows {
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorDeleting", "name", "{globals.terms.template}", "error", pqErrMsg(err)))
	}
	if delID == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, c.i18n.T("templates.cantDeleteDefault"))
	}
	return nil
}
