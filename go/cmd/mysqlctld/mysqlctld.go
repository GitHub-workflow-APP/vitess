/*
Copyright 2019 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// mysqlctld is a daemon that starts or initializes mysqld and provides an RPC
// interface for vttablet to stop and start mysqld from a different container
// without having to restart the container running mysqlctld.
package main

import (
	"context"
	"os"
	"time"

	"github.com/spf13/pflag"

	"vitess.io/vitess/go/acl"
	"vitess.io/vitess/go/exit"
	"vitess.io/vitess/go/vt/dbconfigs"
	"vitess.io/vitess/go/vt/log"
	"vitess.io/vitess/go/vt/logutil"
	"vitess.io/vitess/go/vt/mysqlctl"
	"vitess.io/vitess/go/vt/servenv"
)

var (
	// mysqld is used by the rpc implementation plugin.
	mysqld *mysqlctl.Mysqld
	cnf    *mysqlctl.Mycnf

	mysqlPort   = 3306
	tabletUID   = uint32(41983)
	mysqlSocket string

	// mysqlctl init flags
	waitTime      = 5 * time.Minute
	initDBSQLFile string
)

func init() {
	servenv.RegisterDefaultFlags()
	servenv.RegisterDefaultSocketFileFlags()
	servenv.RegisterFlags()
	servenv.RegisterGRPCServerFlags()
	servenv.RegisterGRPCServerAuthFlags()
	servenv.RegisterServiceMapFlag()
	// mysqlctld only starts and stops mysql, only needs dba.
	dbconfigs.RegisterFlags(dbconfigs.Dba)
	servenv.OnParse(func(fs *pflag.FlagSet) {
		fs.IntVar(&mysqlPort, "mysql_port", mysqlPort, "MySQL port")
		fs.Uint32Var(&tabletUID, "tablet_uid", tabletUID, "Tablet UID")
		fs.StringVar(&mysqlSocket, "mysql_socket", mysqlSocket, "Path to the mysqld socket file")
		fs.DurationVar(&waitTime, "wait_time", waitTime, "How long to wait for mysqld startup or shutdown")
		fs.StringVar(&initDBSQLFile, "init_db_sql_file", initDBSQLFile, "Path to .sql file to run after mysqld initialization")

		acl.RegisterFlags(fs)
	})
}

func main() {
	defer exit.Recover()
	defer logutil.Flush()

	servenv.ParseFlags("mysqlctld")

	// We'll register this OnTerm handler before mysqld starts, so we get notified
	// if mysqld dies on its own without us (or our RPC client) telling it to.
	mysqldTerminated := make(chan struct{})
	onTermFunc := func() {
		close(mysqldTerminated)
	}

	// Start or Init mysqld as needed.
	ctx, cancel := context.WithTimeout(context.Background(), waitTime)
	mycnfFile := mysqlctl.MycnfFile(tabletUID)
	if _, statErr := os.Stat(mycnfFile); os.IsNotExist(statErr) {
		// Generate my.cnf from scratch and use it to find mysqld.
		log.Infof("mycnf file (%s) doesn't exist, initializing", mycnfFile)

		var err error
		mysqld, cnf, err = mysqlctl.CreateMysqldAndMycnf(tabletUID, mysqlSocket, mysqlPort)
		if err != nil {
			log.Errorf("failed to initialize mysql config: %v", err)
			exit.Return(1)
		}
		mysqld.OnTerm(onTermFunc)

		if err := mysqld.Init(ctx, cnf, initDBSQLFile); err != nil {
			log.Errorf("failed to initialize mysql data dir and start mysqld: %v", err)
			exit.Return(1)
		}
	} else {
		// There ought to be an existing my.cnf, so use it to find mysqld.
		log.Infof("mycnf file (%s) already exists, starting without init", mycnfFile)

		var err error
		mysqld, cnf, err = mysqlctl.OpenMysqldAndMycnf(tabletUID)
		if err != nil {
			log.Errorf("failed to find mysql config: %v", err)
			exit.Return(1)
		}
		mysqld.OnTerm(onTermFunc)

		err = mysqld.RefreshConfig(ctx, cnf)
		if err != nil {
			log.Errorf("failed to refresh config: %v", err)
			exit.Return(1)
		}

		// check if we were interrupted during a previous restore
		if !mysqlctl.RestoreWasInterrupted(cnf) {
			if err := mysqld.Start(ctx, cnf); err != nil {
				log.Errorf("failed to start mysqld: %v", err)
				exit.Return(1)
			}
		} else {
			log.Infof("found interrupted restore, not starting mysqld")
		}
	}
	cancel()

	servenv.Init()
	defer servenv.Close()

	// Take mysqld down with us on SIGTERM before entering lame duck.
	servenv.OnTermSync(func() {
		log.Infof("mysqlctl received SIGTERM, shutting down mysqld first")
		ctx := context.Background()
		if err := mysqld.Shutdown(ctx, cnf, true); err != nil {
			log.Errorf("failed to shutdown mysqld: %v", err)
		}
	})

	// Start RPC server and wait for SIGTERM.
	mysqlctldTerminated := make(chan struct{})
	go func() {
		servenv.RunDefault()
		close(mysqlctldTerminated)
	}()

	select {
	case <-mysqldTerminated:
		log.Infof("mysqld shut down on its own, exiting mysqlctld")
	case <-mysqlctldTerminated:
		log.Infof("mysqlctld shut down gracefully")
	}
}
