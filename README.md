#Server-Api

A simple Server is used to comunicate between two raspberry pi pico W.

**Behaviour:**
Client will connect to server with password and what type it is (Input, Output).
Server will wait until a Input and Output-client have connected.
After that it will wait for the Input to send some data. this data will then be printed and sent to the Output. 
If one of the connection dies, a close will be sent to the other one.