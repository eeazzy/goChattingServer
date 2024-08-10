package service

import (
	"chat_socket_server/repository"
	"chat_socket_server/types/schema"
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
)

type Service struct {
	repository *repository.Repository
}

func NewService(repository *repository.Repository) *Service {
	s := &Service{repository: repository}

	return s
}

func (s *Service) PublishServerStatusEvent(ip string, status bool) {
	// kafka에 이벤트 전송
	type ServerInfoEvent struct {
		IP     string
		Status bool
	}

	e := &ServerInfoEvent{
		IP:     ip,
		Status: status,
	}
	ch := make(chan kafka.Event)

	if v, err := json.Marshal(e); err != nil {
		log.Println("Failed to marshal server info")
	} else if result, err := s.PublishEvent("chat", v, ch); err != nil {
		// 카프카에 보내기
		log.Println("Failed to publish event to kafka", "err", err)
	} else {
		log.Println("Published event to kafka", result)
	}

}

func (s *Service) PublishEvent(topic string, value []byte, ch chan kafka.Event) (kafka.Event, error) {
	return s.repository.Kafka.PublishEvent(topic, value, ch)
}

// repository 에 있는 쿼리 활용
func (s *Service) ServerSet(ip string, available bool) error {
	if err := s.repository.ServerSet(ip, available); err != nil {
		log.Println("Failed to set server", "ip", ip, "available", available)
		return err
	} else {
		return nil
	}
}

func (s *Service) InsertChatting(user, message, roomname string) {
	if err := s.repository.InsertChatting(user, message, roomname); err != nil {
		log.Println("Failed to insert chat:", "err", err)
	}
}

func (s *Service) EnterRoom(roomName string) ([]*schema.Chat, error) {
	if res, err := s.repository.GetChatList(roomName); err != nil {
		log.Println("Failed to get chat list", "err", err.Error())
		return nil, err
	} else {
		return res, nil
	}
}

func (s *Service) RoomList() ([]*schema.Room, error) {
	if res, err := s.repository.RoomList(); err != nil {
		log.Println("Failed to get room list", "err", err.Error())
		return nil, err
	} else {
		return res, nil
	}
}

func (s *Service) MakeRoom(name string) error {
	if err := s.repository.MakeRoom(name); err != nil {
		log.Println("Failed to make room", "err", err.Error())
		return err
	} else {
		return nil
	}
}

func (s *Service) Room(name string) (*schema.Room, error) {
	if res, err := s.repository.Room(name); err != nil {
		log.Println("Failed to get room", "err", err.Error())
		return nil, err
	} else {
		return res, nil
	}
}
