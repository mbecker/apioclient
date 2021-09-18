# oclient (v1.0)

OAuth 2.0 client library for Go

The current version, v1.0 matches Medium article:
* OAuth 2.0 in Go

## To Install:

```
go get github.com/exyzzy/oclient
go install $GOPATH/src/github.com/exyzzy/oclient
```

## Notes:

* oclient.go is the library, services.json is the config file for the services. Everything else (main.go and templates/*) is an example of how to use it.
* First you'll need to edit oclient/services.json to match the services for which you have set up api accounts. Look at the curent examples, for these services you will only need to set the client_id and client_secret
For production do not include these in src code, but instead serve them from host env variables. Depending on the api you need you may have to adjust the scope. The redirect_uri is set for localhost, change this to your server when you deploy.

* You may wish to adjust consts: GcPeriod, InitAuthTimeout, MaxState (see oclient/oclient.go)

* See main.go and templates/home.html for an example of how to set up the redirect link and authorization requests.

* See main.go and templates/api.html for an example of how to set up the service api requests.



