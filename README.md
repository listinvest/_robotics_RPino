RPino

Combining the power of Raspberry with the versatility of Arduino

Raspberry provides the OS to run Go, the Wifi connectivity, some GPIO ports and the power for the microcontroller. Arduino reads all sensors and report back to the RPi via a dedicate serial line.

A Golang program will read the sensors values periodically and perform few actions:

Present them on HTTP in json and Prometheus format
Trigger an alarm for low temperature
Read the values to a greenhouse visitor 
Activate power sockets (also possible via API) 


Components

The whole solutions uses few tightly integrated components and this schema illustrates all of them:

Wifi  <— RaspberryPi -->  Arduino Uno —>  custom shield -> Sensors
                            \—>  GPIO

Data Flow

The data are being captured by Arduino and stored in Prometheus .

 Temp sensor (DHT11) —>  Arduino —(serial line)—> RPi —(Go program)—>  Prometheus -> Grafana

Polling every x seconds, where X is configurable via config file and its frequency is the half of Prometheus scraping (to limit 0 reads)



Serial commands:

I decided to create a simple interface over the serial line, which would be easy to extend and debug. So no binary encoding for example, but a simple question / answer protocol. This was chosen to ensure the robustness of the communication over a simple serial line: if a reading fail only one value from a sensor is lost, if a json message fail all sensor reading are missing.

 

Sensor values treatment

The program validate the result for example to ask the temperature, it writes on the serial “T?” and it will read the reply and make sure it is what it asked by looking in the reply of the letter T ( example of response is “T: 23” ).
In case of error you can define the behaviour: use the previous reading or set to zero. 
All analog sensors varies relative slowly and the serial communication may truncate a value, for example instead of 157, the advertised value became 15, so a sanity check must be implemented.
The sensor must be within an interval which boundaries are plus or minus a percentage of the previous reading. If not the previous interval is used.


Web interfaces:

Access via HTTP allows you to gather lots of information

/metrics for Prometheus

/json for compact status

/api/socket for power socket mgmt

/api/arduino_reset for ..guess?

/profiler statistics

