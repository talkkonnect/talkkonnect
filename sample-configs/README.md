# talKKonnect sample xml configuration files

## This Readme file describes the various tested hardware scenarios and provides a corresponding sample configuration xml template


---
### talkkonnect is tested and works on the following patforms

* Raspberry Pi 3 and 3B+
* Raspnerry Pi A+
* Orange Pi Zero
* x86 Linux PC/VM
---
#### Desktop Transceiver Form Factor
#### (Use Case 1) Hitachi HD44780 LCD Display Variants 
* 4 x 20 LCD controller Parallel Interface
* 4 x 20 LCD controller I2C Interface

#### (Use Case 2) OLED Variants 
* 0.96 Inch 4Pin IIC I2C OLED Display Module 12864 LED
* 1.3  Inch 4Pin IIC I2C OLED Display Module 12864 LED

#### (Use Case 3) For use in Datacenters or on Desktop PCs (Without Local Display and no GPIO)
* x86 PCs Running Linux
* AMD PCs Running Linux
---

#### For Raspberry Pi 3, 3B+, A Models with HD44780 4x20 LCD with Parallel Interface use the template
````
/home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/sample-configs/talkkonnect-raspberrypi-hd44780-parallel.xml
`````

#### For Raspberry Pi 3, 3B+, A Models with HD44780 4x20 LCD with I2C Interface use the template
````
/home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/sample-configs/talkkonnect-raspberrypi-hd44780-i2c.xml
````

#### For Raspberry Pi 3, 3B+, A Models with 4Pin IIC I2C OLED Display Module 12864  use the template
````
/home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/sample-configs/talkkonnect-raspberrypi-OLED-i2c.xml
````

#### For PCs/VMs use the template
````
/home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/sample-configs/talkkonnect-pc-nogpio.xml
````


## Contributing 
We invite interested individuals to provide feedback and improvements to the project. Currently we do not have a WIKI so send feedback to <suvir@talkkonnect.com> or open and Issue in github
you can also check my blog  [www.talkkonnect.com](https://www.talkkonnect.com) for updates on the project

Please visit our [blog](www.talkkonnect.com) for our blog or [github](github.com/talkkonnect) for the latest source code and our [facebook](https://www.facebook.com/talkkonnect) page for future updates and information. 
You can also [download](https://talkkonnect.com/wp-content/uploads/2019/01/Readme-13-01-2019.pdf) an "OLDER" PDF version with pictures of this document.

## License 
[talKKonnect](http://www.talkkonnect.com) is open source and available under the MPL V2.00 license.

<suvir@talkkonnect.com> Updated 17/April/2019  talkkonnect version 1.42.13 is the latest release as of this writing.

