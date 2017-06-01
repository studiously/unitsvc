// Copyright © 2017 Meyer Zinn <meyerzinn@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/go-kit/kit/log"
	"github.com/ory/hydra/sdk"
	"github.com/rubenv/sql-migrate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/studiously/classsvc/classsvc"
	"github.com/studiously/classsvc/ddl"
)

var addr string

// hostCmd represents the host command
var hostCmd = &cobra.Command{
	Use:   "host",
	Short: "Start the service.",
	Long: `Starts the service on all transports and connects to a database backend.

This command exposes several environmental variables for controls. You can set environments using "export KEY=VALUE" (Linux/macOS) or "set KEY=VALUE" (Windows). On Linux, you can also set environments by prepending key value pairs: "KEY=VALUE KEY2=VALUE2 classsvc".

All possible controls are listed below. The host process additionally exposes a few flags, which are listed below the controls section.

Core Controls
=============
- DATABASE_DRIVER: The driver to use with the database. Only 'postgres' is currently supported.
- DATABASE_CONFIG: A URL to a persistent backend.

Hydra Controls
==============
A Hydra server is required to perform token introspection and thus authorization. Most endpoints (excepting health and unauthenticated ones) will fail without a valid Hydra server.

- HYDRA_CLIENT_ID: ID for Hydra client.
- HYDRA_CLIENT_SECRET: Secret for Hydra client.
- HYDRA_CLUSTER_URL: URL of Hydra cluster.
- HYDRA_TLS_VERIFY: Whether the client should verify Hydra's TLS.
`,
	Run: func(cmd *cobra.Command, args []string) {
		var logger log.Logger
		{
			logger = log.NewLogfmtLogger(logrus.StandardLogger().Out)
			logger = log.With(logger, "ts", log.DefaultTimestampUTC)
			logger = log.With(logger, "caller", log.DefaultCaller)
		}
		var s classsvc.Service
		{
			// Set up database
			var driver = viper.GetString("database.driver")
			var config = viper.GetString("database.config")

			db, err := sql.Open(driver, config)
			if err != nil {
				logger.Log("msg", "database connection failed", "error", err)
				os.Exit(-1)
			}
			if err := pingDatabase(db); err != nil {
				logger.Log("msg", "database ping attempts failed")
				os.Exit(-1)
			}
			if err := setupDatabase(driver, db); err != nil {
				logger.Log("msg", "database migrations failed", "error", err)
				os.Exit(-1)
			}

			s = classsvc.NewPostgres(db)
		}

		client, err := sdk.Connect(
			sdk.ClientID(viper.GetString("hydra.client.id")),
			sdk.ClientSecret(viper.GetString("hydra.client.secret")),
			sdk.ClusterURL(viper.GetString("hydra.cluster_url")),
			sdk.SkipTLSVerify(viper.GetBool("hydra.tls_verify")),
			sdk.Scopes(),
		)
		if err != nil {
			logger.Log("msg", "could not connect to Hydra cluster", "error", err, "cluster_url", viper.GetString("hydra.cluster_url"))
			os.Exit(-1)
		}

		var h = classsvc.MakeHTTPHandler(s, logger, client)
		errs := make(chan error)
		go func() {
			c := make(chan os.Signal, 100)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			errs <- fmt.Errorf("%s", <-c)
		}()

		go func(address string) {
			logger.Log("transport", "HTTP", "addr", addr)
			errs <- http.ListenAndServe(address, h)
		}(addr)

		logger.Log("exit", <-errs)
	},
}

func init() {
	RootCmd.AddCommand(hostCmd)

	viper.SetDefault("hydra.tls_verify", true)

	hostCmd.Flags().StringVarP(&addr, "addr", "a", ":8080", "HTTP bind address")
}

func setupDatabase(driver string, db *sql.DB) error {
	var migrations = &migrate.AssetMigrationSource{
		Asset:    ddl.Asset,
		AssetDir: ddl.AssetDir,
		Dir:      driver,
	}
	_, err := migrate.Exec(db, driver, migrations, migrate.Up)
	return err
}

func pingDatabase(db *sql.DB) (err error) {
	for i := 0; i < 30; i++ {
		err = db.Ping()
		if err == nil {
			return
		}
		time.Sleep(time.Second)
	}
	return errors.New("database unresponsive")
}
