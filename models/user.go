package models

import "time"

type UserExist struct {
	ID           string `db:"id"`
	PasswordHash string `db:"password_hash"`
	Role         string `db:"role"`
}

type UserRequest struct {
	Name        string `db:"name" json:"name" validate:"required"`
	Email       string `db:"email" json:"email" validate:"required,email"`
	Type        string `db:"type" json:"type" validate:"required,oneof=full_time intern freelancer"`
	PhoneNumber string `db:"phone_number" json:"phoneNumber" validate:"required,min=10"`
	Password    string `db:"password_hash" json:"password" validate:"required,min=6"`
	JoiningDate string `db:"joining_date" json:"joiningDate" validate:"required,datetime=2006-01-02"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AssignRequest struct {
	AssignedTo string `json:"assignedTo" db:"assigned_to" validate:"required"`
}

type AssignRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=admin employee project_manager asset_manager employee_manager"`
}

type UserInfoRequest struct {
	ID           string      `db:"id" json:"id"`
	Name         string      `db:"name" json:"name"`
	Email        string      `db:"email" json:"email"`
	PhoneNumber  string      `db:"phone_number" json:"phoneNumber"`
	Role         string      `db:"role" json:"role"`
	Type         string      `db:"type" json:"type"`
	CreatedAt    time.Time   `db:"created_at" json:"createdAt"`
	AssetDetails []AssetInfo `json:"assetDetails"`
}

type UserAssetRow struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Email       string    `db:"email"`
	PhoneNumber string    `db:"phone_number"`
	Role        string    `db:"role"`
	Type        string    `db:"type"`
	CreatedAt   time.Time `db:"created_at"`

	AssetID   *string `db:"asset_id"`
	Brand     *string `db:"brand"`
	Model     *string `db:"model"`
	Status    *string `db:"status"`
	AssetType *string `db:"asset_type"`
}
