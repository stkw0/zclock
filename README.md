## What is this?

zclock is a small program that prints an emoji according to the [traditional Chinese calendar](https://en.wikipedia.org/wiki/Earthly_Branches#Twelve_branches). It's intended to be used in shell prompts to print the current hour using an animal emoji instead of numbers. The only output returned when executing the program is an emoji from the next list: '🐀', '🐂', '🐅', '🐇', '🐉', '🐍', '🐎', '🐐', '🐒', '🐓', '🐕', '🐖'. See previous link to know which hours correspond to each animal. 

The first execution of `zclock` would require an internet connection.

It also has two options. One which returns the public IP address for your current connection and another one which returns the geolocation coordinates for your public IP address.

## How it works?

`zclock` reads the public IPv4 address from an external service. Then uses the IP2Location LITE data available from http://www.ip2location.com to get the aproximate geolocation for the given IPv4 address. With that the program calculates the solar position for the current longitude which allows to show an accurate hour at the given location. 

The public IPv4 is saved in /tmp directory so we don't need to make too much requests to external services. 
