#include <Wire.h>
#include <TroykaDHT.h>
#include "Arduino.h"
#define S0 9
#define S1 8
#define S2 10
#define S3 12
#define sensorOut 11
int buzzerPin = 2;

// Stores frequency read by the photodiodes
int Frequency = 0;

int DHTPin = 7;
int LightPin = A0;
int light = 0;
int rain = 0;
unsigned long interval = 0;
unsigned long beat = 0;
int pirPin = A1; 
int ledPin = 13; 
char status;
double T,P;
DHT dht(DHTPin, DHT11);

String incoming;
String temp;
String humidity;
String dht_status;
String melody;
int pirValue;
char sz[] = "M;5555;9999;4444;9999";
void(* resetFunc) (void) = 0;//declare reset function at address 0

void setup() {
  pinMode(buzzerPin, OUTPUT);
  digitalWrite(buzzerPin,LOW);
  Serial.begin(9600);
  incoming = "";
  dht.begin();
  Serial.println("RPino started!");
  pinMode(ledPin, OUTPUT);
  pinMode(pirPin, INPUT);
  digitalWrite(ledPin, LOW);
   // Setting the outputs
  pinMode(S0, OUTPUT);
  pinMode(S1, OUTPUT);
  pinMode(S2, OUTPUT);
  pinMode(S3, OUTPUT);
  // Setting the sensorOut as an input
  pinMode(sensorOut, INPUT);
  // Setting frequency scaling to 20%
  digitalWrite(S0,HIGH);
  digitalWrite(S1,LOW);
  
}

void loop() {
  
  if (Serial.available() > 0) {
    incoming = Serial.readStringUntil("\n");
    dht.read();
    if ( dht.getState() == DHT_OK) {
       temp = (int) dht.getTemperatureC();
       humidity = (int) dht.getHumidity();
       dht_status = "ok";
    } else {
        dht_status = "Sensor not connected";
        temp = '0';
        humidity = '0';
    }  
    
    if (incoming == "L?\n") {
       light = analogRead(LightPin);
       Serial.print("L: ");
       Serial.println(light);
       light = 0;
    }
  
    if (incoming == "T?\n") {
       Serial.print("T: ");
       Serial.println(temp);
    }
    if (incoming == "H?\n") {
       Serial.print("H: ");
       Serial.println(humidity);
    }
    if (incoming == "U?\n") {
       pirValue = analogRead(pirPin);
       Serial.print("U: ");
       Serial.println(pirValue);
       pirValue = 0;
    }
    if (incoming == "I?\n") {
       Serial.print("I: ");
       if (interval == 0) {
        Serial.println(0);
       } else {
        beat = millis() - interval;
        Serial.println(beat);
       }
       interval = millis();
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
    if (incoming == "R?\n") {
    // Setting RED (R) filtered photodiodes to be read
    digitalWrite(S2,LOW);
    digitalWrite(S3,LOW);
    // Reading the output frequency
    Frequency = pulseIn(sensorOut, LOW);
    Serial.print("R: ");
    Serial.println(Frequency);
    } 
    if (incoming == "G?\n") {
    // Setting GREEN (R) filtered photodiodes to be read
    digitalWrite(S2,HIGH);
    digitalWrite(S3,HIGH);
    Frequency = pulseIn(sensorOut, LOW);
    Serial.print("G: ");
    Serial.println(Frequency);
    }
    if (incoming == "B?\n") {
    // Setting Blue (R) filtered photodiodes to be read
    digitalWrite(S2,LOW);
    digitalWrite(S3,HIGH);
    Frequency = pulseIn(sensorOut, LOW);
    Serial.print("B: ");
    Serial.println(Frequency);
    }
    if (incoming == "C?\n") {
    // Setting clear () filtered photodiodes to be read
    digitalWrite(S2,HIGH);
    digitalWrite(S3,LOW);
    Frequency = pulseIn(sensorOut, LOW);
    Serial.print("C: ");
    Serial.println(Frequency);
    }      
    if (incoming.startsWith("M;")) {
      Serial.print("music!");
    // Convert from String Object to String.
    char buf[sizeof(sz)];
    incoming.toCharArray(buf, sizeof(buf));
    char *p = buf;
    char *str;
    while ((str = strtok_r(p, ";", &p)) != NULL) // delimiter is the semicolon
      Serial.println(str);
        //tone(atoi(str));
      //tone(4,261);
      //delay(1000);
      //tone(4,277);
      //delay(1000);
    }
  }
  //Serial.println();
  delay(500);
}
