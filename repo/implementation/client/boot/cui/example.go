package cui

import "acfts-client/transaction"

// ExecuteExamples executes sample transactions
func ExecuteExamples() {
	/* Use generalTx */

	// Inside one cluster
	// atxs := []transaction.GeneralTx{
	// 	{From: addrs[0], To: []model.Address{addrs[1]}, Amounts: []int{200}},
	// }
	// transaction.Execute(atxs)

	// atxs := make([]transaction.GeneralTx, 200)
	// tx := transaction.GeneralTx{From: addrs[0], To: []model.Address{addrs[1]}, Amounts: []int{1}}
	// for i := 0; i < 200; i++ {
	// 	atxs[i] = tx
	// }
	// transaction.Execute(atxs)

	// Between two clusters
	// atxs := []transaction.GeneralTx{
	// 	{From: addrs[0], To: []model.Address{addrs[4]}, Amounts: []int{200}},
	// }
	// transaction.Execute(atxs)

	// atxs := make([]transaction.GeneralTx, 200)
	// tx := transaction.GeneralTx{From: addrs[0], To: []model.Address{addrs[4]}, Amounts: []int{1}}
	// for i := 0; i < 200; i++ {
	// 	atxs[i] = tx
	// }
	// transaction.Execute(atxs)

	/* Use insideTx */

	// case 1: 0 -> 1
	itxs := []transaction.InsideTx{
		{From: 0, To: []int{1}, Amounts: []int{200}},
	}
	atxs := transaction.ConvertInsideTxs(itxs)
	transaction.Execute(atxs)

	// case 2: 0 <-> 1
	// tx1 := transaction.InsideTx{From: 0, To: []int{1}, Amounts: []int{200}}
	// tx2 := transaction.InsideTx{From: 1, To: []int{0}, Amounts: []int{200}}
	// itxs := []transaction.InsideTx{}
	// for i := 0; i < 10; i++ {
	// 	itxs = append(itxs, tx1)
	// 	itxs = append(itxs, tx2)
	// }
	// atxs := transaction.ConvertInsideTxs(itxs)
	// transaction.Execute(atxs)

	// case 3: 0 <-> 1 & 2 <-> 3
	// tx0 := transaction.InsideTx{From: 0, To: []int{0, 2}, Amounts: []int{50, 150}}
	// itxs0 := []transaction.InsideTx{}
	// itxs0 = append(itxs0, tx0)
	// atxs0 := transaction.ConvertInsideTxs(itxs0)
	// transaction.Execute(atxs0)
	//
	// tx1 := transaction.InsideTx{From: 0, To: []int{1}, Amounts: []int{50}}
	// tx2 := transaction.InsideTx{From: 1, To: []int{0}, Amounts: []int{50}}
	// tx3 := transaction.InsideTx{From: 2, To: []int{3}, Amounts: []int{150}}
	// tx4 := transaction.InsideTx{From: 3, To: []int{2}, Amounts: []int{150}}
	// itxs1 := []transaction.InsideTx{}
	// for i := 0; i < 10; i++ {
	// 	itxs1 = append(itxs1, tx1)
	// 	itxs1 = append(itxs1, tx2)
	// 	itxs1 = append(itxs1, tx3)
	// 	itxs1 = append(itxs1, tx4)
	// }
	// atxs1 := transaction.ConvertInsideTxs(itxs1)
	// transaction.Execute(atxs1)

	// case 4: random -> random
	// mrand.Seed(time.Now().UnixNano())
	// tx0 := transaction.InsideTx{From: 0, To: []int{0, 1, 2, 3}, Amounts: []int{50, 50, 50, 50}}
	// itxs0 := []transaction.InsideTx{}
	// itxs0 = append(itxs0, tx0)
	// atxs0 := transaction.ConvertInsideTxs(itxs0)
	// transaction.Execute(atxs0)
	//
	// itxs1 := []transaction.InsideTx{}
	// for i := 0; i < 25; i++ {
	// 	from := mrand.Intn(4)
	// 	to := mrand.Intn(4)
	// 	tx1 := transaction.InsideTx{From: from, To: []int{to}, Amounts: []int{1}}
	// 	itxs1 = append(itxs1, tx1)
	// }
	// atxs1 := transaction.ConvertInsideTxs(itxs1)
	// transaction.Execute(atxs1)

	// Wait for receiving outputs of this cluster
	// for {
	// }
}
