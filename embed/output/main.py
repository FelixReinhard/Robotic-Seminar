# OUTPUT 
from machine import Pin, PWM
import network
import socket
from time import sleep

# from micropython import const
# import ustruct


ssid = 'Mate'
password = '12345678'

wlan = network.WLAN(network.STA_IF)
wlan.active(True)
wlan.connect(ssid, password)
while not wlan.isconnected():
    print("waiting to connect to wlan")
    sleep(.25)

addr_info = socket.getaddrinfo("192.168.43.253", 1269) # 192.168.2.203
addr = addr_info[0][-1]
s = socket.socket()
print(addr)
s.connect(addr)

pin_five = Pin(0, Pin.OUT)
pin_hold = Pin(2, Pin.OUT)

pin_vibrator = Pin(3, Pin.OUT)
servo = PWM(Pin(4), freq=50, duty=77)
PAUSE = 0.05

def servo_motion():
    for pos in range(180):
        servo.duty(pos)
        sleep(PAUSE)

    for pos in reversed(range(180)):
        servo.duty(pos)
        sleep(PAUSE)

def vibrator():
    pin_vibrator.value(1)
    sleep(1)
    pin_vibrator.value(0)
# Auth
b = s.recv(1)
print(b == b'm')
s.send(bytes(b'\x00')) 
sleep(1)
try:
    while True:
        r = s.recv(3)
        high_five = r[0] 
        holding_side = r[1]
        holding_interlinked = r[2]
        
        # activates relays
        pin_hold.value(bool(holding_side))
        
        if bool(holding_interlinked):
            servo_motion()
        
        if bool(high_five):
            vibrator()

        print(r)
except:
    pass
s.close()

