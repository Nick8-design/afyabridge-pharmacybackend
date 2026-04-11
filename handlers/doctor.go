package handlers

// import (
// 	"fmt"
// 	"afyabridge-pharmacybackend/database"
// 	"afyabridge-pharmacybackend/model"
// 	"time"

// 	"github.com/gofiber/fiber/v2"
// )

// func GetDashboardStats(c *fiber.Ctx) error {
// 	db := database.DB
// 	doctorID := c.Locals("doctorID").(string)
// 	today := time.Now().Format("2006-01-02")

// 	var totalPatients int64
// 	var totalAppointments int64
// 	var todaysTotal int64
// 	var todaysUpcoming int64
// 	var monthlyEarnings float64

// 	// 1. Count Total Patients (Updated to use doctor_id snake_case)
// 	db.Model(&model.Patient{}).Where("doctor_id = ?", doctorID).Count(&totalPatients)

// 	// 2. Count Total Appointments (Using doctor_id from model)
// 	db.Model(&model.Appointment{}).Where("doctor_id = ?", doctorID).Count(&totalAppointments)

// 	// 3. Today's Schedule Stats
// 	db.Model(&model.Appointment{}).
// 		Where("doctor_id = ? AND date = ?", doctorID, today).
// 		Count(&todaysTotal)
	
// 	// 'confirmed' status matches your Appointment model enum
// 	db.Model(&model.Appointment{}).
// 		Where("doctor_id = ? AND date = ? AND status = ?", doctorID, today, "confirmed").
// 		Count(&todaysUpcoming)

// 	// 4. Calculate Monthly Earnings (Sum of 'charges' field)
// 	currentMonth := time.Now().Format("2006-01")
// 	db.Model(&model.Appointment{}).
// 		Where("doctor_id = ? AND status = ? AND date LIKE ?", doctorID, "completed", currentMonth+"%").
// 		Select("COALESCE(SUM(charges), 0)").
// 		Scan(&monthlyEarnings)

// 	return c.JSON(fiber.Map{
// 		"success": true,
// 		"data": fiber.Map{
// 			"stats": fiber.Map{
// 				"totalAppointments": totalAppointments,
// 				"totalPatients":     totalPatients,
// 				"monthlyEarnings":   fmt.Sprintf("KES %.2f", monthlyEarnings),
// 			},
// 			"todaysSchedule": fiber.Map{
// 				"total":    todaysTotal,
// 				"upcoming": todaysUpcoming,
// 			},
// 		},
// 	})
// }

// // GET /api/v1/doctors/profile
// func GetDoctorProfile(c *fiber.Ctx) error {
// 	db := database.DB
// 	doctorID := c.Locals("doctorID").(string)

// 	var doctor model.User
// 	// Assuming your User model also uses doctor_id or id as primary key
// 	if err := db.Where("id = ?", doctorID).First(&doctor).Error; err != nil {
// 		return c.Status(404).JSON(fiber.Map{"success": false, "message": "Doctor profile not found"})
// 	}

// 	return c.JSON(fiber.Map{
// 		"success": true,
// 		"data":    doctor,
// 	})
// }

// // PUT /api/v1/doctors/profile/personal
// func UpdateDoctorProfile(c *fiber.Ctx) error {
// 	db := database.DB
// 	doctorID := c.Locals("doctorID").(string)
	
// 	var updateData model.User
// 	if err := c.BodyParser(&updateData); err != nil {
// 		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid input data"})
// 	}

// 	// Aligned with the field names and security constraints
// 	result := db.Model(&model.User{}).
// 		Where("id = ?", doctorID).
// 		Updates(map[string]interface{}{
// 			"full_name":        updateData.FullName,
// 			"phone_number":     updateData.PhoneNumber,
// 			"specialty":        updateData.Specialty,
// 			"hospital":         updateData.Hospital,
// 			"consultation_fee": updateData.ConsultationFee,
// 			"updated_at":       time.Now(),
// 		})

// 	if result.Error != nil {
// 		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to update profile"})
// 	}

// 	if result.RowsAffected > 0 {
// 		var updatedDoctor model.User
// 		db.Where("id = ?", doctorID).First(&updatedDoctor)
		
// 		// Run heavy slot generation in a background goroutine
// 		go GenerateSlots(db, updatedDoctor)
// 	}

// 	return c.JSON(fiber.Map{"success": true, "message": "Profile updated successfully"})
// }




// // package handlers

// // import (
// //     "server/database"
// //     "server/model"
// //     "time"
// //     "fmt"
// //     "github.com/gofiber/fiber/v2"
// // )




// // func GetDashboardStats(c *fiber.Ctx) error {
// //     db := database.DB
// //     doctorID := c.Locals("doctorID").(string)
// //     today := time.Now().Format("2006-01-02")

// //     var totalPatients int64
// //     var totalAppointments int64
// //     var todaysTotal int64
// //     var todaysUpcoming int64
// //     var monthlyEarnings float64

// //     // 1. Count Total Patients for this doctor
// //     db.Model(&model.Patient{}).Where("doctorId = ?", doctorID).Count(&totalPatients)

// //     // 2. Count Total Appointments (All time)
// //     db.Model(&model.Appointment{}).Where("doctor_id = ?", doctorID).Count(&totalAppointments)

// //     // 3. Today's Schedule Stats
// //     db.Model(&model.Appointment{}).
// //         Where("doctor_id = ? AND date = ?", doctorID, today).
// //         Count(&todaysTotal)
    
// //     db.Model(&model.Appointment{}).
// //         Where("doctor_id = ? AND date = ? AND status = ?", doctorID, today, "confirmed").
// //         Count(&todaysUpcoming)

// //     // 4. Calculate Monthly Earnings (Sum of fees for completed appointments this month)
// //     // Assuming 'date' format is 2006-01-02
// //     currentMonth := time.Now().Format("2006-01")
// //     db.Model(&model.Appointment{}).
// //         Where("doctor_id = ? AND status = ? AND date LIKE ?", doctorID, "completed", currentMonth+"%").
// //         Select("COALESCE(SUM(charges), 0)").
// //         Scan(&monthlyEarnings)

// //     return c.JSON(fiber.Map{
// //         "success": true,
// //         "data": fiber.Map{
// //             "stats": fiber.Map{
// //                 "totalAppointments": totalAppointments,
// //                 "totalPatients":     totalPatients,
// //                 "monthlyEarnings":   fmt.Sprintf("KES %.2f", monthlyEarnings),
// //             },
// //             "todaysSchedule": fiber.Map{
// //                 "total":    todaysTotal,
// //                 "upcoming": todaysUpcoming,
// //             },
// //         },
// //     })
// // }



// // // GET /api/v1/doctors/profile
// // func GetDoctorProfile(c *fiber.Ctx) error {
// //     db := database.DB
// //     doctorID := c.Locals("doctorID").(string)

// // 	// fmt.Printf("Doctor id is "+doctorID)
    
// //     var doctor model.User
// //     if err := db.Where("doctor_id = ?", doctorID).First(&doctor).Error; err != nil {
// //         return c.Status(404).JSON(fiber.Map{"success": false, "message": "Doctor profile not found"})
// //     }

// //     return c.JSON(fiber.Map{
// //         "success": true,
// //         "data":    doctor,
// //     })
// // }

// // // PUT /api/v1/doctors/profile/personal
// // func UpdateDoctorProfile(c *fiber.Ctx) error {
// //     db := database.DB
// //     doctorID := c.Locals("doctorID").(string)
    
// //     var updateData model.User
// //     if err := c.BodyParser(&updateData); err != nil {
// //         return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid input data"})
// //     }

// //     // Security: Ensure we don't update ID or Password via this endpoint
// //     result := db.Model(&model.User{}).
// //         Where("doctor_id = ?", doctorID).
// //         Updates(map[string]interface{}{
// //             "full_name":        updateData.FullName,
// //             "phone_number":     updateData.PhoneNumber,
// //             "specialty":        updateData.Specialty,
// //             "hospital":         updateData.Hospital,
// //             "consultation_fee": updateData.ConsultationFee,
// //         })

// //     if result.Error != nil {
// //         return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to update profile"})
// //     }


// //     // Inside UpdateDoctorProfile
// // if result.Error == nil {
// //     // Fetch the updated doctor record to get the new WorkingHours/SlotDuration
// //     var updatedDoctor model.User
// //     db.Where("doctor_id = ?", doctorID).First(&updatedDoctor)
    
// //     // Re-generate slots based on new settings
// //     go GenerateSlots(db, updatedDoctor)
// // }

// //     return c.JSON(fiber.Map{"success": true, "message": "Profile updated successfully"})
// // }