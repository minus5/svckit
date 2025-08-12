// Package mongo je server za potrebe pokretanja testova koji koriste mongo bazu
// Djelomicno kopirano iz https://github.com/go-mgo/mgo/blob/v2-unstable/dbtest/dbserver.go
package mongo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Mongo controls a MongoDB server process to be used within test suites.
type Mongo struct {
	session    *mongo.Client
	output     bytes.Buffer
	server     *exec.Cmd
	serverExit chan struct{}
	DbPath     string
	Host       string
}

// New creates new MongoDB server to be used within test suites.
func New() *Mongo {
	dir, err := ioutil.TempDir("", "mongo")
	if err != nil {
		log.Fatal(err)
	}
	m := &Mongo{DbPath: dir, serverExit: make(chan struct{})}
	m.start()
	return m
}

// DBServer controls a MongoDB server process to be used within test suites.
//
// The test server is started when Session is called the first time and should
// remain running for the duration of all tests, with the Wipe method being
// called between tests (before each of them) to clear stored data. After all tests
// are done, the Stop method should be called to stop the test server.
//
// Before the DBServer is used the SetPath method must be called to define
// the location for the database files to be stored.
func (dbs *Mongo) start() {
	if dbs.server != nil {
		panic("DBServer already started")
	}
	if dbs.DbPath == "" {
		panic("DBServer.SetPath must be called before using the server")
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic("unable to listen on a local address: " + err.Error())
	}
	addr := l.Addr().(*net.TCPAddr)
	l.Close()
	dbs.Host = fmt.Sprintf("mongodb://%s", addr.String())

	args := []string{
		"--dbpath", dbs.DbPath,
		"--bind_ip", "127.0.0.1",
		"--port", strconv.Itoa(addr.Port),
	}
	dbs.server = exec.Command("mongod", args...)
	dbs.server.Stdout = &dbs.output
	dbs.server.Stderr = &dbs.output
	if err = dbs.server.Start(); err != nil {
		panic(err)
	}
	dbs.Wipe()
	go dbs.waitServerExit()
}

func (dbs *Mongo) waitServerExit() {
	_, err := dbs.server.Process.Wait()
	if err != nil {
		log.Println(err)
	}
	dbs.server = nil
	close(dbs.serverExit)
}

// Stop stops the test server process, if it is running.
// Uses os.Interrupt signal to stop server (clean exit)
//
// It's okay to call Stop multiple times. After the test server is
// stopped it cannot be restarted.
//
// All database sessions must be closed before or while the Stop method
// is running. Otherwise Stop will panic after a timeout informing that
// there is a session leak.
func (dbs *Mongo) Stop() {
	dbs.stop(false)
}

// Kill kils the test server process, if it is running.
// Uses os.Kill signal to kill server process
//
// It's okay to call Stop multiple times. After the test server is
// stopped it cannot be restarted.
//
// All database sessions must be closed before or while the Stop method
// is running. Otherwise Stop will panic after a timeout informing that
// there is a session leak.
func (dbs *Mongo) Kill() {
	dbs.stop(true)
}

func (dbs *Mongo) stop(kill bool) {
	dbs.closeSession()
	dbs.stopServer(kill)
	os.RemoveAll(dbs.DbPath)
}

func (dbs *Mongo) closeSession() {
	if dbs.session == nil {
		return
	}
	dbs.session.Disconnect(context.Background())
	dbs.session = nil
}

func (dbs *Mongo) stopServer(kill bool) {
	if dbs.server == nil {
		return
	}
	var err error
	if kill {
		err = dbs.server.Process.Kill()
	} else {
		// Iz nekog razloga mongod ignorira Interrupt signal
		//err = dbs.server.Process.Signal(os.Interrupt)
		err = dbs.shutdownServer()
	}
	if err != nil && err != io.EOF {
		log.Printf("%s %T", err, err)
	}
	// Wait for mongo proces to exit
	<-dbs.serverExit
}

func (dbs *Mongo) shutdownServer() error {
	session := dbs.Session()
	return session.Database("admin").RunCommand(context.Background(), bson.D{{"shutdown", 1}}).Err()
}

// Session returns a new session to the server. The returned session
// must be closed after the test is done with it.
//
// The first Session obtained from a DBServer will start it.
func (dbs *Mongo) Session() *mongo.Client {
	if dbs.server == nil {
		dbs.start()
	}
	if dbs.session == nil {
		var err error
		dbs.session, err = mongo.Connect(context.Background(),
			options.Client().ApplyURI(fmt.Sprintf("%s/test", dbs.Host)))
		if err != nil {
			panic(err)
		}
	}
	return dbs.session
}

// Wipe drops all created databases and their data.
//
// The MongoDB server remains running if it was prevoiusly running,
// or stopped if it was previously stopped.
//
// All database sessions must be closed before or while the Wipe method
// is running. Otherwise Wipe will panic after a timeout informing that
// there is a session leak.
func (dbs *Mongo) Wipe() {
	session := dbs.Session()
	names, err := session.ListDatabaseNames(context.Background(), bson.D{{}})
	if err != nil {
		panic(err)
	}
	for _, name := range names {
		switch name {
		case "admin", "local", "config":
		default:
			err = session.Database(name).Drop(context.Background())
			if err != nil {
				panic(err)
			}
		}
	}
}
