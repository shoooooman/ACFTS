package event

import (
	"acfts-client/model"

	"github.com/jinzhu/gorm"
)

func getBalance(db *gorm.DB, addr model.Address) int {
	var balance int
	db.Table("outputs").
		Where("address1 = ? AND address2 = ? AND used = false", addr.Address1, addr.Address2).
		Select("sum(amount)").Row().Scan(&balance)
	return balance
}
