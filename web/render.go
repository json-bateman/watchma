package web

import (
	"net/http"

	appctx "watchma/pkg/context"
	"watchma/web/views/common"

	"github.com/a-h/templ"
)

// RenderPage renders a component WITH the <header> and <footer>
func RenderPage(component templ.Component, title string, w http.ResponseWriter, r *http.Request) {
	renderPage(component, title, w, r, true)
}

// RenderPageNoLayout renders a component WITHOUT the <header> and <footer>
func RenderPageNoLayout(component templ.Component, title string, w http.ResponseWriter, r *http.Request) {
	renderPage(component, title, w, r, false)
}

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
