package repository

import (
	"chat_socket_server/config"
	"chat_socket_server/repository/kafka"
	"chat_socket_server/types/schema"
	"database/sql"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type Repository struct {
	cfg *config.Config

	db *sql.DB

	Kafka *kafka.Kafka
}

const (
	room       = "chatting.room"
	chat       = "chatting.chat"
	serverInfo = "chatting.serverInfo"
)

func NewRepository(cfg *config.Config) (*Repository, error) {
	r := &Repository{cfg: cfg}
	var err error

	if r.db, err = sql.Open(cfg.DB.Database, cfg.DB.URL); err != nil {
		return nil, err
	} else if r.Kafka, err = kafka.NewKafka(r.cfg); err != nil {
		return nil, err
	} else {
		return r, err
	}
}

// 쿼리 작성
func (s *Repository) ServerSet(ip string, available bool) error { // ip에 대해 기존의 데이터가 존재한다면 update만 해라
	_, err := s.db.Exec("INSERT INTO serverInfo(`ip`, `available`) VALUES(?, ?) ON DUPLICATE KEY UPDATE `available` = VALUES(`available`);", ip, available)
	return err
}

func (s *Repository) InsertChatting(user, message, roomname string) error {
	log.Println("Insert chatting", "from", user, "message", message, "room", roomname)

	_, err := s.db.Exec("INSERT INTO chatting.chat(room,name,message) VALUES(?,?,?)", roomname, user, message)

	return err
}

func (s *Repository) GetChatList(roomName string) ([]*schema.Chat, error) { // 이름으로 분류한 채팅방 가져오기
	qs := query([]string{"SELECT * FROM", chat, "WHERE room = ? ORDER BY `when` DESC LIMIT 10"})

	if cursor, err := s.db.Query(qs, roomName); err != nil { // cursor는 메모리를 할당함
		return nil, err
	} else {
		defer cursor.Close() // 메모리 반환

		var results []*schema.Chat

		for cursor.Next() {
			d := new(schema.Chat)
			if err = cursor.Scan(&d.ID, &d.Room, &d.Name, &d.Message, &d.When); err != nil {
				return nil, err
			} else {
				results = append(results, d)
			}
		}

		if len(results) == 0 {
			return []*schema.Chat{}, nil
		} else {
			return results, nil
		}
	}
}

func (s *Repository) RoomList() ([]*schema.Room, error) { // 채팅방을 리스트로 가져옴
	qs := query([]string{"SELECT * FROM", room})

	if cursor, err := s.db.Query(qs); err != nil { // cursor는 메모리를 할당함
		return nil, err
	} else {
		defer cursor.Close() // 메모리 반환

		var results []*schema.Room

		for cursor.Next() {
			d := new(schema.Room)
			if err = cursor.Scan(&d.ID, &d.Name, &d.CreateAt, &d.UpdatedAt); err != nil {
				return nil, err
			} else {
				results = append(results, d)
			}
		}

		if len(results) == 0 {
			return []*schema.Room{}, nil
		} else {
			return results, nil
		}
	}
}

func (s *Repository) MakeRoom(name string) error { // 방 생성
	_, err := s.db.Exec("INSERT INTO chatting.room(name) VALUES (?)", name)
	return err
}

func (s *Repository) Room(name string) (*schema.Room, error) {
	d := new(schema.Room)
	qs := query([]string{"SELECT * FROM", room, "WHERE name = ?"})
	// SELECT * FROM chatting.room WHERE name = ? 이걸 한것

	err := s.db.QueryRow(qs, name).Scan(
		&d.ID,
		&d.Name,
		&d.CreateAt,
		&d.UpdatedAt,
	)

	if err = noResult(err); err != nil {
		return nil, err
	} else {
		return nil, nil
	}
}

func query(qs []string) string {
	return strings.Join(qs, " ") + ";"
}

func noResult(err error) error {
	if strings.Contains(err.Error(), "sql: no rows in result set") {
		return nil
	} else {
		return err
	}
}
