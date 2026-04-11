package handlers

import (
	"afyabridge-pharmacybackend/database"
	"afyabridge-pharmacybackend/model"
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func Login(c *fiber.Ctx) error {
	fmt.Println("JWT_SECRET:", os.Getenv("JWT_SECRET"))
	type LoginInput struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(model.Response{Success: false, Message: "Bad request"})
	}

	var user model.User
	if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		return c.Status(401).JSON(model.Response{Success: false, Message: "Invalid email or password"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return c.Status(401).JSON(model.Response{Success: false, Message: "Invalid email or password"})
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 1).Unix(), // 60 mins as per docs
	})

	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return c.Status(500).JSON(model.Response{Success: false, Message: "Token generation failed"})
	}

	return c.Status(200).JSON(model.Response{
		Success: true,
		Message: "Login successful",
		Data: fiber.Map{
			"access_token":  t,
			"refresh_token": "ref-" + t[:10], // Placeholder refresh logic
			"user":          user,
		},
	})
}

func RegisterComplete(c *fiber.Ctx) error {
	fmt.Printf("Someone is registering\n")

	// 1. PRE-CHECK: Validate email uniqueness before starting transaction
	pharmacistEmail := c.FormValue("pharmacist_email")
	var existingUser model.User
	if err := database.DB.Where("email = ?", pharmacistEmail).First(&existingUser).Error; err == nil {
		return c.Status(400).JSON(model.Response{
			Success: false,
			Message: "Registration failed",
			Errors:  fiber.Map{"pharmacist_email": []string{"A user with this email already exists."}},
		})
	}

	// 2. PRE-CHECK: Validate Business Email uniqueness
	businessEmail := c.FormValue("business_email")
	var existingPharmacy model.Pharmacy
	if err := database.DB.Where("email = ?", businessEmail).First(&existingPharmacy).Error; err == nil {
		return c.Status(400).JSON(model.Response{
			Success: false,
			Message: "Registration failed",
			Errors:  fiber.Map{"business_email": []string{"A pharmacy with this business email is already registered."}},
		})
	}

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 3. Setup IDs
	pharmacyID := uuid.New().String()
	userID := uuid.New().String()
	regID := uuid.New().String()

	// 4. Parse Dates
	licenseExpiry, _ := time.Parse("2006-01-02", c.FormValue("license_expiry"))
	practicingExpiry, _ := time.Parse("2006-01-02", c.FormValue("practicing_expiry"))

	// 5. Create Pharmacy
	pharmacy := model.Pharmacy{
		ID:            pharmacyID,
		Name:          c.FormValue("pharmacy_name_legal"),
		Email:         businessEmail,
		Phone:         c.FormValue("business_phone"),
		AddressLine1:  c.FormValue("physical_address"),
		County:        c.FormValue("county"),
		LicenseNumber: c.FormValue("ppb_license_no"),
		LicenseExpiry: licenseExpiry,
		IsActive:      true,
	}

	if err := tx.Create(&pharmacy).Error; err != nil {
		tx.Rollback()
		return c.Status(400).JSON(model.Response{Success: false, Message: "Database error creating pharmacy"})
	}

	// 6. Create User
	passHash, _ := bcrypt.GenerateFromPassword([]byte(c.FormValue("password")), 14)
	user := model.User{
		ID:            userID,
		FullName:      c.FormValue("pharmacist_name"),
		Email:         pharmacistEmail,
		PasswordHash:  string(passHash),
		Role:          "pharmacist",
		PharmacyID:    &pharmacyID,
		AccountStatus: "active",
		IsActive:      true,
		PhoneNumber:   ptrString(c.FormValue("pharmacist_phone")),
		NationalID:    ptrString(c.FormValue("id_or_passport_no")),
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return c.Status(400).JSON(model.Response{Success: false, Message: "Database error creating pharmacist account"})
	}

	// 7. Create Audit Registration Record
	registration := model.PharmacyRegistration{
		ID:                   regID,
		PharmacyNameLegal:    c.FormValue("pharmacy_name_legal"),
		BusinessRegNo:        c.FormValue("business_reg_no"),
		KraPin:               c.FormValue("kra_pin"),
		PpbLicenseNo:         c.FormValue("ppb_license_no"),
		LicenseExpiry:        licenseExpiry,
		County:               c.FormValue("county"),
		PhysicalAddress:      c.FormValue("physical_address"),
		BusinessPhone:        c.FormValue("business_phone"),
		BusinessEmail:        businessEmail,
		PharmacistName:       c.FormValue("pharmacist_name"),
		IdOrPassportNo:       c.FormValue("id_or_passport_no"),
		PharmacistRegNo:      c.FormValue("pharmacist_reg_no"),
		PracticingLicense:    c.FormValue("practicing_license"),
		PracticingExpiry:     &practicingExpiry,
		PharmacistPhone:      c.FormValue("pharmacist_phone"),
		PharmacistEmail:      pharmacistEmail,
		IdDocument:           c.FormValue("id_document"),
		PracticingLicenseDoc: c.FormValue("practicing_license_doc"),
		OperatingLicenseDoc:  c.FormValue("operating_license_doc"),
		BusinessRegCert:      c.FormValue("business_reg_cert"),
		KraPinCert:           c.FormValue("kra_pin_cert"),
		ProofOfAddressDoc:    c.FormValue("proof_of_address_doc"),
		Status:               "submitted",
		SubmittedAt:          ptrTime(time.Now()),
	}

	if err := tx.Create(&registration).Error; err != nil {
		tx.Rollback()
		return c.Status(400).JSON(model.Response{Success: false, Message: "Database error saving registration audit"})
	}

	tx.Commit()

	return c.Status(201).JSON(model.Response{
		Success: true,
		Message: "Pharmacy registered successfully. Welcome to AfyaBridge.",
		Data: fiber.Map{
			"user": user,
		},
	})
}

// Helper functions for pointers
func ptrString(s string) *string     { return &s }
func ptrTime(t time.Time) *time.Time { return &t }

// Logout - Invalidates the session (Frontend should delete token; backend can blacklist)
func Logout(c *fiber.Ctx) error {
	return c.Status(200).JSON(model.Response{
		Success: true,
		Message: "Logged out successfully",
	})
}

// ChangePassword - Updates password after verifying current one
func ChangePassword(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	var input struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}

	if err := c.BodyParser(&input); err != nil || len(input.NewPassword) < 8 {
		return c.Status(400).JSON(model.Response{Success: false, Message: "Valid new password (min 8 chars) required"})
	}

	var user model.User
	database.DB.First(&user, "id = ?", userID)

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.CurrentPassword)); err != nil {
		return c.Status(400).JSON(model.Response{Success: false, Message: "Current password is incorrect"})
	}

	newHash, _ := bcrypt.GenerateFromPassword([]byte(input.NewPassword), 14)
	database.DB.Model(&user).Update("password_hash", string(newHash))

	return c.Status(200).JSON(model.Response{Success: true, Message: "Password updated successfully"})
}

// ForgotPassword - Initiates reset flow
func ForgotPassword(c *fiber.Ctx) error {
	var input struct {
		Email string `json:"email"`
	}
	c.BodyParser(&input)

	// In production, generate a real token and send an actual email here
	token := uuid.New().String()
	resetURL := fmt.Sprintf("http://localhost:5173/auth/resetPassword?token=%s", token)

	return c.Status(200).JSON(model.Response{
		Success: true,
		Message: "If an account exists for that email, a reset link has been sent",
		Data: fiber.Map{
			"token":     token,
			"reset_url": resetURL,
		},
	})
}

// SendOTP - Mock implementation for sending email/phone OTP
func SendOTP(c *fiber.Ctx) error {
	return c.Status(200).JSON(model.Response{
		Success: true,
		Message: "OTP sent to your email",
	})
}

// VerifyOTP - Validates 6-digit code
func VerifyOTP(c *fiber.Ctx) error {
	var input struct {
		Phone   string `json:"phone"`
		OTPCode string `json:"otp_code"`
	}
	c.BodyParser(&input)

	if input.OTPCode == "123456" { // Mock verification logic
		return c.Status(200).JSON(model.Response{Success: true, Message: "OTP verified successfully"})
	}
	return c.Status(400).JSON(model.Response{Success: false, Message: "Invalid or expired OTP code"})
}

// GetProfile - Return authenticated user's profile with nested pharmacy
func GetProfile(c *fiber.Ctx) error {
    userID, ok := c.Locals("user_id").(string) // Ensure this matches exactly
    if !ok {
         return c.Status(401).JSON(model.Response{Success: false, Message: "Invalid session"})
    }
    
    var user model.User
    if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
        return c.Status(404).JSON(model.Response{Success: false, Message: "User not found"})
    }

    return c.Status(200).JSON(model.Response{
        Success: true,
        Data:    user,
    })
}

// UpdateProfile - Partial update of user details
func UpdateProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	var input struct {
		FullName    string `json:"full_name"`
		PhoneNumber string `json:"phone_number"`
		Bio         string `json:"bio"`
		Gender      string `json:"gender"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(model.Response{Success: false, Message: "Invalid input"})
	}

	var user model.User
	if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
		return c.Status(404).JSON(model.Response{Success: false, Message: "User not found"})
	}

	// Apply updates to specific fields
	database.DB.Model(&user).Updates(input)

	return c.Status(200).JSON(model.Response{
		Success: true,
		Message: "Profile updated",
		Data:    user,
	})
}

// UpdatePhoto - Updates profile_image using a provided link (JSON)
func UpdatePhoto(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	var input struct {
		PhotoURL string `json:"photo_url"`
	}

	if err := c.BodyParser(&input); err != nil || input.PhotoURL == "" {
		return c.Status(400).JSON(model.Response{Success: false, Message: "Valid photo_url is required"})
	}

	if err := database.DB.Model(&model.User{}).Where("id = ?", userID).Update("profile_image", input.PhotoURL).Error; err != nil {
		return c.Status(500).JSON(model.Response{Success: false, Message: "Failed to update photo link"})
	}

	// Fetch updated user to return fresh data
	var user model.User
	database.DB.First(&user, "id = ?", userID)

	return c.Status(200).JSON(model.Response{
		Success: true,
		Message: "Photo updated",
		Data:    user,
	})
}

// DeletePhoto - Removes the user's profile photo link
func DeletePhoto(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	if err := database.DB.Model(&model.User{}).Where("id = ?", userID).Update("profile_image", nil).Error; err != nil {
		return c.Status(500).JSON(model.Response{Success: false, Message: "Failed to remove photo"})
	}

	return c.Status(200).JSON(model.Response{
		Success: true,
		Message: "Photo removed",
		Data:    nil,
	})
}

// ResetPassword - Set a new password using a token (Public)
func ResetPassword(c *fiber.Ctx) error {
	var input struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&input); err != nil || len(input.Password) < 8 {
		return c.Status(400).JSON(model.Response{Success: false, Message: "Invalid token or password too short"})
	}

	// Logic: In production, verify the token against a password_resets table
	// For now, we will assume a valid token for the demo
	passHash, _ := bcrypt.GenerateFromPassword([]byte(input.Password), 14)

	// Mock: finding user by token. In reality, join with a reset_tokens table.
	if err := database.DB.Model(&model.User{}).Where("email = ?", "pedinick@afyabridge.com").Update("password_hash", string(passHash)).Error; err != nil {
		return c.Status(400).JSON(model.Response{Success: false, Message: "This reset link is invalid or has expired."})
	}

	return c.Status(200).JSON(model.Response{
		Success: true,
		Message: "Password reset successfully. You can now log in.",
	})
}
