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
	pharmacyID := c.Locals("pharmacy_id").(string)
	today := time.Now().Format("2006-01-02")

	var stats struct {
		PendingOrders   int64   `json:"pending_orders"`
		LowStockAlerts  int64   `json:"low_stock_alerts"`
		ReadyForPickup  int64   `json:"ready_for_pickup"`
		ActiveDeliveries int64   `json:"active_deliveries"`
		TodayRevenue    float64 `json:"-"` // Internal use for calculation
	}

	// 1. Count Pending Orders
	database.DB.Model(&model.Order{}).Where("pharmacy_id = ? AND status = ?", pharmacyID, "pending").Count(&stats.PendingOrders)

	// 2. Count Low Stock Alerts
	database.DB.Model(&model.Inventory{}).Where("pharmacy_id = ? AND quantity_in_stock <= reorder_level AND is_active = ?", pharmacyID, true).Count(&stats.LowStockAlerts)

	// 3. Count Ready for Pickup
	database.DB.Model(&model.Order{}).Where("pharmacy_id = ? AND status = ?", pharmacyID, "ready").Count(&stats.ReadyForPickup)

	// 4. Count Active Deliveries (Status: assigned, picked_up, in_transit)
	database.DB.Table("deliveries").
		Joins("JOIN orders ON orders.id = deliveries.order_id").
		Where("orders.pharmacy_id = ? AND deliveries.status IN ?", pharmacyID, []string{"assigned", "picked_up", "in_transit"}).
		Count(&stats.ActiveDeliveries)

	// 5. Calculate Today's Revenue (Sum of total_amount for orders completed/delivered today)
	database.DB.Model(&model.Order{}).
		Where("pharmacy_id = ? AND status = ? AND DATE(updated_at) = ?", pharmacyID, "delivered", today).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&stats.TodayRevenue)

	return c.JSON(model.Response{
		Success: true,
		Data: fiber.Map{
			"pending_orders":   stats.PendingOrders,
			"low_stock_alerts": stats.LowStockAlerts,
			"ready_for_pickup": stats.ReadyForPickup,
			"active_deliveries": stats.ActiveDeliveries,
			"today_revenue":    fmt.Sprintf("%.2f", stats.TodayRevenue),
		},
	})
}