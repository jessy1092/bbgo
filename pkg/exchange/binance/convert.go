package binance

import (
	"fmt"
	"strings"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/pkg/errors"

	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
)

func toGlobalMarket(symbol binance.Symbol) types.Market {
	market := types.Market{
		Symbol:          symbol.Symbol,
		LocalSymbol:     symbol.Symbol,
		PricePrecision:  symbol.QuotePrecision,
		VolumePrecision: symbol.BaseAssetPrecision,
		QuoteCurrency:   symbol.QuoteAsset,
		BaseCurrency:    symbol.BaseAsset,
	}

	if f := symbol.MinNotionalFilter(); f != nil {
		market.MinNotional = fixedpoint.MustNewFromString(f.MinNotional)
		market.MinAmount = fixedpoint.MustNewFromString(f.MinNotional)
	}

	// The LOT_SIZE filter defines the quantity (aka "lots" in auction terms) rules for a symbol.
	// There are 3 parts:
	// minQty defines the minimum quantity/icebergQty allowed.
	//	maxQty defines the maximum quantity/icebergQty allowed.
	//	stepSize defines the intervals that a quantity/icebergQty can be increased/decreased by.
	if f := symbol.LotSizeFilter(); f != nil {
		market.MinQuantity = fixedpoint.MustNewFromString(f.MinQuantity)
		market.MaxQuantity = fixedpoint.MustNewFromString(f.MaxQuantity)
		market.StepSize = fixedpoint.MustNewFromString(f.StepSize)
	}

	if f := symbol.PriceFilter(); f != nil {
		market.MaxPrice = fixedpoint.MustNewFromString(f.MaxPrice)
		market.MinPrice = fixedpoint.MustNewFromString(f.MinPrice)
		market.TickSize = fixedpoint.MustNewFromString(f.TickSize)
	}

	return market
}

// TODO: Cuz it returns types.Market as well, merge following to the above function
func toGlobalFuturesMarket(symbol futures.Symbol) types.Market {
	market := types.Market{
		Symbol:          symbol.Symbol,
		LocalSymbol:     symbol.Symbol,
		PricePrecision:  symbol.QuotePrecision,
		VolumePrecision: symbol.BaseAssetPrecision,
		QuoteCurrency:   symbol.QuoteAsset,
		BaseCurrency:    symbol.BaseAsset,
	}

	if f := symbol.MinNotionalFilter(); f != nil {
		market.MinNotional = fixedpoint.MustNewFromString(f.Notional)
		market.MinAmount = fixedpoint.MustNewFromString(f.Notional)
	}

	// The LOT_SIZE filter defines the quantity (aka "lots" in auction terms) rules for a symbol.
	// There are 3 parts:
	// minQty defines the minimum quantity/icebergQty allowed.
	//	maxQty defines the maximum quantity/icebergQty allowed.
	//	stepSize defines the intervals that a quantity/icebergQty can be increased/decreased by.
	if f := symbol.LotSizeFilter(); f != nil {
		market.MinQuantity = fixedpoint.MustNewFromString(f.MinQuantity)
		market.MaxQuantity = fixedpoint.MustNewFromString(f.MaxQuantity)
		market.StepSize = fixedpoint.MustNewFromString(f.StepSize)
	}

	if f := symbol.PriceFilter(); f != nil {
		market.MaxPrice = fixedpoint.MustNewFromString(f.MaxPrice)
		market.MinPrice = fixedpoint.MustNewFromString(f.MinPrice)
		market.TickSize = fixedpoint.MustNewFromString(f.TickSize)
	}

	return market
}

func toGlobalIsolatedUserAsset(userAsset binance.IsolatedUserAsset) types.IsolatedUserAsset {
	return types.IsolatedUserAsset{
		Asset:         userAsset.Asset,
		Borrowed:      fixedpoint.MustNewFromString(userAsset.Borrowed),
		Free:          fixedpoint.MustNewFromString(userAsset.Free),
		Interest:      fixedpoint.MustNewFromString(userAsset.Interest),
		Locked:        fixedpoint.MustNewFromString(userAsset.Locked),
		NetAsset:      fixedpoint.MustNewFromString(userAsset.NetAsset),
		NetAssetOfBtc: fixedpoint.MustNewFromString(userAsset.NetAssetOfBtc),
		BorrowEnabled: userAsset.BorrowEnabled,
		RepayEnabled:  userAsset.RepayEnabled,
		TotalAsset:    fixedpoint.MustNewFromString(userAsset.TotalAsset),
	}
}

func toGlobalIsolatedMarginAsset(asset binance.IsolatedMarginAsset) types.IsolatedMarginAsset {
	return types.IsolatedMarginAsset{
		Symbol:            asset.Symbol,
		QuoteAsset:        toGlobalIsolatedUserAsset(asset.QuoteAsset),
		BaseAsset:         toGlobalIsolatedUserAsset(asset.BaseAsset),
		IsolatedCreated:   asset.IsolatedCreated,
		MarginLevel:       fixedpoint.MustNewFromString(asset.MarginLevel),
		MarginLevelStatus: asset.MarginLevelStatus,
		MarginRatio:       fixedpoint.MustNewFromString(asset.MarginRatio),
		IndexPrice:        fixedpoint.MustNewFromString(asset.IndexPrice),
		LiquidatePrice:    fixedpoint.MustNewFromString(asset.LiquidatePrice),
		LiquidateRate:     fixedpoint.MustNewFromString(asset.LiquidateRate),
		TradeEnabled:      false,
	}
}

func toGlobalIsolatedMarginAssets(assets []binance.IsolatedMarginAsset) (retAssets types.IsolatedMarginAssetMap) {
	retMarginAssets := make(types.IsolatedMarginAssetMap)
	for _, marginAsset := range assets {
		retMarginAssets[marginAsset.Symbol] = toGlobalIsolatedMarginAsset(marginAsset)
	}

	return retMarginAssets
}

//func toGlobalIsolatedMarginAccount(account *binance.IsolatedMarginAccount) *types.IsolatedMarginAccount {
//	return &types.IsolatedMarginAccount{
//		TotalAssetOfBTC:     fixedpoint.MustNewFromString(account.TotalNetAssetOfBTC),
//		TotalLiabilityOfBTC: fixedpoint.MustNewFromString(account.TotalLiabilityOfBTC),
//		TotalNetAssetOfBTC:  fixedpoint.MustNewFromString(account.TotalNetAssetOfBTC),
//		Assets:              toGlobalIsolatedMarginAssets(account.Assets),
//	}
//}

func toGlobalMarginUserAssets(assets []binance.UserAsset) types.MarginAssetMap {
	retMarginAssets := make(types.MarginAssetMap)
	for _, marginAsset := range assets {
		retMarginAssets[marginAsset.Asset] = types.MarginUserAsset{
			Asset:    marginAsset.Asset,
			Borrowed: fixedpoint.MustNewFromString(marginAsset.Borrowed),
			Free:     fixedpoint.MustNewFromString(marginAsset.Free),
			Interest: fixedpoint.MustNewFromString(marginAsset.Interest),
			Locked:   fixedpoint.MustNewFromString(marginAsset.Locked),
			NetAsset: fixedpoint.MustNewFromString(marginAsset.NetAsset),
		}
	}

	return retMarginAssets
}

func toGlobalMarginAccountInfo(account *binance.MarginAccount) *types.MarginAccountInfo {
	return &types.MarginAccountInfo{
		BorrowEnabled:       account.BorrowEnabled,
		MarginLevel:         fixedpoint.MustNewFromString(account.MarginLevel),
		TotalAssetOfBTC:     fixedpoint.MustNewFromString(account.TotalAssetOfBTC),
		TotalLiabilityOfBTC: fixedpoint.MustNewFromString(account.TotalLiabilityOfBTC),
		TotalNetAssetOfBTC:  fixedpoint.MustNewFromString(account.TotalNetAssetOfBTC),
		TradeEnabled:        account.TradeEnabled,
		TransferEnabled:     account.TransferEnabled,
		Assets:              toGlobalMarginUserAssets(account.UserAssets),
	}
}

func toGlobalIsolatedMarginAccountInfo(account *binance.IsolatedMarginAccount) *types.IsolatedMarginAccountInfo {
	return &types.IsolatedMarginAccountInfo{
		TotalAssetOfBTC:     fixedpoint.MustNewFromString(account.TotalAssetOfBTC),
		TotalLiabilityOfBTC: fixedpoint.MustNewFromString(account.TotalLiabilityOfBTC),
		TotalNetAssetOfBTC:  fixedpoint.MustNewFromString(account.TotalNetAssetOfBTC),
		Assets:              toGlobalIsolatedMarginAssets(account.Assets),
	}
}

func toGlobalFuturesAccountInfo(account *futures.Account) *types.FuturesAccountInfo {
	return &types.FuturesAccountInfo{
		Assets:                      toGlobalFuturesUserAssets(account.Assets),
		Positions:                   toGlobalFuturesPositions(account.Positions),
		TotalInitialMargin:          fixedpoint.MustNewFromString(account.TotalInitialMargin),
		TotalMaintMargin:            fixedpoint.MustNewFromString(account.TotalMaintMargin),
		TotalMarginBalance:          fixedpoint.MustNewFromString(account.TotalMarginBalance),
		TotalOpenOrderInitialMargin: fixedpoint.MustNewFromString(account.TotalOpenOrderInitialMargin),
		TotalPositionInitialMargin:  fixedpoint.MustNewFromString(account.TotalPositionInitialMargin),
		TotalUnrealizedProfit:       fixedpoint.MustNewFromString(account.TotalUnrealizedProfit),
		TotalWalletBalance:          fixedpoint.MustNewFromString(account.TotalWalletBalance),
		UpdateTime:                  account.UpdateTime,
	}
}

func toGlobalFuturesBalance(balances []*futures.Balance) types.BalanceMap {
	retBalances := make(types.BalanceMap)
	for _, balance := range balances {
		retBalances[balance.Asset] = types.Balance{
			Currency:  balance.Asset,
			Available: fixedpoint.MustNewFromString(balance.AvailableBalance),
		}
	}
	return retBalances
}

func toGlobalFuturesPositions(futuresPositions []*futures.AccountPosition) types.FuturesPositionMap {
	retFuturesPositions := make(types.FuturesPositionMap)
	for _, futuresPosition := range futuresPositions {
		retFuturesPositions[futuresPosition.Symbol] = types.FuturesPosition{ // TODO: types.FuturesPosition
			Isolated: futuresPosition.Isolated,
			PositionRisk: &types.PositionRisk{
				Leverage: fixedpoint.MustNewFromString(futuresPosition.Leverage),
			},
			Symbol:     futuresPosition.Symbol,
			UpdateTime: futuresPosition.UpdateTime,
		}
	}

	return retFuturesPositions
}

func toGlobalFuturesUserAssets(assets []*futures.AccountAsset) (retAssets types.FuturesAssetMap) {
	retFuturesAssets := make(types.FuturesAssetMap)
	for _, futuresAsset := range assets {
		retFuturesAssets[futuresAsset.Asset] = types.FuturesUserAsset{
			Asset:                  futuresAsset.Asset,
			InitialMargin:          fixedpoint.MustNewFromString(futuresAsset.InitialMargin),
			MaintMargin:            fixedpoint.MustNewFromString(futuresAsset.MaintMargin),
			MarginBalance:          fixedpoint.MustNewFromString(futuresAsset.MarginBalance),
			MaxWithdrawAmount:      fixedpoint.MustNewFromString(futuresAsset.MaxWithdrawAmount),
			OpenOrderInitialMargin: fixedpoint.MustNewFromString(futuresAsset.OpenOrderInitialMargin),
			PositionInitialMargin:  fixedpoint.MustNewFromString(futuresAsset.PositionInitialMargin),
			UnrealizedProfit:       fixedpoint.MustNewFromString(futuresAsset.UnrealizedProfit),
			WalletBalance:          fixedpoint.MustNewFromString(futuresAsset.WalletBalance),
		}
	}

	return retFuturesAssets
}

func toGlobalTicker(stats *binance.PriceChangeStats) (*types.Ticker, error) {
	return &types.Ticker{
		Volume: fixedpoint.MustNewFromString(stats.Volume),
		Last:   fixedpoint.MustNewFromString(stats.LastPrice),
		Open:   fixedpoint.MustNewFromString(stats.OpenPrice),
		High:   fixedpoint.MustNewFromString(stats.HighPrice),
		Low:    fixedpoint.MustNewFromString(stats.LowPrice),
		Buy:    fixedpoint.MustNewFromString(stats.BidPrice),
		Sell:   fixedpoint.MustNewFromString(stats.AskPrice),
		Time:   time.Unix(0, stats.CloseTime*int64(time.Millisecond)),
	}, nil
}

func toLocalOrderType(orderType types.OrderType) (binance.OrderType, error) {
	switch orderType {

	case types.OrderTypeLimitMaker:
		return binance.OrderTypeLimitMaker, nil

	case types.OrderTypeLimit:
		return binance.OrderTypeLimit, nil

	case types.OrderTypeStopLimit:
		return binance.OrderTypeStopLossLimit, nil

	case types.OrderTypeStopMarket:
		return binance.OrderTypeStopLoss, nil

	case types.OrderTypeMarket:
		return binance.OrderTypeMarket, nil
	}

	return "", fmt.Errorf("can not convert to local order, order type %s not supported", orderType)
}

func toLocalFuturesOrderType(orderType types.OrderType) (futures.OrderType, error) {
	switch orderType {

	// case types.OrderTypeLimitMaker:
	// 	return futures.OrderTypeLimitMaker, nil //TODO

	case types.OrderTypeLimit, types.OrderTypeLimitMaker:
		return futures.OrderTypeLimit, nil

	// case types.OrderTypeStopLimit:
	// 	return futures.OrderTypeStopLossLimit, nil //TODO

	// case types.OrderTypeStopMarket:
	// 	return futures.OrderTypeStopLoss, nil //TODO

	case types.OrderTypeMarket:
		return futures.OrderTypeMarket, nil
	}

	return "", fmt.Errorf("can not convert to local order, order type %s not supported", orderType)
}

func toGlobalOrders(binanceOrders []*binance.Order) (orders []types.Order, err error) {
	for _, binanceOrder := range binanceOrders {
		order, err := toGlobalOrder(binanceOrder, false)
		if err != nil {
			return orders, err
		}

		orders = append(orders, *order)
	}

	return orders, err
}

func toGlobalFuturesOrders(futuresOrders []*futures.Order) (orders []types.Order, err error) {
	for _, futuresOrder := range futuresOrders {
		order, err := toGlobalFuturesOrder(futuresOrder, false)
		if err != nil {
			return orders, err
		}

		orders = append(orders, *order)
	}

	return orders, err
}

func toGlobalOrder(binanceOrder *binance.Order, isMargin bool) (*types.Order, error) {
	return &types.Order{
		SubmitOrder: types.SubmitOrder{
			ClientOrderID: binanceOrder.ClientOrderID,
			Symbol:        binanceOrder.Symbol,
			Side:          toGlobalSideType(binanceOrder.Side),
			Type:          toGlobalOrderType(binanceOrder.Type),
			Quantity:      fixedpoint.MustNewFromString(binanceOrder.OrigQuantity),
			Price:         fixedpoint.MustNewFromString(binanceOrder.Price),
			TimeInForce:   types.TimeInForce(binanceOrder.TimeInForce),
		},
		Exchange:         types.ExchangeBinance,
		IsWorking:        binanceOrder.IsWorking,
		OrderID:          uint64(binanceOrder.OrderID),
		Status:           toGlobalOrderStatus(binanceOrder.Status),
		ExecutedQuantity: fixedpoint.MustNewFromString(binanceOrder.ExecutedQuantity),
		CreationTime:     types.Time(millisecondTime(binanceOrder.Time)),
		UpdateTime:       types.Time(millisecondTime(binanceOrder.UpdateTime)),
		IsMargin:         isMargin,
		IsIsolated:       binanceOrder.IsIsolated,
	}, nil
}

func toGlobalFuturesOrder(futuresOrder *futures.Order, isMargin bool) (*types.Order, error) {
	return &types.Order{
		SubmitOrder: types.SubmitOrder{
			ClientOrderID: futuresOrder.ClientOrderID,
			Symbol:        futuresOrder.Symbol,
			Side:          toGlobalFuturesSideType(futuresOrder.Side),
			Type:          toGlobalFuturesOrderType(futuresOrder.Type),
			ReduceOnly:    futuresOrder.ReduceOnly,
			ClosePosition: futuresOrder.ClosePosition,
			Quantity:      fixedpoint.MustNewFromString(futuresOrder.OrigQuantity),
			Price:         fixedpoint.MustNewFromString(futuresOrder.Price),
			TimeInForce:   types.TimeInForce(futuresOrder.TimeInForce),
		},
		Exchange:         types.ExchangeBinance,
		OrderID:          uint64(futuresOrder.OrderID),
		Status:           toGlobalFuturesOrderStatus(futuresOrder.Status),
		ExecutedQuantity: fixedpoint.MustNewFromString(futuresOrder.ExecutedQuantity),
		CreationTime:     types.Time(millisecondTime(futuresOrder.Time)),
		UpdateTime:       types.Time(millisecondTime(futuresOrder.UpdateTime)),
		IsMargin:         isMargin,
	}, nil
}

func millisecondTime(t int64) time.Time {
	return time.Unix(0, t*int64(time.Millisecond))
}

func toGlobalTrade(t binance.TradeV3, isMargin bool) (*types.Trade, error) {
	// skip trade ID that is the same. however this should not happen
	var side types.SideType
	if t.IsBuyer {
		side = types.SideTypeBuy
	} else {
		side = types.SideTypeSell
	}

	price, err := fixedpoint.NewFromString(t.Price)
	if err != nil {
		return nil, errors.Wrapf(err, "price parse error, price: %+v", t.Price)
	}

	quantity, err := fixedpoint.NewFromString(t.Quantity)
	if err != nil {
		return nil, errors.Wrapf(err, "quantity parse error, quantity: %+v", t.Quantity)
	}

	var quoteQuantity fixedpoint.Value
	if len(t.QuoteQuantity) > 0 {
		quoteQuantity, err = fixedpoint.NewFromString(t.QuoteQuantity)
		if err != nil {
			return nil, errors.Wrapf(err, "quote quantity parse error, quoteQuantity: %+v", t.QuoteQuantity)
		}
	} else {
		quoteQuantity = price.Mul(quantity)
	}

	fee, err := fixedpoint.NewFromString(t.Commission)
	if err != nil {
		return nil, errors.Wrapf(err, "commission parse error, commission: %+v", t.Commission)
	}

	return &types.Trade{
		ID:            uint64(t.ID),
		OrderID:       uint64(t.OrderID),
		Price:         price,
		Symbol:        t.Symbol,
		Exchange:      "binance",
		Quantity:      quantity,
		QuoteQuantity: quoteQuantity,
		Side:          side,
		IsBuyer:       t.IsBuyer,
		IsMaker:       t.IsMaker,
		Fee:           fee,
		FeeCurrency:   t.CommissionAsset,
		Time:          types.Time(millisecondTime(t.Time)),
		IsMargin:      isMargin,
		IsIsolated:    t.IsIsolated,
	}, nil
}

func toGlobalFuturesTrade(t futures.AccountTrade) (*types.Trade, error) {
	// skip trade ID that is the same. however this should not happen
	var side types.SideType
	if t.Buyer {
		side = types.SideTypeBuy
	} else {
		side = types.SideTypeSell
	}

	price, err := fixedpoint.NewFromString(t.Price)
	if err != nil {
		return nil, errors.Wrapf(err, "price parse error, price: %+v", t.Price)
	}

	quantity, err := fixedpoint.NewFromString(t.Quantity)
	if err != nil {
		return nil, errors.Wrapf(err, "quantity parse error, quantity: %+v", t.Quantity)
	}

	var quoteQuantity fixedpoint.Value
	if len(t.QuoteQuantity) > 0 {
		quoteQuantity, err = fixedpoint.NewFromString(t.QuoteQuantity)
		if err != nil {
			return nil, errors.Wrapf(err, "quote quantity parse error, quoteQuantity: %+v", t.QuoteQuantity)
		}
	} else {
		quoteQuantity = price.Mul(quantity)
	}

	fee, err := fixedpoint.NewFromString(t.Commission)
	if err != nil {
		return nil, errors.Wrapf(err, "commission parse error, commission: %+v", t.Commission)
	}

	return &types.Trade{
		ID:            uint64(t.ID),
		OrderID:       uint64(t.OrderID),
		Price:         price,
		Symbol:        t.Symbol,
		Exchange:      "binance",
		Quantity:      quantity,
		QuoteQuantity: quoteQuantity,
		Side:          side,
		IsBuyer:       t.Buyer,
		IsMaker:       t.Maker,
		Fee:           fee,
		FeeCurrency:   t.CommissionAsset,
		Time:          types.Time(millisecondTime(t.Time)),
		IsFutures:     true,
	}, nil
}

func toGlobalSideType(side binance.SideType) types.SideType {
	switch side {
	case binance.SideTypeBuy:
		return types.SideTypeBuy

	case binance.SideTypeSell:
		return types.SideTypeSell

	default:
		log.Errorf("can not convert binance side type, unknown side type: %q", side)
		return ""
	}
}

func toGlobalFuturesSideType(side futures.SideType) types.SideType {
	switch side {
	case futures.SideTypeBuy:
		return types.SideTypeBuy

	case futures.SideTypeSell:
		return types.SideTypeSell

	default:
		log.Errorf("can not convert futures side type, unknown side type: %q", side)
		return ""
	}
}

func toGlobalOrderType(orderType binance.OrderType) types.OrderType {
	switch orderType {

	case binance.OrderTypeLimit,
		binance.OrderTypeLimitMaker, binance.OrderTypeTakeProfitLimit:
		return types.OrderTypeLimit

	case binance.OrderTypeMarket:
		return types.OrderTypeMarket

	case binance.OrderTypeStopLossLimit:
		return types.OrderTypeStopLimit

	case binance.OrderTypeStopLoss:
		return types.OrderTypeStopMarket

	default:
		log.Errorf("unsupported order type: %v", orderType)
		return ""
	}
}

func toGlobalFuturesOrderType(orderType futures.OrderType) types.OrderType {
	switch orderType {
	// TODO
	case futures.OrderTypeLimit: // , futures.OrderTypeLimitMaker, futures.OrderTypeTakeProfitLimit:
		return types.OrderTypeLimit

	case futures.OrderTypeMarket:
		return types.OrderTypeMarket
	// TODO
	// case futures.OrderTypeStopLossLimit:
	// 	return types.OrderTypeStopLimit
	// TODO
	// case futures.OrderTypeStopLoss:
	// 	return types.OrderTypeStopMarket

	default:
		log.Errorf("unsupported order type: %v", orderType)
		return ""
	}
}

func toGlobalOrderStatus(orderStatus binance.OrderStatusType) types.OrderStatus {
	switch orderStatus {
	case binance.OrderStatusTypeNew:
		return types.OrderStatusNew

	case binance.OrderStatusTypeRejected:
		return types.OrderStatusRejected

	case binance.OrderStatusTypeCanceled:
		return types.OrderStatusCanceled

	case binance.OrderStatusTypePartiallyFilled:
		return types.OrderStatusPartiallyFilled

	case binance.OrderStatusTypeFilled:
		return types.OrderStatusFilled
	}

	return types.OrderStatus(orderStatus)
}

func toGlobalFuturesOrderStatus(orderStatus futures.OrderStatusType) types.OrderStatus {
	switch orderStatus {
	case futures.OrderStatusTypeNew:
		return types.OrderStatusNew

	case futures.OrderStatusTypeRejected:
		return types.OrderStatusRejected

	case futures.OrderStatusTypeCanceled:
		return types.OrderStatusCanceled

	case futures.OrderStatusTypePartiallyFilled:
		return types.OrderStatusPartiallyFilled

	case futures.OrderStatusTypeFilled:
		return types.OrderStatusFilled
	}

	return types.OrderStatus(orderStatus)
}

func convertSubscription(s types.Subscription) string {
	// binance uses lower case symbol name,
	// for kline, it's "<symbol>@kline_<interval>"
	// for depth, it's "<symbol>@depth OR <symbol>@depth@100ms"
	switch s.Channel {
	case types.KLineChannel:
		return fmt.Sprintf("%s@%s_%s", strings.ToLower(s.Symbol), s.Channel, s.Options.String())
	case types.BookChannel:
		// depth values: 5, 10, 20
		// Stream Names: <symbol>@depth<levels> OR <symbol>@depth<levels>@100ms.
		// Update speed: 1000ms or 100ms
		n := strings.ToLower(s.Symbol) + "@depth"
		switch s.Options.Depth {
		case types.DepthLevel5:
			n += "5"

		case types.DepthLevelMedium:
			n += "20"

		case types.DepthLevelFull:
		default:

		}

		switch s.Options.Speed {
		case types.SpeedHigh:
			n += "@100ms"

		case types.SpeedLow:
			n += "@1000ms"

		}
		return n
	case types.BookTickerChannel:
		return fmt.Sprintf("%s@bookTicker", strings.ToLower(s.Symbol))
	}

	return fmt.Sprintf("%s@%s", strings.ToLower(s.Symbol), s.Channel)
}

func convertPremiumIndex(index *futures.PremiumIndex) (*types.PremiumIndex, error) {
	markPrice, err := fixedpoint.NewFromString(index.MarkPrice)
	if err != nil {
		return nil, err
	}

	lastFundingRate, err := fixedpoint.NewFromString(index.LastFundingRate)
	if err != nil {
		return nil, err
	}

	nextFundingTime := time.Unix(0, index.NextFundingTime*int64(time.Millisecond))
	t := time.Unix(0, index.Time*int64(time.Millisecond))

	return &types.PremiumIndex{
		Symbol:          index.Symbol,
		MarkPrice:       markPrice,
		NextFundingTime: nextFundingTime,
		LastFundingRate: lastFundingRate,
		Time:            t,
	}, nil
}

func convertPositionRisk(risk *futures.PositionRisk) (*types.PositionRisk, error) {
	leverage, err := fixedpoint.NewFromString(risk.Leverage)
	if err != nil {
		return nil, err
	}

	liquidationPrice, err := fixedpoint.NewFromString(risk.LiquidationPrice)
	if err != nil {
		return nil, err
	}

	return &types.PositionRisk{
		Leverage:         leverage,
		LiquidationPrice: liquidationPrice,
	}, nil
}
