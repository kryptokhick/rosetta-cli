// Copyright 2020 Coinbase, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/coinbase/rosetta-cli/pkg/tester"

	"github.com/coinbase/rosetta-sdk-go/fetcher"
	"github.com/coinbase/rosetta-sdk-go/utils"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var (
	checkConstructionCmd = &cobra.Command{
		Use:   "check:construction",
		Short: "Check the correctness of a Rosetta Construction API Implementation",
		Long: `The check:construction command runs an automated test of a
Construction API implementation by creating and broadcasting transactions
on a blockchain. In short, this tool generates new addresses, requests
funds, constructs transactions, signs transactions, broadcasts transactions,
and confirms transactions land on-chain. At each phase, a series of tests
are run to ensure that intermediate representations are correct (i.e. does
an unsigned transaction return a superset of operations provided during
construction?).

Check out the https://github.com/coinbase/rosetta-cli/tree/master/examples
directory for examples of how to configure this test for Bitcoin and
Ethereum.

Right now, this tool only supports transfer testing (for both account-based
and UTXO-based blockchains). However, we plan to add support for testing
arbitrary scenarios (i.e. staking, governance).`,
		Run: runCheckConstructionCmd,
	}
)

func runCheckConstructionCmd(cmd *cobra.Command, args []string) {
	if Config.Construction == nil {
		log.Fatal("construction configuration is missing!")
	}

	ensureDataDirectoryExists()
	ctx, cancel := context.WithCancel(context.Background())

	fetcher := fetcher.New(
		Config.OnlineURL,
		fetcher.WithMaxConnections(Config.MaxOnlineConnections),
		fetcher.WithRetryElapsedTime(time.Duration(Config.RetryElapsedTime)*time.Second),
		fetcher.WithTimeout(time.Duration(Config.HTTPTimeout)*time.Second),
		fetcher.WithMaxRetries(Config.MaxRetries),
	)

	_, _, fetchErr := fetcher.InitializeAsserter(ctx, Config.Network)
	if fetchErr != nil {
		tester.ExitConstruction(
			Config,
			nil,
			nil,
			fmt.Errorf("%w: unable to initialize asserter", fetchErr.Err),
			1,
		)
	}

	_, err := utils.CheckNetworkSupported(ctx, Config.Network, fetcher)
	if err != nil {
		tester.ExitConstruction(
			Config,
			nil,
			nil,
			fmt.Errorf("%w: unable to confirm network is supported", err),
			1,
		)
	}

	constructionTester, err := tester.InitializeConstruction(
		ctx,
		Config,
		Config.Network,
		fetcher,
		cancel,
		&SignalReceived,
	)
	if err != nil {
		tester.ExitConstruction(
			Config,
			nil,
			nil,
			fmt.Errorf("%w: unable to initialize construction tester", err),
			1,
		)
	}

	defer constructionTester.CloseDatabase(ctx)

	if err := constructionTester.PerformBroadcasts(ctx); err != nil {
		log.Fatalf("%s: unable to perform broadcasts", err.Error())
	}

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return constructionTester.StartPeriodicLogger(ctx)
	})

	g.Go(func() error {
		return constructionTester.StartSyncer(ctx, cancel)
	})

	g.Go(func() error {
		return constructionTester.StartConstructor(ctx)
	})

	g.Go(func() error {
		return constructionTester.WatchEndConditions(ctx)
	})

	sigListeners := []context.CancelFunc{cancel}
	go handleSignals(&sigListeners)

	err = g.Wait()
	constructionTester.HandleErr(err, &sigListeners)
}
