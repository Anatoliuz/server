package db

import (
	"gorm.io/gorm"
	"time"
)

type Client struct {
	ID        uint      `gorm:"primaryKey"`          // Unique client ID
	Address   string    `gorm:"not null;unique"`     // Client address (e.g., IP)
	Commands  []Command `gorm:"foreignKey:ClientID"` // One-to-many relationship with Command
	CreatedAt time.Time `gorm:"autoCreateTime"`      // Timestamp of client creation
	UpdatedAt time.Time `gorm:"autoUpdateTime"`      // Timestamp of client update
}

type Command struct {
	ID        uint           `gorm:"primaryKey"`                                    // Unique command ID
	ClientID  uint           `gorm:"not null;index"`                                // Foreign key referencing Client
	Client    Client         `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // Many-to-one relationship
	Command   string         `gorm:"not null"`                                      // Command text
	Status    string         `gorm:"type:varchar;default:queued"`                   // Command status
	Result    *string        `gorm:"type:text"`                                     // Result of command executionclients
	CreatedAt time.Time      `gorm:"autoCreateTime"`                                // Timestamp of creation
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`                                // Timestamp of the last update
	DeletedAt gorm.DeletedAt `gorm:"index"`                                         // Logical deletion
}
