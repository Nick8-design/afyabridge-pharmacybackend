package handlers

import (
	"afyabridge-pharmacybackend/database"
	"afyabridge-pharmacybackend/model"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// GET /deliveries/
func GetDeliveries(c *fiber.Ctx) error {
	// Logic to filter deliveries based on pharmacy_id via linked orders
	var deliveries []model.Delivery
	status := c.Query("status")

	db := database.DB.Table("deliveries").
		Select("deliveries.*").
		Joins("JOIN orders ON orders.id = deliveries.order_id").
		Where("orders.pharmacy_id = ?", c.Locals("pharmacy_id"))

	if status != "" {
		db = db.Where("deliveries.status = ?", status)
	}

	db.Find(&deliveries)
	return c.JSON(model.Response{Success: true, Data: deliveries})
}

// POST /deliveries/:delivery_id/confirm
func ConfirmDelivery(c *fiber.Ctx) error {
	deliveryID := c.Params("delivery_id")
	var input struct {
		OtpCode string `json:"otp_code"`
	}
	c.BodyParser(&input)

	var delivery model.Delivery
	if err := database.DB.First(&delivery, "id = ?", deliveryID).Error; err != nil {
		return c.Status(404).JSON(model.Response{Success: false, Message: "Delivery not found"})
	}

	if delivery.OtpCode != input.OtpCode {
		return c.Status(400).JSON(model.Response{Success: false, Message: "Invalid OTP code"})
	}

	now := time.Now()
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Update Delivery
		tx.Model(&delivery).Updates(map[string]interface{}{
			"status":       "delivered",
			"delivered_at": &now,
		})
		// 2. Update Order
		tx.Model(&model.Order{}).Where("id = ?", delivery.OrderID).Update("status", "delivered")
		return nil
	})

	if err != nil {
		return c.Status(500).JSON(model.Response{Success: false, Message: "Confirmation failed"})
	}

	return c.JSON(model.Response{Success: true, Message: "Delivery confirmed successfully"})
}


// PATCH /deliveries/:delivery_id/status
func UpdateDeliveryStatus(c *fiber.Ctx) error {
	deliveryID := c.Params("delivery_id")
	var input struct {
		Status    string  `json:"status"`
		PickupLat float64 `json:"pickup_lat"`
		PickupLng float64 `json:"pickup_lng"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(model.Response{Success: false, Message: "Invalid status"})
	}

	updates := map[string]interface{}{
		"status": input.Status,
	}

	if input.PickupLat != 0 {
		updates["pickup_lat"] = input.PickupLat
		updates["pickup_lng"] = input.PickupLng
	}

	if err := database.DB.Model(&model.Delivery{}).Where("id = ?", deliveryID).Updates(updates).Error; err != nil {
		return c.Status(500).JSON(model.Response{Success: false, Message: "Update failed"})
	}

	return c.JSON(model.Response{
		Success: true, 
		Message: "Delivery status updated", 
		Data: fiber.Map{"id": deliveryID, "status": input.Status},
	})
}