# Asychronous Consensus-Free Transaction Systems
This repository is for the project of [DISCO](https://disco.ethz.ch/) in ETH Zurich.

ACFTS is a transaction system which avoids double-spending problems without taking consensus.

This repository provides a simulator of the system.
[server](./server) contains the code of the server-side, which works for verification of transactions.
[client](./client) contains the code of the client-side, which creates new transactions and sends them to the servers.
You can get the details by reading READMEs in each directory ([client](./client) and [server](./server)).

[This paper](./docs/thesis/Thesis.pdf) is written for showing the algorithms and experiment results.

## Demo
![demo](https://user-images.githubusercontent.com/32924835/82243449-99959b80-997a-11ea-9e65-2c202dda286b.gif)
