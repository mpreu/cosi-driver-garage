package config

import "errors"

// Config options for the driver.
type Config struct {
	DriverName   string
	COSIEndpoint string
	// GarageAdminEndpoint is the Garage Admin API endpoint.
	GarageAdminEndpoint string
	// GarageAdminToken is the Garage Admin API token.
	GarageAdminToken string
}

// Validate validates a configuration.
func (c *Config) Validate() error {
	if c.DriverName == "" {
		return errors.New("driver name cannot cannot be empty")
	}

	if c.COSIEndpoint == "" {
		return errors.New("COSI endpoint cannot be empty")
	}

	if c.GarageAdminEndpoint == "" {
		return errors.New("Garage admin endpoint cannot be empty")
	}

	if c.GarageAdminToken == "" {
		return errors.New("Garage admin token cannot be empty")
	}

	return nil
}
