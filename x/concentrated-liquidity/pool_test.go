package concentrated_liquidity_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	cltypes "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
	types "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
)

func (s *KeeperTestSuite) TestCalcOutAmtGivenIn() {
	currPrice := sdk.NewDec(5000)
	currSqrtPrice, err := currPrice.ApproxSqrt() // 70.710678118654752440
	s.Require().NoError(err)
	currTick := types.PriceToTick(currPrice) // 85176
	lowerPrice := sdk.NewDec(4545)
	s.Require().NoError(err)
	lowerTick := types.PriceToTick(lowerPrice) // 84222
	upperPrice := sdk.NewDec(5500)
	s.Require().NoError(err)
	upperTick := types.PriceToTick(upperPrice) // 86129

	defaultAmt0 := sdk.NewInt(1000000)
	defaultAmt1 := sdk.NewInt(5000000000)

	swapFee := sdk.ZeroDec()

	tests := map[string]struct {
		positionAmount0  sdk.Int
		positionAmount1  sdk.Int
		addPositions     func(ctx sdk.Context, poolId uint64)
		tokenIn          sdk.Coin
		tokenOutDenom    string
		priceLimit       sdk.Dec
		expectedTokenIn  sdk.Coin
		expectedTokenOut sdk.Coin
		expectedTick     sdk.Int
		newLowerPrice    sdk.Dec
		newUpperPrice    sdk.Dec
		poolLiqAmount0   sdk.Int
		poolLiqAmount1   sdk.Int
	}{
		//  One price range
		//
		//          5000
		//  4545 -----|----- 5500
		"single position within one tick: usdc -> eth": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)
			},
			tokenIn:       sdk.NewCoin("usdc", sdk.NewInt(42000000)),
			tokenOutDenom: "eth",
			priceLimit:    sdk.NewDec(5004),
			// we expect to put 42 usdc in and in return get .008398 eth back
			// due to truncations and precision loss, we actually put in 41.99 usdc and in return get .008396 eth back
			expectedTokenIn:  sdk.NewCoin("usdc", sdk.NewInt(41999999)),
			expectedTokenOut: sdk.NewCoin("eth", sdk.NewInt(8396)),
			expectedTick:     sdk.NewInt(85184),
		},
		"single position within one tick: eth -> usdc": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)
			},
			tokenIn:       sdk.NewCoin("eth", sdk.NewInt(13370)),
			tokenOutDenom: "usdc",
			priceLimit:    sdk.NewDec(4993),
			// we expect to put .01337 eth in and in return get 66.79 usdc back
			expectedTokenIn:  sdk.NewCoin("eth", sdk.NewInt(13370)),
			expectedTokenOut: sdk.NewCoin("usdc", sdk.NewInt(66790908)),
			expectedTick:     sdk.NewInt(85163),
		},
		//  Two equal price ranges
		//
		//          5000
		//  4545 -----|----- 5500
		//  4545 -----|----- 5500
		"two positions within one tick: usdc -> eth": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)

				// add second position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[1], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)
			},
			tokenIn:       sdk.NewCoin("usdc", sdk.NewInt(42000000)),
			tokenOutDenom: "eth",
			priceLimit:    sdk.NewDec(5002),
			// we expect to put 42 usdc in and in return get .008398 eth back
			expectedTokenIn:  sdk.NewCoin("usdc", sdk.NewInt(42000000)),
			expectedTokenOut: sdk.NewCoin("eth", sdk.NewInt(8398)),
			expectedTick:     sdk.NewInt(85180),
			// two positions with same liquidity entered
			poolLiqAmount0: sdk.NewInt(1000000).MulRaw(2),
			poolLiqAmount1: sdk.NewInt(5000000000).MulRaw(2),
		},
		"two positions within one tick: eth -> usdc": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)

				// add second position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[1], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)
			},
			tokenIn:       sdk.NewCoin("eth", sdk.NewInt(13370)),
			tokenOutDenom: "usdc",
			priceLimit:    sdk.NewDec(4996),
			// we expect to put .01337 eth in and in return get 66.79 eth back
			// TODO: look into why we are returning 66.81 instead of 66.79 like the inverse of this test above
			// sure, the above test only has 1 position while this has two positions, but shouldn't that effect the tokenIn as well?
			expectedTokenIn:  sdk.NewCoin("eth", sdk.NewInt(13370)),
			expectedTokenOut: sdk.NewCoin("usdc", sdk.NewInt(66811697)),
			expectedTick:     sdk.NewInt(85169),
			// two positions with same liquidity entered
			poolLiqAmount0: sdk.NewInt(1000000).MulRaw(2),
			poolLiqAmount1: sdk.NewInt(5000000000).MulRaw(2),
		},
		//  Consecutive price ranges
		//
		//          5000
		//  4545 -----|----- 5500
		//             5500 ----------- 6250
		//
		"two positions with consecutive price ranges": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)

				// create second position parameters
				newLowerPrice := sdk.NewDec(5501)
				s.Require().NoError(err)
				newLowerTick := types.PriceToTick(newLowerPrice) // 84222
				newUpperPrice := sdk.NewDec(6250)
				s.Require().NoError(err)
				newUpperTick := types.PriceToTick(newUpperPrice) // 87407

				// add position two with the new price range above
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[2], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), newLowerTick.Int64(), newUpperTick.Int64())
				s.Require().NoError(err)
			},
			tokenIn:       sdk.NewCoin("usdc", sdk.NewInt(10000000000)),
			tokenOutDenom: "eth",
			priceLimit:    sdk.NewDec(6106),
			// we expect to put 10000 usdc in and in return get 1.820536 eth back
			// TODO: see why we don't get 9938.148 usdc and 1.80615 eth
			expectedTokenIn:  sdk.NewCoin("usdc", sdk.NewInt(9999999999)),
			expectedTokenOut: sdk.NewCoin("eth", sdk.NewInt(1820536)),
			expectedTick:     sdk.NewInt(87173),
			newLowerPrice:    sdk.NewDec(5501),
			newUpperPrice:    sdk.NewDec(6250),
		},
	}

	for name, test := range tests {
		s.Run(name, func() {
			s.Setup()
			// create pool
			pool, err := s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(s.Ctx, 1, "eth", "usdc", currSqrtPrice, currTick)
			s.Require().NoError(err)

			// add positions
			test.addPositions(s.Ctx, pool.GetId())

			tokenIn, tokenOut, updatedTick, updatedLiquidity, _, err := s.App.ConcentratedLiquidityKeeper.CalcOutAmtGivenIn(
				s.Ctx,
				test.tokenIn, test.tokenOutDenom,
				swapFee, test.priceLimit, pool.GetId())
			s.Require().NoError(err)

			s.Require().Equal(test.expectedTokenIn.String(), tokenIn.String())
			s.Require().Equal(test.expectedTokenOut.String(), tokenOut.String())
			s.Require().Equal(test.expectedTick.String(), updatedTick.String())

			if test.newLowerPrice.IsNil() && test.newUpperPrice.IsNil() {
				test.newLowerPrice = lowerPrice
				test.newUpperPrice = upperPrice
			}

			newLowerTick := types.PriceToTick(test.newLowerPrice)
			newUpperTick := types.PriceToTick(test.newUpperPrice)

			lowerSqrtPrice, err := types.TickToSqrtPrice(newLowerTick)
			s.Require().NoError(err)
			upperSqrtPrice, err := types.TickToSqrtPrice(newUpperTick)
			s.Require().NoError(err)

			if test.poolLiqAmount0.IsNil() && test.poolLiqAmount1.IsNil() {
				test.poolLiqAmount0 = defaultAmt0
				test.poolLiqAmount1 = defaultAmt1
			}

			expectedLiquidity := types.GetLiquidityFromAmounts(currSqrtPrice, lowerSqrtPrice, upperSqrtPrice, test.poolLiqAmount0, test.poolLiqAmount1)
			s.Require().Equal(expectedLiquidity.TruncateInt(), updatedLiquidity.TruncateInt())
		})
	}
}

// func (s *KeeperTestSuite) TestCalcInAmtGivenOut() {
// 	ctx := s.Ctx
// 	pool, err := s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(s.Ctx, 1, "eth", "usdc", sdk.MustNewDecFromStr("70.710678"), sdk.NewInt(85176))
// 	s.Require().NoError(err)
// 	s.SetupPosition(pool.Id)

// 	// test asset a to b logic
// 	tokenOut := sdk.NewCoin("usdc", sdk.NewInt(4199999999))
// 	tokenInDenom := "eth"
// 	swapFee := sdk.NewDec(0)
// 	minPrice := sdk.NewDec(4500)
// 	maxPrice := sdk.NewDec(5500)

// 	amountIn, _, _, _, err := s.App.ConcentratedLiquidityKeeper.CalcInAmtGivenOut(ctx, tokenOut, tokenInDenom, swapFee, minPrice, maxPrice, pool.Id)
// 	s.Require().NoError(err)
// 	s.Require().Equal(sdk.NewDec(805287), amountIn.Amount.ToDec())

// 	// test asset b to a logic
// 	tokenOut = sdk.NewCoin("eth", sdk.NewInt(133700))
// 	tokenInDenom = "usdc"
// 	swapFee = sdk.NewDec(0)

// 	amountIn, _, _, _, err = s.App.ConcentratedLiquidityKeeper.CalcInAmtGivenOut(ctx, tokenOut, tokenInDenom, swapFee, minPrice, maxPrice, pool.Id)
// 	s.Require().NoError(err)
// 	s.Require().Equal(sdk.NewDec(666975610), amountIn.Amount.ToDec())

// 	// test asset a to b logic
// 	tokenOut = sdk.NewCoin("usdc", sdk.NewInt(4199999999))
// 	tokenInDenom = "eth"
// 	swapFee = sdk.NewDecWithPrec(2, 2)

// 	amountIn, _, _, _, err = s.App.ConcentratedLiquidityKeeper.CalcInAmtGivenOut(ctx, tokenOut, tokenInDenom, swapFee, minPrice, maxPrice, pool.Id)
// 	s.Require().NoError(err)
// 	s.Require().Equal(sdk.NewDec(821722), amountIn.Amount.ToDec())
// }

// func (s *KeeperTestSuite) TestSwapInAmtGivenOut() {
// 	ctx := s.Ctx
// 	pool, err := s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(ctx, 1, "eth", "usdc", sdk.MustNewDecFromStr("70.710678"), sdk.NewInt(85176))
// 	s.Require().NoError(err)
// 	fmt.Printf("%v pool liq pre \n", pool.Liquidity)
// 	lowerTick := int64(84222)
// 	upperTick := int64(86129)
// 	amount0Desired := sdk.NewInt(1)
// 	amount1Desired := sdk.NewInt(5000)

// 	s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, pool.Id, s.TestAccs[0], amount0Desired, amount1Desired, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick, upperTick)

// 	// test asset a to b logic
// 	tokenOut := sdk.NewCoin("usdc", sdk.NewInt(4199999999))
// 	tokenInDenom := "eth"
// 	swapFee := sdk.NewDec(0)
// 	minPrice := sdk.NewDec(4500)
// 	maxPrice := sdk.NewDec(5500)

// 	amountIn, err := s.App.ConcentratedLiquidityKeeper.SwapInAmtGivenOut(ctx, tokenOut, tokenInDenom, swapFee, minPrice, maxPrice, pool.Id)
// 	s.Require().NoError(err)
// 	fmt.Printf("%v amountIn \n", amountIn)
// 	pool = s.App.ConcentratedLiquidityKeeper.GetPoolbyId(ctx, pool.Id)

// // test asset a to b logic
// tokenOut := sdk.NewCoin("usdc", sdk.NewInt(4199999999))
// tokenInDenom := "eth"
// swapFee := sdk.NewDec(0)
// minPrice := sdk.NewDec(4500)
// maxPrice := sdk.NewDec(5500)

// amountIn, _, _, _, err := s.App.ConcentratedLiquidityKeeper.CalcInAmtGivenOut(ctx, tokenOut, tokenInDenom, swapFee, minPrice, maxPrice, pool.Id)
// s.Require().NoError(err)
// s.Require().Equal(sdk.NewDec(805287), amountIn.Amount.ToDec())

// // test asset b to a logic
// tokenOut = sdk.NewCoin("eth", sdk.NewInt(133700))
// tokenInDenom = "usdc"
// swapFee = sdk.NewDec(0)

// amountIn, _, _, _, err = s.App.ConcentratedLiquidityKeeper.CalcInAmtGivenOut(ctx, tokenOut, tokenInDenom, swapFee, minPrice, maxPrice, pool.Id)
// s.Require().NoError(err)
// s.Require().Equal(sdk.NewDec(666975610), amountIn.Amount.ToDec())

// // test asset a to b logic
// tokenOut = sdk.NewCoin("usdc", sdk.NewInt(4199999999))
// tokenInDenom = "eth"
// swapFee = sdk.NewDecWithPrec(2, 2)

// amountIn, _, _, _, err = s.App.ConcentratedLiquidityKeeper.CalcInAmtGivenOut(ctx, tokenOut, tokenInDenom, swapFee, minPrice, maxPrice, pool.Id)
// s.Require().NoError(err)
// s.Require().Equal(sdk.NewDec(821722), amountIn.Amount.ToDec())
// }

func (s *KeeperTestSuite) TestSwapOutAmtGivenIn() {
	currPrice := sdk.NewDec(5000)
	currSqrtPrice, err := currPrice.ApproxSqrt() // 70.710678118654752440
	s.Require().NoError(err)
	currTick := types.PriceToTick(currPrice) // 85176
	lowerPrice := sdk.NewDec(4545)
	s.Require().NoError(err)
	lowerTick := types.PriceToTick(lowerPrice) // 84222
	upperPrice := sdk.NewDec(5500)
	s.Require().NoError(err)
	upperTick := types.PriceToTick(upperPrice) // 86129

	defaultAmt0 := sdk.NewInt(1000000)
	defaultAmt1 := sdk.NewInt(5000000000)

	swapFee := sdk.ZeroDec()

	tests := map[string]struct {
		positionAmount0  sdk.Int
		positionAmount1  sdk.Int
		addPositions     func(ctx sdk.Context, poolId uint64)
		tokenIn          sdk.Coin
		tokenOutDenom    string
		priceLimit       sdk.Dec
		expectedTokenOut sdk.Coin
		expectedTick     sdk.Int
		newLowerPrice    sdk.Dec
		newUpperPrice    sdk.Dec
		poolLiqAmount0   sdk.Int
		poolLiqAmount1   sdk.Int
	}{
		//  One price range
		//
		//          5000
		//  4545 -----|----- 5500
		"single position within one tick: usdc -> eth": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)
			},
			tokenIn:       sdk.NewCoin("usdc", sdk.NewInt(42000000)),
			tokenOutDenom: "eth",
			priceLimit:    sdk.NewDec(5004),
			// we expect to put 42 usdc in and in return get .008398 eth back
			// due to truncations and precision loss, we actually put in 41.99 usdc and in return get .008396 eth back
			expectedTokenOut: sdk.NewCoin("eth", sdk.NewInt(8396)),
			expectedTick:     sdk.NewInt(85184),
		},
		"single position within one tick: eth -> usdc": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)
			},
			tokenIn:       sdk.NewCoin("eth", sdk.NewInt(13370)),
			tokenOutDenom: "usdc",
			priceLimit:    sdk.NewDec(4993),
			// we expect to put .01337 eth in and in return get 66.79 usdc back
			expectedTokenOut: sdk.NewCoin("usdc", sdk.NewInt(66790908)),
			expectedTick:     sdk.NewInt(85163),
		},
		//  Two equal price ranges
		//
		//          5000
		//  4545 -----|----- 5500
		//  4545 -----|----- 5500
		"two positions within one tick: usdc -> eth": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)

				// add second position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[1], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)
			},
			tokenIn:       sdk.NewCoin("usdc", sdk.NewInt(42000000)),
			tokenOutDenom: "eth",
			priceLimit:    sdk.NewDec(5002),
			// we expect to put 42 usdc in and in return get .008398 eth back
			expectedTokenOut: sdk.NewCoin("eth", sdk.NewInt(8398)),
			expectedTick:     sdk.NewInt(85180),
			// two positions with same liquidity entered
			poolLiqAmount0: sdk.NewInt(1000000).MulRaw(2),
			poolLiqAmount1: sdk.NewInt(5000000000).MulRaw(2),
		},
		"two positions within one tick: eth -> usdc": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)

				// add second position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[1], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)
			},
			tokenIn:       sdk.NewCoin("eth", sdk.NewInt(13370)),
			tokenOutDenom: "usdc",
			priceLimit:    sdk.NewDec(4996),
			// we expect to put .01337 eth in and in return get 66.79 eth back
			// TODO: look into why we are returning 66.81 instead of 66.79 like the inverse of this test above
			// sure, the above test only has 1 position while this has two positions, but shouldn't that effect the tokenIn as well?
			expectedTokenOut: sdk.NewCoin("usdc", sdk.NewInt(66811697)),
			expectedTick:     sdk.NewInt(85169),
			// two positions with same liquidity entered
			poolLiqAmount0: sdk.NewInt(1000000).MulRaw(2),
			poolLiqAmount1: sdk.NewInt(5000000000).MulRaw(2),
		},
		//  Consecutive price ranges
		//
		//          5000
		//  4545 -----|----- 5500
		//             5500 ----------- 6250
		//
		"two positions with consecutive price ranges": {
			addPositions: func(ctx sdk.Context, poolId uint64) {
				// add first position
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[0], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), lowerTick.Int64(), upperTick.Int64())
				s.Require().NoError(err)

				// create second position parameters
				newLowerPrice := sdk.NewDec(5501)
				s.Require().NoError(err)
				newLowerTick := types.PriceToTick(newLowerPrice) // 84222
				newUpperPrice := sdk.NewDec(6250)
				s.Require().NoError(err)
				newUpperTick := types.PriceToTick(newUpperPrice) // 87407

				// add position two with the new price range above
				_, _, _, err = s.App.ConcentratedLiquidityKeeper.CreatePosition(ctx, poolId, s.TestAccs[2], defaultAmt0, defaultAmt1, sdk.ZeroInt(), sdk.ZeroInt(), newLowerTick.Int64(), newUpperTick.Int64())
				s.Require().NoError(err)
			},
			tokenIn:       sdk.NewCoin("usdc", sdk.NewInt(10000000000)),
			tokenOutDenom: "eth",
			priceLimit:    sdk.NewDec(6106),
			// we expect to put 10000 usdc in and in return get 1.820536 eth back
			// TODO: see why we don't get 9938.148 usdc and 1.80615 eth
			expectedTokenOut: sdk.NewCoin("eth", sdk.NewInt(1820536)),
			expectedTick:     sdk.NewInt(87173),
			newLowerPrice:    sdk.NewDec(5501),
			newUpperPrice:    sdk.NewDec(6250),
		},
	}

	for name, test := range tests {
		s.Run(name, func() {
			s.Setup()
			// create pool
			pool, err := s.App.ConcentratedLiquidityKeeper.CreateNewConcentratedLiquidityPool(s.Ctx, 1, "eth", "usdc", currSqrtPrice, currTick)
			s.Require().NoError(err)

			// add positions
			test.addPositions(s.Ctx, pool.GetId())

			// execute internal swap function
			tokenOut, err := s.App.ConcentratedLiquidityKeeper.SwapOutAmtGivenIn(
				s.Ctx,
				test.tokenIn, test.tokenOutDenom,
				swapFee, test.priceLimit, pool.GetId())
			s.Require().NoError(err)

			pool, err = s.App.ConcentratedLiquidityKeeper.GetPoolbyId(s.Ctx, pool.GetId())
			s.Require().NoError(err)

			// check that we produced the same token out from the swap function that we expected
			s.Require().Equal(test.expectedTokenOut.String(), tokenOut.String())

			// check that the pool's current tick was updated correctly
			s.Require().Equal(test.expectedTick.String(), pool.GetCurrentTick().String())

			// the following is needed to get the expected liquidity to later compare to what the pool was updated to
			if test.newLowerPrice.IsNil() && test.newUpperPrice.IsNil() {
				test.newLowerPrice = lowerPrice
				test.newUpperPrice = upperPrice
			}

			newLowerTick := types.PriceToTick(test.newLowerPrice)
			newUpperTick := types.PriceToTick(test.newUpperPrice)

			lowerSqrtPrice, err := types.TickToSqrtPrice(newLowerTick)
			s.Require().NoError(err)
			upperSqrtPrice, err := types.TickToSqrtPrice(newUpperTick)
			s.Require().NoError(err)

			if test.poolLiqAmount0.IsNil() && test.poolLiqAmount1.IsNil() {
				test.poolLiqAmount0 = defaultAmt0
				test.poolLiqAmount1 = defaultAmt1
			}

			expectedLiquidity := types.GetLiquidityFromAmounts(currSqrtPrice, lowerSqrtPrice, upperSqrtPrice, test.poolLiqAmount0, test.poolLiqAmount1)
			// check that the pools liquidity was updated correctly
			s.Require().Equal(expectedLiquidity.TruncateInt(), pool.GetLiquidity().TruncateInt())

			// TODO: need to figure out a good way to test that the currentSqrtPrice that the pool is set to makes sense
			// right now we calculate this value through iterations, so unsure how to do this here / if its needed
		})

	}
}

func (s *KeeperTestSuite) TestOrderInitialPoolDenoms() {
	denom0, denom1, err := cltypes.OrderInitialPoolDenoms("axel", "osmo")
	s.Require().NoError(err)
	s.Require().Equal(denom0, "axel")
	s.Require().Equal(denom1, "osmo")

	denom0, denom1, err = cltypes.OrderInitialPoolDenoms("usdc", "eth")
	s.Require().NoError(err)
	s.Require().Equal(denom0, "eth")
	s.Require().Equal(denom1, "usdc")

	denom0, denom1, err = cltypes.OrderInitialPoolDenoms("usdc", "usdc")
	s.Require().Error(err)

}

func (suite *KeeperTestSuite) TestPriceToTick() {
	testCases := []struct {
		name         string
		price        sdk.Dec
		tickExpected string
	}{
		{
			"happy path",
			sdk.NewDec(5000),
			"85176",
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			tick := types.PriceToTick(tc.price)
			suite.Require().Equal(tc.tickExpected, tick.String())
		})
	}
}