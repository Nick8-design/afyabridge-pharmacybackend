package database

import (
	"fmt"
	"log"
	"os"
	// "server/model"

	"gorm.io/driver/mysql" // Changed from postgres
	"gorm.io/gorm"
	// "gorm.io/gorm/schema"
)

var DB *gorm.DB

func ConnectDB() {
    // Your .env DATABASE_URL must now be the TiDB MySQL string
    dsn := os.Getenv("DATABASE_URL") 
    
    // Use mysql.Open instead of postgres.Open
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        DisableForeignKeyConstraintWhenMigrating: true,
        // NamingStrategy: schema.NamingStrategy{
        //     NoLowerCase: true, // This stops the automatic snake_case conversion
        // },
    })

    if err != nil {
        log.Fatalf("Failed to connect to TiDB: %v\n", err)
    }

    fmt.Println("Successfully connected to TiDB Cloud!")

    // TiDB handles AutoMigrate just like Postgres
    // db.AutoMigrate(
    //     &model.Doctor{}, 
    //     &model.Prescription{})
    // db.AutoMigrate( &model.Message{} ,&model.AppointmentSlot{})
    //     // &model.Appointment{},  )
    
    DB = db
}