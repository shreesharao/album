package handler

import (
	"github.com/iris-contrib/swagger/v12"
	"github.com/iris-contrib/swagger/v12/swaggerFiles"
	"github.com/kataras/iris/v12"
	_ "github.com/shreesharao/album/docs"
)

func setupRoutes(app *iris.Application, h *RequestHandler) {

	config := &swagger.Config{
		URL: "http://localhost:5000/swagger/swagger.json", //The url pointing to API definition
	}

	routeApis := app.Party("/api/album")
	{
		routeApis.Post("/v1", h.CreateAlbum)
		routeApis.Delete("/v1/{album:string}", h.DeleteAlbum)
		routeApis.Post("/image/v1", h.CreateImage)
		routeApis.Delete("/image/v1/{album:string}/{image:string}", h.DeleteImage)
		routeApis.Get("/v1/{album:string}", h.GetImages)
		routeApis.Get("/image/v1/{album:string}/{image:string}", h.GetImage)
	}
	app.Get("/swagger/{any:path}", swagger.CustomWrapHandler(config, swaggerFiles.Handler))
	app.Get("/swagger/swagger.json", h.ServeAPIDef)
}
