# rosetta-validator

[![Coinbase](https://circleci.com/gh/coinbase/rosetta-validator/tree/master.svg?style=shield)](https://circleci.com/gh/coinbase/rosetta-validator/tree/master)
[![Coverage Status](https://coveralls.io/repos/github/coinbase/rosetta-validator/badge.svg)](https://coveralls.io/github/coinbase/rosetta-validator)
[![Go Report Card](https://goreportcard.com/badge/github.com/coinbase/rosetta-validator)](https://goreportcard.com/report/github.com/coinbase/rosetta-validator)
[![License](https://img.shields.io/github/license/coinbase/rosetta-validator.svg)](https://github.com/coinbase/rosetta-validator/blob/master/LICENSE.txt)

Once you create a Rosetta Server, you'll need to test its
performance and correctness. This validation tool makes that easy!

## What is Rosetta?
Rosetta is a new project from Coinbase to standardize the process
of deploying and interacting with blockchains. With an explicit
specification to adhere to, all parties involved in blockchain
development can spend less time figuring out how to integrate
with each other and more time working on the novel advances that
will push the blockchain ecosystem forward. In practice, this means
that any blockchain project that implements the requirements outlined
in this specification will enable exchanges, block explorers,
and wallets to integrate with much less communication overhead
and network-specific work.

## Run the Validator
1. Start your Rosetta Server (and the blockchain node it connects to if it is
not a single binary.
2. Start the validator using `make SERVER_URL=<server URL> validate`.
3. Examine processed blocks using `make watch-blocks`. You can also print transactions
by setting `LOG_TRANSACTIONS="true"` in the environment or as a `make` argument.
4. Watch for errors in the processing logs. Any error will cause the validator to stop.
5. Analyze benchmarks from `validator-data/block_benchmarks.csv` and
`validator-data/account_benchmarks.csv` by setting `LOG_BENCHMARKS="true"` in
the environment or as a `make` argument.

### Configuration Options
All configuration options can be set in the call to `make validate`
(ex: `make SERVER_URL=http://localhost:9999 validate`) or altered in
the `Makefile` itself.

_There is no additional setting required to support blockchains with reorgs. This
is handled automatically!_

#### SERVER_URL
_Default: http://localhost:8080_
The URL the validator will use to access the Rosetta Server.

#### LOG_TRANSACTIONS
_Default: true_
All processed transactions will be logged to `transactions.txt`. You can tail
these logs using `watch-transactions`.

#### LOG_BALANCES
_Default: true_
All processed balance changes will be logged to `balances.txt`. You can tail
these logs using `watch-balances`.

#### LOG_RECONCILIATION
_Default: true_
All reconciliation checks will be logged to `reconciliations.txt`. You can tail
these logs using `watch-reconciliations`.

#### BOOTSTRAP_BALANCES
_Default: false_
Blockchains that set balances in genesis must create a `bootstrap_balances.csv`
file in the `/validator-data` directory and pass `BOOTSTRAP_BALANCES=true` as an
argument to make. If balances are not bootsrapped and balances are set in genesis,
reconciliation will fail.

There is an example file in `examples/bootstrap_balances.csv`.

#### LOG_BENCHMARKS
_Default: false_
It can be useful to observe performance characteristics of a Rosetta Server.
When enabled, it is possible to view the latency of block and account fetches.
Note, this naive implementation of benchmarks does not factor in request latency
due to multithreaded requests.

## Development
* `make deps` to install dependencies
* `make test` to run tests
* `make lint` to lint the source code (included generated code)
* `make release` to run one last check before opening a PR

## Correctness Checks
This tool performs a variety of correctness checks using the Rosetta Server. If
any correctness check fails, the validator will exit and print out a detailed
message explaining the error.

### Response Correctness
The validator uses the autogenerated [Go Client package](https://github.com/coinbase/rosetta-sdk-go)
to communicate with the Rosetta Server and assert that responses adhere
to the Rosetta Standard.

### Duplicate Hashes
The validator checks that a block hash or transaction hash is
never duplicated.

### Non-negative Balances
The validator checks that an account balance does not go
negative from any operations.

### Balance Reconciliation
#### Active Addresses
The validator checks that the balance of an account computed by
its operations is equal to the balance of the account according
to the node. If this balance is not identical, the validator will
exit.

#### Inactive Addresses
The validator randomly checks the balances of accounts that aren't
involved in any transactions. The balances of accounts could change
on the blockchain node without being included in an operation
returned by the Rosetta Server. Recall that **ALL** balance-changing
operations must be returned by the Rosetta Server.

## Future Work
* Automatically test the correctness of a Rosetta Client SDK by constructing,
signing, and submitting a transaction. This can be further extended by ensuring
broadcast transactions eventually land in a block.
* Change logging to utilize a more advanced output mechanism than CSV.

## License
This project is available open source under the terms of the [Apache 2.0 License](https://opensource.org/licenses/Apache-2.0).

© 2020 Coinbase
