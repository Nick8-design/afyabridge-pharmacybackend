package handlers

import (
	"afyabridge-pharmacybackend/database"
	"afyabridge-pharmacybackend/model"
	"time"

	"github.com/gofiber/fiber/v2"

)

// GET /deliveries/
// GET /deliveries/
func GetDeliveries(c *fiber.Ctx) error {
	pharmacyID, ok := c.Locals("pharmacy_id").(string)
	if !ok {
		return c.Status(401).JSON(model.Response{Success: false, Message: "Unauthorized"})
	}

	status := c.Query("status")

	type Result struct {
		ID              string     `json:"id"`
		PackageNumber   string     `json:"package_number"`
		OrderID         string     `json:"order_id"`
		RiderID         *string    `json:"rider_id"`
		RiderName       *string    `json:"rider_name"`
		Status          string     `json:"status"`
		PickupLocation  string     `json:"pickup_location"`
		DropoffLocation string     `json:"dropoff_location"`
		OtpCode         string     `json:"otp_code"`
		Charges         float64    `json:"charges"`
		PickupLat       float64    `json:"pickup_lat"`
		PickupLng       float64    `json:"pickup_lng"`
		CreatedAt       time.Time  `json:"created_at"`
		DeliveredAt     *time.Time `json:"delivered_at"`
	}

	var deliveries []Result

	db := database.DB.Table("deliveries").
		Select(`
			deliveries.id,
			deliveries.package_number,
			deliveries.order_id,
			deliveries.rider_id,
			users.name as rider_name,
			deliveries.status,
			deliveries.pickup_location,
			deliveries.dropoff_location,
			deliveries.otp_code,
			deliveries.charges,
			deliveries.pickup_lat,
			deliveries.pickup_lng,
			deliveries.created_at,
			deliveries.delivered_at
		`).
		Joins("JOIN orders ON orders.id = deliveries.order_id").
		Joins("LEFT JOIN users ON users.id = deliveries.rider_id").
		Where("orders.pharmacy_id = ?", pharmacyID)

	if status != "" {
		db = db.Where("deliveries.status = ?", status)
	}

	db.Scan(&deliveries)

	return c.JSON(model.Response{
		Success: true,
		Data:    deliveries,
	})
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
		return c.Status(400).JSON(model.Response{Success: false, Message: "Invalid input"})
	}

	validStatuses := map[string]bool{
		"assigned": true,
		"picked_up": true,
		"in_transit": true,
		"delivered": true,
		"failed": true,
	}

	if !validStatuses[input.Status] {
		return c.Status(400).JSON(model.Response{Success: false, Message: "Invalid status value"})
	}

	updates := map[string]interface{}{
		"status": input.Status,
	}

	if input.PickupLat != 0 && input.PickupLng != 0 {
		updates["pickup_lat"] = input.PickupLat
		updates["pickup_lng"] = input.PickupLng
	}

	if err := database.DB.Model(&model.Delivery{}).
		Where("id = ?", deliveryID).
		Updates(updates).Error; err != nil {

		return c.Status(500).JSON(model.Response{Success: false, Message: "Update failed"})
	}

	return c.JSON(model.Response{
		Success: true,
		Message: "Delivery status updated",
		Data: fiber.Map{
			"id":     deliveryID,
			"status": input.Status,
		},
	})
}