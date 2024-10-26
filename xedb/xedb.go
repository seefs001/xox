package xedb

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/seefs001/xox/xerror"
)

var (
	ErrKeyNotFound  = xerror.New("key not found")
	ErrTypeMismatch = xerror.New("type mismatch")
	ErrMemoryLimit  = xerror.New("memory limit exceeded")
	ErrKeyExists    = xerror.New("key already exists")
	ErrInvalidType  = xerror.New("invalid type")
	ErrInvalidValue = xerror.New("invalid value")
)

// Options represents database configuration options
type Options struct {
	// DataDir specifies the directory for data storage
	DataDir string

	// SyncWrite determines if writes are synchronous
	SyncWrite bool

	// AutoSaveInterval specifies the interval for auto-saving
	AutoSaveInterval time.Duration

	// MaxMemory specifies the maximum memory usage (in bytes)
	MaxMemory int64

	// CompactionThreshold triggers compaction when WAL size exceeds this value
	CompactionThreshold int64

	// EnableAOF enables Append-Only File persistence
	EnableAOF bool

	// LogLevel specifies the logging level
	LogLevel string

	// ValueLogFileSize is the maximum size of each value log file
	ValueLogFileSize int64

	// NumVersionsToKeep is the number of versions to keep per key
	NumVersionsToKeep int

	// CompactionL0Trigger is the number of L0 tables that triggers compaction
	CompactionL0Trigger int

	// EnableVersioning enables multi-version support
	EnableVersioning bool

	// MaxVersions specifies maximum versions to keep per key (0 means unlimited)
	MaxVersions int
}

// DefaultOptions returns default configuration options
func DefaultOptions() Options {
	return Options{
		DataDir:             "data",
		SyncWrite:           true,
		AutoSaveInterval:    time.Minute * 5,
		MaxMemory:           1 << 30, // 1GB
		CompactionThreshold: 1 << 20, // 1MB
		EnableAOF:           false,
		LogLevel:            "info",
		ValueLogFileSize:    1 << 30, // 1GB
		NumVersionsToKeep:   1,
		CompactionL0Trigger: 10,
		EnableVersioning:    false,
		MaxVersions:         10,
	}
}

// Option represents a function that sets an option
type Option func(*Options)

// WithDataDir sets the data directory
func WithDataDir(dir string) Option {
	return func(o *Options) {
		o.DataDir = dir
	}
}

// WithSyncWrite sets the sync write option
func WithSyncWrite(sync bool) Option {
	return func(o *Options) {
		o.SyncWrite = sync
	}
}

// WithAutoSaveInterval sets the auto-save interval
func WithAutoSaveInterval(interval time.Duration) Option {
	return func(o *Options) {
		o.AutoSaveInterval = interval
	}
}

// WithMaxMemory sets the maximum memory usage
func WithMaxMemory(size int64) Option {
	return func(o *Options) {
		o.MaxMemory = size
	}
}

func init() {
	// Register types for gob encoding
	gob.Register([]string{})
	gob.Register(map[string]string{})
	gob.Register(map[string]struct{}{})
	gob.Register([]ZSetMember{})
}

// DataType represents supported data types
type DataType int

const (
	String DataType = iota
	List
	Hash
	Set
	ZSet
)

// ZSetMember represents a sorted set member
type ZSetMember struct {
	Member string
	Score  float64
}

// SetMember represents a set member
type SetMember struct {
	Member string
}

// DB represents the main database structure
type DB struct {
	data      map[string]Entry
	mutex     sync.RWMutex
	expires   map[string]time.Time
	txMutex   sync.Mutex
	dataFile  string
	walFile   string
	aofFile   string
	txCounter uint64
	options   Options
	memUsage  int64

	// Channels for control
	stopChan   chan struct{}
	saveChan   chan struct{}
	commitLock sync.Mutex
	txnLock    sync.Mutex
}

// Entry represents a value stored in the database
type Entry struct {
	Type        DataType
	Value       interface{}
	Version     uint64
	Created     time.Time
	LastUpdated time.Time
	Versions    []VersionedEntry `json:",omitempty"`
}

// DataFile represents the binary format of stored data
type DataFile struct {
	Version   uint32
	TxCounter uint64
	Entries   map[string]Entry
}

// WALEntry represents a write-ahead log entry
type WALEntry struct {
	TxID     uint64
	Commands []Command
}

// StringOp provides string operations
type StringOp struct {
	db  *DB
	key string
}

// ListOp provides list operations
type ListOp struct {
	db  *DB
	key string
}

// HashOp provides hash operations
type HashOp struct {
	db  *DB
	key string
}

// SetOp provides set operations
type SetOp struct {
	db  *DB
	key string
}

// ZSetOp provides sorted set operations
type ZSetOp struct {
	db  *DB
	key string
}

// New creates a new database instance with options
func New(opts ...Option) (*DB, error) {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	db := &DB{
		data:     make(map[string]Entry),
		expires:  make(map[string]time.Time),
		options:  options,
		stopChan: make(chan struct{}),
		saveChan: make(chan struct{}),
	}

	// Initialize paths
	db.dataFile = filepath.Join(options.DataDir, "data.db")
	db.walFile = filepath.Join(options.DataDir, "wal.db")
	if options.EnableAOF {
		db.aofFile = filepath.Join(options.DataDir, "appendonly.aof")
	}

	// Create directory if not exists
	if err := os.MkdirAll(options.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Initialize database
	if err := db.initialize(); err != nil {
		return nil, err
	}

	// Start background tasks
	go db.backgroundTasks()

	return db, nil
}

func (db *DB) initialize() error {
	// Create data directory if not exists
	if err := os.MkdirAll(db.options.DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Create data file if not exists
	if _, err := os.Stat(db.dataFile); os.IsNotExist(err) {
		// Initialize empty database file
		if err := db.writeData(); err != nil {
			return fmt.Errorf("failed to create initial data file: %w", err)
		}
	}

	// Recover from WAL if exists
	if err := db.recoverFromWAL(); err != nil {
		return fmt.Errorf("WAL recovery failed: %w", err)
	}

	// Load data file
	if err := db.loadData(); err != nil {
		return fmt.Errorf("data loading failed: %w", err)
	}

	// Initialize AOF if enabled
	if db.options.EnableAOF {
		if err := db.initAOF(); err != nil {
			return fmt.Errorf("AOF initialization failed: %w", err)
		}
	}

	return nil
}

func (db *DB) backgroundTasks() {
	autoSaveTicker := time.NewTicker(db.options.AutoSaveInterval)
	defer autoSaveTicker.Stop()

	for {
		select {
		case <-db.stopChan:
			return

		case <-autoSaveTicker.C:
			if err := db.Save(); err != nil {
				// log error
			}

		case <-db.saveChan:
			if err := db.Save(); err != nil {
				// log error
			}
		}
	}
}

// Save persists the current state to disk
func (db *DB) Save() error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if err := db.writeData(); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	if db.options.EnableAOF {
		if err := db.rotateAOF(); err != nil {
			return fmt.Errorf("failed to rotate AOF: %w", err)
		}
	}

	return nil
}

// Close gracefully shuts down the database
func (db *DB) Close() error {
	close(db.stopChan)

	// Final save
	if err := db.Save(); err != nil {
		return fmt.Errorf("failed to save on close: %w", err)
	}

	return nil
}

// checkMemoryLimit checks if operation would exceed memory limit
func (db *DB) checkMemoryLimit(additionalBytes int64) error {
	if db.options.MaxMemory > 0 {
		if atomic.LoadInt64(&db.memUsage)+additionalBytes > db.options.MaxMemory {
			return ErrMemoryLimit
		}
	}
	return nil
}

// updateMemUsage updates the memory usage counter
func (db *DB) updateMemUsage(delta int64) {
	atomic.AddInt64(&db.memUsage, delta)
}

// writeWAL writes a transaction to the WAL file
func (db *DB) writeWAL(entry WALEntry) error {
	f, err := os.OpenFile(db.walFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open WAL file: %w", err)
	}
	defer f.Close()

	enc := gob.NewEncoder(f)
	if err := enc.Encode(entry); err != nil {
		return fmt.Errorf("failed to encode WAL entry: %w", err)
	}

	if err := f.Sync(); err != nil {
		return fmt.Errorf("failed to sync WAL file: %w", err)
	}

	return nil
}

// persistData writes the current state to disk
func (db *DB) persistData() error {
	data, err := json.Marshal(db.data)
	if err != nil {
		return err
	}

	// Write to temporary file first
	tmpFile := db.dataFile + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}

	// Atomic rename
	if err := os.Rename(tmpFile, db.dataFile); err != nil {
		return err
	}

	// Clear WAL after successful persistence
	return os.Remove(db.walFile)
}

// loadData loads the database state from disk
func (db *DB) loadData() error {
	f, err := os.OpenFile(db.dataFile, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open data file: %w", err)
	}
	defer f.Close()

	// Check if file is empty
	stat, err := f.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file stats: %w", err)
	}

	// Return if file is empty (newly created)
	if stat.Size() == 0 {
		db.data = make(map[string]Entry)
		return nil
	}

	// Decode existing data
	var df DataFile
	dec := gob.NewDecoder(f)
	if err := dec.Decode(&df); err != nil {
		return fmt.Errorf("failed to decode data file: %w", err)
	}

	db.data = df.Entries
	db.txCounter = df.TxCounter
	return nil
}

// recoverFromWAL recovers the database state from WAL
func (db *DB) recoverFromWAL() error {
	f, err := os.Open(db.walFile)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to open WAL file: %w", err)
	}
	defer f.Close()

	dec := gob.NewDecoder(f)
	for {
		var entry WALEntry
		err := dec.Decode(&entry)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to decode WAL entry: %w", err)
		}

		// Apply WAL entries
		for _, cmd := range entry.Commands {
			db.data[cmd.Key] = Entry{
				Type:    cmd.Type,
				Value:   cmd.Value,
				Version: entry.TxID,
				Created: time.Now(),
			}
		}

		if entry.TxID > db.txCounter {
			db.txCounter = entry.TxID
		}
	}

	return nil
}

// String returns string operations for a key
func (db *DB) String(key string) *StringOp {
	return &StringOp{db: db, key: key}
}

// List returns list operations for a key
func (db *DB) List(key string) *ListOp {
	return &ListOp{db: db, key: key}
}

// Hash returns hash operations for a key
func (db *DB) Hash(key string) *HashOp {
	return &HashOp{db: db, key: key}
}

// Set returns set operations for a key
func (db *DB) Set(key string) *SetOp {
	return &SetOp{db: db, key: key}
}

// ZSet returns sorted set operations for a key
func (db *DB) ZSet(key string) *ZSetOp {
	return &ZSetOp{db: db, key: key}
}

// String operations
func (op *StringOp) Set(value string) error {
	op.db.mutex.Lock()
	defer op.db.mutex.Unlock()

	// Check memory limit before setting value
	valueSize := int64(len(value))
	if err := op.db.checkMemoryLimit(valueSize); err != nil {
		return err
	}

	// Update memory usage
	op.db.updateMemUsage(valueSize)

	op.db.data[op.key] = Entry{
		Type:    String,
		Value:   value,
		Version: op.db.txCounter + 1,
		Created: time.Now(),
	}
	op.db.txCounter++
	return op.db.writeData()
}

func (op *StringOp) Get() (string, bool) {
	op.db.mutex.RLock()
	defer op.db.mutex.RUnlock()

	if entry, ok := op.db.data[op.key]; ok && entry.Type == String {
		return entry.Value.(string), true
	}
	return "", false
}

// List operations
func (op *ListOp) Push(values ...string) error {
	op.db.mutex.Lock()
	defer op.db.mutex.Unlock()

	var list []string
	if entry, ok := op.db.data[op.key]; ok && entry.Type == List {
		list = entry.Value.([]string)
	}

	list = append(list, values...)
	op.db.data[op.key] = Entry{
		Type:    List,
		Value:   list,
		Version: op.db.txCounter + 1,
	}
	op.db.txCounter++
	return op.db.writeData()
}

func (op *ListOp) Pop() (string, bool) {
	op.db.mutex.Lock()
	defer op.db.mutex.Unlock()

	if entry, ok := op.db.data[op.key]; ok && entry.Type == List {
		list := entry.Value.([]string)
		if len(list) == 0 {
			return "", false
		}

		value := list[len(list)-1]
		list = list[:len(list)-1]

		op.db.data[op.key] = Entry{
			Type:    List,
			Value:   list,
			Version: op.db.txCounter + 1,
		}
		op.db.txCounter++
		op.db.writeData()
		return value, true
	}
	return "", false
}

// LPush adds elements to the beginning of the list
func (op *ListOp) LPush(values ...string) error {
	op.db.mutex.Lock()
	defer op.db.mutex.Unlock()

	var list []string
	if entry, ok := op.db.data[op.key]; ok && entry.Type == List {
		list = entry.Value.([]string)
	}

	// Prepend values in correct order
	newList := make([]string, len(values)+len(list))
	copy(newList[len(values):], list)
	for i, v := range values {
		newList[i] = v
	}

	op.db.data[op.key] = Entry{
		Type:    List,
		Value:   newList,
		Version: op.db.txCounter + 1,
	}
	op.db.txCounter++
	return op.db.writeData()
}

// LPop removes and returns the first element
func (op *ListOp) LPop() (string, bool) {
	op.db.mutex.Lock()
	defer op.db.mutex.Unlock()

	if entry, ok := op.db.data[op.key]; ok && entry.Type == List {
		list := entry.Value.([]string)
		if len(list) == 0 {
			return "", false
		}

		value := list[0]
		list = list[1:]

		op.db.data[op.key] = Entry{
			Type:    List,
			Value:   list,
			Version: op.db.txCounter + 1,
		}
		op.db.txCounter++
		op.db.writeData()
		return value, true
	}
	return "", false
}

// Range returns a slice of elements from start to stop index
func (op *ListOp) Range(start, stop int) []string {
	op.db.mutex.RLock()
	defer op.db.mutex.RUnlock()

	if entry, ok := op.db.data[op.key]; ok && entry.Type == List {
		list := entry.Value.([]string)
		length := len(list)

		// Convert negative indices
		if start < 0 {
			start = length + start
		}
		if stop < 0 {
			stop = length + stop
		}

		// Validate indices
		if start < 0 || start >= length || stop < 0 || start > stop {
			return nil
		}

		// Adjust stop if needed
		if stop >= length {
			stop = length - 1
		}

		return list[start : stop+1]
	}
	return nil
}

// Len returns the length of the list
func (op *ListOp) Len() int {
	op.db.mutex.RLock()
	defer op.db.mutex.RUnlock()

	if entry, ok := op.db.data[op.key]; ok && entry.Type == List {
		return len(entry.Value.([]string))
	}
	return 0
}

// Hash operations
func (op *HashOp) Set(field string, value string) error {
	op.db.mutex.Lock()
	defer op.db.mutex.Unlock()

	var hash map[string]string
	if entry, ok := op.db.data[op.key]; ok && entry.Type == Hash {
		hash = entry.Value.(map[string]string)
	} else {
		hash = make(map[string]string)
	}

	hash[field] = value
	op.db.data[op.key] = Entry{
		Type:    Hash,
		Value:   hash,
		Version: op.db.txCounter + 1,
	}
	op.db.txCounter++
	return op.db.writeData()
}

func (op *HashOp) Get(field string) (string, bool) {
	op.db.mutex.RLock()
	defer op.db.mutex.RUnlock()

	if entry, ok := op.db.data[op.key]; ok && entry.Type == Hash {
		hash := entry.Value.(map[string]string)
		value, exists := hash[field]
		return value, exists
	}
	return "", false
}

// Set operations
func (op *SetOp) Add(members ...string) error {
	op.db.mutex.Lock()
	defer op.db.mutex.Unlock()

	var set map[string]struct{}
	if entry, ok := op.db.data[op.key]; ok && entry.Type == Set {
		set = entry.Value.(map[string]struct{})
	} else {
		set = make(map[string]struct{})
	}

	for _, member := range members {
		set[member] = struct{}{}
	}

	op.db.data[op.key] = Entry{
		Type:    Set,
		Value:   set,
		Version: op.db.txCounter + 1,
	}
	op.db.txCounter++
	return op.db.writeData()
}

func (op *SetOp) IsMember(member string) bool {
	op.db.mutex.RLock()
	defer op.db.mutex.RUnlock()

	if entry, ok := op.db.data[op.key]; ok && entry.Type == Set {
		set := entry.Value.(map[string]struct{})
		_, exists := set[member]
		return exists
	}
	return false
}

// ZSet operations
func (op *ZSetOp) Add(score float64, member string) error {
	op.db.mutex.Lock()
	defer op.db.mutex.Unlock()

	var zset []ZSetMember
	if entry, ok := op.db.data[op.key]; ok && entry.Type == ZSet {
		zset = entry.Value.([]ZSetMember)
	}

	// Update or add member
	found := false
	for i, m := range zset {
		if m.Member == member {
			zset[i].Score = score
			found = true
			break
		}
	}

	if !found {
		zset = append(zset, ZSetMember{Member: member, Score: score})
	}

	// Sort by score
	sort.Slice(zset, func(i, j int) bool {
		return zset[i].Score < zset[j].Score
	})

	op.db.data[op.key] = Entry{
		Type:    ZSet,
		Value:   zset,
		Version: op.db.txCounter + 1,
	}
	op.db.txCounter++
	return op.db.writeData()
}

func (op *ZSetOp) Range(start, stop int) []ZSetMember {
	op.db.mutex.RLock()
	defer op.db.mutex.RUnlock()

	if entry, ok := op.db.data[op.key]; ok && entry.Type == ZSet {
		zset := entry.Value.([]ZSetMember)
		length := len(zset)

		// Convert negative indices to positive
		if start < 0 {
			start = length + start
		}
		if stop < 0 {
			stop = length + stop
		}

		// Handle out of bounds cases
		if start < 0 || start >= length || stop < 0 || start > stop {
			return nil
		}

		// Adjust stop if it exceeds length
		if stop >= length {
			stop = length - 1
		}

		// Return slice including stop index
		return zset[start : stop+1]
	}
	return nil
}

// writeData with improved error handling and proper cleanup
func (db *DB) writeData() error {
	if db.dataFile == "" {
		return fmt.Errorf("data file path not set")
	}

	df := DataFile{
		Version:   1,
		TxCounter: db.txCounter,
		Entries:   db.data,
	}

	// Create parent directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(db.dataFile), 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	tmpFile := db.dataFile + ".tmp"
	f, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	// Ensure file is closed and cleaned up in case of errors
	defer func() {
		f.Close()
		if err != nil {
			os.Remove(tmpFile)
		}
	}()

	enc := gob.NewEncoder(f)
	if err := enc.Encode(df); err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}

	if db.options.SyncWrite {
		if err := f.Sync(); err != nil {
			return fmt.Errorf("failed to sync file: %w", err)
		}
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Atomic rename with backup
	backupFile := db.dataFile + ".bak"
	if _, err := os.Stat(db.dataFile); err == nil {
		if err := os.Rename(db.dataFile, backupFile); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	if err := os.Rename(tmpFile, db.dataFile); err != nil {
		// Attempt to restore from backup
		if _, err := os.Stat(backupFile); err == nil {
			os.Rename(backupFile, db.dataFile)
		}
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	// Clean up backup file
	os.Remove(backupFile)
	return nil
}

// initAOF initializes the AOF file
func (db *DB) initAOF() error {
	if !db.options.EnableAOF {
		return nil
	}

	f, err := os.OpenFile(db.aofFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open AOF file: %w", err)
	}
	defer f.Close()

	return nil
}

// rotateAOF performs AOF rewrite
func (db *DB) rotateAOF() error {
	if !db.options.EnableAOF {
		return nil
	}

	// Create new AOF file
	newAOF := db.aofFile + ".new"
	f, err := os.Create(newAOF)
	if err != nil {
		return fmt.Errorf("failed to create new AOF file: %w", err)
	}
	defer f.Close()

	// Write current dataset to new AOF
	enc := gob.NewEncoder(f)
	if err := enc.Encode(db.data); err != nil {
		return fmt.Errorf("failed to encode AOF data: %w", err)
	}

	if db.options.SyncWrite {
		if err := f.Sync(); err != nil {
			return fmt.Errorf("failed to sync AOF file: %w", err)
		}
	}

	// Atomic rename
	if err := os.Rename(newAOF, db.aofFile); err != nil {
		return fmt.Errorf("failed to rename AOF file: %w", err)
	}

	return nil
}

// Command represents a database operation
type Command struct {
	Op      string        // Operation type (SET, GET, etc.)
	Key     string        // Key to operate on
	Value   interface{}   // Value for the operation
	Version uint64        // Version number
	Type    DataType      // Data type
	Field   string        // For hash operations
	Score   float64       // For sorted set operations
	TTL     time.Duration // For key expiration
}

// BatchOp represents a batch operation
type BatchOp struct {
	Op    string
	Key   string
	Value interface{}
}

// ExecuteBatch executes multiple operations atomically
func (db *DB) ExecuteBatch(ops []BatchOp) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	// Prepare WAL entry
	walEntry := WALEntry{
		TxID:     atomic.AddUint64(&db.txCounter, 1),
		Commands: make([]Command, len(ops)),
	}

	// Execute operations
	for i, op := range ops {
		entry := Entry{
			Type:    getTypeFromOp(op.Op),
			Value:   op.Value,
			Version: walEntry.TxID,
			Created: time.Now(),
		}
		db.data[op.Key] = entry

		// Record command in WAL
		walEntry.Commands[i] = Command{
			Op:      op.Op,
			Key:     op.Key,
			Value:   op.Value,
			Version: walEntry.TxID,
			Type:    entry.Type,
		}
	}

	// Write WAL
	if err := db.writeWAL(walEntry); err != nil {
		return fmt.Errorf("failed to write WAL: %w", err)
	}

	return db.writeData()
}

// getTypeFromOp converts operation string to DataType
func getTypeFromOp(op string) DataType {
	switch op {
	case "STRING":
		return String
	case "LIST":
		return List
	case "HASH":
		return Hash
	case "SET":
		return Set
	case "ZSET":
		return ZSet
	default:
		return String // Default to String type
	}
}

// Add new transaction types
type Txn struct {
	db       *DB
	readTs   uint64
	writes   map[string]*Entry
	pending  []*Entry
	conflict map[string]struct{}
	mutex    sync.Mutex
	readOnly bool
}

// Add iterator interface
type Iterator struct {
	db      *DB
	prefix  []byte
	reverse bool
	curr    string
	mutex   sync.RWMutex
}

// Add transaction methods
func (db *DB) NewTransaction(update bool) *Txn {
	db.txnLock.Lock()
	defer db.txnLock.Unlock()

	return &Txn{
		db:       db,
		readTs:   atomic.LoadUint64(&db.txCounter),
		writes:   make(map[string]*Entry),
		conflict: make(map[string]struct{}),
		readOnly: !update,
	}
}

func (txn *Txn) Get(key string) (Entry, error) {
	txn.mutex.Lock()
	defer txn.mutex.Unlock()

	// Check writes first
	if entry, ok := txn.writes[key]; ok {
		return *entry, nil
	}

	// Read from DB with snapshot isolation
	txn.db.mutex.RLock()
	defer txn.db.mutex.RUnlock()

	if entry, ok := txn.db.data[key]; ok {
		if entry.Version <= txn.readTs {
			return entry, nil
		}
	}

	return Entry{}, ErrKeyNotFound
}

func (txn *Txn) Set(key string, entry Entry) error {
	if txn.readOnly {
		return xerror.New("cannot write in read-only transaction")
	}

	txn.mutex.Lock()
	defer txn.mutex.Unlock()

	// Check for write conflicts
	txn.db.mutex.RLock()
	if existing, ok := txn.db.data[key]; ok {
		if existing.Version > txn.readTs {
			txn.db.mutex.RUnlock()
			return xerror.New("write conflict")
		}
	}
	txn.db.mutex.RUnlock()

	entryCopy := entry
	txn.writes[key] = &entryCopy
	return nil
}

func (txn *Txn) Commit() error {
	if txn.readOnly {
		return nil
	}

	txn.mutex.Lock()
	defer txn.mutex.Unlock()

	// Global commit lock to ensure serial commits
	txn.db.commitLock.Lock()
	defer txn.db.commitLock.Unlock()

	// Re-check for conflicts
	txn.db.mutex.RLock()
	for key := range txn.writes {
		if entry, ok := txn.db.data[key]; ok {
			if entry.Version > txn.readTs {
				txn.db.mutex.RUnlock()
				return xerror.New("write conflict")
			}
		}
	}
	txn.db.mutex.RUnlock()

	// Get new transaction ID
	newTxID := atomic.AddUint64(&txn.db.txCounter, 1)

	// Apply changes
	txn.db.mutex.Lock()
	for key, entry := range txn.writes {
		entry.Version = newTxID
		txn.db.data[key] = *entry
	}
	txn.db.mutex.Unlock()

	// Write WAL entry
	walEntry := WALEntry{
		TxID:     newTxID,
		Commands: make([]Command, 0, len(txn.writes)),
	}

	for key, entry := range txn.writes {
		walEntry.Commands = append(walEntry.Commands, Command{
			Key:     key,
			Value:   entry.Value,
			Version: entry.Version,
			Type:    entry.Type,
		})
	}

	if err := txn.db.writeWAL(walEntry); err != nil {
		return fmt.Errorf("failed to write WAL: %w", err)
	}

	return txn.db.writeData()
}

// Add iterator methods
func (db *DB) NewIterator(opts IteratorOptions) *Iterator {
	return &Iterator{
		db:      db,
		prefix:  []byte(opts.Prefix),
		reverse: opts.Reverse,
	}
}

type IteratorOptions struct {
	Prefix  string
	Reverse bool
}

func (it *Iterator) Seek(key string) {
	it.mutex.Lock()
	defer it.mutex.Unlock()

	it.db.mutex.RLock()
	defer it.db.mutex.RUnlock()

	// Collect all matching keys
	var matchingKeys []string
	for k := range it.db.data {
		if bytes.HasPrefix([]byte(k), it.prefix) {
			matchingKeys = append(matchingKeys, k)
		}
	}

	// Sort keys based on iteration direction
	if it.reverse {
		sort.Sort(sort.Reverse(sort.StringSlice(matchingKeys)))
	} else {
		sort.Strings(matchingKeys)
	}

	// Find starting position
	it.curr = ""
	for _, k := range matchingKeys {
		if it.reverse {
			if k <= key {
				it.curr = k
				break
			}
		} else {
			if k >= key {
				it.curr = k
				break
			}
		}
	}

	// If no matching key found and we have keys, use first/last key
	if it.curr == "" && len(matchingKeys) > 0 {
		if it.reverse {
			it.curr = matchingKeys[0]
		} else {
			it.curr = matchingKeys[0]
		}
	}
}

func (it *Iterator) Valid() bool {
	it.mutex.RLock()
	defer it.mutex.RUnlock()

	if it.curr == "" {
		return false
	}

	return bytes.HasPrefix([]byte(it.curr), it.prefix)
}

func (it *Iterator) Next() {
	it.mutex.Lock()
	defer it.mutex.Unlock()

	it.db.mutex.RLock()
	defer it.db.mutex.RUnlock()

	// Get all keys
	var keys []string
	for k := range it.db.data {
		if bytes.HasPrefix([]byte(k), it.prefix) {
			keys = append(keys, k)
		}
	}

	// Sort keys based on direction
	if it.reverse {
		sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	} else {
		sort.Strings(keys)
	}

	// Find next key
	if it.curr == "" {
		if len(keys) > 0 {
			it.curr = keys[0]
		}
		return
	}

	for i, k := range keys {
		if k == it.curr && i+1 < len(keys) {
			it.curr = keys[i+1]
			return
		}
	}
	it.curr = ""
}

func (it *Iterator) Item() *Entry {
	it.mutex.RLock()
	defer it.mutex.RUnlock()

	it.db.mutex.RLock()
	defer it.db.mutex.RUnlock()

	// Return nil if current key is empty
	if it.curr == "" {
		return nil
	}

	// Get entry from DB
	if entry, ok := it.db.data[it.curr]; ok {
		// Return a copy of the entry to prevent modification
		entryCopy := entry
		return &entryCopy
	}
	return nil
}

// ExportToJSON exports the database content as a JSON string
func (db *DB) ExportToJSON() (string, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	exportData := make(map[string]interface{})
	for key, entry := range db.data {
		var value interface{}

		// Format value based on type
		switch entry.Type {
		case String:
			if !db.options.EnableVersioning {
				value = entry.Value.(string)
			} else {
				value = formatVersionedEntry(entry)
			}

		case List:
			if !db.options.EnableVersioning {
				value = entry.Value.([]string)
			} else {
				listValue := map[string]interface{}{
					"type":         entry.Type,
					"value":        entry.Value.([]string),
					"version":      entry.Version,
					"created":      entry.Created,
					"last_updated": entry.LastUpdated,
				}
				if len(entry.Versions) > 0 {
					versions := make([]map[string]interface{}, len(entry.Versions))
					for i, v := range entry.Versions {
						versions[i] = map[string]interface{}{
							"value":        v.Value.([]string),
							"version":      v.Version,
							"created":      v.Created,
							"last_updated": v.LastUpdated,
						}
					}
					listValue["versions"] = versions
				}
				value = listValue
			}

		case Hash:
			if !db.options.EnableVersioning {
				value = entry.Value.(map[string]string)
			} else {
				hashValue := map[string]interface{}{
					"type":         entry.Type,
					"value":        entry.Value.(map[string]string),
					"version":      entry.Version,
					"created":      entry.Created,
					"last_updated": entry.LastUpdated,
				}
				if len(entry.Versions) > 0 {
					versions := make([]map[string]interface{}, len(entry.Versions))
					for i, v := range entry.Versions {
						versions[i] = map[string]interface{}{
							"value":        v.Value.(map[string]string),
							"version":      v.Version,
							"created":      v.Created,
							"last_updated": v.LastUpdated,
						}
					}
					hashValue["versions"] = versions
				}
				value = hashValue
			}

		case Set:
			set := entry.Value.(map[string]struct{})
			members := make([]string, 0, len(set))
			for member := range set {
				members = append(members, member)
			}
			sort.Strings(members)

			if !db.options.EnableVersioning {
				value = members
			} else {
				setValue := map[string]interface{}{
					"type":         entry.Type,
					"value":        members,
					"version":      entry.Version,
					"created":      entry.Created,
					"last_updated": entry.LastUpdated,
				}
				if len(entry.Versions) > 0 {
					versions := make([]map[string]interface{}, len(entry.Versions))
					for i, v := range entry.Versions {
						vset := v.Value.(map[string]struct{})
						vmembers := make([]string, 0, len(vset))
						for member := range vset {
							vmembers = append(vmembers, member)
						}
						sort.Strings(vmembers)
						versions[i] = map[string]interface{}{
							"value":        vmembers,
							"version":      v.Version,
							"created":      v.Created,
							"last_updated": v.LastUpdated,
						}
					}
					setValue["versions"] = versions
				}
				value = setValue
			}

		case ZSet:
			if !db.options.EnableVersioning {
				value = entry.Value.([]ZSetMember)
			} else {
				zsetValue := map[string]interface{}{
					"type":         entry.Type,
					"value":        entry.Value.([]ZSetMember),
					"version":      entry.Version,
					"created":      entry.Created,
					"last_updated": entry.LastUpdated,
				}
				if len(entry.Versions) > 0 {
					versions := make([]map[string]interface{}, len(entry.Versions))
					for i, v := range entry.Versions {
						versions[i] = map[string]interface{}{
							"value":        v.Value.([]ZSetMember),
							"version":      v.Version,
							"created":      v.Created,
							"last_updated": v.LastUpdated,
						}
					}
					zsetValue["versions"] = versions
				}
				value = zsetValue
			}
		}

		exportData[key] = value
	}

	data, err := json.MarshalIndent(exportData, "", "    ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}

	return string(data), nil
}

// Helper function to format versioned string entries
func formatVersionedEntry(entry Entry) map[string]interface{} {
	value := map[string]interface{}{
		"type":         entry.Type,
		"value":        entry.Value,
		"version":      entry.Version,
		"created":      entry.Created,
		"last_updated": entry.LastUpdated,
	}

	if len(entry.Versions) > 0 {
		versions := make([]map[string]interface{}, len(entry.Versions))
		for i, v := range entry.Versions {
			versions[i] = map[string]interface{}{
				"value":        v.Value,
				"version":      v.Version,
				"created":      v.Created,
				"last_updated": v.LastUpdated,
			}
		}
		value["versions"] = versions
	}

	return value
}

// VersionedEntry represents a versioned value
type VersionedEntry struct {
	Value       interface{}
	Version     uint64
	Created     time.Time
	LastUpdated time.Time
}

// WithVersioning enables version control
func WithVersioning(enable bool) Option {
	return func(o *Options) {
		o.EnableVersioning = enable
	}
}

// WithMaxVersions sets maximum versions per key
func WithMaxVersions(max int) Option {
	return func(o *Options) {
		o.MaxVersions = max
	}
}

// StringOp operations with time tracking
func (op *StringOp) SetWithVersion(value string) error {
	op.db.mutex.Lock()
	defer op.db.mutex.Unlock()

	now := time.Now()
	newVersion := op.db.txCounter + 1

	var versions []VersionedEntry
	if existing, ok := op.db.data[op.key]; ok && op.db.options.EnableVersioning {
		// Add current value to version history
		versions = append(existing.Versions, VersionedEntry{
			Value:       existing.Value,
			Version:     existing.Version,
			Created:     existing.Created,
			LastUpdated: existing.LastUpdated,
		})

		// Enforce version limit
		if op.db.options.MaxVersions > 0 && len(versions) > op.db.options.MaxVersions-1 {
			versions = versions[len(versions)-(op.db.options.MaxVersions-1):]
		}
	}

	entry := Entry{
		Type:        String,
		Value:       value,
		Version:     newVersion,
		Created:     now,
		LastUpdated: now,
		Versions:    versions,
	}

	op.db.data[op.key] = entry
	op.db.txCounter++
	return op.db.writeData()
}

// GetVersion retrieves a specific version of a string value
func (op *StringOp) GetVersion(version uint64) (string, bool) {
	op.db.mutex.RLock()
	defer op.db.mutex.RUnlock()

	if entry, ok := op.db.data[op.key]; ok && entry.Type == String {
		// Return current version if it matches
		if entry.Version == version {
			return entry.Value.(string), true
		}

		// Search in version history
		for _, v := range entry.Versions {
			if v.Version == version {
				return v.Value.(string), true
			}
		}
	}
	return "", false
}

// ListVersions returns all available versions for a key
func (op *StringOp) ListVersions() []uint64 {
	op.db.mutex.RLock()
	defer op.db.mutex.RUnlock()

	if entry, ok := op.db.data[op.key]; ok && entry.Type == String {
		versions := make([]uint64, 0, len(entry.Versions)+1)
		versions = append(versions, entry.Version)

		for _, v := range entry.Versions {
			versions = append(versions, v.Version)
		}

		// Sort versions in descending order
		sort.Slice(versions, func(i, j int) bool {
			return versions[i] > versions[j]
		})

		return versions
	}
	return nil
}

// Transaction implementation
type Transaction struct {
	db       *DB
	readOnly bool
	writes   map[string]Entry
	reads    map[string]Entry
	mutex    sync.RWMutex
	started  time.Time
}

func (txn *Transaction) Get(key string) (Entry, error) {
	txn.mutex.RLock()
	defer txn.mutex.RUnlock()

	// First check writes
	if entry, ok := txn.writes[key]; ok {
		return entry, nil
	}

	// Then check reads
	if entry, ok := txn.reads[key]; ok {
		return entry, nil
	}

	// Finally check database
	txn.db.mutex.RLock()
	defer txn.db.mutex.RUnlock()

	if entry, ok := txn.db.data[key]; ok {
		// Store in reads for consistency
		txn.reads[key] = entry
		return entry, nil
	}

	return Entry{}, ErrKeyNotFound
}

func (txn *Transaction) Set(key string, entry Entry) error {
	if txn.readOnly {
		return fmt.Errorf("cannot write in read-only transaction")
	}

	txn.mutex.Lock()
	defer txn.mutex.Unlock()

	// Store in writes
	txn.writes[key] = entry
	return nil
}

func (txn *Transaction) Commit() error {
	if txn.readOnly {
		return nil
	}

	txn.db.txMutex.Lock()
	defer txn.db.txMutex.Unlock()

	// Check for conflicts
	for key := range txn.writes {
		if currentEntry, ok := txn.db.data[key]; ok {
			if readEntry, wasRead := txn.reads[key]; wasRead {
				// Check if the entry was modified after our read
				if currentEntry.Version != readEntry.Version {
					return fmt.Errorf("transaction conflict: key %s has been modified", key)
				}
			}
		}
	}

	// Apply all writes
	now := time.Now()
	for key, entry := range txn.writes {
		entry.Version = txn.db.txCounter + 1
		entry.LastUpdated = now
		if entry.Created.IsZero() {
			entry.Created = now
		}
		txn.db.data[key] = entry
	}

	txn.db.txCounter++
	return nil
}

// Concurrent write control
func (db *DB) writeWithLock(key string, entry Entry) error {
	db.txMutex.Lock()
	defer db.txMutex.Unlock()

	// Update version and timestamps
	entry.Version = db.txCounter + 1
	now := time.Now()
	if entry.Created.IsZero() {
		entry.Created = now
	}
	entry.LastUpdated = now

	db.data[key] = entry
	db.txCounter++
	return nil
}
