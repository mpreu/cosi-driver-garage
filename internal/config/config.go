package config

import "errors"

// Config options for the driver.
type Config struct {
	COSIEndpoint string
	DriverName   string
	Garage       *Garage
}

// Garage settings.
type Garage struct {
	Endpoint           string
	Region             string
	AdminEndpoint      string
	AdminToken         string
	InsecureSkipVerify bool
}

// Validate validates a configuration.
func (c *Config) Validate() error {
	if c.DriverName == "" {
		return errors.New("driver name cannot cannot be empty")
	}

	if c.COSIEndpoint == "" {
		return errors.New("COSI endpoint cannot be empty")
	}

	if c.Garage == nil {
		return errors.New("Garage settings cannot be nil")
	}

	if c.Garage.Endpoint == "" {
		return errors.New("Garage endpoint cannot be empty")
	}

	if c.Garage.Region == "" {
		return errors.New("Garage region cannot be empty")
	}

	if c.Garage.AdminEndpoint == "" {
		return errors.New("Garage admin endpoint cannot be empty")
	}

	if c.Garage.AdminToken == "" {
		return errors.New("Garage admin token cannot be empty")
	}

	return nil
}
