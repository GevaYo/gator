package main

import (
	"context"
	"fmt"
	"gator/internal/database"
	"log"
	"time"

	"github.com/google/uuid"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}

	ctx := context.Background()
	username := cmd.Args[0]
	_, err := s.db.GetUser(ctx, username)
	if err != nil {
		fmt.Println(err)
		log.Fatalf("Username doesn't exist in the database: %s", username)
	}

	err = s.cfg.SetUser(username)
	if err != nil {
		return fmt.Errorf("Couldn't set the current user: %w", err)
	}
	fmt.Println("User switched successfully!")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}

	username := cmd.Args[0]

	ctx := context.Background()
	user, err := s.db.CreateUser(ctx, database.CreateUserParams{uuid.New(), time.Now(), time.Now(), username})
	if err != nil {
		log.Fatalf("User with name '%s' already exists!", username)
	}

	err = s.cfg.SetUser(username)
	if err != nil {
		return fmt.Errorf("Couldn't set the current user: %w", err)
	}
	printUser(user)
	return nil
}

func handlerReset(s *state, cmd command) error {
	if len(cmd.Args) > 0 {
		return fmt.Errorf("usage: %s", cmd.Name)
	}

	ctx := context.Background()
	err := s.db.DeleteUsers(ctx)
	if err != nil {
		log.Fatalf("Failed to delete users!")
	}

	return nil
}

func handlerUsers(s *state, cmd command) error {
	if len(cmd.Args) > 0 {
		return fmt.Errorf("usage: %s", cmd.Name)
	}

	ctx := context.Background()
	users, err := s.db.GetUsers(ctx)
	if err != nil {
		log.Fatalf("Failed to delete users!")
	}

	for _, user := range users {
		fmt.Printf("* %s ", user.Name)
		if user.Name == s.cfg.Current_User_Name {
			fmt.Print("(current)")
		}
		fmt.Println()
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	ctx := context.Background()
	feed, err := fetchFeed(ctx, "https://www.wagslane.dev/index.xml")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v", feed)
	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 2 {
		log.Fatal("Need 2 args")
	}
	ctx := context.Background()
	feed, err := s.db.CreateFeed(ctx, database.CreateFeedParams{uuid.New(), time.Now(), time.Now(), cmd.Args[0], cmd.Args[1], user.ID})
	if err != nil {
		log.Fatal("Couldn't create new feed: %w", err)
	}
	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{uuid.New(), time.Now(), time.Now(), user.ID, feed.ID})
	if err != nil {
		return fmt.Errorf("Couldn't create a new feedFollow record: %w\n", err)
	}

	printFeed(feed, user)

	return nil
}

func handlerFeeds(s *state, cmd command) error {
	ctx := context.Background()
	feeds, err := s.db.GetFeeds(ctx)
	if err != nil {
		log.Fatal("Couldn't fetch feeds: %w", err)
	}

	if len(feeds) == 0 {
		fmt.Println("No feeds found.")
		return nil
	}

	for _, feed := range feeds {
		user, err := s.db.GetUserById(ctx, feed.UserID)
		if err != nil {
			return fmt.Errorf("Couldn't get user: %w", err)
		}
		printFeed(feed, user)
		fmt.Println("=====================================")
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("Only 1 arguments [URL] is allowed!")
	}
	url := cmd.Args[0]
	feed, err := s.db.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return fmt.Errorf("No feed found for URL: %s", url)
	}
	feedFollow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{uuid.New(), time.Now(), time.Now(), user.ID, feed.ID})
	if err != nil {
		return fmt.Errorf("Couldn't create a new feedFollow record: %w\n", err)
	}

	fmt.Printf("Feed Name: %s\n", feedFollow.FeedName)
	fmt.Printf("User Name: %s\n", feedFollow.UserName)
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("Only 1 arguments [URL] is allowed!")
	}
	url := cmd.Args[0]
	feed, err := s.db.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return fmt.Errorf("No feed found for URL: %s", url)
	}
	err = s.db.DeleteFeedFollowRecord(context.Background(), database.DeleteFeedFollowRecordParams{user.ID, feed.ID})
	if err != nil {
		return fmt.Errorf("Couldn't create a new feedFollow record: %w\n", err)
	}
	fmt.Printf("User %s unfollowed \"%s\"\n", user.Name, url)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("No arguments allowed!")
	}
	userFeeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("Couldn't fetch the followed feeds by the user: %w\n", err)
	}

	if len(userFeeds) == 0 {
		fmt.Println("The user isn't following any feed.")
		return nil
	}

	for _, feed := range userFeeds {
		fmt.Printf("* Feed Name: %s\n", feed.FeedName)
	}
	return nil
}
func printFeed(feed database.Feed, user database.User) {
	fmt.Printf("* ID:            %s\n", feed.ID)
	fmt.Printf("* Created:       %v\n", feed.CreatedAt)
	fmt.Printf("* Updated:       %v\n", feed.UpdatedAt)
	fmt.Printf("* Name:          %s\n", feed.Name)
	fmt.Printf("* URL:           %s\n", feed.Url)
	fmt.Printf("* User:          %s\n", user.Name)
}

func printUser(user database.User) {
	fmt.Printf("* ID:            %s\n", user.ID)
	fmt.Printf("* Created:       %v\n", user.CreatedAt)
	fmt.Printf("* Updated:       %v\n", user.UpdatedAt)
	fmt.Printf("* Name:          %s\n", user.Name)
}
