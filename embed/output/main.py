# OUTPUT 
from machine import Pin
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
pin_inter = Pin(3, Pin.OUT)

# Auth
b = s.recv(1)
print(b == b'm')
s.send(bytes(b'\x00')) 
sleep(1)
try:
    while True:
        # recv 2 bytes as an encoding of the 0 = HIGH FIVE, 1 = , 2 = 
        r = s.recv(3)
        high_five = r[0] 
        holding_side = r[1]
        holding_interlinked = r[2]
        
        pin_five.value(bool(high_five))
        pin_hold.value(bool(holding_side))
        pin_inter.value(bool(holding_interlinked))
        
        print(r)
except:
    pass
s.close()


