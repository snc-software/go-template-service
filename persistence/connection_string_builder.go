package persistence

import "fmt"

type ConnectionStringBuilder struct{}

func (ConnectionStringBuilder) Build() string {
	config := DatabaseConfiguration{}.Get()

	return fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		config.Host,
		config.Port,
		config.Name,
		config.User,
		config.Password,
	)
}
