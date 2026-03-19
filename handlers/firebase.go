package handlers

import (
	"net/http"
	"time"

	"firebase.google.com/go/auth"
	"github.com/Shubhouy1/asset-management/database"
	"github.com/Shubhouy1/asset-management/database/dbhelpers"
	"github.com/Shubhouy1/asset-management/models"
	"github.com/Shubhouy1/asset-management/utils"
	"github.com/jmoiron/sqlx"
)

func FirebaseRegisterUser(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	var body models.UserRequest

	err := utils.ParseBody(r, &body)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid body", err)
		return
	}

	params := (&auth.UserToCreate{}).
		Email(body.Email).
		Password(body.Password).
		DisplayName(body.Name)

	userRecord, err := utils.FirebaseAuth.CreateUser(ctx, params)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "firebase user creation failed", err)
		return
	}

	joiningDate, err := time.Parse("2006-01-02", body.JoiningDate)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid joining_date format", err)
		return
	}

	var userID string

	txErr := database.Tx(func(tx *sqlx.Tx) error {

		id, err := dbhelpers.CreateUser(
			tx,
			body.Name,
			body.Email,
			"employee",
			body.Type,
			body.PhoneNumber,
			"",
			joiningDate,
		)

		if err != nil {
			return err
		}

		userID = id
		return nil
	})

	if txErr != nil {
		_ = utils.FirebaseAuth.DeleteUser(ctx, userRecord.UID)

		utils.RespondError(w, http.StatusInternalServerError, "database error", txErr)
		return
	}

	claims := map[string]interface{}{
		"role": "employee",
	}

	err = utils.FirebaseAuth.SetCustomUserClaims(ctx, userRecord.UID, claims)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "failed to set claims", err)
		return
	}
	idToken, err := dbhelpers.FirebaseEmailLogin(body.Email, body.Password)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "auto login failed", err)
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"message":    "user registered successfully",
		"firebaseID": userRecord.UID,
		"userID":     userID,
		"token":      idToken,
	})
}
