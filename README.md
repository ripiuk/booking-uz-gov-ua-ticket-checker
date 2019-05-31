# Tickets checker for https://booking.uz.gov.ua/ 

### Description
This is a tickets checker telegram bot for https://booking.uz.gov.ua/ site. 
You can not only check available tickets with the bot, but also set up some periodic inspection.

Project programming language is [Go 1.12](https://blog.golang.org/go1.12).

### How to run
Install requirements
```bash
make update
```
Build for Linux (`./build/uzTicketsMonitoring` binary script will be created):
```bash
make build
```
Build for Windows (`build/uzTicketsMonitoring.exe` file will be created):
```shell
make build-windows
```
Create new file `credentials.yaml` and paste your personal token:
```yaml
token: PASTE_YOUR_TOKEN_HERE
```
Run this bot locally:
```bash
go run main.go
```

### Tests
Run tests:
```bash
make test
```
Run tests with coverage:
```bash
make test-cov
```
Clean after tests:
```bash
make clean
```