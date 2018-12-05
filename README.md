# Decred Chain Analysis Tool


## Requirements
- [Go](https://golang.org/dl/) 1.10.x or 1.11.x
- Running instance of [dcrd](https://docs.decred.org/advanced/manual-cli-install/) binaries (version 1.3.0 and above).
The dcrd running instance should be synced with the latest best block.


## Setting up the Project
- Install the Go version
- Clone the repository
```git
    git clone github.com/raedahgroup/dcrchainanalysis.git
```
- Cd into the project root folder.
- Copy the `sample-dcrchainalyser.conf` file contents to `dcrchainalyser.conf` in your Appdata folder.
Check the [AppData folder](https://docs.decred.org/getting-started/startup-basics/) path from other common applications.
```bash
    cp ./sample-dcrchainalyser.conf {appData-folder}/dcrchainalyser.conf
```
- Copy the `rpcuser` and `rpcpass` from `dcrd.conf` in your Dcrd Appdata folder into `dcrchainalyser.conf` file as `dcrduser` and `dcrdpass` fields.
- `Mainnet` is activated by default if `Simnet` or `Testnet` environments are not set in the `dcrchainalyser.conf`
- For Chain Analysis Tool to run effectively with dcrd, `txindex=1` flag should be set in the dcrd.conf located at the appdata folder.


## Build the Project.
- `cd ` to the root folder with the cloned repository.
- Run `go build . && ./v1` 

