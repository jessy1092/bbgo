package bpp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/c9s/bbgo/pkg/bbgo"
	"github.com/c9s/bbgo/pkg/exchange/max"
	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/service"
	"github.com/c9s/bbgo/pkg/types"
)

const ID = "bpp"

const stateKey = "state-v1"

var defaultFeeRate = fixedpoint.NewFromFloat(0.001)

var log = logrus.WithField("strategy", ID)

func init() {
	bbgo.RegisterStrategy(ID, &Strategy{})
}

type State struct {
	Position          *bbgo.Position   `json:"position,omitempty"`
	AccumulatedVolume fixedpoint.Value `json:"accumulatedVolume,omitempty"`
	AccumulatedPnL    fixedpoint.Value `json:"accumulatedPnL,omitempty"`
	AccumulatedProfit fixedpoint.Value `json:"accumulatedProfit,omitempty"`
	AccumulatedLoss   fixedpoint.Value `json:"accumulatedLoss,omitempty"`
	AccumulatedSince  int64            `json:"accumulatedSince,omitempty"`
}

type Strategy struct {
	*bbgo.Graceful
	*bbgo.Notifiability
	*bbgo.Persistence

	Symbol    string           `json:"symbol"`
	Interval  types.Interval   `json:"interval"`
	Quantity  fixedpoint.Value `json:"quantity"`
	MinSpread fixedpoint.Value `json:"minSpread"`
	Spread    fixedpoint.Value `json:"spread"`

	session *bbgo.ExchangeSession
	book    *types.StreamOrderBook
	market  types.Market

	state *State

	activeMakerOrders *bbgo.LocalActiveOrderBook
	orderStore        *bbgo.OrderStore
	tradeCollector    *bbgo.TradeCollector

	groupID uint32

	stopC chan struct{}
}

func (s *Strategy) ID() string {
	return ID
}

func (s *Strategy) Subscribe(session *bbgo.ExchangeSession) {
	// session.Subscribe(types.BookChannel, s.Symbol, types.SubscribeOptions{})
	session.Subscribe(types.KLineChannel, s.Symbol, types.SubscribeOptions{
		Interval: string(s.Interval),
	})
}

func (s *Strategy) Validate() error {
	if len(s.Symbol) == 0 {
		return errors.New("symbol is required")
	}

	return nil
}

func (s *Strategy) SaveState() error {
	if err := s.Persistence.Save(s.state, ID, s.Symbol, stateKey); err != nil {
		return err
	} else {
		log.Infof("state is saved => %+v", s.state)
	}
	return nil
}

func (s *Strategy) LoadState() error {
	var state State

	// load position
	if err := s.Persistence.Load(&state, ID, s.Symbol, stateKey); err != nil {
		if err != service.ErrPersistenceNotExists {
			return err
		}

		s.state = &State{}
	} else {
		s.state = &state
		log.Infof("state is restored: %+v", s.state)
	}

	// if position is nil, we need to allocate a new position for calculation
	if s.state.Position == nil {
		s.state.Position = bbgo.NewPositionFromMarket(s.market)
	}

	if s.state.AccumulatedSince == 0 {
		s.state.AccumulatedSince = time.Now().Unix()
	}

	return nil
}

func (s *Strategy) cancelOrders(ctx context.Context) {
	if err := s.session.Exchange.CancelOrders(ctx, s.activeMakerOrders.Orders()...); err != nil {
		log.WithError(err).Errorf("can not cancel %s orders", s.Symbol)
	}

	time.Sleep(30 * time.Millisecond)

	for s.activeMakerOrders.NumOfOrders() > 0 {
		orders := s.activeMakerOrders.Orders()
		log.Warnf("%d orders are not cancelled yet:", len(orders))
		s.activeMakerOrders.Print()

		if err := s.session.Exchange.CancelOrders(ctx, s.activeMakerOrders.Orders()...); err != nil {
			log.WithError(err).Errorf("can not cancel %s orders", s.Symbol)
			continue
		}

		log.Infof("waiting for orders to be cancelled...")

		select {
		case <-time.After(3 * time.Second):

		case <-ctx.Done():
			break

		}

		// verify the current open orders via the RESTful API
		if s.activeMakerOrders.NumOfOrders() > 0 {
			log.Warnf("there are orders not cancelled, using REStful API to verify...")
			openOrders, err := s.session.Exchange.QueryOpenOrders(ctx, s.Symbol)
			if err != nil {
				log.WithError(err).Errorf("can not query %s open orders", s.Symbol)
				continue
			}

			openOrderStore := bbgo.NewOrderStore(s.Symbol)
			openOrderStore.Add(openOrders...)

			for _, o := range s.activeMakerOrders.Orders() {
				// if it does not exist, we should remove it
				if !openOrderStore.Exists(o.OrderID) {
					s.activeMakerOrders.Remove(o)
				}
			}
		}
	}
}

func (s *Strategy) placeOrders(ctx context.Context, orderExecutor bbgo.OrderExecutor) {
	ticker, err := s.session.Exchange.QueryTicker(ctx, s.Symbol)
	if err != nil {
		return
	}

	midPrice := fixedpoint.NewFromFloat((ticker.Buy + ticker.Sell) / 2)

	one := fixedpoint.NewFromFloat(1.0)
	askPrice := midPrice.Mul(one + s.Spread)
	bidPrice := midPrice.Mul(one - s.Spread)
	base := s.state.Position.Base

	sellOrder := types.SubmitOrder{
		Symbol:      s.Symbol,
		Side:        types.SideTypeSell,
		Type:        types.OrderTypeLimitMaker,
		Quantity:    s.Quantity.Float64(),
		Price:       askPrice.Float64(),
		Market:      s.market,
		TimeInForce: "GTC",
		GroupID:     s.groupID,
	}
	buyOrder := types.SubmitOrder{
		Symbol:      s.Symbol,
		Side:        types.SideTypeBuy,
		Type:        types.OrderTypeLimitMaker,
		Quantity:    s.Quantity.Float64(),
		Price:       bidPrice.Float64(),
		Market:      s.market,
		TimeInForce: "GTC",
		GroupID:     s.groupID,
	}

	var submitOrders []types.SubmitOrder

	minQuantity := fixedpoint.NewFromFloat(s.market.MinQuantity)
	if base > -minQuantity && base < minQuantity {
		submitOrders = append(submitOrders, sellOrder, buyOrder)
	} else if base > minQuantity {
		sellOrder.Quantity = base.Float64()
		submitOrders = append(submitOrders, sellOrder)
	} else if base < -minQuantity {
		buyOrder.Quantity = base.Float64()
		submitOrders = append(submitOrders, buyOrder)
	}

	createdOrders, err := orderExecutor.SubmitOrders(ctx, submitOrders...)
	if err != nil {
		log.WithError(err).Errorf("can not place ping pong orders")
	}
	s.orderStore.Add(createdOrders...)
	s.activeMakerOrders.Add(createdOrders...)
}

func (s *Strategy) Run(ctx context.Context, orderExecutor bbgo.OrderExecutor, session *bbgo.ExchangeSession) error {
	// initial required information
	s.session = session

	market, ok := session.Market(s.Symbol)
	if !ok {
		return fmt.Errorf("market %s not found", s.Symbol)
	}
	s.market = market

	// calculate group id for orders
	instanceID := fmt.Sprintf("%s-%s", ID, s.Symbol)
	s.groupID = max.GenerateGroupID(instanceID)
	log.Infof("using group id %d from fnv(%s)", s.groupID, instanceID)

	// restore state
	if err := s.LoadState(); err != nil {
		return err
	}

	s.stopC = make(chan struct{})

	s.activeMakerOrders = bbgo.NewLocalActiveOrderBook()
	s.activeMakerOrders.BindStream(session.UserDataStream)

	s.orderStore = bbgo.NewOrderStore(s.Symbol)
	s.orderStore.BindStream(session.UserDataStream)

	s.tradeCollector = bbgo.NewTradeCollector(s.Symbol, s.state.Position, s.orderStore)
	s.tradeCollector.OnProfit(func(trade types.Trade, profit fixedpoint.Value, netProfit fixedpoint.Value) {
		profitMargin := profit.DivFloat64(trade.QuoteQuantity)
		s.Notify("%s maker profit %s %f %s (%.3f%%), net profit =~ %f %s",
			s.Symbol,
			pnlEmoji(profit),
			profit.Float64(), s.market.QuoteCurrency,
			profitMargin.Float64()*100.0,
			netProfit.Float64(), s.market.QuoteCurrency)
	})
	s.tradeCollector.OnTrade(func(trade types.Trade) {
		s.Notifiability.Notify(trade)
		log.Infof("%s", trade)
	})
	s.tradeCollector.OnPositionUpdate(func(position *bbgo.Position) {
		s.Notifiability.Notify(position)
	})

	s.tradeCollector.BindStream(session.UserDataStream)
	
	// s.tradeCollector.BindStreamForBackground(session.UserDataStream)
	// go s.tradeCollector.Run(ctx)

	session.UserDataStream.OnStart(func() {
		s.placeOrders(ctx, orderExecutor)
	})

	session.MarketDataStream.OnKLineClosed(func(kline types.KLine) {
		if kline.Symbol != s.Symbol {
			return
		}

		s.tradeCollector.Process()
		s.cancelOrders(ctx)
		s.placeOrders(ctx, orderExecutor)
	})

	// s.book = types.NewStreamBook(s.Symbol)
	// s.book.BindStreamForBackground(session.MarketDataStream)

	s.Graceful.OnShutdown(func(ctx context.Context, wg *sync.WaitGroup) {
		defer wg.Done()
		close(s.stopC)

		s.cancelOrders(ctx)

		if err := s.SaveState(); err != nil {
			log.WithError(err).Errorf("can not save state: %+v", s.state)
		}
	})

	return nil
}

// lets move this to the fun package
var lossEmoji = "🔥"
var profitEmoji = "💰"

func pnlEmoji(pnl fixedpoint.Value) string {
	if pnl < 0 {
		return lossEmoji
	}

	if pnl == 0 {
		return ""
	}

	return profitEmoji
}