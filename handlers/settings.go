package handlers

import (
	"afyabridge-pharmacybackend/database"
	"afyabridge-pharmacybackend/model"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GET /settings/pharmacy/
func GetPharmacySettings(c *fiber.Ctx) error {
	pharmacyID := c.Locals("pharmacy_id").(string)
	var pharmacy model.Pharmacy

	if err := database.DB.First(&pharmacy, "id = ?", pharmacyID).Error; err != nil {
		return c.Status(404).JSON(model.Response{Success: false, Message: "Pharmacy not found"})
	}

	return c.JSON(model.Response{Success: true, Data: pharmacy})
}

// PUT /settings/pharmacy/
func UpdatePharmacySettings(c *fiber.Ctx) error {
	pharmacyID := c.Locals("pharmacy_id").(string)
	var pharmacy model.Pharmacy

	if err := database.DB.First(&pharmacy, "id = ?", pharmacyID).Error; err != nil {
		return c.Status(404).JSON(model.Response{Success: false, Message: "Pharmacy not found"})
	}

	if err := c.BodyParser(&pharmacy); err != nil {
		return c.Status(400).JSON(model.Response{Success: false, Message: "Invalid input"})
	}

	database.DB.Save(&pharmacy)
	return c.JSON(model.Response{Success: true, Message: "Pharmacy settings updated", Data: pharmacy})
}

// GET /settings/pharmacy/hours/
func GetPharmacyHours(c *fiber.Ctx) error {
	pharmacyID := c.Locals("pharmacy_id").(string)
	var hours []model.PharmacyHour

	database.DB.Where("pharmacy_id = ?", pharmacyID).Find(&hours)
	return c.JSON(model.Response{Success: true, Data: hours})
}

// PUT /settings/pharmacy/hours/
func UpdatePharmacyHours(c *fiber.Ctx) error {
	pharmacyID := c.Locals("pharmacy_id").(string)
	var input struct {
		Hours []model.PharmacyHour `json:"hours"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(model.Response{Success: false, Message: "Invalid data"})
	}

	// Batch update or replace hours
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// Remove existing to replace with new set
		tx.Where("pharmacy_id = ?", pharmacyID).Delete(&model.PharmacyHour{})
		for i := range input.Hours {
			input.Hours[i].ID = uuid.New().String()
			input.Hours[i].PharmacyID = pharmacyID
		}
		return tx.Create(&input.Hours).Error
	})

	if err != nil {
		return c.Status(500).JSON(model.Response{Success: false, Message: "Failed to update hours"})
	}

	return c.JSON(model.Response{Success: true, Message: "Operating hours updated", Data: input.Hours})
}

// PATCH /settings/pharmacy/logo/
// This version accepts a JSON body with a "logo_url" string instead of a file.
func UploadPharmacyLogo(c *fiber.Ctx) error {
	pharmacyID := c.Locals("pharmacy_id").(string)
	
	var input struct {
		LogoURL string `json:"logo_url"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(model.Response{
			Success: false, 
			Message: "Invalid input. Please provide a valid logo_url",
		})
	}

	if input.LogoURL == "" {
		return c.Status(400).JSON(model.Response{
			Success: false, 
			Message: "logo_url cannot be empty",
		})
	}

	var pharmacy model.Pharmacy
	if err := database.DB.First(&pharmacy, "id = ?", pharmacyID).Error; err != nil {
		return c.Status(404).JSON(model.Response{Success: false, Message: "Pharmacy not found"})
	}

	// Update the logo field (maps to the *string in your model)
	if err := database.DB.Model(&pharmacy).Update("logo", input.LogoURL).Error; err != nil {
		return c.Status(500).JSON(model.Response{Success: false, Message: "Failed to update logo link"})
	}

	return c.JSON(model.Response{
		Success: true,
		Message: "Logo updated successfully",
		Data:    pharmacy,
	})
}