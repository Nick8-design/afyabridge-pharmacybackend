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




    
}



// func SetupRoutes(app *fiber.App) {
// 	api := app.Group("/api/v1/doctors")

// 	api.Get("/ping", func (c *fiber.Ctx) error {
// 		return c.SendString("Doctor api running")
// 	})

// 	// Public Routes
// 	auth := api.Group("/auth")
// 	auth.Post("/login", handlers.Login)
// 	auth.Post("/register", handlers.Register)
//     auth.Post("/forgot-password", handlers.ForgotPassword)
//     auth.Post("/logout", handlers.Logout)


//     // admin := api.Group("/admin", middleware.AdminProtect())
// 	// admin.Get("/doctors", handlers.AdminFetchAllDoctors)
// 	// admin.Get("/doctors/:doctorId", handlers.AdminFetchOneDoctor)
// 	// admin.Put("/doctors/:doctorId/verify", handlers.AdminVerifyDoctor)
// 	// admin.Put("/doctors/:doctorId/suspend", handlers.AdminSuspendDoctor)
// 	// admin.Put("/doctors/:doctorId/lock", handlers.AdminLockDoctor)
// 	// admin.Put("/doctors/:doctorId/activate", handlers.AdminActivateDoctor)
// 	// admin.Delete("/doctors/:doctorId", handlers.AdminDeleteDoctor)


// 	// Protected Routes (Require Token)
// 	api.Use(middleware.Protected())


//     profile := api.Group("/profile")
//     profile.Get("/", handlers.GetDoctorProfile)
//     profile.Put("/personal", handlers.UpdateDoctorProfile)
    
//     profile.Delete("/", handlers.DeleteDoctorAccount)  
//     profile.Patch("/verify", handlers.VerifyDoctor)
 
//     profile.Post("/change-password", handlers.ChangePassword) 
//     profile.Post("/2fa/enable", handlers.Enable2FA)          
//     profile.Post("/signout-all", handlers.SignOutAll)
    
   
   
   
   
//     api.Use(middleware.DoctorServiceGuard())
    
//     api.Get("/search", handlers.GlobalSearch)


// 	// Appointments
//     apt := api.Group("/appointments")
//     apt.Get("/", handlers.GetAppointments)
//     apt.Patch("/:id/status", handlers.UpdateAppointmentStatus)
//     apt.Get("/:id", handlers.GetAppointmentDetails)
//     // apt.Patch("/:id/status", handlers.UpdateAppointmentStatus)
//     apt.Post("/:id/reschedule", handlers.RescheduleAppointment)
//     apt.Post("/:id/cancel", handlers.CancelAppointment)



// 	// Patients
// 	patients := api.Group("/patients")
//     patients.Get("/", handlers.GetPatientsList)   
//     // patients.Post("/", handlers.CreatePatient) 
//     patients.Get("/:id", handlers.GetPatientDetails) 
   
//     patients.Post("/vitals", handlers.UpdateVitals)
//     patients.Get("/:id/vitals", handlers.GetPatientVitals)


// 	patients.Post("/prescriptions", handlers.CreatePrescription) 
//     patients.Get("/:id/current/prescriptions", handlers.GetPatientPrescriptionsCurrentDoctor)
//     patients.Get("/:id/prescriptions", handlers.GetPatientPrescriptionsAll)

//     api.Post("/orders", handlers.CreateOrder) 
//     api.Get("/pharmacies/search", handlers.GetActivePharmacies) 

// // patients.Get("/:id/prescriptions/summary", handlers.GetPrescriptionSummary)


// // patients.Post("/:id/send-to-pharmacy", handlers.SendPrescriptionToExternal)



//     // Clinical - Lab Orders
//     api.Post("/patients/:patientId/lab-orders", handlers.CreateLabOrder)
//     api.Get("/patients/:patientId/lab-orders", handlers.GetLabOrders)

// 	// consultations
// 	consult := api.Group("/consultations")
//     consult.Get("/queue", handlers.GetConsultationQueue)       
//     consult.Post("/start", handlers.StartConsultation)         
 
//     consult.Post("/:consultationId/end", handlers.EndConsultation)
    
   


//     // // Dashboard
// 	  api.Get("/dashboard", handlers.GetDashboardStats) 
	
// // Notifications Group
// notify := api.Group("/notifications")
// notify.Get("/", handlers.GetDoctorNotifications)          // Get doctor's alerts
// notify.Post("/send", handlers.CreateNotification)         // Doctor sends to others
// notify.Patch("/:id/read", handlers.MarkNotificationAsRead) // Mark as read


//       // WebSocket Endpoint (Needs to be outside middleware if token is in query)
//     app.Get("/ws/chat/:id", websocket.New(handlers.ChatWebSocket))

//     // Protected Chat API
//     chat := api.Group("/chat")
//     chat.Get("/history/:patientId", handlers.GetChatHistory)
    
//     // Media Upload for Chat (Images/Audio)
//     chat.Post("/upload", handlers.UploadChatMedia)
// }