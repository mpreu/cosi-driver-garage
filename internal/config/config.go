package config

// Config options for the driver.
type Config struct {
	DriverName   string
	COSIEndpoint string
	// GarageAdminEndpoint is the Garage Admin API endpoint.
	GarageAdminEndpoint string
	// GarageAdminToken is the Garage Admin API token.
	GarageAdminToken string
}
