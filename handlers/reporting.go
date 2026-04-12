package handlers

import (
	"afyabridge-pharmacybackend/database"
	"afyabridge-pharmacybackend/model"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

// GET /reporting/dashboard/
func GetDashboardKPIs(c *fiber.Ctx) error {
	pharmacyID, ok := c.Locals("pharmacy_id").(string)
	if !ok || pharmacyID == "" {
		return c.Status(400).JSON(model.Response{
			Success: false,
			Message: "Your account is not linked to a pharmacy",
		})
	}

	today := time.Now().Format("2006-01-02")

	var stats struct {
		PendingOrders    int64
		LowStockAlerts   int64
		ReadyForPickup   int64
		ActiveDeliveries int64
		TodayRevenue     float64
	}

	// Pending Orders
	database.DB.Model(&model.Order{}).
		Where("pharmacy_id = ? AND status = ?", pharmacyID, "pending").
		Count(&stats.PendingOrders)

	// Low Stock
	database.DB.Model(&model.Inventory{}).
		Where("pharmacy_id = ? AND quantity_in_stock <= reorder_level AND is_active = ?", pharmacyID, true).
		Count(&stats.LowStockAlerts)

	// Ready Orders
	database.DB.Model(&model.Order{}).
		Where("pharmacy_id = ? AND status = ?", pharmacyID, "ready").
		Count(&stats.ReadyForPickup)

	// Active Deliveries
	database.DB.Table("deliveries").
		Joins("JOIN orders ON orders.id = deliveries.order_id").
		Where("orders.pharmacy_id = ? AND deliveries.status IN ?", pharmacyID,
			[]string{"assigned", "picked_up", "in_transit"}).
		Count(&stats.ActiveDeliveries)

	// Revenue
	database.DB.Model(&model.Order{}).
		Where("pharmacy_id = ? AND status = ? AND DATE(updated_at) = ?", pharmacyID, "delivered", today).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&stats.TodayRevenue)

	return c.JSON(model.Response{
		Success: true,
		Data: fiber.Map{
			"pending_orders":    stats.PendingOrders,
			"low_stock_alerts":  stats.LowStockAlerts,
			"ready_for_pickup":  stats.ReadyForPickup,
			"active_deliveries": stats.ActiveDeliveries,
			"today_revenue":     fmt.Sprintf("%.2f", stats.TodayRevenue),
		},
	})
}