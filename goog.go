package goog

import (
  "github.com/gosexy/rest"
  l4g "code.google.com/p/log4go"
)

type DB struct {

}

func Connect(server, database_name, login, password string) (DB, error) {
  l4g.Trace("Starting connection...")
  var db DB
  client, err := rest.New(database_name)

  if err != nil {
    db = DB{}
    client.SetBasicAuth(login, password)

  }

  return db, err
}

