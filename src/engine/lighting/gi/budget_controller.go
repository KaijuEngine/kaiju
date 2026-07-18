/******************************************************************************/
/* budget_controller.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package gi

// BudgetDecision is the per-frame work allocation produced from measured GI
// time. Providers may consume the probe budget, resolve scale, or both.
type BudgetDecision struct {
	MaxProbeUpdates uint32
	ResolveScale    float32
	OverBudget      bool
}

type BudgetController struct {
	settings       Settings
	probeBudget    uint32
	resolveScale   float32
	smoothedTimeMS float32
	initialized    bool
	underBudgetRun uint32
}

func NewBudgetController(settings Settings) BudgetController {
	return BudgetController{
		settings:     settings,
		probeBudget:  settings.MaxProbeUpdatesPerFrame,
		resolveScale: settings.ResolveScale,
	}
}

func (c *BudgetController) Configure(settings Settings) {
	*c = NewBudgetController(settings)
}

// Update uses a fast overload response and a deliberately slow recovery. This
// avoids oscillation while still protecting the developer's frame-time budget.
func (c *BudgetController) Update(measuredTimeMS float32) BudgetDecision {
	if !c.initialized {
		c.smoothedTimeMS = max(0, measuredTimeMS)
		c.initialized = true
	} else {
		c.smoothedTimeMS = c.smoothedTimeMS*0.85 + max(0, measuredTimeMS)*0.15
	}
	budget := c.settings.GPUTimeBudgetMS
	if !c.settings.AdaptiveBudget || budget <= 0 {
		return c.decision(false)
	}
	overBudget := c.smoothedTimeMS > budget*1.05
	if overBudget {
		c.underBudgetRun = 0
		if c.probeBudget > 1 {
			c.probeBudget = max(1, uint32(float32(c.probeBudget)*0.8))
		} else {
			c.resolveScale = max(0.25, c.resolveScale-0.125)
		}
	} else if c.smoothedTimeMS < budget*0.8 {
		c.underBudgetRun++
		if c.underBudgetRun >= 30 {
			c.underBudgetRun = 0
			if c.resolveScale < c.settings.ResolveScale {
				c.resolveScale = min(c.settings.ResolveScale, c.resolveScale+0.125)
			} else if c.probeBudget < c.settings.MaxProbeUpdatesPerFrame {
				step := max(uint32(1), c.settings.MaxProbeUpdatesPerFrame/16)
				c.probeBudget = min(c.settings.MaxProbeUpdatesPerFrame, c.probeBudget+step)
			}
		}
	} else {
		c.underBudgetRun = 0
	}
	return c.decision(overBudget)
}

func (c *BudgetController) decision(overBudget bool) BudgetDecision {
	return BudgetDecision{
		MaxProbeUpdates: c.probeBudget,
		ResolveScale:    c.resolveScale,
		OverBudget:      overBudget,
	}
}
