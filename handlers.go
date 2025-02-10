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
		return fmt.Errorf("couldn't set the current user: %w", err)
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
	user, err := s.db.CreateUser(ctx, database.CreateUserParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: username})
	if err != nil {
		log.Fatalf("User with name '%s' already exists!", username)
	}

	err = s.cfg.SetUser(username)
	if err != nil {
		return fmt.Errorf("couldn't set the current user: %w", err)
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
	if len(cmd.Args) != 1 {
		return fmt.Errorf("only 1 argument is allowed")
	}
	timeBetweenReqs, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("failed to parse the time string argument: %w", err)
	}
	fmt.Printf("Collecting feeds every %v\n", timeBetweenReqs)
	ticker := time.NewTicker(timeBetweenReqs)
	for ; ; <-ticker.C {
		err := scrapeFeeds(s.db)
		fmt.Print("\n\n\n\n")
		if err != nil {
			return err
		}
	}
}

func scrapeFeeds(db *database.Queries) error {
	feed, err := db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("error fetching next feed: %w", err)
	}
	feedContent, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		return fmt.Errorf("failed fetching feed contents: %w", err)
	}
	err = db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		return fmt.Errorf("marking as fetched failed: %w", err)
	}
	for _, p := range feedContent.Channel.Item {
		fmt.Printf("Title: %s\n", p.Title)
	}
	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 2 {
		log.Fatal("Need 2 args")
	}
	ctx := context.Background()
	feed, err := s.db.CreateFeed(ctx, database.CreateFeedParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: cmd.Args[0], Url: cmd.Args[1], UserID: user.ID})
	if err != nil {
		log.Fatal("couldn't create new feed: %w", err)
	}
	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), UserID: user.ID, FeedID: feed.ID})
	if err != nil {
		return fmt.Errorf("couldn't create a new feedFollow record: %w", err)
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
			return fmt.Errorf("couldn't get user: %w", err)
		}
		printFeed(feed, user)
		fmt.Println("=====================================")
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("only 1 arguments [URL] is allowed")
	}
	url := cmd.Args[0]
	feed, err := s.db.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return fmt.Errorf("no feed found for URL: %s", url)
	}
	feedFollow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), UserID: user.ID, FeedID: feed.ID})
	if err != nil {
		return fmt.Errorf("couldn't create a new feedFollow record: %w", err)
	}

	fmt.Printf("Feed Name: %s\n", feedFollow.FeedName)
	fmt.Printf("User Name: %s\n", feedFollow.UserName)
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("only 1 arguments [URL] is allowed")
	}
	url := cmd.Args[0]
	feed, err := s.db.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return fmt.Errorf("no feed found for URL: %s", url)
	}
	err = s.db.DeleteFeedFollowRecord(context.Background(), database.DeleteFeedFollowRecordParams{UserID: user.ID, FeedID: feed.ID})
	if err != nil {
		return fmt.Errorf("couldn't create a new feedFollow record: %w", err)
	}
	fmt.Printf("User %s unfollowed \"%s\"\n", user.Name, url)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("no arguments allowed")
	}
	userFeeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("couldn't fetch the followed feeds by the user: %w", err)
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
