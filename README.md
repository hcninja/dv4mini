# dv4mini

dv4mini is a interfacing library for the dv4mini dongle written in go.

## Serial configuration

```
BaudRate:        115200,
DataBits:        8,
StopBits:        1,
ParityMode:      0,
```

## USB commands

* SETADFQRG (0x01)
* SETADFMODE (0x02)
* FLUSHTXBUF (0x03)
* ADFWRITE (0x04)
* ADFWATCHDOG (0x05)
* ADFGETDATA (0x07)
* ADFGREENLED (0x08)
* ADFSETPOWER (0x09)
* ADFDEBUG (0x10)
* ???????? (0x11)
* STRFWVERSION (0x12)
* ???????? (0x13)
* ???????? (0x14)
* ADFSETSEED (0x17)
* ADFVERSION (0x18)
* ADFSETTXBUF (0x19)
* SETFLASHMODE (0x0b)

### SETADFQRG = 1

```
Set the QRG of Receiver and Transmitter. The RX qrg is set immediately. The TX qrg is set with the next transmission.
Length: 8
Parameters:
    Byte 0 RX-qrg LSB
    Byte 1 RX-qrg
    Byte 2 RX-qrg
    Byte 3 RX-qrg MSB
    Byte 4 TX-qrg LSB
    Byte 5 TX-qrg
    Byte 6 TX-qrg
    Byte 7 TX-qrg MSB
In DMR mode only simplex (RX=TX qrg) is allowed.

 |_________________|435.999.600|435.999.600| < hz
 |_preamble__|C_|S_|0_|1_|2_|3_|4_|5_|6_|7_|8_|9_|10|11|12|13|14|15|16|17|18|19|20|21|22|23|24|
< 71 fe 39 1d 01 08 19 fc d3 70 19 fc d3 70

19 fc d5 70 > 436.000.000
```

### SETADFMODE = 2

```
Set the operating mode
Length: 1
Parameters:
Byte 0 ... mode
mode:
    'D' = Dstar
    'M' = DMR
    'F' = Fusion C4FM
    ''  = P25
    ''  = dPRM
    'T' = TX
    'R' = RX

 |_preamble__|C_|S_|0_|1_|2_|3_|4_|5_|6_|7_|8_|9_|10|11|12|13|14|15|16|17|18|19|20|21|22|23|24|
< 71 fe 39 1d 02 01 44 (D-Star)
< 71 fe 39 1d 02 01 4d (DMR)
< 71 fe 39 1d 02 01 46 (C4FM)
< 71 fe 39 1d 02 01 4d (P25) ???
< 71 fe 39 1d 02 01 4d (dPRM) ???
```

### FLUSHTXBUF = 3

```
Sends all data from the TX buffer
length: 0
no parameters
```

### ADFWRITE = 4

```
Write binary data to the transmit buffer. The data are then transmitted as
soon as the TX buffer is filled (see also ADFSETTXBUF).
The PTT is keyed automatically, a preamble is sent automatically.
Length: 1 to 245
Parameters: binary data stream to be sent. MSB is sent first.

 |_preamble__|C_|S_|0_|1_|2_|3_|4_|5_|6_|7_|8_|9_|10|11|12|13|14|15|16|17|18|19|20|21|22|23|24|
< 71 fe 39 1d 04 24 23 7f f5 9c 4e c8 d2 fc 28 eb bf f5 9c 4e c8 2e 0c 0a 22 0a e8 d0 f8 0e
< 71 fe 39 1d 04 24 23 7d e4 64 93 86 79 79 76 22 d7 e7 41 30 84 2e 0c 0a 22 0a e8 9c f3 73

== Old ==
This command is used to send any data via 70cm transmitter.
First switch ON the transmitter, then start sending data bytes.

Example: to send a D-Star voice stream send 12 byte packets every 20ms. To send a DMR voice stream send 36 byte packets every 30ms.

The data are routed through a fifo, therefore a timing jitter of up to 0,5 seconds is allowed.

When you are done with sending data, switch back to RX. The transmitter will be active until all data from the fifo are sent.
```

### ADFWATCHDOG = 5

```
The DV4mini returns a ADFWATCHDOG message upon receiving this message:
length: 0
no parameter
The DV4mini returns the following data:
Command: ADFWATCHDOG
length: 8
    Byte 0 ... RSSI MSB
    Byte 1 ... RSSI LSB
    Bytes 2,3,4,5,6,7 ... serial number of the DV4mini stick
Used to read the RSSI and also to check if the DV4mini is alive and connected.

 |_preamble__|C_|S_|0_|1_|2_|3_|4_|5_|6_|7_|8_|9_|10|11|12|13|14|15|16|17|18|19|20|21|22|23|24|
< 71 FE 39 1D 05 00
                   |                       8
 |______________|__|__|__|________|________|????
 |_preamble__|C_|S_|0_|1_|2_|3_|4_|5_|6_|7_|8_|9_|10|11|12|13|14|15|16|17|18|19|20|21|22|23|24|
> 71 fe 39 1d 05 28 ff d1 00 01 64 58 87 a0 e8 e6 79 34 55 b5 8d 00 a3 f8 fe bc 41 60 e5 d8 07
  b6 b0 da
> 71 fe 39 1d 05 28 ff 9a 00 01 64 58 87 a0 24 14 3a 4c 8b 59 0e 21 ee db 27 0c 8c 4d bc da 4b
  a8 53 cb
```

### ADFGETDATA = 7

```
Request the contents of the receive buffer.
Length: 0
no parameters
the DV4mini returns an ADFGETDATA message with variable length. The parameter field contains the contents of the RX buffer.

 |_preamble__|C_|S_|0_|1_|2_|3_|4_|5_|6_|7_|8_|9_|10|11|12|13|14|15|16|17|18|19|20|21|22|23|24|
< 71 fe 39 1d 07 00
 |_preamble__|C_|S_|0_|1_|2_|3_|4_|5_|6_|7_|8_|9_|10|11|12|13|14|15|16|17|18|19|20|21|22|23|24|
> 71 fe 39 1d 07 14 c2 04 e8 9a ad 0e aa 6f 91 9f 82 ae ad 7a
> 71 fe 39 1d 07 13 67 29 d5 51 53 54 ef 57 de d9 46 bc b5
> 71 fe 39 1d 07 13 66 3e f1 d4 44 d5 29 55 47 df 75 41 f1
```

### ADFGREENLED = 8

```
Swicth on/of the green LED to show an RX sync
Length: 1
Parameters:
    Byte 0 ... 0=off, 1=on
As soon as the host software recognizes a SYNC it can switch on the green LED.

 |_preamble__|C_|S_|0_|1_|2_|3_|4_|5_|6_|7_|8_|9_|10|11|12|13|14|15|16|17|18|19|20|21|22|23|24|
< 71 fe 39 1d 08 01 01 (On)
 |_preamble__|C_|S_|0_|1_|2_|3_|4_|5_|6_|7_|8_|9_|10|11|12|13|14|15|16|17|18|19|20|21|22|23|24|
< 71 fe 39 1d 08 01 00 (off)
```

### ADFSETPOWER = 9

```
Set the output power of the DV4mini 70cm transmitter.
Length: 1
Parameters:
    Byte 0 ... value from 0(min) to 9(max)

 |_preamble__|C_|S_|0_|1_|2_|3_|4_|5_|6_|7_|8_|9_|10|11|12|13|14|15|16|17|18|19|20|21|22|23|24|
< 71 fe 39 1d 09 01 09
```

### ADFDEBUG = 10

```
This is only sent from DV4mini to the host.
It contains a variable length string with debug messages which can be printed out.
```

### 11

```
NFI

 |_preamble__|C_|S_|0_|1_|2_|3_|4_|5_|6_|7_|8_|9_|10|11|12|13|14|15|16|17|18|19|20|21|22|23|24|
? 71 fe 39 1d 11 04 03 d6 46 ec
```

### GETSTRADFVERSION = 12

Returns string ADF version.

```
 |_preamble__|C_|S_|0_|1_|2_|3_|4_|5_|6_|7_|8_|9_|10|11|12|13|14|15|16|17|18|19|20|21|22|23|24|
< 71 fe 39 1d 12 00
> 71 fe 39 1d 12 07 56 30 31 2e 36 34 00
```

### 13

```
NFI

Is 0f the only value?

 |_preamble__|C_|S_|0_|1_|2_|3_|4_|5_|6_|7_|8_|9_|10|11|12|13|14|15|16|17|18|19|20|21|22|23|24|
? 71 fe 39 1d 13 01 0f
```

### 14

```
NFI

 |_preamble__|C_|S_|0_|1_|2_|3_|4_|5_|6_|7_|8_|9_|10|11|12|13|14|15|16|17|18|19|20|21|22|23|24|
? 71 fe 39 1d 14 02 00 00

 |_preamble__|C_|S_|0_|1_|2_|3_|4_|5_|6_|7_|8_|9_|10|11|12|13|14|15|16|17|18|19|20|21|22|23|24|
? 71 fe 39 1d 14 02 01 00
```



### ADFSETSEED = 17

```
Write a random number to the DV4mini stick which is then used as seed for
random functions. Length: 4
Parameters:
    Byte 0 value LSB Byte 1 value
    Byte 2 value
    Byte 3 value MSB
"value" is a random numer. The host PC usually takes this from the internal clock.

 |_preamble__|C_|S_|0_|1_|2_|3_|4_|5_|6_|7_|8_|9_|10|11|12|13|14|15|16|17|18|19|20|21|22|23|24|

```

### ADFVERSION = 18

```
Length: 0
no parameters
The DV4mini returns an ADFVERSION message including a variable length string with the version number of the firmware.

 |_preamble__|C_|S_|0_|1_|2_|3_|4_|5_|6_|7_|8_|9_|10|11|12|13|14|15|16|17|18|19|20|21|22|23|24|

```

### ADFSETTXBUF = 19

```
Sets the size of the transmit buffer. This buffer is used to handle gaps in the received data stream. Length: 1
Parameters:
    Byte 0 ... buffer size in 100ms steps (min:100ms max:1500ms which is min:1 and max:15)

 |_preamble__|C_|S_|0_|1_|2_|3_|4_|5_|6_|7_|8_|9_|10|11|12|13|14|15|16|17|18|19|20|21|22|23|24|

```

### SETFLASHMODE = 0b

```
Enter into firmware flash mode

Length: 1
params:
    Byte 0 0x01

 |_preamble__|C_|S_|0_|1_|2_|3_|4_|5_|6_|7_|8_|9_|10|11|12|13|14|15|16|17|18|19|20|21|22|23|24|
< 71 fe 39 1d 0b 01 01
```

