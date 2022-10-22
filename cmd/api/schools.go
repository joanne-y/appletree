// Filename: cms/api/schools.go

package main

import (
	"errors"
	"fmt"
	"net/http"

	"appletree.joanneyong.net/internal/data"
	"appletree.joanneyong.net/internal/validator"
)

// createSchoolHandler for the "POST /v1/schools" endpoint
func (app *application) createSchoolHandler(w http.ResponseWriter, r *http.Request) {
	// Our target decode destination
	var input struct {
		Name    string   `json:"name"`
		Level   string   `json:"level"`
		Contact string   `json:"contact"`
		Phone   string   `json:"phone"`
		Email   string   `json:"email"`
		Website string   `json:"website"`
		Address string   `json:"address"`
		Mode    []string `json:"mode"`
	}
	// Initialize a new json.Decoder instance
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Copy the values from the input struct to a new School struct
	school := &data.School{
		Name:    input.Name,
		Level:   input.Level,
		Contact: input.Contact,
		Phone:   input.Phone,
		Email:   input.Email,
		Website: input.Website,
		Address: input.Address,
		Mode:    input.Mode,
	}

	// Initialize a new Validator instance
	v := validator.New()

	// Check the map to determine if there were any validation errors
	if data.ValidateSchool(v, school); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Create a School
	err = app.models.Schools.Insert(school)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	// Create a Location header for the newly created resource/School
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/schools/%d", school.ID))
	// Write the JSON response with 201 - Created status code with the body
	// being the School data and the header being the headers map
	err = app.writeJSON(w, http.StatusCreated, envelope{"school": school}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// showSchoolHandler for the "GET /v1/schools/:id" endpoint
func (app *application) showSchoolHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Fetch the specific school
	school, err := app.models.Schools.Get(id)
	// Handle errors
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Write the data returned by Get()
	err = app.writeJSON(w, http.StatusOK, envelope{"school": school}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateSchoolHandler(w http.ResponseWriter, r *http.Request) {
	// This method does a partial replacement
	// Get the id for the school that needs updating
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Fetch the orginal record from the database
	school, err := app.models.Schools.Get(id)
	// Handle errors
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Create an input struct to hold data read in from the client
	// We update input struct to use pointers because pointers have a
	// default value of nil
	// If a field remains nil then we know that the client did not update it
	var input struct {
		Name    *string  `json:"name"`
		Level   *string  `json:"level"`
		Contact *string  `json:"contact"`
		Phone   *string  `json:"phone"`
		Email   *string  `json:"email"`
		Website *string  `json:"website"`
		Address *string  `json:"address"`
		Mode    []string `json:"mode"`
	}

	// Initialize a new json.Decoder instance
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// Check for updates
	if input.Name != nil {
		school.Name = *input.Name
	}
	if input.Level != nil {
		school.Level = *input.Level
	}
	if input.Contact != nil {
		school.Contact = *input.Contact
	}
	if input.Phone != nil {
		school.Phone = *input.Phone
	}
	if input.Email != nil {
		school.Email = *input.Email
	}
	if input.Website != nil {
		school.Website = *input.Website
	}
	if input.Address != nil {
		school.Address = *input.Address
	}
	if input.Mode != nil {
		school.Mode = input.Mode
	}

	// Perform validation on the updated School. If validation fails, then
	// we send a 422 - Unprocessable Entity respose to the client
	// Initialize a new Validator instance
	v := validator.New()

	// Check the map to determine if there were any validation errors
	if data.ValidateSchool(v, school); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Pass the updated School record to the Update() method
	err = app.models.Schools.Update(school)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Write the data returned by Get()
	err = app.writeJSON(w, http.StatusOK, envelope{"school": school}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteSchoolHandler(w http.ResponseWriter, r *http.Request) {
	// Get the id for the school that needs updating
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Delete the School from the database. Send a 404 Not Found status code to the
	// client if there is no matching record
	err = app.models.Schools.Delete(id)
	// Handle errors
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Return 200 Status OK to the client with a success message
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "school successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// The listSchoolsHandler() allows the client to see a listing of schools
// based on a set of criteria
func (app *application) listSchoolsHandler(w http.ResponseWriter, r *http.Request) {
	// Create an input struct to hold our query parameters
	var input struct {
		Name  string
		Level string
		Mode  []string
		data.Filters
	}
	// Initialize a validator
	v := validator.New()
	// Get the URL values map
	qs := r.URL.Query()
	// Use the helper methods to extract the values
	input.Name = app.readString(qs, "name", "")
	input.Level = app.readString(qs, "level", "")
	input.Mode = app.readCSV(qs, "mode", []string{})
	// Get the page information
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	// Get the sort information
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortList = []string{"id", "name", "level", "-id", "-name", "-level"}
	// Check for validation errors
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Get a listing of all schools
	schools, err := app.models.Schools.GetAll(input.Name, input.Level, input.Mode, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Send a JSON response containg all the schools
	err = app.writeJSON(w, http.StatusOK, envelope{"schools": schools}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
