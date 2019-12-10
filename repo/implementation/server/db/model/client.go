package model

import "github.com/jinzhu/gorm"

// Client is
type Client struct {
	gorm.Model
	ClusterID uint
	Address
}
