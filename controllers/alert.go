package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/laincloud/hagrid/config"
	"github.com/laincloud/hagrid/models"
)

func AddAlertHandler(w http.ResponseWriter, r *http.Request) {
	enabled, _ := strconv.ParseBool(r.FormValue("enabled"))
	if r.FormValue("name") == "" {
		writeResponse(w, "Alert name is empty", http.StatusBadRequest)
		return
	}

	var userName string
	var err error

	if userName, err = getAuthedUserInfo(r); err != nil {
		writeResponse(w, "Authorization failed", http.StatusUnauthorized)
		return
	}

	user := &models.User{}
	if err = models.GetUser(userName, user); err != nil {
		writeResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if models.IsAlertDuplicated(r.FormValue("name")) {
		writeResponse(w, "The alert's name is duplicated", http.StatusConflict)
		return
	}

	newAlert := &models.Alert{
		Name:    r.FormValue("name"),
		Enabled: enabled,
		Admins:  []models.User{*user},
	}
	if err := models.SaveAlert(newAlert); err != nil {
		writeResponse(w, "Add alert error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Add alert successfully"))
}

func UpdateAlertHandler(w http.ResponseWriter, r *http.Request) {
	enabled, _ := strconv.ParseBool(r.FormValue("enabled"))
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	if _, err := authorize(w, r, id); err != nil {
		return
	}

	newAlert := &models.Alert{}

	if err := models.GetAlert(newAlert, id); err != nil {
		writeResponse(w, "Find alert failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	newAlert.Enabled = enabled

	if err := models.SaveAlert(newAlert); err != nil {
		writeResponse(w, "Save alert error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := models.SynchronizeAlert(id); err != nil {
		writeResponse(w, "Synchronize alert to icinga error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Save alert successfully"))
}

func GetAlertHandler(w http.ResponseWriter, r *http.Request) {
	var output []byte
	var err error
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	newAlert := &models.Alert{}

	if err = models.GetDetailedAlert(newAlert, id); err != nil {
		writeResponse(w, "Find alert failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if output, err = json.Marshal(newAlert); err != nil {
		writeResponse(w, "Marshall alert failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(output)
}

func GetAllAlertsHandler(w http.ResponseWriter, r *http.Request) {

	var (
		output   []byte
		err      error
		userName string
	)

	userName, _ = getAuthedUserInfo(r)

	var alerts []models.Alert
	if userName == config.GetSuperUser() {
		models.GetAllAlerts(&alerts)
	} else {
		user := &models.User{}
		models.GetUser(userName, user)
		alerts = user.AdminedAlerts
	}

	if output, err = json.Marshal(alerts); err != nil {
		writeResponse(w, "Marshall alert failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(output)
}

func DeleteAlertHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	if _, err := authorize(w, r, id); err != nil {
		return
	}
	models.DeleteAlert(id)
	writeResponse(w, "Delete alert successfully", http.StatusNoContent)
}
