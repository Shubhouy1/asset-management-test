package dbhelpers

import (
	"fmt"
	"time"

	"github.com/Shubhouy1/asset-management/database"
	"github.com/Shubhouy1/asset-management/models"
	"github.com/jmoiron/sqlx"
)

// make case
func MarkAsDamaged(assetID string) error {
	query := `
		UPDATE assets 
		SET 
			status = 'damaged',
			assigned_to = NULL,
			assigned_by = NULL,
			assigned_at = NULL,
			returned_at = CASE 
				WHEN status = 'assigned' THEN now()
				ELSE returned_at
			END,
			updated_at = now()
		WHERE id = $1
		AND status IN ('available', 'assigned')
		AND archived_at IS NULL
	`

	result, err := database.Asset.Exec(query, assetID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("asset not found or not eligible to mark as damaged")
	}

	return nil
}
func MarkAsWaitingForRepair(assetID string) error {
	query := `UPDATE assets SET status='waiting_for_repair',
              updated_at =now()
              WHERE id = $1
              AND archived_at IS NULL
              AND status ='damaged'`
	result, err := database.Asset.Exec(query, assetID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("asset not found or not available to mark as waiting for repair")
	}
	return nil
}
func MarkAsAvailable(assetID string) error {
	query := `
		UPDATE assets 
		SET 
			status = 'available',
			assigned_to = NULL,
			assigned_by = NULL,
			assigned_at = NULL,
			service_start = NULL,
			service_end = NULL,
			updated_at = now()
		WHERE id = $1
		AND status = 'in_service'
		AND archived_at IS NULL
	`

	result, err := database.Asset.Exec(query, assetID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("asset not found or not eligible to mark as available")
	}
	return nil
}
func ReturnAllAssets(tx *sqlx.Tx, userID string) error {

	query := `
		UPDATE assets
		SET assigned_to = NULL,
		    assigned_by = NULL,
		    assigned_at = NULL,
		    status = 'available',
		    returned_at = now(),
		    updated_at = now()
		WHERE assigned_to = $1
		AND archived_at IS NULL
	`

	_, err := tx.Exec(query, userID)
	if err != nil {
		return err
	}
	return nil

}

func ShowAssets(typeStr, statusStr, ownerStr, brandStr, modelStr, serialNumberStr string, limit, offset int) (models.DashboardData, error) {
	summaryQuery := `SELECT brand, model, type, serial_number, status, owner, created_at
			FROM assets
			WHERE archived_at IS NULL
			AND ($1 = '' OR brand ILIKE '%'||$1||'%')
			AND ($2 = '' OR model ILIKE '%'||$2||'%')
			AND ($3 = '' OR serial_number ILIKE '%'||$3||'%')
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
          COUNT(*) FILTER (WHERE status = 'waiting_for_repair') AS waiting_for_repair,
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
            set brand = $2, model = $3, 	serial_number = $4, type=$5,owner=$6,warranty_start = $7,warranty_end=$8, updated_at =now()
            where id= $1 and archived_at is null `
	_, err := tx.Exec(query, assetID, brand, model, serialNo, assetType, owner, warrantyStart, warrantyEnd)
	if err != nil {
		return err
	}
	return nil

}
func UpdateLaptop(tx *sqlx.Tx, assetID string, laptop *models.LaptopInput) error {
	query := `
	UPDATE laptops
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
	UPDATE mice
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
	UPDATE keyboards
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
func CreateAsset(tx *sqlx.Tx, brand, model, serialNo, assetType, owner string, warrantyStart, warrantyEnd time.Time) (string, error) {
	query := `INSERT INTO assets (brand, model,serial_number,type, owner,warranty_start, warranty_end)
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
             assigned_by= $3,
             assigned_at = now(),
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
		INSERT INTO laptops (
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
		INSERT INTO mice(
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
		INSERT INTO keyboards (
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
	query1 := `SELECT brand,model,serial_number,type,status,owner,created_at
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
              WHERE id=$1 AND archived_at is null and status ='waiting_for_repair'`
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
