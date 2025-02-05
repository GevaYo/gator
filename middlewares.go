package main

import (
	"context"
	"gator/internal/database"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		loggedInUser, err := s.db.GetUser(context.Background(), s.cfg.Current_User_Name)
		if err != nil {
			return err
		}
		return handler(s, cmd, loggedInUser)
	}
}
