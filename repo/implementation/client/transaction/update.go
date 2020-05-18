package transaction

import (
	"acfts-client/config"
	"acfts-client/model"
)

// FIXME: サーバー側もstruct Addressを作ったらaddr1, addr2をまとめる
func updateOutputs(outputs []model.Output, sigs []model.Signature) {
	db := config.DB
	for _, output := range outputs {
		// Make a record if the output is not created yet
		cnt := 0
		db.Where("address1 = ? AND address2 = ? AND previous_hash = ? AND output_index = ?",
			output.Address1, output.Address2, output.PreviousHash, output.Index).
			First(&output).Count(&cnt)
		if cnt == 0 {
			db.Create(&output)
		}
		db.Model(&output).Association("Signatures").Append(sigs)
	}
}
