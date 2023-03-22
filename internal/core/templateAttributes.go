package core

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
	"net/http"
)

// GetTemplateAttributes  получает список всех атрибутов шаблона
func (c *Core) GetTemplateAttributes() ([]models.TemplateAttribute, error) {
	out := []models.TemplateAttribute{}
	if err := c.q.GetTemplateAttributes.Select(&out); err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "{globals.terms.templatesAttributes}", "error", pqErrMsg(err)))
	}
	return out, nil
}

// GetTemplateAttributesByTemplate  получает атрибут шаблона по id шаблона
func (c *Core) GetTemplateAttributesByTemplate(id int) ([]models.TemplateAttribute, error) {
	out := []models.TemplateAttribute{}
	if err := c.q.GetTemplateAttributeByTemplate.Select(&out, id); err != nil {
		return []models.TemplateAttribute{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "{globals.terms.templatesAttributes}", "error", pqErrMsg(err)))
	}
	return out, nil
}

// GetTemplateAttribute получает атрибут шаблона  по номеру id указанному в URL строке
func (c *Core) GetTemplateAttribute(id int) (models.TemplateAttribute, error) {
	var out models.TemplateAttribute
	if err := c.q.GetTemplateAttribute.Select(&out, id); err != nil {
		return models.TemplateAttribute{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "{globals.terms.templatesAttributes}", "error", pqErrMsg(err)))
	}
	return out, nil
}
func (c *Core) DeleteTeplateAttributes(sliceDeleteTemplateAttributes []models.TemplateAttribute, ctx context.Context, tx *sqlx.Tx) error {
	for _, sliceDeleteTemplateAttribute := range sliceDeleteTemplateAttributes {
		deleteString := fmt.Sprintf(c.q.DeleteTemplateAttributes)
		_, err := tx.ExecContext(ctx, deleteString, sliceDeleteTemplateAttribute.ID)
		if err != nil {
			return echo.NewHTTPError(
				http.StatusInternalServerError,
				c.i18n.Ts("globals.messages.errorUpdating", "name", "{globals.templates.DeleteTemplateAttributes}", "error", pqErrMsg(err)))

		}
	}
	return nil
}
