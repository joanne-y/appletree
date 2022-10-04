//Filename: cmd/api/schools.go

package main

import (
	"fmt"
	"net/http"
	"time"

	"appletree.joanneyong.net/internal/data"
	"appletree.joanneyong.net/internal/validator"
)

// createSchoolHandler for the "POST /v1/schools" endpoint
func (app *application) createSchoolHandler(w http.ResponseWriter, r *http.Request) {
	// Our target decvode destination
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
	// Initialize a new Validator instance
	v := validator.New()
	// Use the Check() method to execute our validation checks
	v.Check(input.Name != "", "name", "must be provided")
	v.Check(len(input.Name) <= 200, "name", "must not be more than 200 bytes long")

	v.Check(input.Level != "", "level", "must be provided")
	v.Check(len(input.Level) <= 200, "level", "must not be more than 200 bytes long")

	v.Check(input.Contact != "", "contact", "must be provided")
	v.Check(len(input.Contact) <= 200, "contact", "must not be more than 200 bytes long")

	v.Check(input.Phone != "", "phone", "must be provided")
	v.Check(validator.Matches(input.Phone, validator.PhoneRX), "phone", "must be a valid phone number")

	v.Check(input.Email != "", "email", "must be provided")
	v.Check(validator.Matches(input.Email, validator.EmailRX), "email", "must be a valid email address")

	v.Check(input.Website != "", "website", "must be provided")
	v.Check(validator.ValidWebsite(input.Website), "website", "must be a valid URL")

	v.Check(input.Address != "", "address", "must be provided")
	v.Check(len(input.Address) <= 500, "address", "must not be more than 500 bytes long")

	v.Check(input.Mode != nil, "mode", "must be provided")
	v.Check(len(input.Mode) >= 1, "mode", "must contain at least 1 entry")
	v.Check(len(input.Mode) <= 5, "mode", "must not contain at most 5 entries")
	v.Check(validator.Unique(input.Mode), "mode", "must not contain duplicate entries")
	//Check the map to determine if there were any validation errors
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Display the request
	fmt.Fprintf(w, "%+v\n", input)

}

// showSchoolHnadler for the "GET v1/schools/:id" endpoint
func (app *application) showSchoolHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Create a new instance of the School struct containing the ID we extracted
	// from our URL and some sample data
	school := data.School{
		ID:        id,
		CreatedAt: time.Now(),
		Name:      "Apple Tree",
		Level:     "High School",
		Contact:   "Anna Smith",
		Phone:     "601-4411",
		Address:   "14 Apple street",
		Mode:      []string{"blended", "online"},
		Version:   1,
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"school": school}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
