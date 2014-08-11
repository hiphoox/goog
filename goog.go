package goog

import (
  "github.com/hiphoox/goog/rest"
  l4g "code.google.com/p/log4go"
)

const (
  CONNECT_URL = "/connect/"
  HTTP_PREFIX = "http://"
)

type DataBase struct {
  name       string
  server      string
  client     *rest.Client
}

func Connect(server, database_name, login, password string) (DataBase, error) {
  l4g.Trace("Inside Connect")
  var db DataBase

  // Create client
  client, err := rest.New(HTTP_PREFIX + server)

  if err == nil {
    l4g.Trace("Getting token...")
    client.SetBasicAuth(login, password)
    err = client.GetHeaders(CONNECT_URL + database_name)

    if err == nil {
     l4g.Trace("Creating database structure...")
      db = DataBase{  name: database_name, 
                    server: server,
                    client: client}
    }
  }

  return db, err
}

func (db *DataBase) GetToken() string {
    return db.client.GetHeader("Set-Cookie")
}
