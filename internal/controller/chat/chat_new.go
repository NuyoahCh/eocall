package chat

import (
	"github.com/NuyoahCh/eocall/api/chat"
	"github.com/NuyoahCh/eocall/internal/logic/sse"
)

type ControllerV1 struct {
	service *sse.Service
}

func NewV1() chat.IChatV1 {
	return &ControllerV1{
		service: sse.New(),
	}
}
