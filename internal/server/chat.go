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

package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/googleapis/genai-toolbox/internal/chats"
	"github.com/googleapis/genai-toolbox/internal/server/mcp/jsonrpc"
)

func (s *Server) recordChatMessage(ctx context.Context, sessionID, toolsetName string, payload []byte) {
	if s == nil || s.chatStore == nil {
		return
	}

	direction := chats.DirectionRequest
	var base jsonrpc.BaseMessage
	if err := json.Unmarshal(payload, &base); err == nil && base.Id == nil {
		direction = chats.DirectionNotification
	}

	if err := s.chatStore.RecordMessage(ctx, sessionID, toolsetName, direction, payload); err != nil {
		s.logger.DebugContext(ctx, fmt.Sprintf("unable to record chat message: %v", err))
	}
}

func (s *Server) recordChatResponse(ctx context.Context, sessionID, toolsetName string, payload []byte) {
	if s == nil || s.chatStore == nil {
		return
	}
	if err := s.chatStore.RecordMessage(ctx, sessionID, toolsetName, chats.DirectionResponse, payload); err != nil {
		s.logger.DebugContext(ctx, fmt.Sprintf("unable to record chat response: %v", err))
	}
}
