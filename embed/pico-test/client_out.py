from machine import Pin
from time import sleep 

pin = Pin(0, Pin.OUT)

pin.value(1)
sleep(30)
pin.value(0)