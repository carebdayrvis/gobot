package gpio

import (
	"fmt"
	"gobot.io/x/gobot"
	"time"
)

type RotaryEncoderDriver struct {
	// Position from 0 since start
	Button     *ButtonDriver
	Position   int
	clkPin     string
	dtPin      string
	swPin      string
	name       string
	halt       chan bool
	interval   time.Duration
	connection DigitalReader
	gobot.Eventer
}

func NewRotaryEncoderDriver(a DigitalReader, clkPin, dtPin, swPin string, v ...time.Duration) *RotaryEncoderDriver {
	button := NewButtonDriver(a, swPin)

	r := &RotaryEncoderDriver{
		Button:     button,
		Position:   0,
		connection: a,
		clkPin:     clkPin,
		dtPin:      dtPin,
		swPin:      swPin,
		name:       gobot.DefaultName("Rotary Encoder"),
		Eventer:    gobot.NewEventer(),
		interval:   1 * time.Millisecond,
		halt:       make(chan bool),
	}

	if len(v) > 0 {
		r.interval = v[0]
	}

	r.AddEvent(Rotation)
	r.AddEvent(Error)

	return r
}

// Start starts the RotaryEncoderDriver and polls the state of the rotary encoder at the given interval.
//
// Emits the Events:
//    Turn bool, int - Encoder was turned, clockwise or counter clockwise, position
//    Error error - On encoder error
func (r *RotaryEncoderDriver) Start() (err error) {
	r.Button.Start()

	go func() {

		cLast := 1

		for {
			c, err := r.connection.DigitalRead(r.clkPin)
			if err != nil {
				r.Publish(Error, err)
			}

			d, err := r.connection.DigitalRead(r.dtPin)
			if err != nil {
				r.Publish(Error, err)
			}

			// Detect the rising edge, then publish an event when the corresponding falling edge is detected
			if c != cLast {
				dXor := c ^ d

				if dXor == 1 && c == 0 {
					// If pins differ and c is HI
					fmt.Println("clockwise", cLast, c, d)
				} else if c == 0 {
					// If pins differ and c is LO
					fmt.Println("counter-clockwise", cLast, c, d)
				}

				cLast = c
			}

			if d != 1 || c != 1 {
				//fmt.Printf("C: %v, D: %v, XOR: %v\n", c, d, c^d)
			}

			select {
			case <-time.After(r.interval):
			case <-r.halt:
				return
			}
		}

	}()

	return
}

// Halt stops polling the button for new information
func (b *RotaryEncoderDriver) Halt() (err error) {
	b.halt <- true
	return
}

// Name returns the RotaryEncoderDrivers name
func (r *RotaryEncoderDriver) Name() string { return r.name }

// SetName sets the RotaryEncoderDrivers name
func (r *RotaryEncoderDriver) SetName(n string) { r.name = n }

// Pin returns the RotaryEncoderDrivers pin
//func (b *RotaryEncoderDriver) Pin() string { return b.pin }

// Connection returns the RotaryEncoderDrivers Connection
func (r *RotaryEncoderDriver) Connection() gobot.Connection { return r.connection.(gobot.Connection) }
