package repositories

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination=../mocks/repositories_mock.go -package=mocks github.com/paulrozhkin/otus-microservices-homework/otus-microservices-homework-07/services/user-service/internal/repositories UserRepository,HealthChecker
