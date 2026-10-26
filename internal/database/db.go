package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

//go:embed migrations/001_initial_schema.sql
var initialSchemaSQL string

type DB struct {
	Writer *sql.DB
	Reader *sql.DB
	Path   string
}

// New inicializa a conexão SQLite com WAL mode e pools separados de escrita e leitura.
func New(customPath string) (*DB, error) {
	dbPath := customPath
	if dbPath == "" {
		dbPath = os.Getenv("DB_PATH")
	}
	if dbPath == "" {
		// Se /data existir e for gravável (ex: no container Docker), usa /data/optimanager.db
		if info, err := os.Stat("/data"); err == nil && info.IsDir() {
			dbPath = "/data/optimanager.db"
		} else {
			dbPath = "./data/optimanager.db"
		}
	}

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("erro ao criar diretório para o banco (%s): %w", dir, err)
	}

	dsn := fmt.Sprintf("%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)&_pragma=cache_size(-8000)", dbPath)

	// Pool de Escrita: 1 única conexão aberta para evitar contenção de escrita
	writer, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("falha ao abrir conexão de escrita: %w", err)
	}
	writer.SetMaxOpenConns(1)
	writer.SetMaxIdleConns(1)

	// Pool de Leitura: até 8 conexões concorrentes para leitura simultânea sem bloqueio
	reader, err := sql.Open("sqlite", dsn)
	if err != nil {
		writer.Close()
		return nil, fmt.Errorf("falha ao abrir conexão de leitura: %w", err)
	}
	reader.SetMaxOpenConns(8)
	reader.SetMaxIdleConns(8)

	// Verifica conectividade
	if err := writer.Ping(); err != nil {
		writer.Close()
		reader.Close()
		return nil, fmt.Errorf("falha no ping da conexão de escrita: %w", err)
	}
	if err := reader.Ping(); err != nil {
		writer.Close()
		reader.Close()
		return nil, fmt.Errorf("falha no ping da conexão de leitura: %w", err)
	}

	instance := &DB{
		Writer: writer,
		Reader: reader,
		Path:   dbPath,
	}

	// Executa as migrações iniciais usando o pool de escrita
	if err := instance.migrate(); err != nil {
		instance.Close()
		return nil, fmt.Errorf("falha ao executar migrações: %w", err)
	}

	log.Printf("[DB] SQLite inicializado com sucesso em %s (WAL mode, Writer: 1 conn, Reader: 8 conns)", dbPath)
	return instance, nil
}

func (d *DB) migrate() error {
	_, err := d.Writer.Exec(initialSchemaSQL)
	return err
}

func (d *DB) Close() {
	if d.Writer != nil {
		d.Writer.Close()
	}
	if d.Reader != nil {
		d.Reader.Close()
	}
}

