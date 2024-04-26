package main

import (
	"net/http"

	"awesomeeng/otel-demo/utils"
)

type user struct {
	ID       int64  `json:"id" validate:"-"`
	UserName string `json:"user_name" validate:"required"`
	Account  string `json:"account" validate:"required"`
	Amount   int
}

func getUser(w http.ResponseWriter, r *http.Request) {
	var u user
	_, span := tracer.Start(r.Context(), "get user")
	defer span.End()
	u = user{
		ID:       389023,
		UserName: "user-1",
		Account:  "account-1",
		Amount:   200,
	}
	utils.WriteResponse(w, http.StatusOK, u)
}
