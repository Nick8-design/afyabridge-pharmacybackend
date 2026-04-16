package handlers

import (
	"afyabridge-pharmacybackend/database"
	"afyabridge-pharmacybackend/model"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	// "strings"

	"sync"

	"github.com/google/uuid"

	"github.com/gofiber/fiber/v2"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"time"
)

func GetAvailableGlobalOrders(c *fiber.Ctx) error {
    pharmacyID, ok := c.Locals("pharmacy_id").(string)
    if !ok {
        return c.Status(401).JSON(model.Response{Success: false, Message: "Pharmacy session missing"})
    }
    db := database.DB

    var orders []model.Order
    // 1. Fetch orders and their related prescriptions in TWO queries (Batching)
    // GORM does this efficiently behind the scenes
    // err := db.Preload("Prescription").Where("status = ?", "pending").Find(&orders).Error
	err := db.Preload("Prescription").
    Where("status = ? OR status = ?", "draft", "pending").
    Find(&orders).Error
    if err != nil {
        return c.Status(500).JSON(model.Response{Success: false, Message: "Database error"})
    }

    fulfillableOrders := make([]model.Order, 0) // Initialize to [] instead of null

    // 2. Concurrency Setup
    orderChan := make(chan model.Order, len(orders))
    resultChan := make(chan model.Order, len(orders))
    var wg sync.WaitGroup

    // Start 5 Workers
    for w := 0; w < 5; w++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for order := range orderChan {
                // Now we can access order.Prescription.Items directly!
                hasStock, _ := checkInventoryForPrescription(pharmacyID, order.Prescription.Items)
                if hasStock {
                    resultChan <- order
                }
            }
        }()
    }

    // Send orders to workers
    for _, o := range orders {
        // Only process if the prescription was actually loaded
        if o.Prescription.ID != uuid.Nil || o.PrescriptionID != "" {
            orderChan <- o
        }
    }
    close(orderChan)

    go func() {
        wg.Wait()
        close(resultChan)
    }()

    // 3. Collect results
    for o := range resultChan {
        if o.PharmacyID == pharmacyID {
            db.Model(&model.Order{}).Where("id = ?", o.ID).Update("status", "accepted")
            o.Status = "accepted"
        }
        fulfillableOrders = append(fulfillableOrders, o)
    }

    return c.JSON(model.Response{
        Success: true,
        Data:    fulfillableOrders,
    })
}

func checkInventoryForPrescription(pharmacyID string, itemsJSON datatypes.JSON) (bool, error) {
    var items []struct {
        Name     string `json:"name"`
        Quantity int    `json:"quantity"`
    }

    if err := json.Unmarshal(itemsJSON, &items); err != nil {
        return false, err
    }

    if len(items) == 0 { return false, nil }

    for _, item := range items {
        var count int64
        // Default quantity to 1 if it's not specified in the prescription JSON
        reqQty := item.Quantity
        if reqQty <= 0 { reqQty = 1 }

        // Using .Table("drugs") as per your previous logs
        err := database.DB.Table("drugs").
            Where("pharmacy_id = ? AND drug_name = ? AND quantity_in_stock >= ?",
                  pharmacyID, item.Name, reqQty).
            Count(&count).Error

        if err != nil || count == 0 {
            return false, nil
        }
    }
    return true, nil
}

// func GetAvailableGlobalOrders(c *fiber.Ctx) error {
// 	pharmacyID, ok := c.Locals("pharmacy_id").(string)
// 	if !ok {
// 		return c.Status(401).JSON(model.Response{Success: false, Message: "Pharmacy session missing"})
// 	}
// 	db := database.DB

// 	var orders []model.Order
// 	// 1. Batch fetch pending orders
// 	if err := db.Where("status = ?", "pending").Find(&orders).Error; err != nil {
// 		return c.Status(500).JSON(model.Response{Success: false, Message: "Database error"})
// 	}

// 	if len(orders) == 0 {
// 		return c.JSON(model.Response{Success: true, Data: []model.Order{}})
// 	}

// 	// 2. Extract unique Prescription IDs
// 	prescriptionIDs := []string{}
// 	for _, o := range orders {
// 		if o.PrescriptionID != "" {
// 			prescriptionIDs = append(prescriptionIDs, o.PrescriptionID)
// 		}
// 	}

// 	// 3. Batch fetch all Prescriptions to avoid N+1 query problem
// 	var prescriptions []model.Prescription
// 	db.Where("id IN ?", prescriptionIDs).Find(&prescriptions)

// 	// Map Items for O(1) lookup
// 	prescriptionMap := make(map[string]datatypes.JSON)
//     for _, p := range prescriptions {
//         // Normalize: remove dashes for consistent matching
//         idClean := strings.ReplaceAll(p.ID.String(), "-", "")
//         prescriptionMap[idClean] = p.Items
//     }

// 	// 4. Concurrency Setup (Worker Pool)
// 	type checkTask struct {
// 		Order model.Order
// 		Items datatypes.JSON
// 	}
// 	taskChan := make(chan checkTask, len(orders))
// 	resultChan := make(chan model.Order, len(orders))
// 	var wg sync.WaitGroup

// 	// Start 5 Workers
// 	for w := 0; w < 5; w++ {
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			for task := range taskChan {
// 				// We call the inventory check for each order in parallel
// 				hasStock, _ := checkInventoryForPrescription(pharmacyID, task.Items)
// 				if hasStock {
// 					resultChan <- task.Order
// 				}
// 			}
// 		}()
// 	}

// 	// Send tasks to workers
// 	for _, o := range orders {
// 		orderPIDClean := strings.ReplaceAll(o.PrescriptionID, "-", "")
// 		if items, found := prescriptionMap[orderPIDClean]; found {
//             taskChan <- checkTask{Order: o, Items: items}
//         }
// 	}
// 	close(taskChan)

// 	// Close results channel once all workers are done
// 	go func() {
// 		wg.Wait()
// 		close(resultChan)
// 	}()

// 	// 5. Collect results
// 	fulfillableOrders := make([]model.Order, 0)
// 	for o := range resultChan {
// 		// Auto-accept logic: if the order was specifically for this pharmacy
// 		if o.PharmacyID == pharmacyID {
// 			db.Model(&model.Order{}).Where("id = ?", o.ID).Update("status", "accepted")
// 			o.Status = "accepted"
// 		}
// 		fulfillableOrders = append(fulfillableOrders, o)
// 	}

// 	return c.JSON(model.Response{
// 		Success: true,
// 		Data:    fulfillableOrders,
// 	})
// }



// func checkInventoryForPrescription(pharmacyID string, itemsJSON datatypes.JSON) (bool, error) {
//     var items []struct {
//         Name     string `json:"name"`
//         Quantity int    `json:"quantity"` 
//     }

//     if err := json.Unmarshal(itemsJSON, &items); err != nil {
//         return false, err
//     }

//     for _, item := range items {
//         // If quantity is not in JSON, it defaults to 0. We should check for at least 1.
//         reqQty := item.Quantity
//         if reqQty <= 0 {
//             reqQty = 1 
//         }

//         var count int64
//         // Note: Check if your table is 'inventories' or 'drugs' based on your logs
//         err := database.DB.Table("drugs"). // Use correct table name here
//             Where("pharmacy_id = ? AND drug_name = ? AND quantity_in_stock >= ?", 
//                   pharmacyID, item.Name, reqQty).
//             Count(&count).Error

//         if err != nil || count == 0 {
//             return false, nil 
//         }
//     }
//     return  len(items) > 0,nil
// }



// func checkInventoryForPrescription(pharmacyID string, itemsJSON datatypes.JSON) (bool, error) {
// 	// Anonymous struct to match your specific JSON keys
// 	var items []struct {
// 		Name     string `json:"name"`     // Matches "Vitamin C 1000mg"
// 		Quantity int    `json:"quantity"` // Make sure your JSON has quantity
// 	}

// 	if err := json.Unmarshal(itemsJSON, &items); err != nil {
// 		return false, err
// 	}

// 	// If the prescription has no items, we can't fulfill it
// 	// if len(items) == 0 {
// 	// 	return false, nil
// 	// }

// 	for _, item := range items {
// 		var count int64
// 		// Verify if THIS pharmacy has enough stock of this specific drug name
// 		err := database.DB.Model(&model.Inventory{}).
// 			Where("pharmacy_id = ? AND drug_name = ? AND quantity_in_stock >= ?", 
//                   pharmacyID, item.Name, item.Quantity).
// 			Count(&count).Error

// 		if err != nil || count == 0 {
// 			return false, nil // One missing item fails the whole order check
// 		}
// 	}
// 	return true, nil
// }



// func GetAvailableGlobalOrders(c *fiber.Ctx) error {
//     pharmacyID, ok := c.Locals("pharmacy_id").(string)
//     if !ok {
//         return c.Status(401).JSON(model.Response{Success: false, Message: "Pharmacy session missing"})
//     }
//     db := database.DB

//     var orders []model.Order
//     // 1. Fetch pending orders (using your existing table structure)
//     if err := db.Where("status = ?", "pending").Find(&orders).Error; err != nil {
//         return c.Status(500).JSON(model.Response{Success: false, Message: "Database error"})
//     }

//     if len(orders) == 0 {
//         return c.JSON(model.Response{Success: true, Data: []model.Order{}})
//     }

//     // 2. Extract all unique Prescription IDs from the orders
//     prescriptionIDs := []string{}
//     for _, o := range orders {
//         if o.PrescriptionID != "" {
//             prescriptionIDs = append(prescriptionIDs, o.PrescriptionID)
//         }
//     }

//     // 3. Fetch all relevant prescriptions in ONE batch query (Fast!)
//     var prescriptions []model.Prescription
//     db.Where("id IN ?", prescriptionIDs).Find(&prescriptions)

//     // 4. Create a Map for O(1) lookup: ID -> Prescription Items
//     prescriptionMap := make(map[string]datatypes.JSON)
//     for _, p := range prescriptions {
//         prescriptionMap[p.ID.String()] = p.Items
//     }

//     // 5. Use Channels/Workers to check inventory
//     type checkTask struct {
//         Order model.Order
//         Items datatypes.JSON
//     }
//     taskChan := make(chan checkTask, len(orders))
//     resultChan := make(chan model.Order, len(orders))
//     var wg sync.WaitGroup

//     for w := 0; w < 5; w++ { // 5 Workers
//         wg.Add(1)
//         go func() {
//             defer wg.Done()
//             for task := range taskChan {
//                 hasStock, _ := checkInventoryForPrescription(pharmacyID, task.Items)
//                 if hasStock {
//                     resultChan <- task.Order
//                 }
//             }
//         }()
//     }

//     for _, o := range orders {
//         if items, found := prescriptionMap[o.PrescriptionID]; found {
//             taskChan <- checkTask{Order: o, Items: items}
//         }
//     }
//     close(taskChan)

//     go func() {
//         wg.Wait()
//         close(resultChan)
//     }()

//     var fulfillableOrders []model.Order
//     for o := range resultChan {
//         if o.PharmacyID == pharmacyID {
//             db.Model(&model.Order{}).Where("id = ?", o.ID).Update("status", "accepted")
//             o.Status = "accepted"
//         }
//         fulfillableOrders = append(fulfillableOrders, o)
//     }

//     return c.JSON(model.Response{
//         Success: true,
//         Data:    fulfillableOrders,
//     })
// }

// func checkInventoryForPrescription(pharmacyID string, itemsJSON datatypes.JSON) (bool, error) {
// 	var items []struct {
// 		Name     string `json:"name"` // Matches your CSV "Coartem", "Paracetamol"
// 		Quantity int    `json:"quantity"`
// 	}

// 	if err := json.Unmarshal(itemsJSON, &items); err != nil {
// 		return false, err
// 	}

// 	for _, item := range items {
// 		var count int64
// 		// Verify stock in your Inventory table
// 		err := database.DB.Model(&model.Inventory{}).
// 			Where("pharmacy_id = ? AND drug_name = ? AND quantity_in_stock >= ?", 
//                   pharmacyID, item.Name, item.Quantity).
// 			Count(&count).Error

// 		if err != nil || count == 0 {
// 			return false, nil 
// 		}
// 	}
// 	return true, nil
// }




// func GetAvailableGlobalOrders(c *fiber.Ctx) error {
// 	pharmacyID, ok := c.Locals("pharmacy_id").(string)
// 	if !ok {
// 		return c.Status(401).JSON(model.Response{Success: false, Message: "Pharmacy session missing"})
// 	}
// 	db := database.DB

// 	var orders []model.Order
// 	// Preload Prescription to get Items directly
// 	err := db.Preload("Prescription").Where("status = ?", "pending").Find(&orders).Error
// 	if err != nil {
// 		return c.Status(500).JSON(model.Response{Success: false, Message: "Database error"})
// 	}

// 	orderChan := make(chan model.Order, len(orders))
// 	resultChan := make(chan model.Order, len(orders))
// 	var wg sync.WaitGroup

// 	// Create 5 workers (adjust based on your CPU/Database capacity)
// 	for w := 1; w <= 5; w++ {
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			for order := range orderChan {
// 				// Check inventory concurrently
// 				hasStock, _ := checkInventoryForPrescription(pharmacyID, order.Prescription.Items)
// 				if hasStock {
// 					resultChan <- order
// 				}
// 			}
// 		}()
// 	}

// 	// Send orders to workers
// 	for _, o := range orders {
// 		orderChan <- o
// 	}
// 	close(orderChan)

// 	// Wait for workers and close results
// 	go func() {
// 		wg.Wait()
// 		close(resultChan)
// 	}()

// 	var fulfillableOrders []model.Order
// 	for o := range resultChan {
// 		// Auto-accept logic
// 		if o.PharmacyID == pharmacyID {
// 			db.Model(&o).Update("status", "accepted")
// 			o.Status = "accepted"
// 		}
// 		fulfillableOrders = append(fulfillableOrders)
// 	}

// 	return c.JSON(model.Response{
// 		Success: true,
// 		Message: fmt.Sprintf("Processed %d orders", len(orders)),
// 		Data:    fulfillableOrders,
// 	})
// }




type DrugItem struct {
    DrugName  string `json:"name"`
    Dosage    string `json:"dosage"`
    Frequency string `json:"frequency"` // Added to capture "Twice daily"
    Duration  string `json:"duration"`
    Quantity  int    `json:"quantity"`
}
func safeFloat64(val *float64) float64 {
    if val == nil {
        return 0.0
    }
    return *val
}

func ServeOrder(c *fiber.Ctx) error {
	orderID := c.Params("order_id")
	pharmacyID := c.Locals("pharmacy_id").(string)
	db := database.DB

	// 1. Fetch Order with Patient context
	var order model.Order
	if err := db.First(&order, "id = ?", orderID).Error; err != nil {
		return c.Status(404).JSON(model.Response{Success: false, Message: "Order not found"})
	}

	// 2. Fetch Prescription
	var prescription model.Prescription
	if err := db.First(&prescription, "id = ?", order.PrescriptionID).Error; err != nil {
		return c.Status(404).JSON(model.Response{Success: false, Message: "Prescription not found"})
	}

	// 3. Fetch Pharmacy details (for Pickup Info)
	// var pharmacy model.Pharmacy // Assuming you have a Pharmacy model
	// db.First(&pharmacy, "id = ?", pharmacyID)

	// 3. Fetch Pharmacy details
var pharmacy model.Pharmacy
if err := db.First(&pharmacy, "id = ?", pharmacyID).Error; err != nil {
    return c.Status(404).JSON(model.Response{
        Success: false, 
        Message: "Pharmacy profile incomplete or not found",
    })
}

	var items []DrugItem
	json.Unmarshal(prescription.Items, &items)

	err := db.Transaction(func(tx *gorm.DB) error {
		// --- [Inventory & Medication Logic from previous steps goes here] ---
		// (Assume inventory is deducted and patient_medications created)

		// 4. Create or Update Comprehensive Delivery Record
		var delivery model.Delivery
		err := tx.Where("order_id = ?", order.ID).First(&delivery).Error

		if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
			now := time.Now()
			
			// Generate 6-digit OTP
			otp := fmt.Sprintf("%06d", rand.Intn(1000000))

			newDelivery := model.Delivery{
				ID:            uuid.New().String(),
				PackageNumber: "PKG-" + now.Format("20060102") + "-" + uuid.New().String()[:5],
				OrderID:       order.ID,
				RiderID:       nil, // Unassigned
				Status:        "pending",
				AcceptStatus:  false,

				// Pickup Details (From Pharmacy)
				PickupLocation: pharmacy.AddressLine1,
PickupLat:      safeFloat64(pharmacy.GpsLat), // Use a helper to avoid panic
PickupLng:      safeFloat64(pharmacy.GpsLng), // Use a helper to avoid panic
PickupContact:  pharmacy.Phone,

				// Dropoff Details (From Order/Patient)
				DropoffLocation: order.PatientAddress,

				DropoffLat : safeFloat64(&order.PatientLat)     ,
	            DropoffLng :safeFloat64(&order.PatientLng),
				ReceiverContact:  order.PatientPhone,
				// DropoffLat/Lng should come from your patient's saved geocoding if available
				
				// Requirements & Logistics
				Requirement:           "Handle with care - Temperature sensitive",
				EstimatedDeliveryTime: "30-60 mins",
				Distance:              0.0, // Should be calculated via Maps API
				Charges:               150.00, // Fixed or calculated
				DeliveryZone:          "Nairobi Central",
				DeliveryNotes:         "Contact patient upon arrival",
				
				// Security
				OtpCode:               otp,
				
				// Verification Flags
				PackageSealed:        true,
				LabeledCorrectly:     true,
				VerifiedWithPharmacy: true,

				// Timestamps
				DateApproved: &now,
				CreatedAt:    now,
				UpdatedAt:    now,
			}

			if err := tx.Create(&newDelivery).Error; err != nil {
				return err
			}
		} else if err == nil {
			// Update existing record to 'pending' if it was failed/cancelled before
			tx.Model(&delivery).Updates(map[string]interface{}{
				"status":     "pending",
				"updated_at": time.Now(),
			})
		}

		// 5. Update Order Status
		return tx.Model(&order).Updates(map[string]interface{}{
			"status":      "ready",
			"pharmacy_id": pharmacyID,
		}).Error
	})

	if err != nil {
		return c.Status(400).JSON(model.Response{Success: false, Message: err.Error()})
	}

	return c.JSON(model.Response{
		Success: true, 
		Message: "Order served and delivery record initiated",
	})
}

// POST /orders/:order_id/serve
// func ServeOrder(c *fiber.Ctx) error {
//     orderID := c.Params("order_id")
//     pharmacyID := c.Locals("pharmacy_id").(string)
//     db := database.DB

//     var order model.Order
//     if err := db.First(&order, "id = ?", orderID).Error; err != nil {
//         return c.Status(404).JSON(model.Response{Success: false, Message: "Order not found"})
//     }

//     var prescription model.Prescription
//     if err := db.First(&prescription, "id = ?", order.PrescriptionID).Error; err != nil {
//         return c.Status(404).JSON(model.Response{Success: false, Message: "Prescription not found"})
//     }

//     var items []DrugItem
//     if err := json.Unmarshal(prescription.Items, &items); err != nil {
//         return c.Status(500).JSON(model.Response{Success: false, Message: "Failed to parse prescription items"})
//     }

//     err := db.Transaction(func(tx *gorm.DB) error {
//         // 1. Update Order Status
//         if err := tx.Model(&order).Updates(map[string]interface{}{
//             "status":      "ready",
//             "pharmacy_id": pharmacyID,
//         }).Error; err != nil {
//             return err
//         }

//         prescriptionIDStr := prescription.ID.String()

//         // 2. Process Inventory and Medications
// 		// ... (previous setup code)

// // 2. Process Inventory and Medications
// for _, item := range items {
//     qtyToDeduct := item.Quantity
//     if qtyToDeduct <= 0 {
//         qtyToDeduct = 1
//     }

//     var drug model.Inventory
//     if err := tx.Table("drugs").Where("pharmacy_id = ? AND drug_name = ?", pharmacyID, item.DrugName).First(&drug).Error; err != nil {
//         return fmt.Errorf("drug %s not found in your inventory", item.DrugName)
//     }

//     if drug.QuantityInStock < qtyToDeduct {
//         return fmt.Errorf("insufficient stock for %s", item.DrugName)
//     }

//     if err := tx.Table("drugs").Where("id = ?", drug.ID).Update("quantity_in_stock", gorm.Expr("quantity_in_stock - ?", qtyToDeduct)).Error; err != nil {
//         return err
//     }

//     // FIX: Provide valid defaults for the Medication table ENUMs
//     newMed := model.Medication{
//         ID:                uuid.New().String(),
//         PatientID:         order.PatientID,
//         PrescriptionID:    &prescriptionIDStr,
//         PharmacyID:        &pharmacyID,
//         DrugName:          item.DrugName,
//         Dosage:            item.Dosage,
//         // Match these strings to your database ENUM options (e.g., 'tablet', 'syrup', 'capsule')
//         // If unsure, check your MySQL/TiDB schema for `dosage_form`
//         DosageForm:        "tablet", 
//         Frequency:         item.Frequency, 
//         QuantityDispensed: qtyToDeduct,
//         Status:            "active",
//         StartDate:         time.Now(),
//         CreatedAt:         time.Now(),
//         UpdatedAt:         time.Now(),
//     }
    
//     if err := tx.Create(&newMed).Error; err != nil {
//         return err // This is where the 1265 error triggers
//     }
// }

// // ... (remaining code)
//         // for _, item := range items {
//         //     // FIX: If quantity is missing in JSON (0), default to 1 unit
//         //     qtyToDeduct := item.Quantity
//         //     if qtyToDeduct <= 0 {
//         //         qtyToDeduct = 1
//         //     }

//         //     var drug model.Inventory
//         //     // Use the correct table name 'drugs' as per your TiDB logs
//         //     if err := tx.Table("drugs").Where("pharmacy_id = ? AND drug_name = ?", pharmacyID, item.DrugName).First(&drug).Error; err != nil {
//         //         return fmt.Errorf("drug %s not found in your inventory", item.DrugName)
//         //     }

//         //     if drug.QuantityInStock < qtyToDeduct {
//         //         return fmt.Errorf("insufficient stock for %s (Available: %d)", item.DrugName, drug.QuantityInStock)
//         //     }

//         //     // Update Stock
//         //     if err := tx.Table("drugs").Where("id = ?", drug.ID).Update("quantity_in_stock", gorm.Expr("quantity_in_stock - ?", qtyToDeduct)).Error; err != nil {
//         //         return err
//         //     }

//         //     // Add to Patient Medication Log
//         //     newMed := model.Medication{
//         //         ID:                uuid.New().String(),
//         //         PatientID:         order.PatientID,
//         //         PrescriptionID:    &prescriptionIDStr,
//         //         PharmacyID:        &pharmacyID,
//         //         DrugName:          item.DrugName,
//         //         Dosage:            item.Dosage,
//         //         QuantityDispensed: qtyToDeduct,
//         //         Status:            "active",
//         //         StartDate:         time.Now(),
//         //         CreatedAt:         time.Now(),
//         //         UpdatedAt:         time.Now(),
//         //     }
//         //     if err := tx.Create(&newMed).Error; err != nil {
//         //         return err
//         //     }
//         // }

//         // 3. Create Delivery Record
//        // 3. Create or Update Delivery Record
// var delivery model.Delivery
// // Check if a delivery already exists for this order
// err := tx.Where("order_id = ?", order.ID).First(&delivery).Error

// if err != nil && err == gorm.ErrRecordNotFound {
//     // No delivery exists, create a new one
//     newDelivery := model.Delivery{
//         ID:            uuid.New().String(),
//         PackageNumber: "PKG-" + time.Now().Format("20060102") + "-" + uuid.New().String()[:5],
//         OrderID:       order.ID,
//         Status:        "pending",
//         CreatedAt:     time.Now(),
//         UpdatedAt:     time.Now(),
//     }
//     if err := tx.Create(&newDelivery).Error; err != nil {
//         return err
//     }
// } else if err == nil {
//     // Delivery already exists, just update the status/timestamp if needed
//     if err := tx.Model(&delivery).Update("updated_at", time.Now()).Error; err != nil {
//         return err
//     }
// } else {
//     return err
// }

// return nil
//     })

//     if err != nil {
//         return c.Status(400).JSON(model.Response{Success: false, Message: err.Error()})
//     }

//     return c.JSON(model.Response{
//         Success: true,
//         Message: "Order served successfully. Inventory updated.",
//     })
// }

// checkInventoryForPrescription implements the logic to verify if a pharmacy can fulfill an entire order



// POST /orders/:order_id/serve
// func ServeOrder(c *fiber.Ctx) error {
//     orderID := c.Params("order_id")
//     pharmacyID := c.Locals("pharmacy_id").(string)
//     db := database.DB

//     var order model.Order
//     if err := db.First(&order, "id = ?", orderID).Error; err != nil {
//         return c.Status(404).JSON(model.Response{Success: false, Message: "Order not found"})
//     }

//     var prescription model.Prescription
//     if err := db.First(&prescription, "id = ?", order.PrescriptionID).Error; err != nil {
//         return c.Status(404).JSON(model.Response{Success: false, Message: "Prescription not found"})
//     }

//     // Use a transaction to ensure all-or-nothing execution
//     err := db.Transaction(func(tx *gorm.DB) error {
        
//         // 1. Update Order Status and Link to fulfilling Pharmacy
//         if err := tx.Model(&order).Updates(map[string]interface{}{
//             "status":      "ready",
//             "pharmacy_id": pharmacyID,
//         }).Error; err != nil {
//             return err
//         }

//         // 2. Process Items (Assuming Items JSON is a list of drug names and quantities)
//         // Note: You may need to parse prescription.Items based on your specific JSON structure
//         var items []struct {
//             DrugName string `json:"drug_name"`
//             Quantity int    `json:"quantity"`
//         }
//         // ... (Unmarshal prescription.Items into items) ...

//         for _, item := range items {
//             // A. Remove from Pharmacy Inventory
//             var drug model.Inventory
//             if err := tx.Where("pharmacy_id = ? AND drug_name = ?", pharmacyID, item.DrugName).First(&drug).Error; err != nil {
//                 return fmt.Errorf("drug %s not found in inventory", item.DrugName)
//             }

//             if drug.QuantityInStock < item.Quantity {
//                 return fmt.Errorf("insufficient stock for %s", item.DrugName)
//             }

//             if err := tx.Model(&drug).Update("quantity_in_stock", drug.QuantityInStock-item.Quantity).Error; err != nil {
//                 return err
//             }

//             // B. Add to Patient Inventory (Medication table)
//             newMed := model.Medication{
//                 ID:                uuid.New().String(),
//                 PatientID:         order.PatientID,
//                 PrescriptionID:    &prescription.ID,
//                 PharmacyID:        &pharmacyID,
//                 DrugName:          item.DrugName,
//                 QuantityDispensed: item.Quantity,
//                 Status:            "active",
//                 StartDate:         time.Now(),
//                 CreatedAt:         time.Now(),
//             }
//             if err := tx.Create(&newMed).Error; err != nil {
//                 return err
//             }
//         }

//         // 3. Create Delivery Record (Unassigned rider as requested)
//         delivery := model.Delivery{
//             ID:            uuid.New().String(),
//             PackageNumber: "PKG-" + time.Now().Format("20060102") + "-" + order.ID[:4],
//             OrderID:       order.ID,
//             Status:        "pending", // Unassigned
//             RiderID:       nil,       // Leaves it for any rider to see
//             CreatedAt:     time.Now(),
//         }

//         if err := tx.Create(&delivery).Error; err != nil {
//             return err
//         }

//         return nil
//     })

//     if err != nil {
//         return c.Status(500).JSON(model.Response{Success: false, Message: err.Error()})
//     }

//     return c.JSON(model.Response{Success: true, Message: "Order served and sent to delivery"})
// }

// // Helper function to check stock
// func checkInventoryForPrescription(pharmacyID string, itemsJSON datatypes.JSON) (bool, error) {
//     // Implement parsing logic here to verify if 
//     // database.DB.Where("pharmacy_id = ? AND drug_name = ? AND quantity_in_stock >= ?", ...)
//     return true, nil 
// }








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