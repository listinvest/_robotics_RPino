
#include <TroykaDHT.h>

int DHTPin = 10;
int LightPin = A0;
int light = 0;
int RainPin = A2;
int rain = 0;
int pirPin = 16;
int ledPin = 13;
DHT dht(DHTPin, DHT11);
String incoming;
String temp;
String humidity;
String dht_status;
int pirValue;

void(* resetFunc) (void) = 0;//declare reset function at address 0

void setup() {
  Serial.begin(9600);
  incoming = "";
  dht.begin();
  Serial.println("RPino ready!");
  pinMode(ledPin, OUTPUT);
  pinMode(pirPin, INPUT);
  digitalWrite(ledPin, LOW);
}

void loop() {
  
  if (Serial.available() > 0) {
    incoming = Serial.readStringUntil("\n");
    dht.read();
    switch(dht.getState()) {
      case DHT_OK:
       temp = (int) dht.getTemperatureC();
       humidity = (int) dht.getHumidity();
       dht_status = "ok";
       break;
     case DHT_ERROR_CHECKSUM:
       dht_status = "Checksum error";
       break;
     case DHT_ERROR_TIMEOUT:
       dht_status = "Time out error";
       break;
     case DHT_ERROR_NO_REPLY:
        dht_status = "Sensor not connected";
       break;
    }  
    
    if (incoming == "L?\n") {
       light = analogRead(LightPin);
       Serial.print("L: ");
       Serial.println(light);
    }
    if (incoming == "R?\n") {
       rain = analogRead(RainPin);
       Serial.print("R: ");
       Serial.println(rain);
    }  
    if (incoming == "T?\n") {
       Serial.print("T: ");
       Serial.println(temp);
    }
    if (incoming == "H?\n") {
       Serial.print("H: ");
       Serial.println(humidity);
    }
    if (incoming == "P?\n") {
       pirValue = digitalRead(pirPin);
       Serial.print("P: ");
       Serial.println(pirValue);
    }
    if (incoming == "U?\n") {
       Serial.print("U: ");
       Serial.println(millis());
    }
    if (incoming == "X?\n") {
        Serial.print("resetting..");
        resetFunc();
    }
    if (incoming == "S?\n") {
       Serial.print("S: ");
       Serial.println(dht_status);
       digitalWrite(ledPin, HIGH);
      delay(1000);
      digitalWrite(ledPin, LOW);
    }
  }  
  //Serial.println();
  delay(1000);
}
