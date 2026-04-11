package handlers

import (
    "afyabridge-pharmacybackend/database"
    "afyabridge-pharmacybackend/model"
    "time"
    "github.com/gofiber/fiber/v2"
)

// GetDoctorNotifications - Fetch all notifications for the logged-in doctor
func GetDoctorNotifications(c *fiber.Ctx) error {
    db := database.DB
    doctorID := c.Locals("doctorID").(string)

    var notifications []model.Notification
    err := db.Where("user_id = ?", doctorID).
        Order("created_at DESC").
        Limit(50).
        Find(&notifications).Error

    if err != nil {
        return c.Status(500).JSON(fiber.Map{"success": false, "message": "Could not fetch notifications"})
    }

    return c.JSON(fiber.Map{
        "success": true,
        "data":    notifications,
    })
}

// CreateNotification - Allows doctor to send a notification to a patient or staff
func CreateNotification(c *fiber.Ctx) error {
    db := database.DB
    
    var req struct {
        TargetUserID     string  `json:"target_user_id"` // Who receives it
        Title            string  `json:"title"`
        Message          string  `json:"message"`
        NotificationType string  `json:"type"`          // e.g., 'prescription', 'appointment'
        ReferenceID      *string `json:"reference_id"`
    }

    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid request"})
    }

    newNotification := model.Notification{
        UserID:           req.TargetUserID,
        Title:            req.Title,
        Message:          req.Message,
        NotificationType: req.NotificationType,
        ReferenceID:      req.ReferenceID,
    }

    if err := db.Create(&newNotification).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to send notification"})
    }

    return c.JSON(fiber.Map{
        "success": true, 
        "message": "Notification sent successfully",
        "id":      newNotification.ID,
    })
}

// MarkNotificationAsRead - Simple toggle for the UI
func MarkNotificationAsRead(c *fiber.Ctx) error {
    db := database.DB
    id := c.Params("id")
    doctorID := c.Locals("doctorID").(string)

    now := time.Now()
    err := db.Model(&model.Notification{}).
        Where("id = ? AND user_id = ?", id, doctorID).
        Updates(map[string]interface{}{
            "is_read": true,
            "read_at": &now,
        }).Error

    if err != nil {
        return c.Status(500).JSON(fiber.Map{"success": false, "message": "Update failed"})
    }

    return c.JSON(fiber.Map{"success": true})
}