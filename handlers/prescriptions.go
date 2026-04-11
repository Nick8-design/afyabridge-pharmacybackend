package handlers

import (
	"errors"
	"afyabridge-pharmacybackend/database"
	"afyabridge-pharmacybackend/model"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// POST /api/v1/orders
func CreateOrder(c *fiber.Ctx) error {
	// Grab the ID of the person creating this order (Doctor or Pharmacist)
	// This ensures we satisfy the 'prepared_by' Foreign Key constraint
	creatorID := c.Locals("doctorID").(string) 
	db := database.DB

	var input struct {
		PrescriptionID string `json:"prescription_id"`
		PharmacyID     string `json:"pharmacy_id"`
		DeliveryType   string `json:"delivery_type"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid input"})
	}

	// 1. Fetch the existing prescription to grab patient data snapshots
	var rx model.Prescription
	if err := db.First(&rx, "id = ?", input.PrescriptionID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "message": "Prescription not found"})
	}

	// 2. Fetch Pharmacy from the PHARMACIES table (not users table)
	var facility model.Pharmacy
	if err := db.Where("id = ? AND is_active = ?", input.PharmacyID, true).First(&facility).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"success": false, "message": "The selected pharmacy is either inactive or does not exist"})
	}

	// 3. Create the Order with all required ENUMs and Foreign Keys
	newOrder := model.Order{
		OrderNumber:    "ORD-" + time.Now().Format("20060102") + "-" + uuid.New().String()[:5],
		PrescriptionID: rx.ID.String(),
		PharmacyID:     facility.ID, // Using the ID from the pharmacy facility
		PatientID:      rx.PatientID.String(),
		PatientName:    rx.PatientName,
		PatientPhone:   stringValue(rx.PatientPhone),
        PatientAddress: stringValue(rx.PatientAddress),
		
		
		// satisfy ENUM and Foreign Key constraints
		PreparedBy:     creatorID, // Satisfies fk_3
		DeliveryType:   input.DeliveryType,
		Status:         "pending",
		PaymentStatus:  "unpaid",
		PaymentMethod:  "mpesa", // Default to prevent truncation error
		TotalAmount:    decimal.NewFromInt(0),
	}

	// newOrder := model.Order{
	// 	OrderNumber:    "ORD-" + time.Now().Format("20060102") + "-" + uuid.New().String()[:5],
	// 	PrescriptionID: rx.ID.String(),
	
	// 	PharmacyID:     facility.ID,
		
	// 	// FULL SNAPSHOT
	// 	PatientID:      rx.PatientID.String(),
	// 	PatientName:    rx.PatientName,
	// 	PatientPhone:   *rx.PatientPhone,
	// 	PatientAddress: *rx.PatientAddress,
	
	// 	// ADD THIS
	// 	PharmacyName:   facility.Name,
	// 	PharmacyPhone:  facility.Phone,
	// 	PharmacyAddress: facility.AddressLine1,
	
	// 	PreparedBy:     creatorID,
	
	// 	DeliveryType:   input.DeliveryType,
	// 	Status:         "pending",
	// 	PaymentStatus:  "unpaid",
	// 	PaymentMethod:  "mpesa",
	// 	TotalAmount:    decimal.NewFromInt(0),
	// }

	// Handle default delivery type if empty
	if newOrder.DeliveryType == "" {
		newOrder.DeliveryType = "pickup"
	}

	if err := db.Create(&newOrder).Error; err != nil {
		message := err.Error()
		if strings.Contains(message, "1452") {
			message = "Database Integrity Error: Ensure the Pharmacy and Creator IDs are valid."+err.Error()
		}
		return c.Status(500).JSON(fiber.Map{"success": false, "message": message})
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"message": "Order created successfully",
		"order":   newOrder,
	})
}


// POST /api/v1/doctors/patients/:id/prescriptions
func CreatePrescription(c *fiber.Ctx) error {
    doctorIDStr := c.Locals("doctorID").(string)
    db := database.DB

    var prescription model.Prescription
    if err := c.BodyParser(&prescription); err != nil {
        return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid input"})
    }

    // Prepare metadata
    prescription.ID = uuid.New()
    pNum := "RX-" + time.Now().Format("020106-150405")
    prescription.PrescriptionNumber = &pNum
    prescription.IssueDate = time.Now()
    prescription.DoctorID, _ = uuid.Parse(doctorIDStr)




	
    orderCreated := false
    var orderNote string

    // START TRANSACTION
    err := db.Transaction(func(tx *gorm.DB) error {
        var patient model.User
if err := tx.Where("id = ?", prescription.PatientID).First(&patient).Error; err != nil {
    return errors.New("patient not found")
}

var doctor model.User
if err := tx.Select("full_name").Where("id = ?", doctorIDStr).First(&doctor).Error; err != nil {
    return errors.New("doctor not found")
}
prescription.PatientName = patient.FullName
prescription.PatientPhone = patient.PhoneNumber
prescription.PatientAddress = patient.Address
prescription.DoctorName = doctor.FullName



        if prescription.MakeOrder {
            // SCENARIO 1: Order Requested
			if prescription.PharmacyID == nil || *prescription.PharmacyID == uuid.Nil {
				return errors.New("pharmacy_id is required when make_order is true")
			}

            // Verify the pharmacy exists in the Users/Pharmacies table
            // var pharmacy model.User
            // if err := tx.Where("id = ?", prescription.PharmacyID).First(&pharmacy).Error; err != nil {
            //     return errors.New("selected pharmacy does not exist in our records")
            // }

			var facility model.Pharmacy
    if err := tx.Where("id = ? AND is_active = ?", prescription.PharmacyID, true).First(&facility).Error; err != nil {
        return errors.New("the selected pharmacy is either inactive or does not exist")
    }

            // Create the Prescription (with pharmacy_id)
            if err := tx.Create(&prescription).Error; err != nil {
                return err
            }

      
newOrder := model.Order{
    PrescriptionID: prescription.ID.String(),
    PharmacyID:     prescription.PharmacyID.String(), 
    PatientID:      prescription.PatientID.String(),
    
    // Use the doctorID from Locals to satisfy the 'prepared_by' Foreign Key
	
    PatientName:    prescription.PatientName,
	PatientPhone:   stringValue(prescription.PatientPhone),
    PatientAddress: stringValue(prescription.PatientAddress),
    PreparedBy:     doctorIDStr, 
	Priority:       prescription.Priority,
    Status:         "pending",
    TotalAmount:    decimal.NewFromInt(0),
    PaymentMethod:  "mpesa",   // Added to prevent ENUM truncation error
    DeliveryType:   "home_delivery",  // Added to satisfy ENUM defaults
    PaymentStatus:  "unpaid",
}

if err := tx.Create(&newOrder).Error; err != nil {
    return err
}
            orderCreated = true

        } else {
            // SCENARIO 2: No Order Requested
            // Force pharmacy_id to nil so the Foreign Key constraint is bypassed
            prescription.PharmacyID = nil 
            
            if err := tx.Create(&prescription).Error; err != nil {
                return err
            }
            orderNote = "Prescription saved safely without an order."
        }

        return nil
    })

    if err != nil {
		// Custom error handling for the Foreign Key issue
		message := err.Error()
		if strings.Contains(message, "1452") {
			message = "The selected pharmacy ID is invalid or not registered. : "+err.Error()
		}
        return c.Status(400).JSON(fiber.Map{"success": false, "message": message})
    }

    return c.Status(201).JSON(fiber.Map{
        "success":       true,
        "data":          prescription,
        "order_created": orderCreated,
        "message":       orderNote,
    })
}



// GET /api/v1/pharmacies/search?county=Nairobi
func GetActivePharmacies(c *fiber.Ctx) error {
	db := database.DB
	county := c.Query("county")

	var pharmacies []model.Pharmacy
	query := db.Where("is_active = ?", true)

	if county != "" {
		query = query.Where("county = ?", county)
	}

	if err := query.Find(&pharmacies).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Could not fetch pharmacies"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    pharmacies,
	})
}


// Helper to safely dereference string pointers
func stringValue(s *string) string {
    if s == nil {
        return ""
    }
    return *s
}

// func CreatePrescription(c *fiber.Ctx) error {
// 	doctorIDStr := c.Locals("doctorID").(string)
// 	db := database.DB

// 	var prescription model.Prescription
// 	if err := c.BodyParser(&prescription); err != nil {
// 		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid input"})
// 	}

// 	// Fetch Doctor & Patient for snapshots
// 	var doctor model.User
// 	var patient model.User
// 	db.First(&doctor, "id = ?", doctorIDStr)
// 	if err := db.First(&patient, "id = ?", prescription.PatientID).Error; err != nil {
// 		return c.Status(404).JSON(fiber.Map{"success": false, "message": "Patient not found"})
// 	}

// 	// Prepare Metadata
// 	prescription.ID = uuid.New()
// 	pNum := "RX-" + time.Now().Format("020106-150405")
// 	prescription.PrescriptionNumber = &pNum
// 	prescription.IssueDate = time.Now()
// 	prescription.DoctorID, _ = uuid.Parse(doctorIDStr)
// 	prescription.DoctorName = doctor.FullName
// 	prescription.PatientName = patient.FullName
// 	prescription.PatientPhone = patient.PhoneNumber
// 	prescription.PatientAddress = patient.Address

// 	orderCreated := false
// 	var orderError string

// 	// START TRANSACTION
// 	err := db.Transaction(func(tx *gorm.DB) error {
// 		// 1. Create Prescription
// 		if err := tx.Create(&prescription).Error; err != nil {
// 			return err
// 		}

// 		// 2. Logic for Order Creation
// 		if prescription.MakeOrder {
// 			// Validate Pharmacy Existence
// 			var pharmacy model.User
// 			if err := tx.Where("id = ? AND role = 'pharmacist'", prescription.PharmacyID).First(&pharmacy).Error; err != nil {
// 				orderError = "Pharmacy not found. Order skipped."
// 				return errors.New("invalid pharmacy id") 
// 			}

// 			// Create the Order
// 			newOrder := model.Order{
// 				PrescriptionID: prescription.ID.String(),
// 				PharmacyID:     prescription.PharmacyID.String(),
// 				PatientID:      prescription.PatientID.String(),
// 				PatientName:    prescription.PatientName,
// 				PatientPhone:   *prescription.PatientPhone,
// 				PatientAddress: *prescription.PatientAddress,
// 				Status:         "pending",
// 				TotalAmount:    decimal.NewFromInt(0), // Set real amount if available
// 			}

// 			if err := tx.Create(&newOrder).Error; err != nil {
// 				return err
// 			}
// 			orderCreated = true
// 		}
// 		return nil
// 	})

// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{
// 			"success": false, 
// 			"message": "Failed to complete request: " + err.Error(),
// 		})
// 	}

// 	return c.Status(201).JSON(fiber.Map{
// 		"success":       true,
// 		"data":          prescription,
// 		"order_created": orderCreated,
// 		"order_note":    orderError,
// 	})
// }










// func CreatePrescription(c *fiber.Ctx) error {
// 	doctorIDStr := c.Locals("doctorID").(string)
// 	doctorID, _ := uuid.Parse(doctorIDStr)
// 	db := database.DB

// 	var prescription model.Prescription
// 	if err := c.BodyParser(&prescription); err != nil {
// 		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid input"})
// 	}

// 	// 1. Fetch Doctor Info for the snapshot
// 	var doctor model.User
// 	// Assuming your User model uses 'id' as the primary key
// 	db.Select("full_name").Where("id = ?", doctorIDStr).First(&doctor)

// 	// 2. Fetch Patient Info for the snapshot
// 	var patient model.User
// 	if err := db.Where("id = ?", prescription.PatientID).First(&patient).Error; err != nil {
// 		return c.Status(404).JSON(fiber.Map{"success": false, "message": "Patient not found"})
// 	}

// 	// 3. Populate Snapshot & Metadata
// 	prescription.ID = uuid.New()
// 	pNum := "RX-" + time.Now().Format("020106-150405")
// 	prescription.PrescriptionNumber = &pNum
	
// 	prescription.IssueDate = time.Now()
// 	prescription.DoctorID = doctorID
// 	prescription.DoctorName = doctor.FullName
	
// 	prescription.PatientName = patient.FullName
// 	prescription.PatientPhone = patient.PhoneNumber
// 	prescription.PatientAddress = patient.Address
	
// 	// Set default status if not provided
// 	if prescription.Status == "" {
// 		prescription.Status = "pending"
// 	}

// 	// 4. Save to prescriptions table
// 	if err := db.Create(&prescription).Error; err != nil {
// 		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to create prescription: " + err.Error()})
// 	}

// 	return c.Status(201).JSON(fiber.Map{"success": true, "data": prescription})
// }

// GET /api/v1/doctors/patients/:id/prescriptions
func GetPatientPrescriptionsCurrentDoctor(c *fiber.Ctx) error {
	patientID := c.Params("id")
	doctorID := c.Locals("doctorID").(string)
	db := database.DB

	var prescriptions []model.Prescription
	err := db.Where("patient_id = ? AND doctor_id = ?", patientID, doctorID).
		Order("created_at desc").
		Find(&prescriptions).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Database error"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"prescriptions": prescriptions,
		},
	})
}

func GetPatientPrescriptionsAll(c *fiber.Ctx) error {
	patientID := c.Params("id")
	db := database.DB

	var prescriptions []model.Prescription
	err := db.Where("patient_id = ?", patientID).
		Order("created_at desc").
		Find(&prescriptions).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Database error"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"prescriptions": prescriptions,
		},
	})
}

// GET /api/v1/doctors/patients/:id/prescriptions/summary (Grouped by Date)
func GetPrescriptionSummary(c *fiber.Ctx) error {
	patientID := c.Params("id")
	doctorID := c.Locals("doctorID").(string)
	db := database.DB

	var prescriptions []model.Prescription
	// Fixed: doctorID -> doctor_id to match SQL naming
	db.Where("patient_id = ? AND doctor_id = ?", patientID, doctorID).Find(&prescriptions)

	// Grouping logic (using IssueDate)
	summary := make(map[string][]model.Prescription)
	for _, p := range prescriptions {
		dateKey := p.IssueDate.Format("2006-01-02")
		summary[dateKey] = append(summary[dateKey], p)
	}

	return c.JSON(fiber.Map{"success": true, "data": summary})
}




// package handlers

// import (
// 	"server/database"
// 	"server/model"
// 	"time"

// 	"github.com/gofiber/fiber/v2"
// )

// // POST /api/v1/doctors/patients/:id/prescriptions
// func CreatePrescription(c *fiber.Ctx) error {
// 	// id := c.Params("id") // Patient UID
// 	doctorID := c.Locals("doctorID").(string)
// 	db := database.DB

// 	var prescription model.Prescription
// 	if err := c.BodyParser(&prescription); err != nil {
// 		return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid input"})
// 	}

// 	// 1. Fetch Doctor Info for the snapshot
// 	var doctor model.User
// 	db.Select("full_name").Where("doctor_id = ?", doctorID).First(&doctor)

// 	// 2. Fetch Patient Info for the snapshot
// 	var patient model.Patient
// 	if err := db.Where("id = ?",prescription.PatientID).First(&patient).Error; err != nil {
// 		return c.Status(404).JSON(fiber.Map{"success": false, "message": "Patient not found"})
// 	}

// 	// 3. Populate Snapshot & Metadata
// 	prescription.PrescriptionId = "RX-" + time.Now().Format("020106-150405")
// 	prescription.PrescriptionDate = time.Now().Format("2006-01-02")
// 	prescription.DoctorID = doctorID
// 	prescription.DoctorName = doctor.FullName
// 	prescription.PatientID = patient.PatientID
// 	prescription.PatientName = patient.FullName
// 	prescription.PatientsPhone = patient.PhoneNumber
// 	prescription.PatientsAddress = patient.Address

// 	// 4. Save to prescriptions table (Standalone row)
// 	if err := db.Create(&prescription).Error; err != nil {
// 		return c.Status(500).JSON(fiber.Map{"success": false, "message": err.Error()})
// 	}

// 	return c.Status(201).JSON(fiber.Map{"success": true, "data": prescription})
// }

// // GET /api/v1/doctors/patients/:id/prescriptions
// func GetPatientPrescriptionsCurrentDoctor(c *fiber.Ctx) error {
// 	id := c.Params("id")
// 	doctorID := c.Locals("doctorID").(string)
// 	db := database.DB

// 	var prescriptions []model.Prescription
// 	// Fetch directly from the prescriptions table
// 	err := db.Where("patient_id = ? AND doctor_id = ?", id, doctorID).
// 		Order("created_at desc").
// 		Find(&prescriptions).Error

// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Database error"})
// 	}

// 	return c.JSON(fiber.Map{
// 		"success": true,
// 		"data": fiber.Map{
// 			"prescriptions": prescriptions,
// 		},
// 	})
// }

// func GetPatientPrescriptionsAll(c *fiber.Ctx) error {
// 	id := c.Params("id")
	
// 	db := database.DB

// 	var prescriptions []model.Prescription
// 	// Fetch directly from the prescriptions table
// 	err := db.Where("patient_id = ? ", id).
// 		Order("created_at desc").
// 		Find(&prescriptions).Error

// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{"success": false, "message": "Database error"})
// 	}

// 	return c.JSON(fiber.Map{
// 		"success": true,
// 		"data": fiber.Map{
// 			"prescriptions": prescriptions,
// 		},
// 	})
// }


// // GET /api/v1/doctors/patients/:id/prescriptions/summary (Grouped by Date)
// func GetPrescriptionSummary(c *fiber.Ctx) error {
// 	id := c.Params("id")
// 	doctorID := c.Locals("doctorID").(string)
// 	db := database.DB

// 	var prescriptions []model.Prescription
// 	db.Where("patient_id = ? AND doctorId = ?", id, doctorID).Find(&prescriptions)

// 	// Grouping logic
// 	summary := make(map[string][]model.Prescription)
// 	for _, p := range prescriptions {
// 		summary[p.PrescriptionDate] = append(summary[p.PrescriptionDate], p)
// 	}

// 	return c.JSON(fiber.Map{"success": true, "data": summary})
// }

// // POST /api/v1/doctors/patients/:id/send-to-pharmacy?date=2026-02-28
// // func SendPrescriptionToExternal(c *fiber.Ctx) error {
// // 	db := database.DB
// // 	id := c.Params("id")
// // 	date := c.Query("date")
// // 	doctorID := c.Locals("doctorID").(string)

// // 	var dailyPrescriptions []model.Prescription
// // 	// Query the prescriptions table instead of the patient slice
// // 	err := db.Where("patient_id = ? AND doctorId = ? AND prescription_date = ?", id, doctorID, date).
// // 		Find(&dailyPrescriptions).Error

// // 	if err != nil || len(dailyPrescriptions) == 0 {
// // 		return c.Status(404).JSON(fiber.Map{"success": false, "message": "No prescriptions found for this date"})
// // 	}

// // 	// Logic for sending to external API would go here...

// // 	return c.JSON(fiber.Map{
// // 		"success": true,
// // 		"message": "Prescriptions for " + date + " retrieved and ready for external sync",
// // 		"count":   len(dailyPrescriptions),
// // 		"data":    dailyPrescriptions,
// // 	})
// // }