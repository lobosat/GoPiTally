#!/usr/bin/python
import RPi.GPIO as GPIO


def led(color):
    # color can be red, green, yellow, all, off(default)

    # definitions for the GPIO pins on the PiTraffic hat
    red_pins = [11, 16, 29, 36]
    yellow_pins = [13, 18, 31, 38]
    green_pins = [15, 22, 33, 40]
    north_pins = [29, 31, 33]
    south_pins = [11, 13, 15]
    east_pins = [36, 38, 40]
    west_pins = [16, 18, 22]

    # Turn off all LEDs
    for pin in green_pins:
        GPIO.setup(pin, GPIO.OUT)
        GPIO.output(pin, GPIO.LOW)

    for pin in red_pins:
        GPIO.setup(pin, GPIO.OUT)
        GPIO.output(pin, GPIO.LOW)

    for pin in yellow_pins:
        GPIO.setup(pin, GPIO.OUT)
        GPIO.output(pin, GPIO.LOW)

    if color == 'red' or color == 'all':
        for pin in red_pins:
            GPIO.setup(pin, GPIO.OUT)
            GPIO.output(pin, GPIO.HIGH)

    if color == 'yellow' or color == 'all':
        for pin in yellow_pins:
            GPIO.setup(pin, GPIO.OUT)
            GPIO.output(pin, GPIO.HIGH)

    if color == 'green' or color == 'all':
        for pin in green_pins:
            GPIO.setup(pin, GPIO.OUT)
            GPIO.output(pin, GPIO.HIGH)

    if color == 'north':
        for pin in north_pins:
            GPIO.setup(pin, GPIO.OUT)
            GPIO.output(pin, GPIO.HIGH)


GPIO.setmode(GPIO.BOARD)
GPIO.setwarnings(False)

GPIO.setup(7, GPIO.IN, pull_up_down=GPIO.PUD_UP)

led("all")
