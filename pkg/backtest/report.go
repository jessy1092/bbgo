package backtest

import (
	"fmt"
	"strings"
	"time"

	"github.com/c9s/bbgo/pkg/accounting/pnl"
	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
)

// SummaryReport is the summary of the back-test session
type SummaryReport struct {
	StartTime            time.Time        `json:"startTime"`
	EndTime              time.Time        `json:"endTime"`
	Sessions             []string         `json:"sessions"`
	InitialTotalBalances types.BalanceMap `json:"initialTotalBalances"`
	FinalTotalBalances   types.BalanceMap `json:"finalTotalBalances"`
}

// SessionSymbolReport is the report per exchange session
// trades are merged, collected and re-calculated
type SessionSymbolReport struct {
	StartTime       time.Time                 `json:"startTime"`
	EndTime         time.Time                 `json:"endTime"`
	Symbol          string                    `json:"symbol,omitempty"`
	LastPrice       fixedpoint.Value          `json:"lastPrice,omitempty"`
	StartPrice      fixedpoint.Value          `json:"startPrice,omitempty"`
	PnLReport       *pnl.AverageCostPnlReport `json:"pnlReport,omitempty"`
	InitialBalances types.BalanceMap          `json:"initialBalances,omitempty"`
	FinalBalances   types.BalanceMap          `json:"finalBalances,omitempty"`
	Manifests       Manifests                 `json:"manifests,omitempty"`
}

const SessionTimeFormat = "2006-01-02T15_04"

// FormatSessionName returns the back-test session name
func FormatSessionName(sessions []string, symbols []string, startTime, endTime time.Time) string {
	return fmt.Sprintf("%s_%s_%s-%s",
		strings.Join(sessions, "-"),
		strings.Join(symbols, "-"),
		startTime.Format(SessionTimeFormat),
		endTime.Format(SessionTimeFormat),
	)
}
