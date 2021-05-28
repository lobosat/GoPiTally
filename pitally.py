#!/usr/bin/python
import socket
import time
import json
import RPi.GPIO as GPIO
import os

# Read in configuration file
config_file = open("/usr/local/etc/pitally/tally_config.json")
config = json.load(config_file)
config_file.close()
ip = config["ip"]
port = config["port"]
poll_interval = config["poll_interval"]
input_num = config["input_num"]


def connect():
    try:
        # print('Trying to connect\n')
        global s
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        s.settimeout(5)
        s.connect((ip, port))
        # print('Connected!')
        return True

    except:
        return False


def get_tally():
    try:
        s.send(b"TALLY\r\n")
        data = s.recv(2048)

    except Exception as e:
        return [False, e]

    lines = data.split(b'\r\n')
    for line in range(len(lines)):
        if lines[line].startswith(b'TALLY OK'):
            return [True, lines[line].split(b' ')[2].decode("UTF-8")]
        else:
            return [True, '']


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


def button_callback(pin):
    os.system('/usr/local/bin/doap.sh on')
    led('north')
    time.sleep(3)
    os.system('reboot')


# Initialize GPIO
GPIO.setmode(GPIO.BOARD)
GPIO.setwarnings(False)

GPIO.setup(7, GPIO.IN, pull_up_down=GPIO.PUD_UP)
GPIO.add_event_detect(7, GPIO.FALLING, callback=button_callback, bouncetime=2000)

# Try to connect to the vMix API.  If unable wait 10 seconds and try again
led('all')
c_status = connect()
while c_status is False:
    time.sleep(10)
    c_status = connect()
led('off')

# Begin a loop to start polling for tally info
tally = [True, '']
while tally[0]:
    tally = get_tally()

    if tally[0] == True and tally[1] != '':
        tally_list = list(tally[1])

        if tally_list[input_num - 1] == '0':
            # print('Input is inactive')
            led('off')
        elif tally_list[input_num - 1] == '1':
            # print('Input is active')
            led('all')
        elif tally_list[input_num - 1] == '2':
            # print('Input is preview')
            led('off')

    time.sleep(poll_interval)

    # Try to reconnect if tally is false
    while tally[0] is False:
        led('all')
        time.sleep(5)
        connect()
        tally = get_tally()
        if tally[0]:
            led('off')

# if we exited the above while loop something broke.  End the script and let systemd try to bring it back up again
print('Something broke')
