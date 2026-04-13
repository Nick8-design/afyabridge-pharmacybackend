package handlers

import (
	"afyabridge-pharmacybackend/database"
	"afyabridge-pharmacybackend/model"
	"time"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GET /inventory/expiring
func GetExpiringDrugs(c *fiber.Ctx) error {
	pharmacyID := c.Locals("pharmacy_id").(string)
	days := c.QueryInt("days", 30)

	var results []fiber.Map

	err := database.DB.Table("stock_batches").
		Select("drugs.id, drugs.drug_name, stock_batches.batch_number, stock_batches.expiry_date, stock_batches.quantity_remaining").
		Joins("JOIN drugs ON drugs.id = stock_batches.drug_id").
		Where("drugs.pharmacy_id = ? AND stock_batches.expiry_date <= ?", pharmacyID, time.Now().AddDate(0, 0, days)).
		Scan(&results).Error

	if err != nil {
		return c.Status(500).JSON(model.Response{Success: false, Message: "Error fetching expiring drugs"})
	}

	return c.JSON(model.Response{Success: true, Data: results})
}

// GET /inventory/:drug_id
func GetDrugDetails(c *fiber.Ctx) error {
	drugID := c.Params("drug_id")
	var drug model.Inventory

	if err := database.DB.First(&drug, "id = ?", drugID).Error; err != nil {
		return c.Status(404).JSON(model.Response{Success: false, Message: "Drug not found"})
	}

	return c.JSON(model.Response{Success: true, Data: drug})
}

// PUT /inventory/:drug_id
func UpdateDrug(c *fiber.Ctx) error {
	drugID := c.Params("drug_id")
	var drug model.Inventory
	if err := database.DB.First(&drug, "id = ?", drugID).Error; err != nil {
		return c.Status(404).JSON(model.Response{Success: false, Message: "Drug not found"})
	}

	if err := c.BodyParser(&drug); err != nil {
		return c.Status(400).JSON(model.Response{Success: false, Message: "Invalid input"})
	}

	database.DB.Save(&drug)
	return c.JSON(model.Response{Success: true, Message: "Drug updated", Data: drug})
}

// DELETE /inventory/:drug_id
func DeleteDrug(c *fiber.Ctx) error {
	drugID := c.Params("drug_id")
	// Soft delete by setting is_active to false
	if err := database.DB.Model(&model.Inventory{}).Where("id = ?", drugID).Update("is_active", false).Error; err != nil {
		return c.Status(500).JSON(model.Response{Success: false, Message: "Failed to remove drug"})
	}

	return c.JSON(model.Response{Success: true, Message: "Drug removed from active inventory"})
}

// GET /inventory/low-stock
func GetLowStock(c *fiber.Ctx) error {
	pharmacyID := c.Locals("pharmacy_id").(string)
	var drugs []model.Inventory
	
	database.DB.Where("pharmacy_id = ? AND quantity_in_stock <= reorder_level AND is_active = ?", pharmacyID, true).
		Find(&drugs)

	return c.JSON(model.Response{Success: true, Data: drugs})
}





// GET /inventory/
func GetInventory(c *fiber.Ctx) error {
	pharmacyID := c.Locals("pharmacy_id").(string)
	category := c.Query("category")
	search := c.Query("q")

	var drugs []model.Inventory
	db := database.DB.Where("pharmacy_id = ? AND is_active = ?", pharmacyID, true)

	if category != "" {
		db = db.Where("category = ?", category)
	}
	if search != "" {
		db = db.Where("drug_name LIKE ?", "%"+search+"%")
	}

	db.Find(&drugs)

	return c.JSON(model.Response{
		Success: true,
		Data: fiber.Map{
			"count":   len(drugs),
			"results": drugs,
		},
	})
}

// POST /inventory/
func AddDrug(c *fiber.Ctx) error {
	pharmacyID := c.Locals("pharmacy_id").(string)
	var drug model.Inventory
	if err := c.BodyParser(&drug); err != nil {
		return c.Status(400).JSON(model.Response{Success: false, Message: "Invalid input"})
	}

	drug.ID = uuid.New().String()
	drug.PharmacyID = pharmacyID
	drug.IsActive = true

	if err := database.DB.Create(&drug).Error; err != nil {
		return c.Status(500).JSON(model.Response{Success: false, Message: "Could not add drug"})
	}

	return c.Status(201).JSON(model.Response{Success: true, Message: "Drug added to inventory", Data: drug})
}

// POST /inventory/:drug_id/restock
func RestockDrug(c *fiber.Ctx) error {
	drugID := c.Params("drug_id")
	var input struct {
		Quantity   int    `json:"quantity"`
		BatchNo    string `json:"batch_no"`
		ExpiryDate string `json:"expiry_date"`
		SupplierID string `json:"supplier_id"`
	}
	c.BodyParser(&input)

	expiry, _ := time.Parse("2006-01-02", input.ExpiryDate)

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Create Batch record
		batch := model.StockBatch{
			ID:                uuid.New().String(),
			DrugID:            drugID,
			BatchNumber:       input.BatchNo,
			QuantityReceived:  input.Quantity,
			QuantityRemaining: input.Quantity,
			ExpiryDate:        expiry,
		}
		if err := tx.Create(&batch).Error; err != nil {
			return err
		}

		// 2. Increment main inventory quantity
		return tx.Model(&model.Inventory{}).Where("id = ?", drugID).
			UpdateColumn("quantity_in_stock", gorm.Expr("quantity_in_stock + ?", input.Quantity)).Error
	})

	if err != nil {
		return c.Status(500).JSON(model.Response{Success: false, Message: "Restock failed"})
	}
	return c.JSON(model.Response{
		Success: true,
		Message: "Stock updated",
		Data: fiber.Map{
			"drug_id": drugID,
			"batch_no": input.BatchNo,
			"expiry_date": input.ExpiryDate,
			"added_quantity": input.Quantity,
		},
	})
}

// GET /inventory/dashboard/
func GetInventoryDashboard(c *fiber.Ctx) error {
	pharmacyID := c.Locals("pharmacy_id").(string)
	db := database.DB.Where("pharmacy_id = ? AND is_active = ?", pharmacyID, true)

	var stats struct {
		TotalSkus     int64 `json:"total_skus"`
		LowStockCount int64 `json:"low_stock_count"`
		CriticalCount int64 `json:"critical_count"`
		ExpiringCount int64 `json:"expiring_count"`
	}

	db.Model(&model.Inventory{}).Count(&stats.TotalSkus)
	db.Model(&model.Inventory{}).Where("quantity_in_stock <= reorder_level").Count(&stats.LowStockCount)
	db.Model(&model.Inventory{}).Where("quantity_in_stock <= critical_level").Count(&stats.CriticalCount)

	// Expiring within 30 days
	database.DB.Model(&model.StockBatch{}).
		Joins("JOIN inventories ON stock_batches.drug_id = inventories.id").
		Where("inventories.pharmacy_id = ? AND stock_batches.expiry_date <= ?", pharmacyID, time.Now().AddDate(0, 0, 30)).
		Count(&stats.ExpiringCount)

	return c.JSON(model.Response{Success: true, Data: stats})
}