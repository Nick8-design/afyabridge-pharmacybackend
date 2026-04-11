package routes

import (
	"afyabridge-pharmacybackend/handlers"
	"afyabridge-pharmacybackend/middleware"

	"github.com/gofiber/fiber/v2"
)


// import (
// 	"afyabridge/handlers"
// 	"afyabridge/middleware"
// 	"github.com/gofiber/fiber/v2"
// )

func SetupRoutes(app *fiber.App) {
         
        api := app.Group("/api")



        auth := api.Group("/auth")
        // Public Routes
        auth.Post("/login", handlers.Login)
        auth.Post("/register/complete/", handlers.RegisterComplete)
        auth.Post("/forgot-password", handlers.ForgotPassword)
        auth.Post("/reset-password", handlers.ResetPassword)
        auth.Post("/otp/send", handlers.SendOTP)
        auth.Post("/otp/verify", handlers.VerifyOTP)
    
        // Protected Routes (Require Token)
        api.Use(middleware.Protected())
        
        api.Post("/logout", handlers.Logout)
        api.Get("/profile", handlers.GetProfile)
        api.Put("/profile", handlers.UpdateProfile)
        api.Patch("/profile/photo", handlers.UpdatePhoto)
        api.Delete("/profile/photo", handlers.DeletePhoto) // Set profile_image to NULL
        api.Put("/change-password", handlers.ChangePassword)


        orders := api.Group("/orders")

        orders.Get("/", handlers.GetOrders)
        orders.Get("/today", handlers.GetOrdersToday)
        orders.Get("/ready", handlers.GetOrdersReady)
        orders.Get("/riders/available", handlers.GetAvailableRiders)
        orders.Get("/:order_id", handlers.GetOrderDetails)
        orders.Patch("/:order_id/status", handlers.UpdateOrderStatus)
        orders.Post("/:order_id/cancel", handlers.CancelOrder)
        orders.Post("/:order_id/dispense", handlers.DispenseOrder)
        orders.Post("/:order_id/assign-rider", handlers.AssignRider)
        orders.Get("/patient/:patient_id/history", handlers.GetPatientOrderHistory)

// Inventory Group
inventory := api.Group("/inventory")
{
    inventory.Get("/", handlers.GetInventory)
    inventory.Post("/", handlers.AddDrug)
    inventory.Get("/dashboard", handlers.GetInventoryDashboard)
    inventory.Post("/:drug_id/restock", handlers.RestockDrug)
    inventory.Get("/:drug_id", handlers.GetDrugDetails) // Implement similarly to GetOrderDetails
    inventory.Put("/:drug_id", handlers.UpdateDrug)    // Implement using tx.Save()
    inventory.Delete("/:drug_id", handlers.DeleteDrug) // Soft delete: tx.Model().Update("is_active", false)
}

// Delivery Group
deliveries := api.Group("/deliveries")
{
    deliveries.Get("/", handlers.GetDeliveries)
    deliveries.Patch("/:delivery_id/status", handlers.UpdateDeliveryStatus)
    deliveries.Post("/:delivery_id/confirm", handlers.ConfirmDelivery)
}




// Reporting
api.Get("/reporting/dashboard", handlers.GetDashboardKPIs)

// Settings Group
settings := api.Group("/settings/pharmacy")
{
    settings.Get("/", handlers.GetPharmacySettings)
    settings.Put("/", handlers.UpdatePharmacySettings)
    settings.Get("/hours", handlers.GetPharmacyHours)
    settings.Put("/hours", handlers.UpdatePharmacyHours)
    // For logo upload, ensure you have a multipart form handler
    settings.Patch("/logo", handlers.UploadPharmacyLogo) 
}
    
}




