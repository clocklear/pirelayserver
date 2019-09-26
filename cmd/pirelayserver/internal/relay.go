package internal

import (
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/robfig/cron/v3"
	"github.com/stianeikeland/go-rpio/v4"
)

type RelayController struct {
	relayPins []uint8
	scheduler *cron.Cron
	logger    log.Logger
	cfger     Configurer
}

func NewRelayController(l log.Logger, relayPins []uint8, cfger Configurer) (*RelayController, error) {
	c := RelayController{
		relayPins: relayPins,
		logger:    l,
		scheduler: cron.New(),
	}
	cfg, err := cfger.Get()
	if err != nil {
		return nil, err
	}
	err = c.ApplyConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *RelayController) ApplyConfig(cfg Config) error {
	// Handle schedules
	c.scheduler.Stop()
	c.clearScheduler()
	for _, s := range cfg.Schedules {
		f := c.createToggleFunction(s.Relay, s.Action)
		_, err := c.scheduler.AddFunc(s.Expression, f)
		if err != nil {
			return err
		}
	}
	c.scheduler.Start()
	return nil
}

func (c *RelayController) clearScheduler() {
	for _, e := range c.scheduler.Entries() {
		c.scheduler.Remove(e.ID)
	}
}

func (c *RelayController) createToggleFunction(relay uint8, action Action) func() {
	ret := func() {}
	switch action {
	case On:
		ret = func() {
			c.logger.Log("msg", "Switching relay to On", "relay", relay)
			c.On(relay)
		}
	case Off:
		ret = func() {
			c.logger.Log("msg", "Switching relay to Off", "relay", relay)
			c.Off(relay)
		}
	}
	return ret
}

func (c *RelayController) Status() (Status, error) {
	r := Status{}
	states := []State{}
	if err := rpio.Open(); err != nil {
		return r, err
	}
	defer rpio.Close()
	for k, v := range c.relayPins {
		pin := rpio.Pin(v)
		s := pin.Read()
		states = append(states, State{
			Relay: uint8(k) + 1,
			State: uint8(s),
		})
	}
	r.States = states
	return r, nil
}

func (c *RelayController) Toggle(relay uint8) error {
	if int(relay) > len(c.relayPins) {
		return fmt.Errorf("invalid relay. must be uint between 1 and %v", len(c.relayPins))
	}
	if err := rpio.Open(); err != nil {
		return err
	}
	defer rpio.Close()
	pin := rpio.Pin(c.relayPins[relay-1])
	pin.Output()
	pin.Toggle()
	return nil
}

func (c *RelayController) On(relay uint8) error {
	if int(relay) > len(c.relayPins) {
		return fmt.Errorf("invalid relay. must be uint between 1 and %v", len(c.relayPins))
	}
	if err := rpio.Open(); err != nil {
		return err
	}
	defer rpio.Close()
	pin := rpio.Pin(c.relayPins[relay-1])
	pin.Output()
	pin.High()
	return nil
}

func (c *RelayController) Off(relay uint8) error {
	if int(relay) > len(c.relayPins) {
		return fmt.Errorf("invalid relay. must be uint between 1 and %v", len(c.relayPins))
	}
	if err := rpio.Open(); err != nil {
		return err
	}
	defer rpio.Close()
	pin := rpio.Pin(c.relayPins[relay-1])
	pin.Output()
	pin.Low()
	return nil
}