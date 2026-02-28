package infrastructure

import (
	"log"
)

type ConsoleEmailService struct{}

func NewConsoleEmailService() *ConsoleEmailService {
	return &ConsoleEmailService{}
}

func (s *ConsoleEmailService) SendEmail(to, subject, body string) error {
	log.Printf("================ EMAIL NOTIFICATION ================")
	log.Printf("To: %s", to)
	log.Printf("Subject: %s", subject)
	log.Printf("Body: %s", body)
	log.Printf("====================================================")
	return nil
}
