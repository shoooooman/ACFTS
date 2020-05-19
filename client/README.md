# Client-side

## Setup

### mysql (e.g. cluster number 0)

```
$ mysql.server start
$ mysql -u root -p
mysql> create database acfts_client_0
```

## Run (e.g. run with port 3000)

```
$ go run client.go
...
[GIN-debug] Listening and serving HTTP on :3000/
```

## Usage

You can switch the GUI mode and CUI mode by changing `IsGUI` in  `config/config.go`.

### GUI Mode

If you choose the GUI mode, a gotron browser will open after running the command.

You may run as much clients as you want, but make sure that each client waits for everyone else to finish the configuration pages.

If all the clients finish the configuration page, please press the ready button and start making transactions!

### CUI Mode

If you are intersted in making transaction manually, you should use the CUI mode.

You can create a request of a new transaction by writing a query in `boot/cui/example.go`.

There are two types of transactions, which are `transaction.GeneralTx` and `transaction.InsideTx`.

### GeneralTx

`GeneralTx` is defined in `transaction/request.go`.

```go
type GeneralTx struct {
	From    model.Address
	To      []model.Address
	Amounts []int
}
```

You can designate the sender (`From`), the receivers (`To`) and amounts for each receiver (`Amounts`).

The orders of `To` and `Amounts` must correspond to each other.

### InsideTx

On the other hand, `InsideTx` represents a transaction inside one cluster, which is also defined in `transaction/request.go`.

```go
type InsideTx struct {
	From    int
	To      []int
	Amounts []int
}
```

`InsedeTx` has `From`, `To` and `Amounts` as well as `GeneralTx`, but the types of the first two are not `model.Address`, but `int` for simplicity.

It means you can designate the sender and the receivers by the client indices in one cluster.

You need to convert `InsideTx` to `GeneralTx` with `transaction.ConvertInsideTxs` before making a request as the following step.

### Execution

In `boot/cui/example.go`, queries can be written in the following format. 

You can designate the sender and the receivers with addresses. Please use `boot.GetAllAddrs` to get the addresses.

```go
atxs := []transaction.GeneralTx{
  {From: addrs[0], To: []model.Address{addrs[1]}, Amounts: []int{200}},
}
transaction.Execute(atxs)
```

In this case, transactions (addrs[0] → addrs[1], 200) will be created.

Make sure that the sender has the enough amount of assets.

## [WIP] Benchmarking

You can benchmark the programs with `go test` command.

You need to prepare benchmark programs first.

```go
func BenchmarkScenario1(b *testing.B) {
	serverURLs := []string{
		"http://localhost:8080",
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
	}
	addrs := benchmarkSetup(serverURLs, 1, 4, 0)
	defer db.Close()

	atxs1 := []generalTx{
		{From: addrs[0], To: []model.Address{addrs[1]}, Amounts: []int{1}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executeTxs(serverURLs, atxs1)
	}
	manners.Close()
}
```

By default, 5 scenarios are prepared in `client_test.go`.

In those secarios, client {0, 1, 2, 3} are in cluster0 and client {4, 5, 6, 7} are in cluster1.

The genesis is bound to client 0.

Transaction: (from → to, amount)

- case 1: (0 → 1, 1)
- case 2: (0 → 1, 1), (1 → 0, 1)
- case 3: (0 → 1, 500000), (0 → 1, 500000), (2 → 3, 500000), (3 → 2, 500000)
- case 4: (random → random, 1)
- case 5: (0 → 4, 1)

In order to run the benchmark, execute the following command.

```
$ go test -bench Scenario1 -benchtime 10000x -timeout 24h
```

If you want to benchmark transactions between different clusters, you need to run the other client programs before executing the benchmark command.

#### Options

`-bench regexp`: Choose scenario(s) that you want to run benchmark matching a regular expression.

`-benchtime t`: Set the number of iterations (`b.N`). If you put `Nx` as t,  each benchmark will be run N times.

`-timeout d`: Set the limit time for benchmarks. The default is 10 mins.

You can see the details of the options [here](https://golang.org/cmd/go/#hdr-Testing_flags).

### pprof

pprof is a tool for profiling and visualization of programs. pprof provides various information such as flame graphs.

As setup, the server program needs to listen and serve the specified port for pprof.

```go
go func() {
	log.Println(http.ListenAndServe("localhost:7000", nil))
}()
```

When you want to profile localhost:7000, run the following command during the execution.

```
$ go tool pprof -http=":7777" -seconds 60 localhost:7000
```

When finishing the profiling, your browser will automatically open and show the result.

#### Options

`-http`: Specify host:post on which you can get an interactive web interface

`-seconds`: Set duration for time-based profile collection

#### Result

After finishing the profiling, the result is saved in your computer.

You can review the result on localhost:7777 whenever you want by the following command.

```
$ go tool pprof -http="localhost:7777" /path/to/pprof/pprof.samples.cpu.001.pb.gz
```

