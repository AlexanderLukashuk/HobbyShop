package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *Application) routes() http.Handler {

	router := httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/item", app.createItemHandler)
	router.HandlerFunc(http.MethodGet, "/v1/items/:id", app.showItemHandler)
	router.HandlerFunc(http.MethodGet, "/v1/items", app.listItemsHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/items/:id", app.updateItemHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/items/:id", app.requireActivatedUser(app.deleteItemHandler))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	router.HandlerFunc(http.MethodPost, "/v1/basket", app.createBasketHandler)
	router.HandlerFunc(http.MethodGet, "/v1/basket/:id", app.showBasketHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/basket/:id", app.updateBasketHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/basket/:id", app.deleteBasketHandler)

	router.HandlerFunc(http.MethodGet, "/v1/get-user", app.getUserInfoHandler)

	return app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router))))
}
