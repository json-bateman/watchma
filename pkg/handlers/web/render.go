package web

import (
	"net/http"
	"watchma/view/common"

	"github.com/a-h/templ"
)

func (h *WebHandler) RenderPage(component templ.Component, title string, w http.ResponseWriter, r *http.Request) {
	h.renderPage(component, title, w, r, true)
}

func (h *WebHandler) RenderPageNoLayout(component templ.Component, title string, w http.ResponseWriter, r *http.Request) {
	h.renderPage(component, title, w, r, false)
}

func (h *WebHandler) renderPage(component templ.Component, title string, w http.ResponseWriter, r *http.Request, headerFooter bool) {
	user := h.GetUserFromContext(r)

	pc := common.PageContext{
		Title:        title,
		HeaderFooter: headerFooter,
		User:         user,
		Content:      component,
	}

	templ.Handler(common.Layout(pc)).ServeHTTP(w, r)
}
