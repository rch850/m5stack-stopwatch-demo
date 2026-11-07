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
	for {
		if err := display.Init(); err != nil {
			println("StopWatch display/power initialization failed; retrying:", err.Error())
			time.Sleep(500 * time.Millisecond)
			continue
		}
		break
	}
	display.Draw(0, 0, crab.Size, crab.Size, 0, 0, 0)
	// Audio capture runs separately so waiting for DMA never blocks the buttons.
	levels := make(chan int, 1)
	go captureVolume(levels)
	var leftButton, rightButton crab.Button
	left, right := 0, 0
	mouth, targetMouth := 0, 0
	start := time.Now()
	nextFrame := time.Duration(0)
	for {
		select {
		case targetMouth = <-levels:
		default:
		}
		elapsed := time.Since(start)
		l := leftButton.Update(!leftPin.Get(), elapsed.Milliseconds())
		r := rightButton.Update(!rightPin.Get(), elapsed.Milliseconds())
		if elapsed >= nextFrame {
			nl, nr := crab.Approach(left, l), crab.Approach(right, r)
			if nl != left {
				display.Draw(58, 102, 120, 120, nl, nr, targetMouth)
				display.Draw(172, 370, 16, 16, nl, nr, targetMouth)
			}
			if nr != right {
				display.Draw(288, 102, 120, 120, nl, nr, targetMouth)
				display.Draw(278, 370, 16, 16, nl, nr, targetMouth)
			}
			if targetMouth != mouth {
				display.Draw(crab.MouthX, crab.MouthY, crab.MouthWidth, crab.MouthHeight, nl, nr, targetMouth)
				mouth = targetMouth
			}
			left, right = nl, nr
			nextFrame = time.Since(start) + 30*time.Millisecond
		}
		time.Sleep(3 * time.Millisecond)
	}
}

func captureVolume(levels chan<- int) {
	var mic stopwatch.Microphone
	if err := mic.Init(); err != nil {
		println("Microphone initialization failed:", err.Error())
		return // Keep the crab and buttons usable if audio fails.
	}
	var samples [256]uint16 // 16 ms at 16 kHz
	var volume crab.Volume
	warmup := time.Now()
	announced := false
	for {
		n, err := mic.Read(samples[:])
		if err != nil {
			println("Microphone read failed:", err.Error())
			levels <- 0
			return
		}
		// Drain DMA while the analog stage warms up; do not animate startup pops.
		if time.Since(warmup) < time.Second {
			continue
		}
		if !announced {
			println("Microphone capture active (16000 Hz)")
			announced = true
		}
		level := volume.Update(samples[:n])
		select {
		case levels <- level:
		default: // Keep draining audio while the display is busy.
		}
	}
}
