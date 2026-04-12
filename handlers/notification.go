package handlers

import (
	"afyabridge-pharmacybackend/database"
	"afyabridge-pharmacybackend/model"
	"time"

	"github.com/gofiber/fiber/v2"
)






// GetNotifications - Fetch notifications for the logged-in user and their pharmacy
func GetNotifications(c *fiber.Ctx) error {
	db := database.DB
	
	// 1. Get Personal User ID from middleware
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(401).JSON(model.Response{
			Success: false,
			Message: "User context missing",
		})
	}

	// 2. Get Pharmacy ID from middleware (set during login/auth)
	// If you don't have this in locals, you'd fetch it from the user model
	pharmacyID, _ := c.Locals("pharmacy_id").(string)

	var notifications []model.Notification

	// 3. Query for notifications belonging to the individual OR the pharmacy
	query := db.Where("user_id = ?", userID)
	
	if pharmacyID != "" {
		// Use an OR condition to include pharmacy-wide alerts
		query = query.Or("user_id = ?", pharmacyID)
	}

	err := query.Order("created_at DESC").
		Limit(100).
		Find(&notifications).Error

	if err != nil {
		return c.Status(500).JSON(model.Response{
			Success: false,
			Message: "Could not fetch notifications",
		})
	}

	return c.JSON(model.Response{
		Success: true,
		Message: "Notifications retrieved successfully",
		Data:    notifications,
	})
}

// CreateNotification - Enhanced to handle pharmacy targets
func CreateNotification(c *fiber.Ctx) error {
	db := database.DB

	var req struct {
		TargetID         string  `json:"target_id"` // Can be a User ID or a Pharmacy ID
		Title            string  `json:"title"`
		Message          string  `json:"message"`
		NotificationType string  `json:"type"` 
		ReferenceID      *string `json:"reference_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(model.Response{
			Success: false,
			Message: "Invalid request body",
		})
	}

	newNotification := model.Notification{
		UserID:           req.TargetID, // Maps to the business ID or personal ID
		Title:            req.Title,
		Message:          req.Message,
		NotificationType: req.NotificationType,
		ReferenceID:      req.ReferenceID,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := db.Create(&newNotification).Error; err != nil {
		return c.Status(500).JSON(model.Response{
			Success: false,
			Message: "Failed to send notification",
		})
	}

	return c.JSON(model.Response{
		Success: true,
		Message: "Notification sent successfully",
		Data:    newNotification,
	})
}



// // GetDoctorNotifications - Fetch all notifications for the logged-in doctor
// func GetDoctorNotifications(c *fiber.Ctx) error {
// 	db := database.DB
// 	doctorID, ok := c.Locals("user_id").(string)
// 	if !ok {
// 		return c.Status(401).JSON(model.Response{
// 			Success: false,
// 			Message: "User context missing",
// 		})
// 	}

// 	var notifications []model.Notification
// 	err := db.Where("user_id = ?", doctorID).
// 		Order("created_at DESC").
// 		Limit(50).
// 		Find(&notifications).Error

// 	if err != nil {
// 		return c.Status(500).JSON(model.Response{
// 			Success: false,
// 			Message: "Could not fetch notifications",
// 		})
// 	}

// 	return c.JSON(model.Response{
// 		Success: true,
// 		Message: "Notifications retrieved successfully",
// 		Data:    notifications,
// 	})
// }



// MarkNotificationAsRead - Simple toggle for the UI
func MarkNotificationAsRead(c *fiber.Ctx) error {
	db := database.DB
	id := c.Params("id")
	doctorID, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(401).JSON(model.Response{
			Success: false,
			Message: "User context missing",
		})
	}

	now := time.Now()
	result := db.Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", id, doctorID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		})

	if result.Error != nil {
		return c.Status(500).JSON(model.Response{
			Success: false,
			Message: "Database error while updating notification",
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(404).JSON(model.Response{
			Success: false,
			Message: "Notification not found or access denied",
		})
	}

	return c.JSON(model.Response{
		Success: true,
		Message: "Notification marked as read",
	})
}