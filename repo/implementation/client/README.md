# Client-side

## Setup

### mysql

```
$ mysql.server start
$ mysql -u root -p
mysql> create database acfts_client
```

## Run

```
$ go run client.go
```

## Usage

Initially, four clients (i.e. four public keys) are generated and verification requests are sent to localhost with port number 8080, 8081, 8082 and 8083.

There are six example cases now. Transaction: (from → to, amount)

Genesis is bound to client 0 by default.

- case 1: (0 → 1, 200)
- case 2: (0 ⇄ 1, 200) × 10
- case 3: (0 ⇄ 1, 50) × 10,  (2 ⇄ 3, 200) × 10 (parallelly)
- case 4.1: (0 ⇄ 1, 50) × 25,  (1 ⇄ 2, 50) × 25, (2 ⇄ 3, 50) × 25, (3 ⇄ 0, 50) × 25 (parallelly)
- case 4.2: (0 ⇄ 1, 100) × 25,  (0 ⇄ 1, 100) × 25 (parallelly)
- case 5: (random → random, 1) × 25

You can create original transactions with `struct simpleTx`.

### SimpleTx

`simpleTx` is a golang struct which has `From`, `To` and `Amounts`.

`From` is `int`. `To` and `Amounts` are arrays of int, which means one transaction represented by `simpleTx` can have multiple outputs although the number of inputs is one. Each element of `Amounts` corresponds to each element of `To` in the same order.

```go
tx := simpleTx{From: x, To: []int{y, z}, Amounts: []int{50, 150}}
```

In this case, transactions (x → y, 50) and (x → z, 150) are created.

#### Execution

To send transactions to servers to get their signatures, call `executeTxs()`.

```go
func executeTxs(baseURLs []string, txs []simpleTx, async bool, finished chan bool) {...}
```

`txs` represents a set of transactions which will be verified by servers.

E.g. **case 1**

```
txs1 := []simpleTx{
	{From: 0, To: []int{1}, Amounts: []int{200}},
}
executeTxs(baseURLs, txs1, false, nil)
```

#### Execution with multi-thread

If you want to execute sets of transactions asynchrously, the argument `async` should be `true` and pass `chan bool` for mutual exclusion. Then, call `executeTxs`s  with multi-thread using `go`.

E.g. **case 3**

```go
finished := make(chan bool)

tx0 := simpleTx{From: 0, To: []int{0, 2}, Amounts: []int{50, 150}}
txs0 := []simpleTx{}
txs0 = append(txs0, tx0)
executeTxs(baseURLs, txs0, false, nil)

tx1 := simpleTx{From: 0, To: []int{1}, Amounts: []int{50}}
tx2 := simpleTx{From: 1, To: []int{0}, Amounts: []int{50}}
tx3 := simpleTx{From: 2, To: []int{3}, Amounts: []int{150}}
tx4 := simpleTx{From: 3, To: []int{2}, Amounts: []int{150}}
txs1 := []simpleTx{}
txs2 := []simpleTx{}
for i := 0; i < 10; i++ {
  txs1 = append(txs1, tx1)
  txs1 = append(txs1, tx2)
  txs2 = append(txs2, tx3)
  txs2 = append(txs2, tx4)
}

go executeTxs(baseURLs, txs1, true, finished)
go executeTxs(baseURLs, txs2, true, finished)
<-finished
<-finished
```

