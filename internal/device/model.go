package device

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Type string

const (
	TypeORAN = "oran"
	TypeIOT  = "iot"
)

type Status string

const (
	StatusProvisioning   Status = "provisioning"
	StatusOnline         Status = "online"
	StatusOffline        Status = "offline"
	StatusMaintenance    Status = "maintenance"
	StatusDecommissioned Status = "decommissioned"
)

// Device represents a physical or logical devices maintained by NetPilot
type Device struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	DeviceID string             `bson:"device_id"     json:"device_id"`
	Name     string             `bson:"name"          json:"name"`
	Type     Type               `bson:"type"          json:"type"`

	Manufacturer string `bson:"manufacturer,omitempty"  json:"manufacturer,omitempty"`
	Model        string `bson:"model,omitempty"         json:"model,omitempty"`
	SerialNumber string `bson:"serial_number,omitempty" json:"serial_number,omitempty"`

	IPAddress  string `bson:"ip_address,omitempty"  json:"ip_address,omitempty"`
	MACAddress string `bson:"mac_address,omitempty" json:"mac_address,omitempty"`

	FirmwareVersion string `bson:"firmware_version,omitempty" json:"firmware_version,omitempty"`

	Status Status `bson:"status" json:"status"`

	LastSeenAt *time.Time `bson:"last_seen_at,omitempty" json:"last_seen_at,omitempty"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}
