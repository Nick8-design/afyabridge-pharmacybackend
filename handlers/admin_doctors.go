package handlers

import (
	"fmt"
	"afyabridge-pharmacybackend/database"
	"afyabridge-pharmacybackend/model"
	"time"

	"github.com/gofiber/fiber/v2"
)

// GET /api/v1/doctors/admin/doctors?status=verified|pending_verification|rejected&accountStatus=active|suspended|locked
func AdminFetchAllDoctors(c *fiber.Ctx) error {
	db := database.DB

	status := c.Query("status")
	accountStatus := c.Query("accountStatus")

	q := db.Model(&model.User{}).Where("role = ?", "doctor")

	if status != "" {
		q = q.Where("verification_status = ?", status)
	}
	if accountStatus != "" {
		q = q.Where("account_status = ?", accountStatus)
	}

	var doctors []model.User
	if err := q.Order("created_at desc").Find(&doctors).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to fetch doctors",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"count":   len(doctors),
		"data":    doctors,
	})
}

// GET /api/v1/doctors/admin/doctors/:doctorId
func AdminFetchOneDoctor(c *fiber.Ctx) error {
	db := database.DB
	doctorId := c.Params("doctorId")

	var doctor model.User
	if err := db.Where("id = ? AND role = ?", doctorId, "doctor").First(&doctor).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"message": "Doctor not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    doctor,
	})
}

// PUT /api/v1/doctors/admin/doctors/:doctorId/verify
func AdminVerifyDoctor(c *fiber.Ctx) error {
	db := database.DB
	doctorId := c.Params("doctorId")

	var input struct {
		Status     string `json:"status"`
		Reason     string `json:"reason"`
		VerifiedBy string `json:"verifiedBy"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid input",
		})
	}

	allowed := map[string]bool{
		"verified":             true,
		"rejected":             true,
		"pending_verification": true,
	}

	if !allowed[input.Status] {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Invalid status value",
		})
	}

	var doctor model.User
	if err := db.Where("id = ? AND role = ?", doctorId, "doctor").First(&doctor).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"message": "Doctor not found",
		})
	}

	now := time.Now()

	doctor.VerificationStatus = &input.Status
	doctor.VerifiedAt = &now

	if input.VerifiedBy != "" {
		doctor.VerifiedBy = &input.VerifiedBy
	}

	if input.Status == "rejected" && input.Reason != "" {
		doctor.StatusReason = &input.Reason
	}

	if err := db.Save(&doctor).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update doctor",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Doctor verification updated",
		"data":    doctor,
	})
}

// PUT /api/v1/doctors/admin/doctors/:doctorId/suspend
func AdminSuspendDoctor(c *fiber.Ctx) error {
	db := database.DB
	doctorId := c.Params("doctorId")

	var input struct {
		Reason string `json:"reason"`
	}
	c.BodyParser(&input)

	var doctor model.User
	if err := db.Where("id = ? AND role = ?", doctorId, "doctor").First(&doctor).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"message": "Doctor not found",
		})
	}

	doctor.AccountStatus = "suspended"

	if input.Reason != "" {
		doctor.StatusReason = &input.Reason
	} else {
		defaultReason := "Your account is suspended. Contact support."
		doctor.StatusReason = &defaultReason
	}

	if err := db.Save(&doctor).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to suspend doctor",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Doctor suspended",
		"data":    doctor,
	})
}

// PUT /api/v1/doctors/admin/doctors/:doctorId/lock
func AdminLockDoctor(c *fiber.Ctx) error {
	db := database.DB
	doctorId := c.Params("doctorId")

	var input struct {
		Reason string `json:"reason"`
	}
	c.BodyParser(&input)

	var doctor model.User
	if err := db.Where("id = ? AND role = ?", doctorId, "doctor").First(&doctor).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"message": "Doctor not found",
		})
	}

	doctor.AccountStatus = "locked"

	if input.Reason != "" {
		doctor.StatusReason = &input.Reason
	} else {
		defaultReason := "Your account is locked. Contact support."
		doctor.StatusReason = &defaultReason
	}

	if err := db.Save(&doctor).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to lock doctor",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Doctor locked",
		"data":    doctor,
	})
}

// PUT /api/v1/doctors/admin/doctors/:doctorId/activate
func AdminActivateDoctor(c *fiber.Ctx) error {
	db := database.DB
	doctorId := c.Params("doctorId")

	var input struct {
		Reason string `json:"reason"`
	}
	c.BodyParser(&input)

	var doctor model.User
	if err := db.Where("id = ? AND role = ?", doctorId, "doctor").First(&doctor).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"message": "Doctor not found",
		})
	}

	doctor.AccountStatus = "active"

	if input.Reason != "" {
		doctor.StatusReason = &input.Reason
	} else {
		doctor.StatusReason = nil
	}

	if err := db.Save(&doctor).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to activate doctor",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Doctor activated",
		"data":    doctor,
	})
}

// DELETE /api/v1/doctors/admin/doctors/:doctorId
func AdminDeleteDoctor(c *fiber.Ctx) error {
	db := database.DB
	doctorId := c.Params("doctorId")

	result := db.Where("id = ? AND role = ?", doctorId, "doctor").Delete(&model.User{})

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Could not delete doctor",
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"message": "Doctor not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("Doctor %s deleted", doctorId),
	})
}