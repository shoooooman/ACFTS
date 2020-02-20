# Asychronous Consensus-Free Transaction Systems
This repository is for the project of [DISCO](https://disco.ethz.ch/) in ETH Zurich.

ACFTS is a transaction system which can solve double-spending problems without consensus.
You can see the details of the system at [the paper](./thesis/Thesis.pdf).

This repository also contains a simulator of the system at [implementation](./implementation).
[server](./implementation/server) contains the code of the server-side which works for verification of transactions.
[client](./implementation/client) contains the code of the client-side which creates new transactions and sends them to the server for the simulation.
You can also get the details in READMEs in each directory.
