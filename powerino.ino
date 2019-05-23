#include "PinChangeInterrupt.h"
#include "avr/wdt.h"
#define meter_pin 4 //it will trigger the interrupt
#define acs_pin 0 //measure Ampere
#define zmpt_pin 0 //measure Volts
volatile byte burp=0;
String incoming;
double raw_value = 0;
double voltage = 0;
double amps = 0;
//double Vcc;
//unsigned int ADCValue;

void setup() {
  Serial.begin(9600);
  Serial.println("Powerino started!");
  pinMode(meter_pin, INPUT_PULLUP);
  pinMode(LED_BUILTIN, OUTPUT);
   wdt_enable(WDTO_1S);
  // Attach the new PinChangeInterrupt and enable event function below
  attachPCINT(digitalPinToPCINT(meter_pin), counter, FALLING);
}

void counter(void) {
  // Switch Led state
  digitalWrite(LED_BUILTIN, !digitalRead(LED_BUILTIN));
  burp++;
}

void loop() {
  incoming = Serial.readStringUntil("\n");
  if (incoming=="P?\n")
  {
    Serial.print("P: ");
    Serial.println(burp, DEC);
    burp = 0;
  }
  if (incoming=="A?\n")
  {
    //DC calculations
    //raw_value = analogRead(acs_pin);
    //voltage = (raw_value / 1024.0) * 5000; 
    //amps = ((voltage - 2500) / 100);
    //AC calculations
    raw_value = getVPP(acs_pin);
    voltage = (raw_value / 2.0) * 0.707;
    amps = voltage * 10;
    Serial.print("A: ");
    Serial.println(amps,3);
  }
  if (incoming=="V?\n")
  {
    raw_value = getVPP(zmpt_pin);
    Serial.print("V: ");
    Serial.println(raw_value,3);
  }
  if (incoming == "S?\n") {
       Serial.println("S: ok");
       digitalWrite(LED_BUILTIN, HIGH);
      delay(1000);
      digitalWrite(LED_BUILTIN, LOW);
   }
   wdt_reset();  
}

float getVPP(int pin)
{
  float result;
  int readValue;             //value read from the sensor
  int maxValue = 0;          // store max value here
  int minValue = 1024;          // store min value here
  uint32_t start_time = millis();
  while((millis()-start_time) < 1000) //sample for 1 Sec
  {
      readValue = analogRead(pin);
      // see if you have a new maxValue
      if (readValue > maxValue) 
      {
          /*record the maximum sensor value*/
         maxValue = readValue;
      }
      if (readValue < minValue) 
      {
          /*record the maximum sensor value*/
         minValue = readValue;
      }
      delayMicroseconds(2);
  }
  // Subtract min from max
  if (pin == acs_pin) {
    result = ((maxValue - minValue) * 5.0)/1024.0;
  } else {  
    result = maxValue - minValue;
  }
 return result;
 }



