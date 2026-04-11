package middleware

import (
	"afyabridge-pharmacybackend/database"
	"afyabridge-pharmacybackend/model"
	"fmt"

	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)



func Protected() fiber.Handler {
    return func(c *fiber.Ctx) error {
        authHeader := c.Get("Authorization")
        if authHeader == "" || len(authHeader) < 8 {
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

        // FIX 1: always convert to string safely
        userID := fmt.Sprintf("%v", claims["user_id"])

        // Save user_id
        c.Locals("user_id", userID)
        c.Locals("role", claims["role"])

        // FIX 2: Fetch user and attach pharmacy_id
        var user model.User
        if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
            return c.Status(401).JSON(model.Response{Success: false, Message: "User not found"})
        }

        if user.PharmacyID == nil {
            return c.Status(400).JSON(model.Response{
                Success: false,
                Message: "Your account is not linked to a pharmacy",
            })
        }

        // Attach pharmacy_id globally
        c.Locals("pharmacy_id", *user.PharmacyID)

        return c.Next()
    }
}
// func Protected() fiber.Handler {
//     return func(c *fiber.Ctx) error {
//         authHeader := c.Get("Authorization")
//         if authHeader == "" || len(authHeader) < 8 {
//             return c.Status(401).JSON(model.Response{Success: false, Message: "Unauthorized"})
//         }

//         tokenString := authHeader[7:]
//         token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
//             return []byte(os.Getenv("JWT_SECRET")), nil
//         })

//         if err != nil || !token.Valid {
//             return c.Status(401).JSON(model.Response{Success: false, Message: "Invalid token"})
//         }

//         claims := token.Claims.(jwt.MapClaims)
        
//         // FIX: Match the key used in handlers/auth.go
//         c.Locals("user_id", claims["user_id"]) 
//         c.Locals("role", claims["role"])
        
		
		
		


//         return c.Next()
//     }
// }
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