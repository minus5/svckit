# Service Kit

Your swiss army knife for microservices in Go.

## Motivation and goals
The goal of this library is to provide comprehensive toolkit over basic Go primitives and idioms and thus enable creation of complex backends designed as microservices.

## Library design
Library provides abstraction for many functional areas and technologies commonly used for backend creation, e.g.:  

* metrics & health-checks support
* logging
* messaging
* service mesh
* common patterns like fault tolerance (leader/follower), saga, etc.

## How to get started
Before getting really started familiarize yourself with following core building blocks:

* [goreman](https://github.com/mattn/goreman) process manager 
* [nsq](https://nsq.io/) realtime messaging
* [consul](https://www.consul.io/) service mesh

Once you have some basic idea, have a look at [example services](https://github.com/minus5/examples-services) utilizing these technologies. To see even more follow the **Examples** section below.

## Examples
Practical usage of the library can be found [here](./example).