package web

import (
	"net/http"
	"watchma/pkg/utils"
	"watchma/view/common"

	"github.com/a-h/templ"
)

type PageResponse struct {
	Component templ.Component
	Title     string
}

// NewPageResponse creates a standard page response
func NewPageResponse(component templ.Component, title string) PageResponse {
	return PageResponse{
		Component: component,
		Title:     title,
	}
}

// RenderPage handles layout wrapping and rendering of common page elements such as page head and menu
func (h *WebHandler) RenderPage(response PageResponse, w http.ResponseWriter, r *http.Request) {
	user := utils.GetUserFromContext(r)
	pc := common.PageContext{
		Title: response.Title,
		User:  user,
	}

	component := common.Layout(pc, response.Component)
	templ.Handler(component).ServeHTTP(w, r)
}
