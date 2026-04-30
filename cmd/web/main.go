package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/mattn/go-sqlite3"
	"notes.fritz.box/internal/models"
)

type application struct {
	logger         *slog.Logger
	notes          models.NoteModelInterface
	users          models.UserModelInterface
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
	debugMode      bool
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	dbPath := flag.String("db", "", "Path to the SQLite database")
	tlsCert := flag.String("cert", "", "Path to the TLS cert")
	tlsKey := flag.String("key", "", "Path to the TLS key")
	debug := flag.Bool("debug", false, "Debug mode with more detailed logging")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if len(*dbPath) == 0 {
		logger.Error("database command line argument was not set")
		os.Exit(1)
	}
	if len(*tlsCert) == 0 || len(*tlsKey) == 0 {
		logger.Error("tls cert and/or key command line argument was not set")
		os.Exit(1)
	}

	db, err := openDB(*dbPath)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	app := &application{
		logger:         logger,
		notes:          &models.NoteModel{DB: db},
		users:          &models.UserModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
		debugMode:      *debug,
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	srv := &http.Server{
		Addr:         *addr,
		Handler:      app.routes(),
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Info("starting server", "addr", *addr)

	err = srv.ListenAndServeTLS(*tlsCert, *tlsKey)
	logger.Error(err.Error())
	os.Exit(1)
}

func openDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	sqlStmt :=
		`CREATE TABLE IF NOT EXISTS notes (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, 
		title TEXT NOT NULL,
		content TEXT NOT NULL
		);`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	sqlStmt =
		`CREATE TABLE IF NOT EXISTS users (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, 
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		hashed_password TEXT NOT NULL,
		created TEXT NOT NULL
		);`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	return db, nil
}
