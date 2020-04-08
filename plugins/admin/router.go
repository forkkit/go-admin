package admin

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/auth"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/errors"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/modules/page"
	"github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/types"
	"net/http"
)

// initRouter initialize the router and return the context.
func (admin *Admin) initRouter() *Admin {
	app := context.NewApp()

	route := app.Group(config.Prefix(), admin.globalErrorHandler)

	// auth
	route.GET(config.GetLoginUrl(), admin.handler.ShowLogin)
	route.POST("/signin", admin.handler.Auth)

	// auto install
	route.GET("/install", admin.handler.ShowInstall)
	route.POST("/install/database/check", admin.handler.CheckDatabase)

	for _, path := range template.Get(config.GetTheme()).GetAssetList() {
		route.GET("/assets"+path, admin.handler.Assets)
	}

	for _, path := range template.GetComponentAssetLists() {
		route.GET("/assets"+path, admin.handler.Assets)
	}

	authRoute := route.Group("/", auth.Middleware(admin.conn))

	// auth
	authRoute.GET("/logout", admin.handler.Logout)

	authPrefixRoute := route.Group("/", auth.Middleware(admin.conn), admin.guardian.CheckPrefix)

	// menus
	authRoute.POST("/menu/delete", admin.guardian.MenuDelete, admin.handler.DeleteMenu).Name("menu_delete")
	authRoute.POST("/menu/new", admin.guardian.MenuNew, admin.handler.NewMenu).Name("menu_new")
	authRoute.POST("/menu/edit", admin.guardian.MenuEdit, admin.handler.EditMenu).Name("menu_edit")
	authRoute.POST("/menu/order", admin.handler.MenuOrder).Name("menu_order")
	authRoute.GET("/menu", admin.handler.ShowMenu).Name("menu")
	authRoute.GET("/menu/edit/show", admin.handler.ShowEditMenu).Name("menu_edit_show")
	authRoute.GET("/menu/new", admin.handler.ShowNewMenu).Name("menu_new_show")

	// add delete modify query
	authPrefixRoute.GET("/info/:__prefix/detail", admin.handler.ShowDetail).Name("detail")
	authPrefixRoute.GET("/info/:__prefix/edit", admin.guardian.ShowForm, admin.handler.ShowForm).Name("show_edit")
	authPrefixRoute.GET("/info/:__prefix/new", admin.guardian.ShowNewForm, admin.handler.ShowNewForm).Name("show_new")
	authPrefixRoute.POST("/edit/:__prefix", admin.guardian.EditForm, admin.handler.EditForm).Name("edit")
	authPrefixRoute.POST("/new/:__prefix", admin.guardian.NewForm, admin.handler.NewForm).Name("new")
	authPrefixRoute.POST("/delete/:__prefix", admin.guardian.Delete, admin.handler.Delete).Name("delete")
	authPrefixRoute.POST("/export/:__prefix", admin.guardian.Export, admin.handler.Export).Name("export")
	authPrefixRoute.GET("/info/:__prefix", admin.handler.ShowInfo).Name("info")

	authPrefixRoute.POST("/update/:__prefix", admin.guardian.Update, admin.handler.Update).Name("update")

	route.ANY("/operation/:__goadmin_op_id", auth.Middleware(admin.conn), admin.handler.Operation)

	// TODO: crud json apis
	//apiRoute := route.Group("/api", auth.Middleware(admin.conn), admin.guardian.CheckPrefix)
	//apiRoute.GET("/list/:__prefix")
	//apiRoute.GET("/detail/:__prefix")
	//apiRoute.DELETE("/delete/:__prefix")
	//apiRoute.POST("/update/:__prefix")
	//apiRoute.POST("/create/:__prefix")

	admin.app = app
	return admin
}

func (admin *Admin) globalErrorHandler(ctx *context.Context) {
	defer admin.handler.GlobalDeferHandler(ctx)

	// Check site offline flag
	if config.GetSiteOff() {
		if ctx.WantsHTML() {
			page.SetPageContent(ctx, auth.Auth(ctx), func(ctx interface{}) (types.Panel, error) {
				return template.WarningPanel(errors.SiteOff), nil
			}, admin.conn)
		} else {
			ctx.JSON(http.StatusForbidden, map[string]interface{}{
				"code": http.StatusForbidden,
				"msg":  language.Get(errors.SiteOff),
			})
		}
	}

	ctx.Next()
}
