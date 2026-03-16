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
             FROM user_session where id=$1
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
	query := `INSERT INTO users (name, email,role,type,phone_no,password_hash,joining_date)
               values ($1, trim(lower($2)), $3, $4, $5, $6, $7) RETURNING id`
	var userID string
	err := tx.Get(&userID, query, name, email, role, userType, phoneNo, passwordHash, joiningDate)
	if err != nil {
		return "", err
	}
	return userID, nil
}
func CreateUserSession(tx *sqlx.Tx, userID string) (string, error) {
	query := `INSERT INTO user_session (user_id)
              values ($1)RETURNING id`
	var userSessionID string
	err := tx.Get(&userSessionID, query, userID)
	if err != nil {
		return "", err
	}
	return userSessionID, nil
}

func ArchivedSession(sessionID string) error {
	query := `UPDATE user_session 
             SET archived_at=Now() where id=$1 and archived_at is null`
	_, err := database.Asset.Exec(query, sessionID)
	if err != nil {
		return err
	}
	return nil
}
func CreateAsset(tx *sqlx.Tx, brand, model, serialNo, assetType, owner string, warrantyStart, warrantyEnd time.Time) (string, error) {
	query := `INSERT INTO assets (brand, model, serial_no,type, owner,warranty_start, warranty_end)
	          Values($1, $2, $3, $4, $5, $6, $7) returning id`
	var assetID string
	err := tx.Get(&assetID, query, brand, model, serialNo, assetType, owner, warrantyStart, warrantyEnd)
	if err != nil {
		return "", err
	}
	return assetID, nil

}
func AssignAsset(assetID, assignedTo, assignedBy string) error {
	query := `UPDATE assets SET
             status ='assigned',
             assigned_to = $2 ,
             assigned_by_id = $3,
             assigned_on = now(),
             updated_at = now()
             where id = $1
             and archived_at is null
             and status = 'available'`
	result, err := database.Asset.Exec(query, assetID, assignedTo, assignedBy)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("asset not found or not available for assignment")
	}
	return nil
}
func InsertLaptop(tx *sqlx.Tx, assetID string, laptop *models.LaptopInput) error {
	query := `
		INSERT INTO laptop (
			asset_id,
			processor,
			ram,
			storage,
			os,
			charger,
			password
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := tx.Exec(
		query,
		assetID,
		laptop.Processor,
		laptop.RAM,
		laptop.Storage,
		laptop.OS,
		laptop.Charger,
		laptop.Password,
	)

	return err
}

func InsertMouse(tx *sqlx.Tx, assetID string, mouse *models.MouseInput) error {
	query := `
		INSERT INTO mouse (
			asset_id,
			dpi,
			connectivity
		)
		VALUES ($1, $2, $3)
	`

	_, err := tx.Exec(
		query,
		assetID,
		mouse.Dpi,
		mouse.Connectivity,
	)

	return err
}

func InsertKeyboard(tx *sqlx.Tx, assetID string, keyboard *models.KeyboardInput) error {
	query := `
		INSERT INTO keyboard (
			asset_id,
			layout,
			connectivity
		)
		VALUES ($1, $2, $3)
	`

	_, err := tx.Exec(
		query,
		assetID,
		keyboard.Layout,
		keyboard.Connectivity,
	)

	return err
}

func InsertMobile(tx *sqlx.Tx, assetID string, mobile *models.MobileInput) error {
	query := `
		INSERT INTO mobile (
			asset_id,
			os,
			ram,
			storage,
			charger,
			password
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := tx.Exec(
		query,
		assetID,
		mobile.OS,
		mobile.RAM,
		mobile.Storage,
		mobile.Charger,
		mobile.Password,
	)

	return err
}
func FindTotalAssetById(userID string) (models.DashboardUserData, error) {
	query := `SELECT COUNT(*) as active_asset from assets 
                where assigned_to=$1 and status='assigned'
                AND archived_at is null`
	var dummy models.DashboardUserData
	var summary models.DashboardUserSummary
	err := database.Asset.Get(&summary, query, userID)
	if err != nil {
		return dummy, err
	}
	assetInfo := make([]models.Asset, 0)
	query1 := `SELECT brand,model,serial_no,type,status,owner,created_at
               from assets 
               where assigned_to=$1 and archived_at is null`
	err = database.Asset.Select(&assetInfo, query1, userID)
	if err != nil {
		return dummy, err
	}
	return models.DashboardUserData{
		Summary: summary,
		Assets:  assetInfo,
	}, nil
}
func SentToService(assetID string, serviceStart, serviceEnd time.Time) error {
	query := `UPDATE assets SET status='in_service',service_start=$2,service_end=$3,updated_at=now()
              WHERE id=$1 AND archived_at is null and status ='available'`
	result, err := database.Asset.Exec(query, assetID, serviceStart, serviceEnd)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("asset not available for service")
	}

	return nil
}
func ShowAssets(typeStr, statusStr, ownerStr, brandStr, modelStr, serialNumberStr string, limit, offset int) (models.DashboardData, error) {
	summaryQuery := `SELECT brand, model, type, serial_no, status, owner, created_at
			FROM assets
			WHERE archived_at IS NULL
			AND ($1 = '' OR brand ILIKE '%'||$1||'%')
			AND ($2 = '' OR model ILIKE '%'||$2||'%')
			AND ($3 = '' OR serial_no ILIKE '%'||$3||'%')
			AND ($4 = '' OR type::text ILIKE '%'||$4||'%')
			AND ($5 = '' OR status::text ILIKE '%'||$5||'%')
			AND ($6 = '' OR owner::text ILIKE '%'||$6||'%')
			ORDER BY created_at DESC
			LIMIT $7 OFFSET $8
          `

	assets := make([]models.Asset, 0)
	var summary models.DashboardSummary

	Sql := `SELECT
          COUNT(*) AS total,
          COUNT(*) FILTER (WHERE status = 'available') AS available,
          COUNT(*) FILTER (WHERE status = 'assigned') AS assigned,
          COUNT(*) FILTER (WHERE status = 'for_repair') AS waiting_for_repair,
          COUNT(*) FILTER (WHERE status = 'in_service') AS in_service,
          COUNT(*) FILTER (WHERE status = 'damaged') AS damaged
       FROM assets
       WHERE archived_at IS NULL`

	var res models.DashboardData
	DashboardErr := database.Asset.Get(&summary, Sql)
	if DashboardErr != nil {
		return res, DashboardErr
	}

	err := database.Asset.Select(&assets, summaryQuery, brandStr, modelStr, serialNumberStr, typeStr, statusStr, ownerStr, limit, offset)
	if err != nil {
		return res, err
	}
	return models.DashboardData{
		Summary: summary,
		Assets:  assets,
	}, nil
}
func UpdateAsset(tx *sqlx.Tx, assetID, brand, model, serialNo, assetType, owner string, warrantyStart, warrantyEnd time.Time) error {
	query := `UPDATE assets
            set brand = $2, model = $3, serial_no = $4, type=$5,owner=$6,warranty_start = $7,warranty_end=$8, updated_at =now()
            where id= $1 and archived_at is null `
	_, err := tx.Exec(query, assetID, brand, model, serialNo, assetType, owner, warrantyStart, warrantyEnd)
	if err != nil {
		return err
	}
	return nil

}
func UpdateLaptop(tx *sqlx.Tx, assetID string, laptop *models.LaptopInput) error {
	query := `
	UPDATE laptop
	SET
	    processor = $2,
	    ram = $3,
	    storage = $4,
	    os = $5,
	    charger = $6,
	    password = $7
	WHERE asset_id = $1
	`

	result, err := tx.Exec(query,
		assetID,
		laptop.Processor,
		laptop.RAM,
		laptop.Storage,
		laptop.OS,
		laptop.Charger,
		laptop.Password,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("laptop asset not found")
	}
	return nil
}
func UpdateMouse(tx *sqlx.Tx, assetID string, mouse *models.MouseInput) error {
	query := `
	UPDATE mouse
	SET
	    dpi = $2,
	    connectivity = $3
	WHERE asset_id = $1
	`

	result, err := tx.Exec(query, assetID, mouse.Dpi, mouse.Connectivity)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("mouse asset not found")
	}
	return nil
}
func UpdateKeyboard(tx *sqlx.Tx, assetID string, keyboard *models.KeyboardInput) error {
	query := `
	UPDATE keyboard
	SET
	    layout = $2,
	    connectivity = $3
	WHERE asset_id = $1
	`

	result, err := tx.Exec(query, assetID, keyboard.Layout, keyboard.Connectivity)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("keyboard asset not found")
	}
	return nil
}
func UpdateMobile(tx *sqlx.Tx, assetID string, mobile *models.MobileInput) error {
	query := `
	UPDATE mobile
	SET
	    os = $2,
	    ram = $3,
	    storage = $4,
	    charger = $5,
	    password = $6
	WHERE asset_id = $1
	`

	result, err := tx.Exec(
		query,
		assetID,
		mobile.OS,
		mobile.RAM,
		mobile.Storage,
		mobile.Charger,
		mobile.Password,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("mobile asset not found")
	}

	return nil
}
func GetAssetInfo(userID, assetStatus string) ([]models.AssetInfo, error) {

	query := `
		SELECT id, brand, model, status, type
		FROM assets
		WHERE assigned_to = $1
		AND archived_at IS NULL
		AND ($2 = '' OR status::TEXT = $2)
	`

	assetDetails := make([]models.AssetInfo, 0)

	err := database.Asset.Select(&assetDetails, query, userID, assetStatus)
	return assetDetails, err
}

func GetUserInfo(name, role, userType, assetStatus string) ([]models.UserInfoRequest, error) {

	query := `
		SELECT id, name, email, phone_no, role, type, created_at
		FROM users
		WHERE archived_at IS NULL
		AND ($1 = '' OR name ILIKE '%' || $1 || '%')
		AND ($2 = '' OR role::TEXT = $2)
		AND ($3 = '' OR type::TEXT = $3)
	`

	users := make([]models.UserInfoRequest, 0)

	err := database.Asset.Select(&users, query, name, role, userType)
	if err != nil {
		return users, err
	}

	filteredUsers := make([]models.UserInfoRequest, 0)

	for _, user := range users {

		assetDetails, err := GetAssetInfo(user.ID, assetStatus)
		if err != nil {
			return users, err
		}

		if assetStatus != "" && len(assetDetails) == 0 {
			continue
		}

		user.AssetDetails = assetDetails
		filteredUsers = append(filteredUsers, user)
	}

	return filteredUsers, nil
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
	query := `update user_session
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
func ReturnAllAssets(tx *sqlx.Tx, userID string) error {

	query := `
		UPDATE assets
		SET assigned_to = NULL,
		    assigned_by_id = NULL,
		    assigned_on = NULL,
		    status = 'available',
		    returned_on = now(),
		    updated_at = now()
		WHERE assigned_to = $1
		AND archived_at IS NULL
	`

	result, err := tx.Exec(query, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no assets assigned to user")
	}
	return nil

}
