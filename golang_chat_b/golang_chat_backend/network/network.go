package network

import (
	"chat_socket_server/service"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type Server struct {
	engine *gin.Engine

	service *service.Service

	port string
	ip   string
}

func NewServer(service *service.Service, port string) *Server {
	s := &Server{engine: gin.New(), service: service, port: port}

	s.engine.Use(gin.Logger())
	s.engine.Use(gin.Recovery())
	s.engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		AllowHeaders:     []string{"ORIGIN", "Content-Length", "Content-Type", "Access-Control-Allow-Headers", "Access-Control-Allow-Origin", "Authorization", "X-Requested-With", "expires"},
		ExposeHeaders:    []string{"ORIGIN", "Content-Length", "Content-Type", "Access-Control-Allow-Headers", "Access-Control-Allow-Origin", "Authorization", "X-Requested-With", "expires"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
	}))

	registerServer(s)

	return s
}

func (s *Server) setServerInfo() {
	// IP를 가져오고
	// IP를 기반으로, mySQL serverInfo 테이블 변경함
	if addrs, err := net.InterfaceAddrs(); err != nil {
		panic(err.Error())
	} else { // ip주소를 가져온다
		var ip net.IP
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
					ip = ipnet.IP
					break
				}
			}
		}

		if ip == nil {
			panic("no valid ip address")
		} else {
			if err := s.service.ServerSet(ip.String()+s.port, true); err != nil {
				panic(err)
			} else {
				s.ip = ip.String()
			}
		}
	}
}

func (s *Server) StartServer() error {
	s.setServerInfo() // 서버시작 시 정보 저장

	channel := make(chan os.Signal, 1)     // 이벤트를 받을 수 있는 변수
	signal.Notify(channel, syscall.SIGINT) // 서버가 죽었을 때 메세지 전송함

	go func() {
		<-channel // 여기에 값이 들어오면 뒤의 로직이 실행된다
		// db변경
		if err := s.service.ServerSet(s.ip+s.port, false); err != nil {
			log.Println("Failed to set server info when close", "err", err)
		}

		// kafka에 이벤트 전송
		type ServerInfoEvent struct {
			IP     string
			Status bool
		}

		e := &ServerInfoEvent{
			IP:     s.ip + s.port,
			Status: false,
		}
		ch := make(chan kafka.Event)

		if v, err := json.Marshal(e); err != nil {
			log.Println("Failed to marshal server info")
		} else if result, err := s.service.PublishEvent("chat", v, ch); err != nil {
			// 카프카에 보내기
			log.Println("Failed to publish event to kafka", "err", err)
		} else {
			log.Println("Published event to kafka", result)
		}

		os.Exit(1)
	}()

	log.Printf("Start Tx Server")
	return s.engine.Run(s.port)
}
