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
You can also start its fabric appchain by carrying parameter `--chain fabric`.
## Usage
```shell script
goduck [global options] command [command options] [arguments...]
```
#### command
- `deploy`         Deploy BitXHub and pier
- `version`         Components version  
- `init`          Init config home for GoDuck
- `status`          List the status of instantiated components  
- `key`          Create and show key information
- `bitxhub`          Start or stop BitXHub nodes
- `pier`          Operation about pier  
- `playground`          Set up and experience interchain system smoothly
- `info`          Show basic info about interchain system
- `prometheus`          Start or stop prometheus
- `help, h`          Shows a list of commands or help for one command

#### global options
- `--repo value`          Goduck storage repo path
- `--help, -h`

See 
[usage documentation](https://github.com/meshplus/goduck/wiki/%E9%83%A8%E7%BD%B2%E5%B7%A5%E5%85%B7goduck%E4%BD%BF%E7%94%A8%E6%96%87%E6%A1%A3)
 in the wiki.