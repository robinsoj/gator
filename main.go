package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/robinsoj/gator/internal/config"
	"github.com/robinsoj/gator/internal/database"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name   string
	params []string
}

// var commands
type commands struct {
	return_map map[string]func(*state, command) error
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.return_map[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	cmdFunc, err := c.return_map[cmd.name]
	if !err {
		return fmt.Errorf("command %s not found", cmd.name)
	}
	return cmdFunc(s, cmd)
}

func commandLogin(s *state, cmd command, user database.User) error {
	if len(cmd.params) == 0 {
		return errors.New("missing required parameters for login command")
	}
	user, err := s.db.GetUser(context.Background(), cmd.params[0])
	if err != nil {
		return errors.New("cannot login an unregistered user")
	}
	s.cfg.SetUser(user.Name)
	fmt.Println("User has been set in the configuration file.")
	return nil
}

func commandRegister(s *state, cmd command) error {
	if len(cmd.params) == 0 {
		return errors.New("missing required parameters for register command")
	}
	var usrParams database.CreateUserParams
	usrParams.ID = uuid.New()
	usrParams.CreatedAt = time.Now()
	usrParams.UpdatedAt = time.Now()
	usrParams.Name = cmd.params[0]
	user, err := s.db.CreateUser(context.Background(), usrParams)
	if err != nil {
		return err
	}
	s.cfg.SetUser(user.Name)
	fmt.Println("User has been successfully registered.  Name:", user.Name, "ID:", user.ID, "Created at:", user.CreatedAt, "Updated at:", user.UpdatedAt)
	return nil
}

func commandReset(s *state, _ command) error {
	err := s.db.DeleteUsers(context.Background())
	if err == nil {
		fmt.Println("User database successfully reset")
	}
	return err
}

func commandUsers(s *state, _ command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range users {
		cur := ""
		if user == s.cfg.CurrentUserName {
			cur = "(current)"
		}
		fmt.Println("*", user, cur)
	}
	return nil
}

func commandAgg(s *state, cmd command) error {
	if len(cmd.params) == 0 {
		return errors.New("missing the required parameters for agg command")
	}
	time_between_req := cmd.params[0]
	timeDur, err := time.ParseDuration(time_between_req)
	if err != nil {
		return err
	}
	fmt.Println("Collecting feeds every", timeDur)
	ticker := time.NewTicker(timeDur)
	defer ticker.Stop()
	for ; ; <-ticker.C {
		scrapFeed(s)
	}
}

func commandAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.params) != 2 {
		return errors.New("missing required parameters for add feed command")
	}
	var feedParams database.CreateFeedParams
	feedParams.ID = uuid.New()
	feedParams.CreatedAt = time.Now()
	feedParams.UpdatedAt = time.Now()
	feedParams.Name = cmd.params[0]
	feedParams.Url = cmd.params[1]
	feedParams.UserID = user.ID
	feed, err := s.db.CreateFeed(context.Background(), feedParams)

	if err != nil {
		return err
	}
	fmt.Println("Feed has been successfully added.  Name:", feed.Name, "ID:", feed.ID, "Created at:", feed.CreatedAt, "Updated at:", feed.UpdatedAt,
		"URL:", feed.Url, "User ID:", feed.UserID)
	cmdParams := command{
		name:   "follow",
		params: []string{feed.Url},
	}
	commandFollow(s, cmdParams, user)
	return nil
}

func commandListFeeds(s *state, _ command) error {
	feeds, err := s.db.ListFeeds(context.Background())
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		fmt.Println(feed.Name, feed.Url, feed.Name_2)
	}
	return nil
}

func commandFollow(s *state, cmd command, user database.User) error {
	if len(cmd.params) == 0 {
		return errors.New("missing required parameters for follow command")
	}
	feeds, err := s.db.ListFeedsByURL(context.Background(), cmd.params[0])
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		var ff database.CreateFeedFollowParams
		ff.ID = uuid.New()
		ff.CreatedAt = time.Now()
		ff.UpdatedAt = time.Now()
		ff.UserID = user.ID
		ff.FeedID = feed.ID
		feedfollow, err := s.db.CreateFeedFollow(context.Background(), ff)
		if err != nil {
			return err
		}
		fmt.Println("FeedFollow has been successfully added.  Name:", feedfollow.FeedName, "Username:", feedfollow.UserName)
	}
	return nil
}

func commandFollowing(s *state, _ command) error {
	feedfollows, err := s.db.GetFeedFollowsForUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return err
	}
	for _, follows := range feedfollows {
		fmt.Println("Feed Name:", follows.Feedname)
	}
	return nil
}

func commandUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.params) != 1 {
		return errors.New("incorrect number of parameters for unfollow")
	}
	var deleteParams database.DeleteFeedFollowForUserParams
	deleteParams.Url = cmd.params[0]
	deleteParams.UserID = user.ID
	err := s.db.DeleteFeedFollowForUser(context.Background(), deleteParams)
	if err != nil {
		return err
	}
	return err
}

func commandBrowse(s *state, cmd command, user database.User) error {
	var limit int
	var err error
	if len(cmd.params) == 0 {
		limit = 2
	} else {
		limit, err = strconv.Atoi(cmd.params[0])
		if err != nil {
			return err
		}
	}
	postParm := database.GetPostsForUserParams{
		Limit: int32(limit),
		Name:  user.Name,
	}
	posts, err := s.db.GetPostsForUser(context.Background(), postParm)
	if err != nil {
		return nil
	}
	for _, post := range posts {
		fmt.Println(post.Title, post.Description)
	}
	return nil
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gator")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var rssFeed RSSFeed
	err = xml.Unmarshal(body, &rssFeed)
	if err != nil {
		return nil, err
	}

	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)
	for i, item := range rssFeed.Channel.Item {
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
		rssFeed.Channel.Item[i] = item
	}
	return &rssFeed, nil
}

func loggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return err
		}
		return handler(s, cmd, user)
	}

}

func scrapFeed(s *state) error {
	next_feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}
	err = s.db.MarkFeedFetched(context.Background(), next_feed.ID)
	if err != nil {
		return err
	}
	rssFeed, err := fetchFeed(context.Background(), next_feed.Url)
	if err != nil {
		return err
	}
	for _, feed := range rssFeed.Channel.Item {
		var postParams database.CreatePostParams
		postParams.ID = uuid.New()
		postParams.CreatedAt = time.Now()
		postParams.UpdatedAt = sql.NullTime{}
		postParams.Title = feed.Title
		postParams.Url = next_feed.Url
		postParams.Description = feed.Description
		t, err := time.Parse("yyyy-mm-dd", feed.PubDate)
		var timestamp time.Time
		if err != nil {
			timestamp = time.Now()
		} else {
			timestamp = t.UTC()
		}
		postParams.PublishedAt = timestamp
		postParams.FeedID = next_feed.ID
		_ = s.db.CreatePost(context.Background(), postParams)

		//fmt.Println(feed.Title)
	}
	return nil
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("Error encountered: ", err)
		os.Exit(1)
	}
	dbURL := cfg.DbURL
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println("Error encountered: ", err)
		os.Exit(1)
	}
	dbQueries := database.New(db)
	st := &state{cfg: cfg, db: dbQueries}
	cmds := &commands{return_map: make(map[string]func(*state, command) error)}
	cmds.register("login", loggedIn(commandLogin))
	cmds.register("register", commandRegister)
	cmds.register("reset", commandReset)
	cmds.register("users", commandUsers)
	cmds.register("agg", commandAgg)
	cmds.register("addfeed", loggedIn(commandAddFeed))
	cmds.register("feeds", commandListFeeds)
	cmds.register("follow", loggedIn(commandFollow))
	cmds.register("following", commandFollowing)
	cmds.register("unfollow", loggedIn(commandUnfollow))
	cmds.register("browse", loggedIn(commandBrowse))

	args := os.Args

	if len(args) < 2 {
		fmt.Println("Error ecounters: not enough arguments")
		os.Exit(1)
	}

	run_cmd_str := args[1]
	run_cmd := command{run_cmd_str, args[2:]}
	err = cmds.run(st, run_cmd)

	if err != nil {
		fmt.Println("Error encounterd: ", err)
		os.Exit(1)
	}
	os.Exit(0)
}
