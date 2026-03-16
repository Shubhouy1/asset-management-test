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
	Role        string `db:"role" json:"role" validate:"required"`
	Type        string `db:"type" json:"type" validate:"required"`
	PhoneNumber string `db:"phone_no" json:"phoneNo" validate:"required,min=10"`
	Password    string `db:"password_hash" json:"password" validate:"required,min=6"`
	JoiningDate string `db:"joining_date" json:"joiningDate" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AssignRequest struct {
	AssignedTo string `json:"assignedTo" db:"assigned_to" validate:"required"`
}

type UserInfoRequest struct {
	ID           string      `db:"id" json:"id"`
	Name         string      `db:"name" json:"name"`
	Email        string      `db:"email" json:"email"`
	PhoneNo      string      `db:"phone_no" json:"phoneNo"`
	Role         string      `db:"role" json:"role"`
	Type         string      `db:"type" json:"type"`
	CreatedAt    time.Time   `db:"created_at" json:"createdAt"`
	AssetDetails []AssetInfo `json:"assetDetails"`
}
