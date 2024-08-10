package service

import (
	"chat_controller_server/repository"
	"chat_controller_server/types/table"
	"fmt"
	. "github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
)

type Service struct {
	repository *repository.Repository

	AvgServerList map[string]bool // ip주소와 사용가능여부
}

func NewService(repository *repository.Repository) *Service {
	s := &Service{repository: repository, AvgServerList: make(map[string]bool)}

	s.setServerInfo()

	if err := s.repository.Kafka.RegisterSubTopic("chat"); err != nil {
		panic(err)
	} else {
		go s.loopSubKafka() // 백그라운드 실행
	}

	return s
}

func (s *Service) loopSubKafka() {
	for {
		ev := s.repository.Kafka.Pool(100)
		switch event := ev.(type) {
		case *Message:
			fmt.Println(event)
		case *Error:
			log.Println("Failed to pooling event", event.Error())
		}
	}
}

func (s *Service) GetAvgServerList() []string {
	var res []string

	for ip, available := range s.AvgServerList {
		if available {
			res = append(res, ip)
		}
	}

	return res
}

func (s *Service) setServerInfo() {
	if serverList, err := s.GetAvailableServerList(); err != nil {
		panic(err)
	} else {

		for _, server := range serverList {
			s.AvgServerList[server.IP] = true
		}
	}
}

func (s *Service) GetAvailableServerList() ([]*table.ServerInfo, error) {
	return s.repository.GetAvailableServerList()
}
