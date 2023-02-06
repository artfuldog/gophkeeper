# GophKeeper

## About

GophKeeper is private data managment client-server solution.

It was developed as graduation diploma of [Yandex Practicum's "Advanded Go-Developer" course.](https://practicum.yandex.ru/go-advanced/)

## Features:

- Four types of secret items for storing private data - Login, Card, Note, Data
- No pre-requirements for client
- Graphical terminal user interface (GTUI) for Windows, MacOS, Linux
- Two-factor authentication
- Server-client TLS authenticaion and ecryption 
- Generation TOTP Verification codes
- Client and server side encryption
- Server horizontal scaling

### Secret Items

1) Login - personal login information (username, password, authenticato key)
2) Card - credit card information (number, cardholder, month and year expiration, cvv)
3) Note - text information
4) Data - binary files

All items may have notes field for store related to item information.

All items may have custom fields for storing additional information (**currently not supported in GTUI**):
- Text - key-value in plain text
- Hidden - same as text, but value is considered as sensitive infomation
- Bool - flag

Login items may have URI fields for storing related web-pages information (**currently not supported in GTUI**).

Detailed description of items field is provided on DB schema below.

### Data Safety

Connections between client and server is secured by server-side TLS. All data is encrypted, only server needs to provide its certitcate to client.

All privite data (items), such as notes, secret's information, custom fields and URIs store encrypted with AES256-GCM in server's database. Data is encrypted by user's personal encryption key, which also stores encrypted in server's database. Encryption key in turn encrypted with user's secret key. This secret key is used only on client and is stored in client's config.

In general, to access private data client after succesfull login receives encryption key from server, decrypts this key with it's own secret key, then client is able to decrypt private data with encryption key.

User login process described in following figure:

![UserLoginProcess](./doc/user_login_process.drawio.svg)

Processes with items described in following figure:

![ItemIteraction](./doc/item_iteraction.drawio.svg)


## Client

Client is a ready-to-use terminal application, provided graphical user interface for interacting with server.
Graphical interface support keyboard and mouse control.

Client's configuration parameters:
- Username
- Server address
- Secret key
- E-mail
- Show sensitive (when enabled all item's value is shown by default, when disabled sensitive data is hidden)
- CA certificate path - path to custom CA root certificate in case server's certificate is signed by unknown authority of self-signed certificates used.

**First time setup and user registration process:**

![first_setup_and_login](./doc/first_setup_and_login.gif)


**Main menu:**
![main_menu](./doc/main_menu.gif)


**CRUD operations:**
![vault_browse](./doc/vault_browse.gif)


## Server

Server provides gRPC-API for users and items operations. For data store server uses external database.

Several server instances can be launched for distribute workload. In such cases all instances should use same server password, which used for generation AuthTokens.

### Database

Database driver is based on pgx4 package, so supported database type - PostgesSQL 10 and higher.

General DB Schema:
![DBSchema](./doc/db_scheme.drawio.svg)

### AuthTokens and TLS authentication/encryption.

Currently server supports PASETO tokens for authentication and authorization user's request. Token expiration period is configurable parameter (by default equals 1800 seconds).

For TLS valid certificate and key should be passed via flags or envvars. For testing purposes TLS can be disabled. Also you can generate self-signed with `make certs` command

### Configuration parameters

Server's configuration parameters is described in documentation and can be viewed by `--help` option.

Parameters can be passed via environmental variables and cli arguments (flags). However, general recommendation is pass credentials **only** via envvars.

## **Roadmap, currently not implemeneted**
- Custom fields and URIs (for Login items) support in GUI
- Reprompt password to show sensitive information for flagged items
- Client local working mode, with storing data encrypted locally

## Documentation

Packages documentation is available in godoc format.

To run godoc server:

```
go install -v golang.org/x/tools/cmd/godoc@latest
godoc -http=:6060
```

Then open in browser - http://localhost:6060/pkg/github.com/artfuldog/gophkeeper/?m=all

## Development
All repo can be launched in MS VScode devcontainer.

Some of useful commands, such as generate protofiles, mocks, perform test and run client/server, provided in Makefile. To view it run `make help`.