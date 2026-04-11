package handlers

import (
	"afyabridge-pharmacybackend/database"
	"afyabridge-pharmacybackend/model"
	"fmt"
	"math/rand"
    "github.com/google/uuid"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"


	"time"

)

// POST /orders/:order_id/cancel
func CancelOrder(c *fiber.Ctx) error {
	orderID := c.Params("order_id")
	var input struct {
		Reason string `json:"reason"`
	}
	c.BodyParser(&input)

	db := database.DB
	var order model.Order
	if err := db.First(&order, "id = ?", orderID).Error; err != nil {
		return c.Status(404).JSON(model.Response{Success: false, Message: "Order not found"})
	}

	// Business Rule: Cannot cancel delivered or dispatched orders
	if order.Status == "delivered" || order.Status == "dispatched" {
		return c.Status(400).JSON(model.Response{
			Success: false,
			Message: "Cannot cancel an order that has already been dispatched or delivered",
		})
	}

	if err := db.Model(&order).Update("status", "cancelled").Error; err != nil {
		return c.Status(500).JSON(model.Response{Success: false, Message: "Failed to cancel order"})
	}

	return c.JSON(model.Response{
		Success: true,
		Message: "Order cancelled",
		Data:    fiber.Map{"id": order.ID, "status": "cancelled"},
	})
}

// GET /orders/patient/:patient_id/history
func GetPatientOrderHistory(c *fiber.Ctx) error {
	patientID := c.Params("patient_id")
	pharmacyID := c.Locals("pharmacy_id").(string)

	var orders []model.Order
	// Fetching orders that reached terminal success status
	err := database.DB.Preload("Prescription").
		Where("patient_id = ? AND pharmacy_id = ? AND status IN ('delivered', 'dispatched')", patientID, pharmacyID).
		Order("created_at desc").
		Find(&orders).Error

	if err != nil {
		return c.Status(500).JSON(model.Response{Success: false, Message: "Database error"})
	}

	return c.JSON(model.Response{
		Success: true,
		Data:    orders,
	})
}





// GET /orders/ - List all orders with items fetched from Prescriptions
func GetOrders(c *fiber.Ctx) error {
	pharmacyID := c.Locals("pharmacy_id").(string)
	status := c.Query("status")
	priority := c.Query("priority")
	search := c.Query("q")

	var orders []model.Order
	// Preload the Prescription to get the drug items
	db := database.DB.Preload("Prescription").Where("pharmacy_id = ?", pharmacyID)

	if status != "" {
		db = db.Where("status = ?", status)
	}
	if priority != "" {
		db = db.Where("priority = ?", priority)
	}
	if search != "" {
		db = db.Where("patient_name LIKE ?", "%"+search+"%")
	}

	db.Order("created_at desc").Find(&orders)

	return c.JSON(model.Response{
		Success: true,
		Data: fiber.Map{
			"count":   len(orders),
			"results": orders,
		},
	})
}

// GET /orders/:order_id - Single order details
func GetOrderDetails(c *fiber.Ctx) error {
	orderID := c.Params("order_id")
	var order model.Order

	if err := database.DB.First(&order, "id = ?", orderID).Error; err != nil {
		return c.Status(404).JSON(model.Response{Success: false, Message: "Order not found"})
	}

	return c.JSON(model.Response{
		Success: true,
		Data:    order,
	})
}

// POST /orders/:order_id/dispense
func DispenseOrder(c *fiber.Ctx) error {
	orderID := c.Params("order_id")
	db := database.DB

	var order model.Order
	// Preload Prescription to verify items exist
	if err := db.First(&order, "id = ?", orderID).Error; err != nil {
		return c.Status(404).JSON(model.Response{Success: false, Message: "Order not found"})
	}

	if order.Status != "ready" || order.DeliveryType != "pickup" {
		return c.Status(400).JSON(model.Response{
			Success: false,
			Message: fmt.Sprintf("Order must be 'ready' and 'pickup'. Current status: %s", order.Status),
		})
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&order).Update("status", "delivered").Error; err != nil {
			return err
		}
		// You can now access items via order.Prescription.Items (or however they are stored in Prescription)
		return nil
	})

	if err != nil {
		return c.Status(500).JSON(model.Response{Success: false, Message: "Dispensing failed"})
	}

	return c.JSON(model.Response{
		Success: true,
		Message: fmt.Sprintf("Order dispensed successfully to %s", order.PatientName),
		Data: fiber.Map{
			"order_id": order.ID,
			"status":   "delivered",
		},
	})
}

// GET /orders/today - Filtered list
func GetOrdersToday(c *fiber.Ctx) error {
	pharmacyID := c.Locals("pharmacy_id").(string)
	var orders []model.Order
	
	today := time.Now().Format("2006-01-02")
	database.DB.Preload("Prescription").
		Where("pharmacy_id = ? AND DATE(created_at) = ?", pharmacyID, today).
		Find(&orders)

	return c.JSON(model.Response{
		Success: true,
		Data: fiber.Map{
			"count":   len(orders),
			"results": orders,
		},
	})
}

// GET /orders/ready - List orders ready for pickup or dispatch
func GetOrdersReady(c *fiber.Ctx) error {
	pharmacyID := c.Locals("pharmacy_id").(string)
	var orders []model.Order

	database.DB.Preload("Prescription").
		Where("pharmacy_id = ? AND status = ?", pharmacyID, "ready").
		Find(&orders)

	return c.JSON(model.Response{
		Success: true,
		Data: fiber.Map{
			"count":   len(orders),
			"results": orders,
		},
	})
}

// PATCH /orders/:order_id/status - Update Lifecycle
func UpdateOrderStatus(c *fiber.Ctx) error {
	orderID := c.Params("order_id")
	var input struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(model.Response{Success: false, Message: "Invalid status"})
	}

	if err := database.DB.Model(&model.Order{}).Where("id = ?", orderID).Update("status", input.Status).Error; err != nil {
		return c.Status(500).JSON(model.Response{Success: false, Message: "Update failed"})
	}

	return c.JSON(model.Response{
		Success: true,
		Message: "Order status updated",
		Data: fiber.Map{"id": orderID, "status": input.Status},
	})
}







// POST /orders/:order_id/assign-rider
func AssignRider(c *fiber.Ctx) error {
    orderID := c.Params("order_id")
    var input struct {
        RiderID         string  `json:"rider_id"`
        DeliveryNotes   string  `json:"delivery_notes"`
        Charges         float64 `json:"charges"`
        PickupLocation  string  `json:"pickup_location"`
        DropoffLocation string  `json:"dropoff_location"`
    }

    if err := c.BodyParser(&input); err != nil {
        return c.Status(400).JSON(model.Response{Success: false, Message: "Invalid input"})
    }

    db := database.DB
    var order model.Order
    if err := db.First(&order, "id = ?", orderID).Error; err != nil {
        return c.Status(404).JSON(model.Response{Success: false, Message: "Order not found"})
    }

	

    if order.Status != "ready" {
        return c.Status(400).JSON(model.Response{
            Success: false, 
            Message: fmt.Sprintf("Order must be 'ready' to assign a rider. Current status: %s", order.Status),
        })
    }

    otp := fmt.Sprintf("%06d", rand.Intn(1000000))
    
    // Initialize the struct with correct field names and types
    delivery := model.Delivery{
        ID:              uuid.New().String(),
        OrderID:         order.ID,
        // FIX 1: Use &input.RiderID because the struct expects *string
        RiderID:         &input.RiderID, 
        // FIX 2: Use OtpCode (case sensitive)
        OtpCode:         otp, 
        PackageNumber:   "PKG-" + order.OrderNumber[4:],
        DeliveryNotes:   input.DeliveryNotes,
        Charges:         input.Charges,
        Status:          "pending",
        PickupLocation:  input.PickupLocation,
        DropoffLocation: input.DropoffLocation,
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
    }

	var rider model.User
	if err := db.First(&rider, "id = ?", input.RiderID).Error; err != nil {
		return c.Status(404).JSON(model.Response{Success: false, Message: "Rider not found"})
	}
	
	if !rider.OnDuty {
		return c.Status(400).JSON(model.Response{
			Success: false,
			Message: "Rider is not currently on duty",
		})
	}

    err := db.Transaction(func(tx *gorm.DB) error {
        // Create the delivery record
        if err := tx.Create(&delivery).Error; err != nil {
            return err
        }
        // Update order status to dispatched
        if err := tx.Model(&order).Update("status", "dispatched").Error; err != nil {
            return err
        }
        return nil
    })

    if err != nil {
        return c.Status(500).JSON(model.Response{Success: false, Message: "Assignment failed: " + err.Error()})
    }

    return c.Status(200).JSON(model.Response{
        Success: true,
        Message: "Rider assigned successfully",
        Data: fiber.Map{
            "order_id":     order.ID,
            "order_number": order.OrderNumber,
            "status":       "dispatched",
            "delivery_id":  delivery.ID,
            "otp_code":     otp,
        },
    })
}

// GET /orders/riders/available
func GetAvailableRiders(c *fiber.Ctx) error {
	var riders []model.User
	// Filtering by role and on_duty status
	database.DB.Where("role = ? AND on_duty = ? AND is_active = ?", "rider", true, true).Find(&riders)

	return c.JSON(model.Response{
		Success: true,
		Data:    riders,
	})
}