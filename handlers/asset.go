package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Shubhouy1/asset-management/database"
	"github.com/Shubhouy1/asset-management/database/dbhelpers"
	"github.com/Shubhouy1/asset-management/middleware"
	"github.com/Shubhouy1/asset-management/models"
	"github.com/Shubhouy1/asset-management/utils"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

func CreateAsset(w http.ResponseWriter, r *http.Request) {

	var body models.CreateAssetRequest
	var assetID string

	if err := utils.ParseBody(r, &body); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid body", err)
		return
	}

	if err := validate.Struct(&body); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "validation failed", err)
		return
	}

	warrantyStart, err := time.Parse("2006-01-02", body.WarrantyStart)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid date", err)
		return
	}

	warrantyEnd, err := time.Parse("2006-01-02", body.WarrantyEnd)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid date", err)
		return
	}

	if warrantyEnd.Before(warrantyStart) {
		utils.RespondError(w, http.StatusBadRequest, "invalid warranty range", fmt.Errorf("warranty end before start"))
		return
	}

	txErr := database.Tx(func(tx *sqlx.Tx) error {

		var err error
		assetID, err = dbhelpers.CreateAsset(tx, body.Brand, body.Model, body.SerialNo, body.Type, body.Owner, warrantyStart, warrantyEnd)
		if err != nil {
			return err
		}

		switch body.Type {

		case "laptop":
			if body.Laptop == nil {
				return fmt.Errorf("laptop details required")
			}
			return dbhelpers.InsertLaptop(tx, assetID, body.Laptop)

		case "mouse":
			if body.Mouse == nil {
				return fmt.Errorf("mouse details required")
			}
			return dbhelpers.InsertMouse(tx, assetID, body.Mouse)

		case "keyboard":
			if body.Keyboard == nil {
				return fmt.Errorf("keyboard details required")
			}
			return dbhelpers.InsertKeyboard(tx, assetID, body.Keyboard)

		case "mobile":
			if body.Mobile == nil {
				return fmt.Errorf("mobile details required")
			}
			return dbhelpers.InsertMobile(tx, assetID, body.Mobile)

		default:
			return fmt.Errorf("unsupported asset type")
		}
	})

	if txErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, "failed to create asset", txErr)
		return
	}

	utils.RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"assetID": assetID,
	})
}

func AssignAsset(w http.ResponseWriter, r *http.Request) {

	assetID := chi.URLParam(r, "id")

	if assetID == "" {
		utils.RespondError(w, http.StatusBadRequest, "asset id is required", fmt.Errorf("empty asset id"))
		return
	}

	var body models.AssignRequest

	err := utils.ParseBody(r, &body)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid body", err)
		return
	}

	validateErr := validate.Struct(&body)
	if validateErr != nil {
		utils.RespondError(w, http.StatusBadRequest, "failed to validate body", validateErr)
		return
	}

	auth, ok := middleware.GetAuthContext(r)
	if !ok {
		utils.RespondError(w, http.StatusUnauthorized, "unauthorized", fmt.Errorf("missing auth context"))
		return
	}

	userID := auth.UserID

	err = dbhelpers.AssignAsset(assetID, body.AssignedTo, userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "failed to assign asset", err)
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "asset assigned successfully",
	})
}

func GetTotalAssets(w http.ResponseWriter, r *http.Request) {

	auth, ok := middleware.GetAuthContext(r)
	if !ok {
		utils.RespondError(w, http.StatusUnauthorized, "unauthorized", fmt.Errorf("missing auth context"))
		return
	}

	userID := auth.UserID

	totalAssets, err := dbhelpers.FindTotalAssetById(userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "failed to find total assets", err)
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"totalAssets": totalAssets,
	})
}

func SendAssetToService(w http.ResponseWriter, r *http.Request) {

	assetID := chi.URLParam(r, "id")

	if assetID == "" {
		utils.RespondError(w, http.StatusBadRequest, "invalid id", fmt.Errorf("empty asset id"))
		return
	}

	var body models.SentServiceRequest

	err := utils.ParseBody(r, &body)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid body", err)
		return
	}

	err = validate.Struct(&body)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "validation failed", err)
		return
	}

	serviceStart, err := time.Parse("2006-01-02", body.StartDate)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid service date", err)
		return
	}

	serviceEnd, err := time.Parse("2006-01-02", body.EndDate)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid service end date", err)
		return
	}

	if serviceEnd.Before(serviceStart) {
		utils.RespondError(w, http.StatusBadRequest, "end date must be after start date", fmt.Errorf("service end before start"))
		return
	}

	err = dbhelpers.SentToService(assetID, serviceStart, serviceEnd)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "failed to send asset for service", err)
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "asset sent for service successfully",
	})
}

func GetAssets(w http.ResponseWriter, r *http.Request) {

	typeFilter := r.URL.Query().Get("type")
	statusFilter := r.URL.Query().Get("status")
	ownerFilter := r.URL.Query().Get("owner")
	brandFilter := r.URL.Query().Get("brand")
	modelFilter := r.URL.Query().Get("model")
	serialNumberFilter := r.URL.Query().Get("serialNumber")

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 5

	if pageStr != "" {
		pageValue, err := strconv.Atoi(pageStr)
		if err != nil || pageValue <= 0 {
			utils.RespondError(w, http.StatusBadRequest, "invalid page", err)
			return
		}
		page = pageValue
	}

	if limitStr != "" {
		limitValue, err := strconv.Atoi(limitStr)
		if err != nil || limitValue <= 0 {
			utils.RespondError(w, http.StatusBadRequest, "invalid limit", err)
			return
		}
		limit = limitValue
	}

	offset := (page - 1) * limit

	assets, err := dbhelpers.ShowAssets(typeFilter, statusFilter, ownerFilter, brandFilter, modelFilter, serialNumberFilter, limit, offset)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "failed to fetch assets", err)
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"assets": assets,
	})
}

func UpdateAsset(w http.ResponseWriter, r *http.Request) {

	assetID := chi.URLParam(r, "id")

	if assetID == "" {
		utils.RespondError(w, http.StatusBadRequest, "invalid id", fmt.Errorf("empty asset id"))
		return
	}

	var body models.UpdateAssetRequest

	err := utils.ParseBody(r, &body)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid body", err)
		return
	}

	validateErr := validate.Struct(&body)
	if validateErr != nil {
		utils.RespondError(w, http.StatusBadRequest, "failed to validate body", validateErr)
		return
	}

	warrantyStart, err := time.Parse("2006-01-02", body.WarrantyStart)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid warrantyStart", err)
		return
	}

	warrantyEnd, err := time.Parse("2006-01-02", body.WarrantyEnd)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "invalid warrantyEnd", err)
		return
	}

	if warrantyEnd.Before(warrantyStart) {
		utils.RespondError(w, http.StatusBadRequest, "invalid warranty range", fmt.Errorf("warranty end before start"))
		return
	}

	txErr := database.Tx(func(tx *sqlx.Tx) error {

		err := dbhelpers.UpdateAsset(tx, assetID, body.Brand, body.Model, body.SerialNo, body.Type, body.Owner, warrantyStart, warrantyEnd)
		if err != nil {
			return err
		}

		switch body.Type {

		case "laptop":
			if body.Laptop == nil {
				return fmt.Errorf("laptop details required")
			}
			return dbhelpers.UpdateLaptop(tx, assetID, body.Laptop)

		case "mouse":
			if body.Mouse == nil {
				return fmt.Errorf("mouse details required")
			}
			return dbhelpers.UpdateMouse(tx, assetID, body.Mouse)

		case "keyboard":
			if body.Keyboard == nil {
				return fmt.Errorf("keyboard details required")
			}
			return dbhelpers.UpdateKeyboard(tx, assetID, body.Keyboard)

		case "mobile":
			if body.Mobile == nil {
				return fmt.Errorf("mobile details required")
			}
			return dbhelpers.UpdateMobile(tx, assetID, body.Mobile)

		default:
			return fmt.Errorf("unsupported asset type")
		}
	})

	if txErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, "failed to update asset", txErr)
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "asset updated",
	})
}
