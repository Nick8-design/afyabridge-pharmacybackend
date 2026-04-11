package middleware

import (
	"afyabridge-pharmacybackend/database"
	"afyabridge-pharmacybackend/model"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Protected verifies the JWT and sets the user context
func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(401).JSON(model.Response{Success: false, Message: "Unauthorized"})
		}

		tokenString := authHeader[7:]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			return c.Status(401).JSON(model.Response{Success: false, Message: "Invalid token"})
		}

		claims := token.Claims.(jwt.MapClaims)
		c.Locals("userID", claims["user_id"])
		c.Locals("role", claims["role"])
		return c.Next()
	}
}

// PharmacyServiceGuard ensures the user is a verified pharmacist and account is active
func PharmacyServiceGuard() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("userID").(string)
		if !ok || userID == "" {
			return c.Status(401).JSON(model.Response{Success: false, Message: "Unauthorized"})
		}

		var user model.User
		err := database.DB.Where("id = ?", userID).First(&user).Error

		if err != nil {
			return c.Status(401).JSON(model.Response{Success: false, Message: "User not found"})
		}

		// Role Check
		if user.Role != "pharmacist" {
			return c.Status(403).JSON(model.Response{Success: false, Message: "Access denied: Pharmacists only"})
		}

		// Account Status Check
		if user.AccountStatus == "suspended" || user.AccountStatus == "locked" {
			return c.Status(403).JSON(model.Response{
				Success: false,
				Message: "Your account is " + user.AccountStatus + ". Contact support.",
			})
		}

		// Verification Check
		if user.VerificationStatus == nil || *user.VerificationStatus != "verified" {
			return c.Status(403).JSON(model.Response{
				Success: false,
				Message: "Your account is pending verification.",
			})
		}

		return c.Next()
	}
}