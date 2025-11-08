package web

import (
	"net/http"

	appctx "watchma/pkg/context"
	"watchma/view/common"

	"github.com/a-h/templ"
)

// RenderPage renders a component with the common layout (header/footer)
func RenderPage(component templ.Component, title string, w http.ResponseWriter, r *http.Request) {
	renderPage(component, title, w, r, true)
}

// RenderPageNoLayout renders a component without the common layout
func RenderPageNoLayout(component templ.Component, title string, w http.ResponseWriter, r *http.Request) {
	renderPage(component, title, w, r, false)
}

// renderPage is the internal helper that handles both cases
func renderPage(component templ.Component, title string, w http.ResponseWriter, r *http.Request, headerFooter bool) {
	user := appctx.GetUserFromRequest(r)

	pc := common.PageContext{
		Title:        title,
		HeaderFooter: headerFooter,
		User:         user,
		Content:      component,
	}

	templ.Handler(common.Layout(pc)).ServeHTTP(w, r)
}
