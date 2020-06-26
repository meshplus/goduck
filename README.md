# Go Duck
Goduck is a command-line management tool that can help to run BitXHub.
## Quick Start
### Installation
```shell script
git clone git@github.com:meshplus/goduck
cd goduck
sudo make install
```
### Initialization
```shell script
goduck init
```
### Start BitXHub
```shell script
goduck bitxhub start
```
The command will initialize and start BitXHub nodes in solo mode.
### Start Pier
```shell script
goduck pier start
```
The command will start pier and its ethereum appchain.   
You can also start its fabric appchain by carrying parameter "--chain fabric".
## Usage
```shell script
goduck [global options] command [command options] [arguments...]
```
#### command
- `version`         Goduck version  
- `init`          init config home for goduck  
- `status`          List the status of instantiated components  
- `fabric`          Operation about fabric network
- `ether`          Operation about ethereum chain]
- `key`          Create and show key information
- `bitxhub`          start or stop BitXHub nodes
- `pier`          Operation about pier  
- `help, h`          Shows a list of commands or help for one command

#### global options
- `--repo value`          GoDuck storage repo path
- `--help, -h`