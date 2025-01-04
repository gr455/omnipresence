package storage

import (
	"encoding/json"
	"fmt"
	pb "github.com/gr455/omnipresence/raft/service/genproto"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

type RaftStorage struct {
	LogfilePath             string
	PersistentMetaPath      string
	LogFileWriteLock        sync.Mutex
	PersistentMetaWriteLock sync.Mutex
}

type PersistentStorage struct {
	CurrentTerm int64 `json:"current_term"`
	// String peerId who was voted for last term by this peer
	CurrentTermVotedFor string `json:"current_term_voted_for"`
}

type LogStorage struct {
	Log []*pb.LogEntry `json:"log"`
	// Log is snapshotted till log_start_idx - 1
	LogStartIdx int64 `json:"log_start_idx"`
}

func NewRaftStorage(logfilePath, persistentMetaPath string) *RaftStorage {
	// Check that persistent storage files exist else create
	_, err := os.Stat(logfilePath)
	if os.IsNotExist(err) {
		fmt.Printf("Log file does not exist, creating new...")
		CreateDefaultLogFile(logfilePath)
	}
	_, err = os.Stat(persistentMetaPath)
	if os.IsNotExist(err) {
		fmt.Printf("Log file does not exist, creating new...")
		CreateDefaultPersistentMeta(persistentMetaPath)
	}
	return &RaftStorage{
		LogfilePath:        logfilePath,
		PersistentMetaPath: persistentMetaPath,
	}
}

func CreateDefaultLogFile(logfilePath string) {
	file, err := os.Create(logfilePath)
	if err != nil {
		log.Fatalf("Err: Could not create log file")
		return
	}

	defer file.Close()

	_, err = file.WriteString("{\"log\": [], \"log_start_idx\": 0}")
	if err != nil {
		log.Fatalf("Error writing to file:", err)
		return
	}
}

func CreateDefaultPersistentMeta(persistentMetaPath string) {
	file, err := os.Create(persistentMetaPath)
	if err != nil {
		log.Fatalf("Err: Could not create persistent metadata file")
		return
	}

	defer file.Close()

	_, err = file.WriteString("{\"current_term\": 0, \"current_term_voted_for\": \"\"}")
	if err != nil {
		log.Fatalf("Error writing to file:", err)
		return
	}
}

func (storage *RaftStorage) ReadLog() ([]*pb.LogEntry, int64, error) {
	var log LogStorage

	logdata, err := ioutil.ReadFile(storage.LogfilePath)
	if err != nil {
		fmt.Println("Err: Error reading logfile %v\n", err)
		return nil, 0, err
	}

	err = json.Unmarshal(logdata, &log)
	if err != nil {
		fmt.Println("Err: Error unmarshaling log %v\n", err)
		return nil, 0, err
	}

	return log.Log, log.LogStartIdx, nil
}

func (storage *RaftStorage) ReadPersistent() (*PersistentStorage, error) {
	var persistentMeta PersistentStorage

	meta, err := ioutil.ReadFile(storage.PersistentMetaPath)
	if err != nil {
		fmt.Println("Err: Error reading persistent %v\n", err)
		return nil, err
	}

	err = json.Unmarshal(meta, &persistentMeta)
	if err != nil {
		fmt.Println("Err: Error unmarshaling persistent %v\n", err)
		return nil, err
	}

	return &persistentMeta, nil
}

func (storage *RaftStorage) WriteToLog(msg string, term, msgIdx int64) error {
	storage.LogFileWriteLock.Lock()
	defer storage.LogFileWriteLock.Unlock()

	log, startIdx, err := storage.ReadLog()
	if err != nil {
		fmt.Printf("Err: Error reading logfile while writing %v\n", err)
		return err
	}

	if msgIdx < startIdx+int64(len(log)) {
		log[msgIdx] = &pb.LogEntry{Entry: msg, Term: term}
	} else if msgIdx == startIdx+int64(len(log)) {
		log = append(log, &pb.LogEntry{Entry: msg, Term: term})
	}

	logJson, err := json.Marshal(log)
	if err != nil {
		fmt.Printf("Err: Error marshaling log: %v\n", err)
		return err
	}

	err = ioutil.WriteFile(storage.LogfilePath, logJson, 0644)
	if err != nil {
		fmt.Printf("Err: Error writing log: %v\n", err)
		return err
	}

	return nil
}

func (storage *RaftStorage) WritePersistent(term int64, votedFor string) error {
	storage.PersistentMetaWriteLock.Lock()
	defer storage.PersistentMetaWriteLock.Unlock()

	persistent, err := storage.ReadPersistent()
	if err != nil {
		fmt.Printf("Err: Error reading persistent while writing %v\n", err)
		return err
	}

	persistent.CurrentTerm = term
	persistent.CurrentTermVotedFor = votedFor

	persistentJson, err := json.Marshal(persistent)
	if err != nil {
		fmt.Printf("Err: Error marshaling persistent: %v\n", err)
		return err
	}

	err = ioutil.WriteFile(storage.PersistentMetaPath, persistentJson, 0644)
	if err != nil {
		fmt.Printf("Err: Error writing persistent: %v\n", err)
		return err
	}

	return nil
}
