# PGP Server

PGP Server application based on Ethereum address identity. 

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

- Install Go

### Installing

Use go to clone the project

```
go get git.proxeus.com/web/pgp_service
```

## Running the tests

There are unit tests at test package. To run it:

```
cd test
go test
```

## Run the server

To run the server:

```
go run main.go
```

Some flags can also be passed to the previous comand. To see a list of possible flags and a description:

```
go run main.go -h
```

### Available flags

Available flags are:
- **storageDir** - to change the directory and name of the database (default is the current directory with 'database' name)
- **serverAddress** - to change the host and port where the server will run (default is ':8080')
- **contractAddress** - ProxeusFS contract address (default is current directory)

Example: to change databaseName to 'anotherName':

```
go run main.go -storageDir=anotherName
```

## How to use

To add a public key:

- Submit a public key with sign validation
    - Front-End makes a HTTP GET request for a challenge
    - Front-End creates a signature from the received challenge
    - Front-End makes a HTTP POST request providing the signature in query string and public key in body form data
    - Server checks if the ethereum address provided in signature matches the identity of the public key
	- If everything goes well - server will record it on a database in a key-value format, with the key being the identity and value the submited key
	
- To find a public key
    - Front-End makes a HTTP GET request with the ethereum address in query string
    - Server will look for the input at the database
    - If everything goes well - server will reply with the corresponding public key

Open the file **test.html** (test purpose only) in a browser. There there is two options:

- Submit a public key without sign validation (test purposes only - remove when production ready)
    - Add the public key in the corresponding field and press submit button
	- Server will use OpenPGP to read the identify from the submited public key and will record it on a database in a key-value format, with the key being the identity and value the submited key
	- If everything goes well - server will reply with the corresponding identity
			
- Find a public key
	- Fill the 'Ethereum address' field with the identity to look for at the database
	- Server will look for the submited input at the database
	- If everything goes well - server will reply with the corresponding public key
	

Note: In order to facilitate the test an identity was already added at the html form. The public key for the corresponding identity was already added to the database. 

## Built With

* [Bolt](https://github.com/boltdb/bolt) - An embedded key/value database for Go
* [Echo](https://echo.labstack.com) - High performance, minimalist Go web framework
