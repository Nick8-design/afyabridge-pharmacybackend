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
	pharmacyID, ok := c.Locals("pharmacy_id").(string)
	if !ok {
		return c.Status(400).JSON(model.Response{
			Success: false,
			Message: "Your account is not linked to a pharmacy",
		})
	}

	var input struct {
		Name          *string   `json:"name"`
		Phone         *string   `json:"phone"`
		Email         *string   `json:"email"`
		DeliveryZones *[]string `json:"delivery_zones"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(model.Response{Success: false, Message: "Invalid input"})
	}

	updates := map[string]interface{}{}

	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Phone != nil {
		updates["phone"] = *input.Phone
	}
	if input.Email != nil {
		updates["email"] = *input.Email
	}
	if input.DeliveryZones != nil {
		updates["delivery_zones"] = input.DeliveryZones
	}

	if len(updates) == 0 {
		return c.Status(400).JSON(model.Response{Success: false, Message: "No fields to update"})
	}

	if err := database.DB.Model(&model.Pharmacy{}).
		Where("id = ?", pharmacyID).
		Updates(updates).Error; err != nil {

		return c.Status(500).JSON(model.Response{Success: false, Message: "Update failed"})
	}

	var updated model.Pharmacy
	database.DB.First(&updated, "id = ?", pharmacyID)

	return c.JSON(model.Response{
		Success: true,
		Message: "Pharmacy settings updated",
		Data:    updated,
	})
}

func GetPharmacyHours(c *fiber.Ctx) error {
	pharmacyID := c.Locals("pharmacy_id").(string)

	var hours []model.PharmacyHour
	database.DB.Where("pharmacy_id = ?", pharmacyID).Find(&hours)

	// Ensure all 7 days exist
	days := []string{"MON","TUE","WED","THU","FRI","SAT","SUN"}

	result := []model.PharmacyHour{}

	for _, day := range days {
		found := false
		for _, h := range hours {
			if h.DayOfWeek == day {
				result = append(result, h)
				found = true
				break
			}
		}

		if !found {
			result = append(result, model.PharmacyHour{
				DayOfWeek: day,
				IsClosed:  true,
			})
		}
	}

	return c.JSON(model.Response{Success: true, Data: result})
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