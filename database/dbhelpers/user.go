package dbhelpers

import (
	"fmt"
	"time"

	"github.com/Shubhouy1/asset-management/database"
	"github.com/Shubhouy1/asset-management/models"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

func GetUserIDFromSession(sessionID string) (userID string, err error) {
	query := `SELECT user_id 
             FROM user_sessions where id=$1
             AND archived_at is null`

	err = database.Asset.Get(&userID, query, sessionID)
	if err != nil {
		return "", err
	}
	return userID, nil
}
func GetUserByEmail(tx *sqlx.Tx, email, password string) (string, string, error) {
	query := `SELECT id, password_hash,role
             FROM users 
             WHERE TRIM(LOWER(email)) = LOWER($1)
             AND archived_at is null`
	var result models.UserExist
	err := tx.Get(&result, query, email)
	if err != nil {
		return "", "", err
	}
	if err := bcrypt.CompareHashAndPassword(
		[]byte(result.PasswordHash),
		[]byte(password)); err != nil {
		return "", "", err
	}
	return result.ID, result.Role, nil
}
func IsUserExist(email string) (bool, error) {
	query := `SELECT COUNT(*)>0
             FROM users 
             WHERE TRIM(LOWER(email)) = TRIM(lower($1))
             AND archived_at is null`
	var exist bool
	err := database.Asset.Get(&exist, query, email)
	if err != nil {
		return false, err
	}
	return exist, nil
}
func CreateUser(tx *sqlx.Tx, name, email, role, userType, phoneNo, passwordHash string, joiningDate time.Time) (string, error) {
	query := `INSERT INTO users (name, email,role,type,phone_number,password_hash,joining_date)
               values ($1, trim(lower($2)), $3, $4, $5, $6, $7) RETURNING id`
	var userID string
	err := tx.Get(&userID, query, name, email, role, userType, phoneNo, passwordHash, joiningDate)
	if err != nil {
		return "", err
	}
	return userID, nil
}
func CreateUserSession(tx *sqlx.Tx, userID string) (string, error) {
	query := `INSERT INTO user_sessions (user_id)
              values ($1)RETURNING id`
	var userSessionID string
	err := tx.Get(&userSessionID, query, userID)
	if err != nil {
		return "", err
	}
	return userSessionID, nil
}

func ArchivedSession(sessionID string) error {
	query := `UPDATE user_sessions
             SET archived_at=Now() where id=$1 and archived_at is null`
	_, err := database.Asset.Exec(query, sessionID)
	if err != nil {
		return err
	}
	return nil
}

//func GetAssetInfo(userID, assetStatus string) ([]models.AssetInfo, error) {
//
//	query := `
//		SELECT id, brand, model, status, type
//		FROM assets
//		WHERE assigned_to = $1
//		AND archived_at IS NULL
//		AND ($2 = '' OR status::TEXT = $2)
//	`
//
//	assetDetails := make([]models.AssetInfo, 0)
//
//	err := database.Asset.Select(&assetDetails, query, userID, assetStatus)
//	return assetDetails, err
//}

func GetUserInfo(name, role, userType, assetStatus string) ([]models.UserInfoRequest, error) {

	query := `
	SELECT 
		u.id,
		u.name,
		u.email,
		u.phone_number,
		u.role,
		u.type,
		u.created_at,
		a.id AS asset_id,
		a.brand,
		a.model,
		a.status,
		a.type AS asset_type
	FROM users u
	LEFT JOIN assets a 
		ON a.assigned_to = u.id
		AND a.archived_at IS NULL
		AND ($4 = '' OR a.status::TEXT = $4)
	WHERE u.archived_at IS NULL
	AND ($1 = '' OR u.name ILIKE '%' || $1 || '%')
	AND ($2 = '' OR u.role::TEXT = $2)
	AND ($3 = '' OR u.type::TEXT = $3)
`

	users := make([]models.UserAssetRow, 0)

	err := database.Asset.Select(&users, query, name, role, userType, assetStatus)
	if err != nil {
		return nil, err
	}

	userMap := make(map[string]*models.UserInfoRequest)

	for _, user := range users {

		if _, exists := userMap[user.ID]; !exists {
			userMap[user.ID] = &models.UserInfoRequest{
				ID:           user.ID,
				Name:         user.Name,
				Email:        user.Email,
				PhoneNumber:  user.PhoneNumber,
				Role:         user.Role,
				Type:         user.Type,
				CreatedAt:    user.CreatedAt,
				AssetDetails: make([]models.AssetInfo, 0),
			}
		}

		if user.AssetID != nil {
			userMap[user.ID].AssetDetails = append(
				userMap[user.ID].AssetDetails,
				models.AssetInfo{
					ID:     *user.AssetID,
					Brand:  *user.Brand,
					Model:  *user.Model,
					Status: *user.Status,
					Type:   *user.AssetType,
				},
			)
		}
	}

	result := make([]models.UserInfoRequest, 0)

	for _, user := range userMap {
		if assetStatus != "" && len(user.AssetDetails) == 0 {
			continue
		}
		result = append(result, *user)
	}

	return result, nil
}
func DeleteUser(tx *sqlx.Tx, userID string) error {
	query := `Update users set archived_at = now()
             where id = $1 and archived_at is null`
	result, err := tx.Exec(query, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not found or already archived")
	}
	return nil

}
func CountActiveAssets(tx *sqlx.Tx, userID string) (int, error) {
	query := `Select count(*)
              from assets 
              where assigned_to = $1
              and archived_at IS NULL`
	var count int
	err := tx.Get(&count, query, userID)
	if err != nil {
		return 0, err
	}
	return count, nil
}
func ArchiveUserSession(tx *sqlx.Tx, userID string) error {
	query := `update user_sessions
              set archived_at = now()
              where user_id = $1
              and archived_at is null`
	result, err := tx.Exec(query, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no active session found for user")
	}
	return nil

}

func AssignRole(userID, role string) error {
	query := `UPDATE users SET role =$1
              WHERE id = $2 AND archived_at IS NULL`
	result, err := database.Asset.Exec(query, role, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not found or already archived")
	}
	return nil
}
