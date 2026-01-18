// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package chats

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	DirectionRequest      = "request"
	DirectionNotification = "notification"
	DirectionResponse     = "response"
)

type Message struct {
	SessionID string          `json:"sessionId"`
	Toolset   string          `json:"toolset,omitempty"`
	Direction string          `json:"direction"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

type Chat struct {
	SessionID string    `json:"sessionId"`
	Archived  bool      `json:"archived"`
	Messages  []Message `json:"messages"`
}

type Export struct {
	ExportedAt    time.Time `json:"exportedAt"`
	ActiveChats   int       `json:"activeChats"`
	ArchivedChats int       `json:"archivedChats"`
	Chats         []Chat    `json:"chats"`
}

type Store struct {
	rootDir string
	mu      sync.Mutex
	now     func() time.Time
}

func NewStore(rootDir string) (*Store, error) {
	if strings.TrimSpace(rootDir) == "" {
		return nil, nil
	}
	if err := os.MkdirAll(filepath.Join(rootDir, "active"), 0o700); err != nil {
		return nil, fmt.Errorf("unable to create chat storage directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(rootDir, "archived"), 0o700); err != nil {
		return nil, fmt.Errorf("unable to create chat archive directory: %w", err)
	}
	return &Store{rootDir: rootDir, now: time.Now}, nil
}

func (s *Store) RecordMessage(ctx context.Context, sessionID, toolset, direction string, payload []byte) error {
	if s == nil {
		return nil
	}
	if sessionID == "" {
		return fmt.Errorf("session id is required")
	}
	message := Message{
		SessionID: sessionID,
		Toolset:   toolset,
		Direction: direction,
		Timestamp: s.now(),
		Payload:   json.RawMessage(payload),
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("unable to marshal message: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	path := s.sessionPath(sessionID, false)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("unable to open chat session file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	if _, err := writer.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("unable to write chat message: %w", err)
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("unable to flush chat message: %w", err)
	}
	return nil
}

func (s *Store) ArchiveSession(sessionID string) error {
	if s == nil {
		return nil
	}
	if sessionID == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	source := s.sessionPath(sessionID, false)
	dest := s.sessionPath(sessionID, true)
	if _, err := os.Stat(source); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("unable to stat chat session file: %w", err)
	}

	if _, err := os.Stat(dest); err == nil {
		if err := appendFile(dest, source); err != nil {
			return err
		}
		if err := os.Remove(source); err != nil {
			return fmt.Errorf("unable to remove chat session file: %w", err)
		}
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("unable to stat chat archive file: %w", err)
	}

	if err := os.Rename(source, dest); err != nil {
		return fmt.Errorf("unable to archive chat session file: %w", err)
	}
	return nil
}

func (s *Store) ExportAll(ctx context.Context, outputPath string) error {
	if s == nil {
		return fmt.Errorf("chat storage is not configured")
	}
	if strings.TrimSpace(outputPath) == "" {
		return fmt.Errorf("output path is required")
	}

	activeChats, err := s.loadChats(ctx, false)
	if err != nil {
		return err
	}
	archivedChats, err := s.loadChats(ctx, true)
	if err != nil {
		return err
	}

	export := Export{
		ExportedAt:    s.now(),
		ActiveChats:   len(activeChats),
		ArchivedChats: len(archivedChats),
		Chats:         append(activeChats, archivedChats...),
	}

	sort.Slice(export.Chats, func(i, j int) bool {
		return export.Chats[i].SessionID < export.Chats[j].SessionID
	})

	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal chat export: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o700); err != nil {
		return fmt.Errorf("unable to create export directory: %w", err)
	}
	if err := os.WriteFile(outputPath, data, 0o600); err != nil {
		return fmt.Errorf("unable to write export file: %w", err)
	}
	return nil
}

func (s *Store) loadChats(ctx context.Context, archived bool) ([]Chat, error) {
	dir := filepath.Join(s.rootDir, "active")
	if archived {
		dir = filepath.Join(s.rootDir, "archived")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to read chat directory: %w", err)
	}

	chats := make([]Chat, 0, len(entries))
	for _, entry := range entries {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}
		sessionID := strings.TrimSuffix(entry.Name(), ".jsonl")
		path := filepath.Join(dir, entry.Name())
		messages, err := readMessages(path)
		if err != nil {
			return nil, err
		}
		chats = append(chats, Chat{SessionID: sessionID, Archived: archived, Messages: messages})
	}
	return chats, nil
}

func (s *Store) sessionPath(sessionID string, archived bool) string {
	cleanID := sanitizeSessionID(sessionID)
	dir := filepath.Join(s.rootDir, "active")
	if archived {
		dir = filepath.Join(s.rootDir, "archived")
	}
	return filepath.Join(dir, cleanID+".jsonl")
}

func sanitizeSessionID(sessionID string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '-' || r == '_' || r == '.':
			return r
		default:
			return '_'
		}
	}, sessionID)
}

func appendFile(dest, source string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("unable to read chat session file: %w", err)
	}
	file, err := os.OpenFile(dest, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("unable to open chat archive file: %w", err)
	}
	defer file.Close()
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("unable to append chat archive file: %w", err)
	}
	return nil
}

func readMessages(path string) ([]Message, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open chat session file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	var messages []Message
	for scanner.Scan() {
		var msg Message
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			return nil, fmt.Errorf("unable to decode chat message: %w", err)
		}
		messages = append(messages, msg)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("unable to scan chat session file: %w", err)
	}
	return messages, nil
}
