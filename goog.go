package goog

import (
  "github.com/hiphoox/goog/rest"
  "net/url"
  l4g "code.google.com/p/log4go"
)

type DataBase struct {
  url        string
  name       string
  basicToken string
}

func Connect(server, database_name, login, password string) (DataBase, error) {
  l4g.Trace("Starting connection...")
  var db DataBase

  // We don't need any GET vars.
  requestVariables := url.Values{}

  // Create client
  client, err := rest.New(server)

  if err != nil {
    db = DataBase{}
    client.SetBasicAuth(login, password)
    err = client.Get("", "/connect/" + database_name, requestVariables)
    if err != nil {
      l4g.Trace(err)
    }
  }

  return db, err
}

