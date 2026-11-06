//go:build tinygo && esp32s3

package main

import (
	"example.com/stopwatch-crab/internal/crab"
	"example.com/stopwatch-crab/internal/stopwatch"
	"machine"
	"time"
)

func main() {
	leftPin, rightPin := machine.GPIO2, machine.GPIO1
	leftPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	rightPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	var display stopwatch.Display
	if err := display.Init(); err != nil {
		for {
			println("StopWatch display/power initialization failed:", err.Error())
			time.Sleep(time.Second)
		}
	}
	display.Draw(0, 0, crab.Size, crab.Size, 0, 0)
	var leftButton, rightButton crab.Button
	left, right := 0, 0
	start := time.Now()
	nextFrame := time.Duration(0)
	for {
		elapsed := time.Since(start)
		l := leftButton.Update(!leftPin.Get(), elapsed.Milliseconds())
		r := rightButton.Update(!rightPin.Get(), elapsed.Milliseconds())
		if elapsed >= nextFrame {
			nl, nr := crab.Approach(left, l), crab.Approach(right, r)
			if nl != left {
				display.Draw(58, 102, 120, 120, nl, nr)
				display.Draw(172, 370, 16, 16, nl, nr)
			}
			if nr != right {
				display.Draw(288, 102, 120, 120, nl, nr)
				display.Draw(278, 370, 16, 16, nl, nr)
			}
			left, right = nl, nr
			nextFrame = time.Since(start) + 30*time.Millisecond
		}
		time.Sleep(3 * time.Millisecond)
	}
}
