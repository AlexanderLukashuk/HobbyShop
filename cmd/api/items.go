package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"fainal.net/internal/data"
	"fainal.net/internal/validator"
)

func (app *application) createItemHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Price       int32    `json:"price"`
		Category    []string `json:"category"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
	}
	authorizationHeader := r.Header.Get("Authorization")

	headerParts := strings.Split(authorizationHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		app.invalidAuthenticationTokenResponse(w, r)
		return
	}
	v := validator.New()
	token := headerParts[1]

	if data.ValidateTokenPlaintext(v, token); !v.Valid() {
		app.invalidAuthenticationTokenResponse(w, r)
		return
	}
	user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidAuthenticationTokenResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	if user.Role != "admin" {
		app.notPermittedResponse(w, r)
		return
	}
	item := &data.Item{
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Category:    input.Category,
	}

	err = app.models.Items.Insert(item)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/items/%d", item.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"item": item}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) showItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	item, err := app.models.Items.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"item": item}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) updateItemHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	item, err := app.models.Items.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Name        *string  `json:"name"`
		Description *string  `json:"description"`
		Price       *int32   `json:"price"`
		Category    []string `json:"category"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Name != nil {
		item.Name = *input.Name
	}
	if input.Description != nil {
		item.Description = *input.Description
	}
	if input.Price != nil {
		item.Price = *input.Price
	}
	if input.Category != nil {
		item.Category = input.Category
	}

	v := validator.New()
	if data.ValidateItem(v, item); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Items.Update(item)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"item": item}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteItemHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Items.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "item successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listItemsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string
		Category []string
		data.Filters
	}
	v := validator.New()
	qs := r.URL.Query()
	input.Name = app.readString(qs, "name", "")
	input.Category = app.readCSV(qs, "category", []string{})
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "name", "price", "-id", "-name", "-price"}
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	items, err := app.models.Items.GetAll(input.Name, input.Category, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"items": items}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
